import { client } from "@purl/lib"

const BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:9090"

export const API_KEY_STORAGE_KEY = "purl_api_key"

export function getApiKey() {
  return localStorage.getItem(API_KEY_STORAGE_KEY) ?? ""
}

export function setApiKey(key: string) {
  localStorage.setItem(API_KEY_STORAGE_KEY, key)
}

// Configure the generated API client — runs once when this module is first imported.
// auth is called per-request so key changes (login/logout) take effect immediately.
client.setConfig({ auth: getApiKey, baseUrl: BASE_URL })
