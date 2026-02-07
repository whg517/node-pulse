/**
 * Application constants
 *
 * Centralized location for magic numbers and configuration values
 */

// Token management (JWT-based authentication)
export const ACCESS_TOKEN_EXPIRY_MINUTES = 15
export const REFRESH_TOKEN_EXPIRY_DAYS = 7
export const TOKEN_PRE_REFRESH_THRESHOLD_SECONDS = 30 // Refresh token 30 seconds before expiry
export const TOKEN_REFRESH_CHECK_INTERVAL_SECONDS = 60 // Check every minute
export const MAX_REFRESH_RETRY_COUNT = 1 // Maximum retry attempts for failed refresh

// Session management (deprecated - kept for backwards compatibility)
export const SESSION_EXPIRY_HOURS = 24
export const SESSION_COOKIE_NAME = 'session_id'

// API configuration
// Production environment should set VITE_API_BASE_URL in .env file
// Development uses Vite proxy (empty string = relative path)
export const API_BASE_URL = import.meta.env.MODE === 'development'
  ? ''
  : (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080')

// Dashboard defaults
export const DEFAULT_REFRESH_INTERVAL = 5 // seconds
export const DEFAULT_TIME_RANGE: '24h' | '7d' | '30d' = '24h'
