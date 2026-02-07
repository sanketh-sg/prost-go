<script setup lang="ts">
// ==========================================
// Root App Component
// The main entry point of the application
// ==========================================

import { useAuthStore } from '@/stores/authStore'

// Get auth store
const authStore = useAuthStore()

// Initialize authentication from localStorage when app starts
// This restores the user session if they were logged in before
authStore.initializeAuth()
</script>

<template>
  <!-- Router outlet - this is where Vue Router renders components -->
  <!-- Depending on the URL, different components will be shown here -->
  <!-- /login → LoginForm.vue -->
  <!-- /dashboard → Dashboard.vue -->
  <!-- /oauth/callback → OAuthCallback.vue -->
  <RouterView />
</template>

<style>
/* ==========================================
   Global Styles
   Apply to entire application
   ========================================== */

body {
  margin: 0;
  padding: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

#app {
  width: 100%;
  min-height: 100vh;
}
</style>