import type {
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  RegisterResponse,
  ErrorResponse,
  AuthError,
  User
} from '@/types/auth'

const API_URL = import.meta.env.VITE_API_URL

class AuthServiceClass {
  private async handleResponse<T>(response: Response): Promise<T> {
    // Check if response is OK (status 200-299)
    if (!response.ok) {
      // Response failed, get error message
      const error: ErrorResponse = await response.json().catch(() => ({
        error: 'Unknown error',
        message: response.statusText,
        code: response.status,
      }))

      // Create an error with the backend's message
      const authError: AuthError = new Error(error.message || 'Request failed')
      authError.code = error.error
      authError.statusCode = response.status
      throw authError
    }

    // Success! Parse and return the JSON response
    return response.json()
  }

  // New user creates account  
  async login(email: string, password: string): Promise<LoginResponse> {
    // Step 1: Create the request body (what we send to backend)
    const payload: LoginRequest = { email, password }

    // Step 2: Make HTTP POST request to /login endpoint
    const response = await fetch(`${API_URL}/login`, {
      method: 'POST', // POST = we're sending data
      headers: {
        'Content-Type': 'application/json', // Telling backend it's JSON
      },
      body: JSON.stringify(payload), // Convert object to JSON string
    })

    // Step 3: Handle the response (success or error)
    return this.handleResponse<LoginResponse>(response)
    // This returns: { user, access_token, refresh_token, expires_in, token_type }
  }

    async register(email: string, username: string, password: string): Promise<RegisterResponse> {
    const payload: RegisterRequest = { email, username, password }

    const response = await fetch(`${API_URL}/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
    })

    return this.handleResponse<RegisterResponse>(response)
  }

// User clicks "Login with Auth0" button
  
  initiateOAuthLogin(): void {
    // Redirect to backend's OAuth endpoint
    // Backend will redirect to Auth0
    window.location.href = `${API_URL}/oauth/login`
  }


    // When access_token expires, get a new one
  
  async refreshToken(refreshToken: string): Promise<{ access_token: string; expires_in: number; token_type: string }> {
    const response = await fetch(
      `${API_URL}/oauth/refresh?refresh_token=${encodeURIComponent(refreshToken)}`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      },
    )

    return this.handleResponse(response)
  }


    // Constants for localStorage keys (prevents typos)
  private readonly TOKEN_KEY = 'access_token'
  private readonly REFRESH_TOKEN_KEY = 'refresh_token'
  private readonly USER_KEY = 'user'

  // Save tokens to localStorage
  setTokens(accessToken: string, refreshToken: string): void {
    localStorage.setItem(this.TOKEN_KEY, accessToken)
    localStorage.setItem(this.REFRESH_TOKEN_KEY, refreshToken)
  }

  // Get access token from localStorage
  getAccessToken(): string | null {
    return localStorage.getItem(this.TOKEN_KEY)
  }

  // Get refresh token from localStorage
  getRefreshToken(): string | null {
    return localStorage.getItem(this.REFRESH_TOKEN_KEY)
  }

  // Save user info to localStorage
  setUser(user: User): void {
    localStorage.setItem(this.USER_KEY, JSON.stringify(user))
  }

  // Get user info from localStorage
  getUser(): User | null { 
    try {
      const userJson = localStorage.getItem(this.USER_KEY)
      if (!userJson) {
        return null
      }
      const parsed = JSON.parse(userJson)
      // Validate it's actually a User object
      return this.validateUser(parsed) ? parsed : null
    } catch (error) {
      console.error('Failed to parse user from localStorage:', error)
      return null
    }
  }

  // Delete all stored tokens and user
  clearTokens(): void {
    localStorage.removeItem(this.TOKEN_KEY)
    localStorage.removeItem(this.REFRESH_TOKEN_KEY)
    localStorage.removeItem(this.USER_KEY)
  }

  // Check if user has a token (is logged in)
  isAuthenticated(): boolean {
    return !!this.getAccessToken()
  }

    // Type guard: Verify object is actually a User
  private validateUser(obj: unknown): obj is User {
    if (typeof obj !== 'object' || obj === null) {
      return false
    }

    const user = obj as Record<string, unknown>
    return (
      typeof user.id === 'string' &&
      typeof user.email === 'string' &&
      typeof user.username === 'string' &&
      typeof user.created_at === 'string'
    )
  }
}

export default new AuthServiceClass()