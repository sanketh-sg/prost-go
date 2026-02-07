<script setup lang="ts">
// ==========================================
// Register/Signup Component
// Let new users create an account
// ==========================================

import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/authStore'

// Get router and store
const router = useRouter()
const authStore = useAuthStore()

// ========== REACTIVE STATE ==========

// Form inputs
const email = ref<string>('')
const username = ref<string>('')
const password = ref<string>('')
const confirmPassword = ref<string>('')

// Loading state (true while sending request to backend)
const isLoading = ref<boolean>(false)

// Error message
const localError = ref<string>('')

// ========== VALIDATION ==========

// Check if form is valid before submitting
const validateForm = (): boolean => {
  // Check email is not empty
  if (!email.value.trim()) {
    localError.value = 'Email is required'
    return false
  }

  // Check email format (must have @)
  if (!email.value.includes('@')) {
    localError.value = 'Please enter a valid email address'
    return false
  }

  // Check username is not empty
  if (!username.value.trim()) {
    localError.value = 'Username is required'
    return false
  }

  // Check username is at least 3 characters
  if (username.value.length < 3) {
    localError.value = 'Username must be at least 3 characters'
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

  // Check passwords match
  if (password.value !== confirmPassword.value) {
    localError.value = 'Passwords do not match'
    return false
  }

  return true
}

// ========== METHODS ==========

// Handle form submission (when user clicks "Create Account")
const handleRegister = async (): Promise<void> => {
  // Validate form first
  if (!validateForm()) {
    return
  }

  // Show loading state
  isLoading.value = true
  localError.value = ''

  try {
    // Call store's register action
    // This sends email, username, password to backend
    // Backend creates the user in database
    await authStore.register(email.value, username.value, password.value)

    // Registration successful! Go to dashboard
    router.push('/dashboard')
  } catch (err) {
    // Show error message from backend (e.g., "Email already exists")
    localError.value = err instanceof Error ? err.message : 'Registration failed'
  } finally {
    isLoading.value = false
  }
}

// Clear error message when user starts typing
const clearError = (): void => {
  localError.value = ''
}

// Go back to login page
const goToLogin = (): void => {
  router.push('/login')
}
</script>

<template>
  <div class="register-container">
    <!-- Card wrapper -->
    <div class="register-card">
      <!-- Header -->
      <div class="register-header">
        <h1>Create Account</h1>
        <p class="subtitle">Join Prost Today</p>
      </div>

      <!-- Registration Form -->
      <form @submit.prevent="handleRegister" class="register-form">
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
          <p class="input-hint">We'll use this to sign you in</p>
        </div>

        <!-- Username Input -->
        <div class="form-group">
          <label for="username">Username</label>
          <input
            id="username"
            v-model="username"
            type="text"
            placeholder="your_username"
            :disabled="isLoading"
            @focus="clearError"
            class="form-input"
          />
          <p class="input-hint">At least 3 characters, no spaces</p>
        </div>

        <!-- Password Input -->
        <div class="form-group">
          <label for="password">Password</label>
          <input
            id="password"
            v-model="password"
            type="password"
            placeholder="Create a strong password"
            :disabled="isLoading"
            @focus="clearError"
            class="form-input"
          />
          <p class="input-hint">At least 6 characters recommended</p>
        </div>

        <!-- Confirm Password Input -->
        <div class="form-group">
          <label for="confirm-password">Confirm Password</label>
          <input
            id="confirm-password"
            v-model="confirmPassword"
            type="password"
            placeholder="Confirm your password"
            :disabled="isLoading"
            @focus="clearError"
            class="form-input"
          />
        </div>

        <!-- Error Message -->
        <div v-if="localError" class="error-message">
          <span class="error-icon">✕</span>
          <span>{{ localError }}</span>
        </div>

        <!-- Submit Button -->
        <button
          type="submit"
          :disabled="isLoading || !email || !username || !password || !confirmPassword"
          class="btn btn-primary"
        >
          <!-- Show spinner while loading -->
          <span v-if="isLoading" class="spinner"></span>
          {{ isLoading ? 'Creating Account...' : 'Create Account' }}
        </button>
      </form>

      <!-- Login Link -->
      <div class="login-link">
        <p>
          Already have an account?
          <button @click="goToLogin" class="link-button">Sign in here</button>
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
* {
  box-sizing: border-box;
}

.register-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}

.register-card {
  background: white;
  border-radius: 12px;
  padding: 40px;
  max-width: 450px;
  width: 100%;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.15);
  animation: slideUp 0.3s ease-out;
}

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

/* Header */
.register-header {
  text-align: center;
  margin-bottom: 32px;
}

.register-header h1 {
  font-size: 28px;
  font-weight: 700;
  color: #333;
  margin: 0;
}

.subtitle {
  font-size: 16px;
  color: #666;
  margin: 8px 0 0 0;
}

/* Form */
.register-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* Form Group */
.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-size: 14px;
  font-weight: 600;
  color: #333;
}

.form-input {
  padding: 12px 14px;
  border: 1.5px solid #ddd;
  border-radius: 8px;
  font-size: 15px;
  font-family: inherit;
  transition: all 0.2s;
}

.form-input:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.form-input:disabled {
  background: #f5f5f5;
  cursor: not-allowed;
  color: #999;
}

/* Input Hint */
.input-hint {
  font-size: 12px;
  color: #999;
  margin: 0;
  font-weight: 500;
}

/* Error Message */
.error-message {
  background: #ffebee;
  border: 1px solid #ff4757;
  border-radius: 8px;
  padding: 12px 14px;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #d32f2f;
  font-size: 14px;
  animation: slideDown 0.2s ease-out;
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

.error-icon {
  font-weight: bold;
  flex-shrink: 0;
}

/* Buttons */
.btn {
  padding: 12px 16px;
  border: none;
  border-radius: 8px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(102, 126, 234, 0.3);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Spinner */
.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* Login Link */
.login-link {
  text-align: center;
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid #f0f0f0;
}

.login-link p {
  font-size: 14px;
  color: #666;
  margin: 0;
}

.link-button {
  background: none;
  border: none;
  color: #667eea;
  font-weight: 600;
  cursor: pointer;
  padding: 0;
  text-decoration: underline;
  transition: color 0.2s;
}

.link-button:hover {
  color: #764ba2;
}

/* Responsive */
@media (max-width: 480px) {
  .register-card {
    padding: 24px;
  }

  .register-header h1 {
    font-size: 24px;
  }

  .form-input {
    font-size: 16px; /* Prevents zoom on mobile */
  }
}
</style>