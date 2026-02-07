export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  user: User
  access_token: string
  refresh_token: string
  expires_in: number
  token_type: string
}

export interface User {
  id: string
  email: string
  username: string
  created_at: string
  updated_at?: string
  deleted_at?: string | null
}

export interface RegisterRequest {
  email: string
  username: string
  password: string
}

export interface RegisterResponse {
  message: string
  user: User
}

export interface ErrorResponse {
  error: string // Error type (e.g., "invalid_credentials")
  message: string // Human-readable error message
  code: number // HTTP status code (401, 400, etc.)
}

export interface AuthError extends Error {
  code?: string       // Error code from backend
  statusCode?: number // HTTP status code
}

export interface RefreshTokenResponse {
  access_token: string
  expires_in: number
  token_type: string
}