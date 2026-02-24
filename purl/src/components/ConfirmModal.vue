<template>
  <Transition name="confirm-fade">
    <div v-if="visible" class="confirm-backdrop" @click.self="emit('cancel')">
      <div class="confirm-panel" role="alertdialog" aria-modal="true">
        <p class="confirm-title">{{ title }}</p>
        <p v-if="message" class="confirm-message">{{ message }}</p>
        <div class="confirm-actions">
          <button class="confirm-btn confirm-btn--back" @click="emit('cancel')">Go back</button>
          <button class="confirm-btn confirm-btn--delete" @click="emit('confirm')">Delete</button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  visible: boolean
  title?: string
  message?: string
}>(), {
  title: "Are you sure?",
})

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()
</script>

<style scoped>
.confirm-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.55);
  backdrop-filter: blur(6px);
  z-index: 130;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.confirm-panel {
  width: 100%;
  max-width: 320px;
  background: rgba(15, 23, 42, 0.97);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  padding: 24px 24px 20px;
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.5);
  animation: confirmUp 0.18s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes confirmUp {
  from { opacity: 0; transform: translateY(10px) scale(0.97); }
}

.confirm-title {
  font-size: 15px;
  font-weight: 700;
  color: #f1f5f9;
  margin: 0 0 6px;
}

.confirm-message {
  font-size: 13px;
  color: rgba(148, 163, 184, 0.65);
  margin: 0 0 20px;
  line-height: 1.5;
}

.confirm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.confirm-btn {
  padding: 8px 16px;
  border-radius: 9px;
  font-size: 13px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  border: none;
  transition: background 0.15s, color 0.15s;
}

.confirm-btn--back {
  background: rgba(255, 255, 255, 0.06);
  color: #94a3b8;
}

.confirm-btn--back:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #e2e8f0;
}

.confirm-btn--delete {
  background: rgba(239, 68, 68, 0.15);
  color: #fca5a5;
}

.confirm-btn--delete:hover {
  background: rgba(239, 68, 68, 0.25);
  color: #fecaca;
}

.confirm-fade-enter-active,
.confirm-fade-leave-active {
  transition: opacity 0.15s ease;
}

.confirm-fade-enter-from,
.confirm-fade-leave-to {
  opacity: 0;
}
</style>
