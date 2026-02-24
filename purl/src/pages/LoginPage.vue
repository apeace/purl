<script setup lang="ts">
import { ref } from "vue"
import { useRouter } from "vue-router"
import { setApiKey } from "../utils/api"
import { useKanbanStore } from "../stores/useKanbanStore"
import { useTicketStore } from "../stores/useTicketStore"

const router = useRouter()
const { loadBoards } = useKanbanStore()
const { reloadTickets } = useTicketStore()
const apiKey = ref("")
const error = ref("")

function submit() {
  const key = apiKey.value.trim()
  if (!key) {
    error.value = "Please enter an API key."
    return
  }
  setApiKey(key)
  loadBoards()
  reloadTickets()
  router.push("/")
}
</script>

<template>
  <div class="login-page">
    <div class="card">
      <div class="logo">
        <span class="logo-mark">P</span>
        <span class="logo-name">Purl</span>
      </div>

      <h1 class="title">Enter your API key</h1>
      <p class="subtitle">Your key will be stored locally and sent with every request.</p>

      <form @submit.prevent="submit" class="form">
        <textarea
          v-model="apiKey"
          class="key-input"
          :class="{ error: error }"
          placeholder="Paste your API key…"
          rows="3"
          spellcheck="false"
          autocomplete="off"
          @input="error = ''"
        />
        <p v-if="error" class="error-msg">{{ error }}</p>
        <button type="submit" class="btn">Continue</button>
      </form>
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

.form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.key-input {
  width: 100%;
  box-sizing: border-box;
  padding: 12px 14px;
  border: 1.5px solid #e4e4e7;
  border-radius: 10px;
  font-family: monospace;
  font-size: 13px;
  color: #18181b;
  background: #fafaf9;
  resize: none;
  outline: none;
  transition: border-color 0.15s;
}

.key-input:focus {
  border-color: #18181b;
  background: #fff;
}

.key-input.error {
  border-color: #ef4444;
}

.error-msg {
  font-size: 12px;
  color: #ef4444;
  margin: 0;
}

.btn {
  padding: 11px;
  background: #18181b;
  color: #fff;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s, transform 0.1s;
}

.btn:hover {
  background: #27272a;
}

.btn:active {
  transform: scale(0.98);
}
</style>
