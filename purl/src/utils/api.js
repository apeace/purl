const BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:9090"

export const API_KEY_STORAGE_KEY = "purl_api_key"

export function getApiKey() {
  return localStorage.getItem(API_KEY_STORAGE_KEY) ?? ""
}

export function setApiKey(key) {
  localStorage.setItem(API_KEY_STORAGE_KEY, key)
}

export function apiFetch(path, options = {}) {
  return fetch(`${BASE_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      "x-api-key": getApiKey(),
      ...options.headers,
    },
  })
}
