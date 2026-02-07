<script setup lang="ts">
// ==========================================
// OAuth Callback Component
// Handles the Auth0 redirect after login
// ==========================================

import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/authStore'
import authService from '@/api/authService'
import type { User } from '@/types/auth'

// Get router, route, and store
const router = useRouter()
const authStore = useAuthStore()

// ========== REACTIVE STATE ==========

// Is the component processing the callback?
const isProcessing = ref<boolean>(true)

// Did the callback succeed?
const isSuccess = ref<boolean>(false)

// Error message if callback failed
const errorMessage = ref<string>('')

// ========== LIFECYCLE ==========
// When component mounts (appears on screen), process OAuth callback

onMounted(async (): Promise<void> => {
  try {
    console.log('Full URL:', window.location.href)

    // ========== STEP 1: Extract tokens from URL query params ==========
    // Backend redirected here with tokens already processed
    const params = new URLSearchParams(window.location.search)
    const accessToken = params.get('access_token')
    const refreshToken = params.get('refresh_token')
    const userId = params.get('user_id')
    const email = params.get('email')
    const username = params.get('username')

    console.log('Extracted params:', {
      accessToken: !!accessToken,
      refreshToken: !!refreshToken,
      userId,
      email,
      username,
    })

    // ========== STEP 2: Check for Auth0 errors ==========
    // Auth0 might reject for various reasons

    const errorParam = params.get('error')
    const errorDescription = params.get('error_description')

    if (errorParam) {
      throw new Error(errorDescription || `Auth0 error: ${errorParam}`)
    }

    // ========== STEP 3: Validate required tokens and user info ==========
    if (!accessToken || !refreshToken || !userId || !email) {
      console.error('Missing required parameters:', {
        accessToken: !!accessToken,
        refreshToken: !!refreshToken,
        userId,
        email,
      })
      throw new Error('Missing required tokens or user information from OAuth callback')
    }

    // ========== STEP 4: Create user object ==========
    const user: User = {
      id: userId,
      email: email,
      username: username || email,
      created_at: new Date().toISOString(),
    }
    console.log('User object created:', user)

    // ========== STEP 5: Save tokens and user to store & localStorage ==========
    // The response contains:
    // - access_token: short-lived token for API requests
    // - refresh_token: long-lived token to get new access tokens
    // - user: the logged-in user's information

    // Save to Pinia store (reactive state)
    authStore.setOAuthTokens(accessToken, refreshToken, user)
    // Also save to localStorage (persistent storage)
    authService.setTokens(accessToken, refreshToken)
    authService.setUser(user)

    console.log('Token and users saved')
    // Mark as successful
    isSuccess.value = true

    // ========== STEP 6: Redirect to dashboard ==========
    // After 1 second (let user see the success message),
    // navigate to the dashboard
    setTimeout(() => {
      console.log('Redirecting to Dashboard...')
      router.push('/dashboard')
    }, 1000)
  } catch (error) {
    // Something went wrong
    isSuccess.value = false
    errorMessage.value =
      error instanceof Error ? error.message : 'An unexpected error occurred during OAuth callback'

    // Log the full error for debugging
    console.error('OAuth callback error:', error)
  } finally {
    // Finished processing (whether success or error)
    isProcessing.value = false
  }
})
</script>

<template>
  <div class="callback-container">
    <!-- Processing State (show while waiting for callback) -->
    <div v-if="isProcessing" class="callback-state processing-state">
      <div class="spinner"></div>
      <h1>Processing Login</h1>
      <p>Completing OAuth authentication...</p>
      <div class="steps">
        <div class="step active">
          <span class="step-number">1</span>
          <span class="step-label">Received auth code from Auth0</span>
        </div>
        <div class="step">
          <span class="step-number">2</span>
          <span class="step-label">Processing tokens</span>
        </div>
        <div class="step">
          <span class="step-number">3</span>
          <span class="step-label">Saving user session</span>
        </div>
      </div>
    </div>

    <!-- Success State (show if callback succeeded) -->
    <div v-else-if="isSuccess" class="callback-state success-state">
      <div class="success-icon">✓</div>
      <h1>Login Successful!</h1>
      <p>Welcome back! Redirecting to dashboard...</p>
      <div class="steps">
        <div class="step completed">
          <span class="step-number">✓</span>
          <span class="step-label">Received auth code from Auth0</span>
        </div>
        <div class="step completed">
          <span class="step-number">✓</span>
          <span class="step-label">Processed tokens</span>
        </div>
        <div class="step completed">
          <span class="step-number">✓</span>
          <span class="step-label">User session saved</span>
        </div>
      </div>
    </div>

    <!-- Error State (show if callback failed) -->
    <div v-else class="callback-state error-state">
      <div class="error-icon">✕</div>
      <h1>Login Failed</h1>
      <div class="error-box">
        <p class="error-title">Error:</p>
        <p class="error-message">{{ errorMessage }}</p>
      </div>
      <div class="actions">
        <router-link to="/login" class="btn btn-primary"> Back to Login </router-link>
        <p class="help-text">If you continue to have issues, please contact support.</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
* {
  box-sizing: border-box;
}

.callback-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
  font-family:
    -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}

.callback-state {
  background: white;
  border-radius: 12px;
  padding: 48px 32px;
  max-width: 500px;
  width: 100%;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
  animation: slideUp 0.3s ease-out;
  text-align: center;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Processing State */
.processing-state {
  animation: slideUp 0.3s ease-out;
}

.spinner {
  width: 50px;
  height: 50px;
  border: 4px solid #f0f0f0;
  border-top-color: #667eea;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 24px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* Success State */
.success-state {
  animation: slideUp 0.3s ease-out;
}

.success-icon {
  width: 60px;
  height: 60px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  font-weight: bold;
  margin: 0 auto 24px;
  animation: scaleIn 0.4s ease-out;
}

@keyframes scaleIn {
  from {
    opacity: 0;
    transform: scale(0.5);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

/* Error State */
.error-state {
  animation: slideUp 0.3s ease-out;
}

.error-icon {
  width: 60px;
  height: 60px;
  background: #ffebee;
  color: #ff4757;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  font-weight: bold;
  margin: 0 auto 24px;
  animation: scaleIn 0.4s ease-out;
}

.error-box {
  background: #ffebee;
  border: 1px solid #ff4757;
  border-radius: 8px;
  padding: 16px;
  margin: 24px 0;
  text-align: left;
}

.error-title {
  font-weight: 600;
  color: #ff4757;
  margin: 0 0 8px 0;
  font-size: 14px;
}

.error-message {
  color: #d32f2f;
  margin: 0;
  font-size: 14px;
  word-break: break-word;
}

/* Text */
.callback-state h1 {
  font-size: 28px;
  font-weight: 700;
  color: #333;
  margin: 0 0 12px 0;
}

.callback-state p {
  font-size: 16px;
  color: #666;
  margin: 0;
}

/* Steps */
.steps {
  margin-top: 32px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.step {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #f8f9ff;
  border-radius: 6px;
  opacity: 0.6;
  transition: all 0.3s ease;
}

.step.active {
  opacity: 1;
  background: #e8ebff;
  border-left: 3px solid #667eea;
}

.step.completed {
  opacity: 1;
  background: #e8f5e9;
  border-left: 3px solid #2e7d32;
}

.step-number {
  width: 24px;
  height: 24px;
  background: #667eea;
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  flex-shrink: 0;
}

.step.completed .step-number {
  background: #2e7d32;
  font-size: 14px;
}

.step-label {
  font-size: 14px;
  color: #333;
  font-weight: 500;
}

/* Actions */
.actions {
  margin-top: 32px;
}

.btn {
  display: inline-block;
  padding: 12px 24px;
  border: none;
  border-radius: 6px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  text-decoration: none;
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 16px rgba(102, 126, 234, 0.4);
}

.help-text {
  font-size: 13px;
  color: #999;
  margin-top: 16px;
}

/* Responsive */
@media (max-width: 480px) {
  .callback-state {
    padding: 32px 20px;
  }

  .callback-state h1 {
    font-size: 22px;
  }

  .callback-state p {
    font-size: 14px;
  }

  .error-box {
    padding: 12px;
  }
}
</style>
