/**
 * Application constants
 *
 * Centralized location for magic numbers and configuration values
 */

// Token management (JWT-based authentication)
export const ACCESS_TOKEN_EXPIRY_MINUTES = 15
export const REFRESH_TOKEN_EXPIRY_DAYS = 7

// API configuration
// Production environment should set VITE_API_BASE_URL in .env file
// Development uses Vite proxy (empty string = relative path)
export const API_BASE_URL = import.meta.env.MODE === 'development'
  ? ''
  : (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080')

// Dashboard defaults
export const DEFAULT_REFRESH_INTERVAL = 5 // seconds
export const DEFAULT_TIME_RANGE: '24h' | '7d' | '30d' = '24h'
