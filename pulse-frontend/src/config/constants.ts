/**
 * Application constants
 *
 * Centralized location for magic numbers and configuration values
 */

// Session management
export const SESSION_EXPIRY_HOURS = 24
export const SESSION_COOKIE_NAME = 'session_id'

// API configuration
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

// Dashboard defaults
export const DEFAULT_REFRESH_INTERVAL = 5 // seconds
export const DEFAULT_TIME_RANGE: '24h' | '7d' | '30d' = '24h'
