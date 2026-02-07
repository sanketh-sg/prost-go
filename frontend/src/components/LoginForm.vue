<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/authStore'

// Get router for navigation between pages
const router = useRouter()
const authStore = useAuthStore()

// Component state
const email = ref<string>('')
const password = ref<string>('')
const isLoading = ref<boolean>(false)
const localError = ref<string>('')

// ========== VALIDATION ==========
const validateForm = (): boolean => {
  // Check email is not empty
  if (!email.value) {
    localError.value = 'Email is required'
    return false
  }

  // Check email format is valid
  if (!email.value.includes('@')) {
    localError.value = 'Please enter a valid email'
    return false
  }

  // Check password is not empty
  if (!password.value) {
    localError.value = 'Password is required'
    return false
  }

  // Check password is long enough
  if (password.value.length < 6) {
    localError.value = 'Password must be at least 6 characters'
    return false
  }

  return true
}

// ========== METHODS ==========

// Handle form submission
const handleLogin = async (): Promise<void> => {
  // Validate first
  if (!validateForm()) {
    return
  }

  isLoading.value = true
  localError.value = ''

  try {
    // Call store's login action
    // This talks to backend, gets tokens, and stores them
    await authStore.login(email.value, password.value)

    // Login successful! Go to dashboard
    router.push('/dashboard')
  } catch (err) {
    // Show error message from backend
    localError.value = err instanceof Error ? err.message : 'Login failed'
  } finally {
    isLoading.value = false
  }
}

// Handle OAuth login button click
const handleOAuthLogin = (): void => {
  // This redirects to backend's /oauth/login
  // Backend then redirects to Auth0
  authStore.initiateOAuthLogin()
}

const handleGmailLogin = (): void => {
  window.location.href = `${import.meta.env.VITE_API_URL}/oauth/login/gmail`
}

// Clear error when user starts typing
const clearError = (): void => {
  localError.value = ''
}

// Go to register page
const goToRegister = (): void => {
  router.push('/register')
}
</script>

<template>
  <!-- Main container -->
  <div class="login-container">
    <!-- Card wrapper -->
    <div class="login-card">
      <!-- Header -->
      <div class="login-header">
        <h1>Prost</h1>
        <p class="subtitle">E-commerce Platform</p>
      </div>

      <!-- Login Form -->
      <form @submit.prevent="handleLogin" class="login-form">
        <!-- Email Input -->
        <div class="form-group">
          <label for="email">Email Address</label>
          <input
            id="email"
            v-model="email"
            type="email"
            placeholder="you@example.com"
            :disabled="isLoading"
            @focus="clearError"
            class="form-input"
          />
        </div>

        <!-- Password Input -->
        <div class="form-group">
          <label for="password">Password</label>
          <input id="password" 
          v-model="password" 
          type="password" 
          placeholder="Enter your password" 
          :disabled="isLoading" 
          @focus="clearError" 
          class="form-input"
          />
        </div>

        <!-- Submit Button -->
        <button
          type="submit"
          :disabled="isLoading || !email || !password"
          class="btn btn-primary"
        >
          <!-- Show spinner while loading -->
          <span v-if="isLoading" class="spinner"></span>
          {{ isLoading ? 'Logging in...' : 'Login' }}
        </button>
      </form>

      <!-- Divider -->
      <div class="divider">
        <span>OR</span>
      </div>

  <!-- OAuth Buttons -->
  <div class="oauth-buttons">
    <!-- Gmail Button -->
    <button @click="handleGmailLogin" :disabled="isLoading" class="btn btn-gmail">
      <img src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24'%3E%3Cpath fill='%234285F4' d='M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z'/%3E%3Cpath fill='%3434A853' d='M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z'/%3E%3Cpath fill='%23FBBC05' d='M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z'/%3E%3Cpath fill='%23EA4335' d='M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z'/%3E%3C/svg%3E" alt="Google" class="oauth-icon">
      Continue with Gmail
    </button>

    <!-- Auth0 Button -->
    <button @click="handleOAuthLogin" :disabled="isLoading" class="btn btn-oauth">
      Continue with Auth0
    </button>
  </div>

      <!-- Register Link -->
      <div class="register-link">
        <p>
          Don't have an account?
          <button
            type="button"
            @click="goToRegister"
            class="link-button"
          >
            Create one here
          </button>
        </p>
      </div>

      <!-- Error Message -->
      <div v-if="localError" class="error-message">
        <span class="error-icon">⚠️</span>
        {{ localError }}
      </div>
    </div>
  </div>
</template>
<style scoped>
* {
  box-sizing: border-box;
}

/* Main container - fills screen */
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}

/* White card in center */
.login-card {
  background: white;
  border-radius: 12px;
  padding: 48px 40px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  width: 100%;
  max-width: 420px;
  animation: slideUp 0.3s ease-out;
}

/* Slide up animation when page loads */
@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Header section */
.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-header h1 {
  font-size: 32px;
  font-weight: 700;
  color: #333;
  margin: 0 0 8px 0;
}

.subtitle {
  font-size: 14px;
  color: #999;
  margin: 0;
}

/* Form container */
.login-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* Form group (label + input) */
.form-group {
  display: flex;
  flex-direction: column;
}

.form-group label {
  font-size: 14px;
  font-weight: 600;
  color: #333;
  margin-bottom: 6px;
}

/* Text input styling */
.form-input {
  padding: 12px 14px;
  border: 2px solid #e0e0e0;
  border-radius: 6px;
  font-size: 15px;
  transition: all 0.2s;
}

/* Input focus state */
.form-input:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

/* Input disabled state */
.form-input:disabled {
  background-color: #f5f5f5;
  color: #999;
  cursor: not-allowed;
}

/* Buttons */
.btn {
  padding: 12px 16px;
  border: none;
  border-radius: 6px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

/* Primary button (Login) */
.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  margin-top: 8px;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 16px rgba(102, 126, 234, 0.4);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* OAuth buttons */
.oauth-buttons {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.btn-gmail {
  background: white;
  color: #333;
  border: 1px solid #ddd;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.btn-gmail:hover:not(:disabled) {
  background: #f8f8f8;
  border-color: #999;
}

.oauth-icon {
  width: 18px;
  height: 18px;
}

.btn-oauth:hover:not(:disabled) {
  background: #f0f2ff;
  border-color: #667eea;
  transform: translateY(-2px);
}

.btn-oauth:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Loading spinner */
.spinner {
  display: inline-block;
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* Divider between login and OAuth */
.divider {
  text-align: center;
  color: #999;
  margin: 24px 0;
  position: relative;
  font-size: 13px;
}

.divider::before,
.divider::after {
  content: '';
  position: absolute;
  top: 50%;
  width: 35%;
  height: 1px;
  background: #e0e0e0;
}

.divider::before {
  left: 0;
}

.divider::after {
  right: 0;
}

.divider span {
  position: relative;
  background: white;
  padding: 0 12px;
}

/* Register link */
.register-link {
  text-align: center;
  margin-top: 20px;
}

.register-link p {
  color: #666;
  font-size: 14px;
  margin: 0;
}

.link-button {
  background: none;
  border: none;
  color: #667eea;
  cursor: pointer;
  font-weight: 600;
  padding: 0;
  font-size: inherit;
  transition: color 0.2s;
}

.link-button:hover {
  color: #764ba2;
  text-decoration: underline;
}

/* Error message */
.error-message {
  margin-top: 16px;
  padding: 12px 14px;
  background: #fee;
  color: #c33;
  border-radius: 6px;
  font-size: 14px;
  border-left: 4px solid #c33;
  animation: slideDown 0.2s ease-out;
  display: flex;
  align-items: center;
  gap: 8px;
}

.error-icon {
  font-size: 16px;
  flex-shrink: 0;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Responsive design for mobile */
@media (max-width: 480px) {
  .login-card {
    padding: 32px 24px;
  }

  .login-header h1 {
    font-size: 28px;
  }

  /* Prevent zoom on mobile when focusing input */
  .form-input,
  .btn {
    font-size: 16px;
  }
}
</style>