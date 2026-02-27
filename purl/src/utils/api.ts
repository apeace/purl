import { client } from "@purl/lib"

const BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:9090"

export const API_KEY_STORAGE_KEY = "purl_api_key"
export const SESSION_TOKEN_KEY = "purl_session_token"

export function getApiKey() {
  return localStorage.getItem(API_KEY_STORAGE_KEY) ?? ""
}

export function setApiKey(key: string) {
  localStorage.setItem(API_KEY_STORAGE_KEY, key)
}

export function getSessionToken() {
  return localStorage.getItem(SESSION_TOKEN_KEY) ?? ""
}

export function setSessionToken(token: string) {
  localStorage.setItem(SESSION_TOKEN_KEY, token)
}

export function clearSessionToken() {
  localStorage.removeItem(SESSION_TOKEN_KEY)
}

// Configure the generated API client — runs once when this module is first imported.
client.setConfig({ baseUrl: BASE_URL })

// Add session token as Bearer auth when available, fall back to API key
client.interceptors.request.use((request) => {
  const token = getSessionToken()
  if (token) {
    request.headers.set("Authorization", `Bearer ${token}`)
  } else {
    const key = getApiKey()
    if (key) {
      request.headers.set("x-api-key", key)
    }
  }
  return request
})

// Redirect to login on 401
client.interceptors.response.use((response) => {
  if (response.status === 401 && !response.url.includes("/auth/")) {
    clearSessionToken()
    localStorage.removeItem(API_KEY_STORAGE_KEY)
    window.location.href = "/login"
  }
  return response
})
