<template>
  <Transition name="modal-fade">
    <div v-if="ticketId" class="modal-backdrop" @click="$emit('close')" @keydown.esc="$emit('close')">
      <div class="modal-panel" @click.stop>
        <button class="modal-close" @click="$emit('close')">&times;</button>
        <TicketDetail :ticket-id="ticketId" @resolve="$emit('close')" />
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, watch } from "vue"
import TicketDetail from "./TicketDetail.vue"

const props = withDefaults(defineProps<{
  ticketId?: string | null
}>(), {
  ticketId: null,
})

const emit = defineEmits<{
  close: []
}>()

function onEsc(e: KeyboardEvent) {
  if (e.key === "Escape" && props.ticketId) emit("close")
}

watch(() => props.ticketId, (val) => {
  if (val) {
    document.addEventListener("keydown", onEsc)
  } else {
    document.removeEventListener("keydown", onEsc)
  }
}, { immediate: true })

onMounted(() => {
  if (props.ticketId) document.addEventListener("keydown", onEsc)
})

onUnmounted(() => {
  document.removeEventListener("keydown", onEsc)
})
</script>

<style scoped>
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(8px);
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}

.modal-panel {
  position: relative;
  width: 100%;
  max-width: 860px;
  height: 85vh;
  background: #0f172a;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 20px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: 0 24px 80px rgba(0, 0, 0, 0.5);
  animation: modal-up 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes modal-up {
  from {
    opacity: 0;
    transform: translateY(16px) scale(0.97);
  }
}

.modal-close {
  position: absolute;
  top: 12px;
  right: 12px;
  z-index: 10;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.06);
  color: rgba(148, 163, 184, 0.6);
  font-size: 20px;
  line-height: 1;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s, color 0.15s;
}

.modal-close:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #e2e8f0;
}

/* ── Transition ──────────────────────────────────────────── */

.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.2s ease;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}
</style>
