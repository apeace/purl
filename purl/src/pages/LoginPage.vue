<script setup lang="ts">
import { onMounted, ref } from "vue"
import { useRouter } from "vue-router"
import { setSessionToken } from "../utils/api"
import { useKanbanStore } from "../stores/useKanbanStore"
import { useTicketStore } from "../stores/useTicketStore"
import { useUserStore } from "../stores/useUserStore"

const BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:9090"
const GOOGLE_CLIENT_ID = import.meta.env.VITE_GOOGLE_CLIENT_ID ?? ""

const router = useRouter()
const { loadBoards } = useKanbanStore()
const { reloadTickets } = useTicketStore()
const userStore = useUserStore()
const error = ref("")
const loading = ref(false)

async function handleGoogleCallback(response: { credential: string }) {
  loading.value = true
  error.value = ""

  try {
    const res = await fetch(`${BASE_URL}/auth/google`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ id_token: response.credential }),
    })

    if (!res.ok) {
      const text = await res.text()
      error.value = text.trim() || "Sign in failed."
      loading.value = false
      return
    }

    const data = await res.json()
    setSessionToken(data.token)
    userStore.name = data.user.name
    userStore.email = data.user.email
    userStore.orgName = data.user.org.name
    userStore.isAuthenticated = true
    loadBoards()
    reloadTickets()
    router.push("/")
  } catch {
    error.value = "Something went wrong. Please try again."
    loading.value = false
  }
}

function loadGoogleScript() {
  if (document.getElementById("google-gsi")) return
  const script = document.createElement("script")
  script.id = "google-gsi"
  script.src = "https://accounts.google.com/gsi/client"
  script.async = true
  script.onload = initGoogle
  document.head.appendChild(script)
}

function initGoogle() {
  const google = (window as any).google
  if (!google) return
  google.accounts.id.initialize({
    client_id: GOOGLE_CLIENT_ID,
    callback: handleGoogleCallback,
  })
  google.accounts.id.renderButton(
    document.getElementById("google-signin-btn"),
    {
      type: "standard",
      theme: "outline",
      size: "large",
      width: 348,
      text: "signin_with",
      shape: "pill",
    },
  )
}

onMounted(() => {
  if ((window as any).google) {
    initGoogle()
  } else {
    loadGoogleScript()
  }
})
</script>

<template>
  <div class="login-page">
    <div class="card">
      <div class="logo">
        <span class="logo-mark">P</span>
        <span class="logo-name">Purl</span>
      </div>

      <h1 class="title">Sign in to Purl</h1>
      <p class="subtitle">Use your Google account to continue.</p>

      <div class="google-wrap">
        <div id="google-signin-btn" />
        <div v-if="loading" class="loading-text">Signing in...</div>
      </div>
      <p v-if="error" class="error-msg">{{ error }}</p>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100dvh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f8f8f7;
  padding: 24px;
}

.card {
  background: #fff;
  border: 1px solid #e8e8e6;
  border-radius: 16px;
  padding: 40px 36px;
  width: 100%;
  max-width: 420px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.06);
  animation: rise 0.3s ease;
}

@keyframes rise {
  from { opacity: 0; transform: translateY(12px); }
  to   { opacity: 1; transform: translateY(0); }
}

.logo {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 28px;
}

.logo-mark {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: #18181b;
  color: #fff;
  font-size: 15px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-name {
  font-size: 16px;
  font-weight: 600;
  color: #18181b;
}

.title {
  font-size: 20px;
  font-weight: 700;
  color: #18181b;
  margin: 0 0 6px;
}

.subtitle {
  font-size: 13px;
  color: #71717a;
  margin: 0 0 24px;
  line-height: 1.5;
}

.google-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.loading-text {
  font-size: 13px;
  color: #71717a;
}

.error-msg {
  font-size: 12px;
  color: #ef4444;
  margin: 12px 0 0;
  text-align: center;
}
</style>
