import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import authService from '@/api/authService'
import type { User, LoginResponse, RegisterResponse } from '@/types/auth'


export const useAuthStore = defineStore('auth', () => {
// State has set of variables when they change Vue updates the UI automatically
  const user = ref<User | null>(null)
  const accessToken = ref<string | null>(null)
  const refreshToken = ref<string | null>(null)
  const isLoading = ref<boolean>(false)
  const error = ref<string | null>(null)

//computed value obtains the value from the state variables, it also changes when state changes.
  const isAuthenticated = computed(():boolean => {
    return !!accessToken.value
  })

//Actions are the methods that change the state
  const initializeAuth = (): void => {
    const savedUser = authService.getUser()
    const savedAccessToken = authService.getAccessToken()
    const savedRefreshToken = authService.getRefreshToken()

    if (savedUser) {
      user.value = savedUser
    }
    if (savedAccessToken) {
      accessToken.value = savedAccessToken
    }
    if (savedRefreshToken) {
      refreshToken.value = savedRefreshToken
    }
  }

  const login = async (email: string, password: string): Promise<LoginResponse> => {
    isLoading.value = true;
    error.value = null;
    try{
      const response = await authService.login(email,password)

      user.value = response.user
      accessToken.value = response.access_token
      refreshToken.value = response.refresh_token

      authService.setTokens(response.access_token, response.refresh_token)
      authService.setUser(response.user)
      return response
    } catch(err) {
      const message = err instanceof Error ? err.message : 'Login Fialed'
      error.value = message
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const register = async (email:string, username:string, password:string): Promise<RegisterResponse> => {
    isLoading.value = true;
    error.value = null;
    try{
      const response = await authService.register(email,username,password)
      return response
    } catch(err){
      const message = err instanceof Error ? err.message : 'Registration Failed'
      error.value = message
      throw err
    } finally {
      isLoading.value = false
    }
  }

  // Set tokens after OAuth login (callback from Auth0)
const setOAuthTokens = (access_token: string, refresh_token: string, userData: User): void => {
    user.value = userData
    accessToken.value = access_token
    refreshToken.value = refresh_token

    authService.setTokens(access_token, refresh_token)
    authService.setUser(userData)
  }

  // Get a new access token using refresh token
  const refreshAccessToken = async (): Promise<void> => {
    if (!refreshToken.value) {
      throw new Error('No refresh token available')
    }

    try {
      const response = await authService.refreshToken(refreshToken.value)
      accessToken.value = response.access_token
      authService.setTokens(response.access_token, refreshToken.value)
    } catch (err) {
      // If refresh fails, logout completely
      clearAuth()
      throw err
    }
  }

      // Clear auth state (helper for other methods)
  const clearAuth = (): void => {
    logout()
  }

    // Logout - clear everything
  const logout = (): void => {
    user.value = null
    accessToken.value = null
    refreshToken.value = null
    error.value = null
    authService.clearTokens()
  }
 

  // Start OAuth login (redirects to Auth0)
  const initiateOAuthLogin = (): void => {
    authService.initiateOAuthLogin()
  }

  // Clear error message
  const clearError = (): void => {
    error.value = null
  }

  // These are the things components can use
  return {
    // State (components can read these)
    user,
    accessToken,
    refreshToken,
    isLoading,
    error,

    // Computed (components can read these too)
    isAuthenticated,

    // Methods (components call these)
    initializeAuth,
    login,
    register,
    setOAuthTokens,
    refreshAccessToken,
    logout,
    clearAuth,
    initiateOAuthLogin,
    clearError,
  }
})