<template>
  <Transition name="picker-fade">
    <div v-if="visible" class="picker-backdrop" @click.self="emit('close')">
      <div class="picker-panel">
        <div class="picker-header">
          <h3 class="picker-title">Add to {{ boardName }}</h3>
          <button class="picker-close" @click="emit('close')">
            <X :size="14" />
          </button>
        </div>
        <div class="picker-subtitle">Choose a column</div>
        <div class="picker-list">
          <button
            v-for="stage in stages"
            :key="stage.id"
            class="picker-option"
            @click="handlePick(stage.id)"
          >
            <span class="picker-dot" :style="{ background: stage.color }" />
            <span class="picker-name">{{ stage.name }}</span>
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { X } from "lucide-vue-next"
import type { BoardStage } from "../composables/useKanbanBoards"

defineProps<{
  visible: boolean
  boardName: string
  stages: BoardStage[]
}>()

const emit = defineEmits<{
  close: []
  pick: [stageId: string]
}>()

function handlePick(stageId: string) {
  emit("pick", stageId)
}
</script>

<style scoped>
.picker-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.55);
  backdrop-filter: blur(6px);
  z-index: 110;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.picker-panel {
  width: 100%;
  max-width: 300px;
  background: rgba(15, 23, 42, 0.97);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 14px;
  overflow: hidden;
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.5);
  animation: slideUp 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(12px) scale(0.97); }
}

.picker-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px 0;
}

.picker-title {
  font-size: 15px;
  font-weight: 700;
  color: #f1f5f9;
  margin: 0;
}

.picker-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.04);
  color: #64748b;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.picker-close:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #94a3b8;
}

.picker-subtitle {
  font-size: 12px;
  color: rgba(148, 163, 184, 0.4);
  padding: 4px 16px 10px;
}

.picker-list {
  padding: 0 6px 6px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.picker-option {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 10px 12px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: #e2e8f0;
  font-size: 14px;
  font-weight: 500;
  font-family: inherit;
  cursor: pointer;
  text-align: left;
  transition: background 0.15s;
}

.picker-option:hover {
  background: rgba(255, 255, 255, 0.06);
}

.picker-option:active {
  background: rgba(99, 102, 241, 0.12);
}

.picker-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.picker-name {
  flex: 1;
}

.picker-fade-enter-active,
.picker-fade-leave-active {
  transition: opacity 0.15s ease;
}

.picker-fade-enter-from,
.picker-fade-leave-to {
  opacity: 0;
}
</style>
