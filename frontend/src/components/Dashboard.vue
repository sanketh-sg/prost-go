<script setup lang="ts">
// ==========================================
// Dashboard Component
// Shows after user logs in
// ==========================================

import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/authStore'
import type { User } from '@/types/auth'

// Get router and store
const router = useRouter()
const authStore = useAuthStore()

// Get current user (computed - updates when store changes)
const user = computed((): User | null => authStore.user)

// Get access token
const accessToken = computed((): string | null => authStore.accessToken)

// ========== METHODS ==========

// Format date nicely
const formatDate = (dateString: string): string => {
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}

// Show first 50 characters of token
const getTokenPreview = (token: string | null): string => {
  if (!token) return 'No token'
  return `${token.substring(0, 50)}...`
}

// Copy text to clipboard
const copyToClipboard = (text: string): void => {
  navigator.clipboard.writeText(text)
}

// Logout and go to login page
const handleLogout = (): void => {
  authStore.logout()
  router.push('/login')
}
</script>

<template>
  <div class="dashboard-container">
    <div class="dashboard-layout">
      <!-- Header -->
      <header class="dashboard-header">
        <div class="header-content">
          <h1>Prost Dashboard</h1>
          <button @click="handleLogout" class="btn btn-logout">
            🚪 Logout
          </button>
        </div>
      </header>

      <!-- Main Content -->
      <main class="dashboard-content">
        <!-- Welcome Card -->
        <section class="card welcome-card">
          <div class="card-header">
            <h2>Welcome back! 👋</h2>
          </div>
          <div class="card-body">
            <p v-if="user" class="welcome-message">
              Hello <strong>{{ user.username || user.email }}</strong>
            </p>
            <p class="welcome-subtext">
              You are successfully authenticated and logged in.
            </p>
          </div>
        </section>

        <!-- User Information Card -->
        <section v-if="user" class="card info-card">
          <div class="card-header">
            <h3>📋 Your Information</h3>
          </div>
          <div class="card-body">
            <div class="info-grid">
              <div class="info-item">
                <label>User ID</label>
                <div class="info-value">
                  <code>{{ user.id }}</code>
                  <button 
                    @click="copyToClipboard(user.id)" 
                    class="copy-btn"
                    title="Copy to clipboard"
                  >
                    📋
                  </button>
                </div>
              </div>

              <div class="info-item">
                <label>Email</label>
                <div class="info-value">
                  <span>{{ user.email }}</span>
                </div>
              </div>

              <div class="info-item">
                <label>Username</label>
                <div class="info-value">
                  <span>{{ user.username }}</span>
                </div>
              </div>

              <div class="info-item">
                <label>Account Created</label>
                <div class="info-value">
                  <span>{{ formatDate(user.created_at) }}</span>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- Token Information Card -->
        <section v-if="accessToken" class="card token-card">
          <div class="card-header">
            <h3>🔐 Access Token</h3>
            <span class="badge badge-success">Active</span>
          </div>
          <div class="card-body">
            <div class="token-info">
              <p class="token-label">Token Preview</p>
              <div class="token-container">
                <code class="token-text">{{ getTokenPreview(accessToken) }}</code>
                <button 
                  @click="copyToClipboard(accessToken)" 
                  class="copy-btn copy-btn-large"
                  title="Copy full token"
                >
                  📋 Copy
                </button>
              </div>
              <p class="token-note">
                ✓ Token stored securely in browser
              </p>
            </div>
          </div>
        </section>

        <!-- Status Card -->
        <section class="card status-card">
          <div class="card-header">
            <h3>✅ Status</h3>
          </div>
          <div class="card-body">
            <div class="status-list">
              <div class="status-item">
                <span class="status-icon">✓</span>
                <span>Authenticated</span>
              </div>
              <div class="status-item">
                <span class="status-icon">✓</span>
                <span>Session Active</span>
              </div>
              <div class="status-item">
                <span class="status-icon">✓</span>
                <span>Tokens Valid</span>
              </div>
            </div>
          </div>
        </section>
      </main>
    </div>
  </div>
</template>

<style scoped>
* {
  box-sizing: border-box;
}

.dashboard-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}

.dashboard-layout {
  max-width: 1000px;
  margin: 0 auto;
}

/* Header */
.dashboard-header {
  margin-bottom: 32px;
  animation: slideDown 0.3s ease-out;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
}

.dashboard-header h1 {
  font-size: 32px;
  font-weight: 700;
  color: white;
  margin: 0;
}

/* Content Grid */
.dashboard-content {
  display: grid;
  gap: 20px;
}

/* Cards */
.card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
  animation: fadeIn 0.3s ease-out;
  overflow: hidden;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.card-header {
  padding: 20px 24px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h2,
.card-header h3 {
  font-size: 18px;
  font-weight: 600;
  color: #333;
  margin: 0;
}

.card-header h2 {
  font-size: 24px;
}

.card-body {
  padding: 24px;
}

/* Welcome Card */
.welcome-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.welcome-card .card-header {
  border-bottom-color: rgba(255, 255, 255, 0.2);
}

.welcome-card .card-header h2 {
  color: white;
}

.welcome-card .card-body {
  color: white;
}

.welcome-message {
  font-size: 18px;
  margin: 0 0 8px 0;
}

.welcome-subtext {
  font-size: 14px;
  opacity: 0.9;
  margin: 0;
}

/* Badge */
.badge {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}

.badge-success {
  background: #e8f5e9;
  color: #2e7d32;
}

/* Info Grid */
.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
}

.info-item {
  display: flex;
  flex-direction: column;
}

.info-item label {
  font-size: 12px;
  font-weight: 600;
  color: #999;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 8px;
}

.info-value {
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-value code {
  background: #f5f5f5;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 13px;
  font-family: 'Courier New', monospace;
  color: #333;
  flex: 1;
  word-break: break-all;
}

.info-value span {
  font-size: 15px;
  color: #333;
}

.copy-btn {
  background: none;
  border: none;
  font-size: 16px;
  cursor: pointer;
  padding: 4px 8px;
  transition: transform 0.2s;
}

.copy-btn:hover {
  transform: scale(1.2);
}

/* Token Card */
.token-info {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.token-label {
  font-size: 13px;
  font-weight: 600;
  color: #999;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin: 0;
}

.token-container {
  display: flex;
  gap: 8px;
  align-items: center;
}

.token-text {
  background: #f5f5f5;
  padding: 12px 14px;
  border-radius: 6px;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  color: #333;
  flex: 1;
  overflow: hidden;
  word-break: break-all;
}

.copy-btn-large {
  padding: 8px 16px;
  background: #667eea;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.2s;
}

.copy-btn-large:hover {
  background: #764ba2;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.token-note {
  font-size: 13px;
  color: #666;
  margin: 0;
}

/* Status Card */
.status-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #f8f9ff;
  border-radius: 6px;
  font-size: 15px;
  color: #333;
}

.status-icon {
  font-size: 18px;
  flex-shrink: 0;
}

/* Buttons */
.btn {
  padding: 10px 16px;
  border: none;
  border-radius: 6px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-logout {
  background: #ff4757;
  color: white;
}

.btn-logout:hover {
  background: #ff3838;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(255, 71, 87, 0.3);
}

/* Responsive */
@media (max-width: 768px) {
  .header-content {
    flex-direction: column;
    align-items: flex-start;
  }

  .dashboard-header h1 {
    font-size: 24px;
  }

  .info-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 480px) {
  .dashboard-container {
    padding: 12px;
  }

  .dashboard-header h1 {
    font-size: 20px;
  }

  .card {
    border-radius: 8px;
  }

  .card-header {
    padding: 16px;
  }

  .card-body {
    padding: 16px;
  }
}
</style>