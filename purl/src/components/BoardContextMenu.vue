<template>
  <Teleport to="body">
    <div v-if="visible" class="ctx-backdrop" @click="emit('close')" @contextmenu.prevent="emit('close')">
      <div
        class="ctx-menu"
        :style="{ top: `${y}px`, left: `${x}px` }"
        @click.stop
      >
        <button class="ctx-item" @click="emit('rename')">
          <Pencil :size="14" />
          <span>Rename</span>
        </button>
        <button class="ctx-item ctx-item--danger" @click="emit('delete')">
          <Trash2 :size="14" />
          <span>Delete</span>
        </button>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { Pencil, Trash2 } from "lucide-vue-next"

defineProps<{
  visible: boolean
  x: number
  y: number
}>()

const emit = defineEmits<{
  close: []
  rename: []
  delete: []
}>()
</script>

<style scoped>
.ctx-backdrop {
  position: fixed;
  inset: 0;
  z-index: 120;
}

.ctx-menu {
  position: fixed;
  min-width: 150px;
  background: rgba(15, 23, 42, 0.97);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 10px;
  padding: 4px;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.5);
  animation: ctxPop 0.12s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes ctxPop {
  from { opacity: 0; transform: scale(0.95); }
}

.ctx-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 9px 12px;
  border: none;
  border-radius: 7px;
  background: transparent;
  color: #e2e8f0;
  font-size: 13px;
  font-weight: 500;
  font-family: inherit;
  cursor: pointer;
  text-align: left;
  transition: background 0.12s;
}

.ctx-item:hover {
  background: rgba(255, 255, 255, 0.06);
}

.ctx-item--danger:hover {
  background: rgba(239, 68, 68, 0.12);
  color: #fca5a5;
}
</style>
