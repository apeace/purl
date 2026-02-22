const BASE_URL = "http://localhost:8080"

export const API_KEY_STORAGE_KEY = "pipeline_api_key"

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
