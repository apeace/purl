import { defineStore } from "pinia"
import { ref } from "vue"
import { clearSessionToken, getSessionToken } from "../utils/api"

const BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:9090"

interface UserInfo {
  id: string
  name: string
  email: string
  org: { id: string; name: string }
}

export const useUserStore = defineStore("user", () => {
  const name = ref("")
  const email = ref("")
  const orgName = ref("")
  const isAuthenticated = ref(false)

  async function loadUser() {
    const token = getSessionToken()
    if (!token) {
      isAuthenticated.value = false
      return
    }

    try {
      const resp = await fetch(`${BASE_URL}/auth/me`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (!resp.ok) {
        isAuthenticated.value = false
        return
      }
      const user: UserInfo = await resp.json()
      name.value = user.name
      email.value = user.email
      orgName.value = user.org.name
      isAuthenticated.value = true
    } catch {
      isAuthenticated.value = false
    }
  }

  async function logout() {
    const token = getSessionToken()
    if (token) {
      try {
        await fetch(`${BASE_URL}/auth/logout`, {
          method: "POST",
          headers: { Authorization: `Bearer ${token}` },
        })
      } catch {
        // Best-effort — clear local state regardless
      }
    }
    clearSessionToken()
    name.value = ""
    email.value = ""
    orgName.value = ""
    isAuthenticated.value = false
  }

  return {
    email,
    isAuthenticated,
    loadUser,
    logout,
    name,
    orgName,
  }
})
