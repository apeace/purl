<template>
  <Transition name="modal-fade">
    <div v-if="visible" class="modal-backdrop" @click.self="emit('close')">
      <div class="modal-panel">
        <div class="modal-header">
          <h2 class="modal-title">Create Board</h2>
          <button class="modal-close" @click="emit('close')">
            <X :size="16" />
          </button>
        </div>

        <div class="modal-body">
          <label class="field-label">Board name</label>
          <input
            ref="nameInput"
            v-model="boardName"
            class="field-input"
            :placeholder="namePlaceholder"
            @keydown.enter="handleCreate"
          />

          <label class="field-label field-label--stages">Columns</label>
          <div class="stages-list">
            <div
              v-for="(stage, i) in stages"
              :key="stage.color + i"
              class="stage-row"
              :class="{
                'stage-row--dragging': dragIndex === i,
                'stage-row--drop-above': dropTarget === i && dragIndex !== null && dragIndex > i,
                'stage-row--drop-below': dropTarget === i && dragIndex !== null && dragIndex < i,
              }"
              draggable="true"
              @dragstart="onStageDragStart($event, i)"
              @dragend="onStageDragEnd"
              @dragover.prevent="onStageDragOver(i)"
              @dragleave="onStageDragLeave(i)"
              @drop.prevent="onStageDrop(i)"
            >
              <div class="drag-handle">
                <GripVertical :size="14" />
              </div>
              <button
                class="color-dot"
                :style="{ background: stage.color }"
                @click="cycleColor(i)"
              />
              <input
                v-model="stage.name"
                class="stage-input"
                placeholder="Column name…"
                @keydown.enter="handleCreate"
              />
              <button
                v-if="stages.length > 1"
                class="stage-remove"
                @click="stages.splice(i, 1)"
              >
                <Minus :size="14" />
              </button>
            </div>
          </div>
          <button class="add-stage-btn" @click="addStage">
            <Plus :size="14" />
            <span>Add column</span>
          </button>
        </div>

        <div class="modal-actions">
          <button class="btn btn--ghost" @click="emit('close')">Cancel</button>
          <button class="btn btn--primary" :disabled="!canCreate" @click="handleCreate">
            Create Board
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { GripVertical, Minus, Plus, X } from "lucide-vue-next"
import { computed, nextTick, ref, watch } from "vue"
import { useKanbanStore } from "../stores/useKanbanStore"

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{
  close: []
  created: [boardId: string]
}>()

const { createBoard } = useKanbanStore()

const PALETTE = [
  "#38bdf8", "#6366f1", "#a855f7", "#ec4899",
  "#f97316", "#f59e0b", "#34d399", "#60a5fa",
  "#ef4444", "#14b8a6", "#84cc16", "#94a3b8",
]

const nameInput = ref<HTMLInputElement | null>(null)
const boardName = ref("")
const stages = ref<{ name: string; color: string }[]>([])
const dragIndex = ref<number | null>(null)
const dropTarget = ref<number | null>(null)

const namePlaceholder = computed(() => {
  const d = new Date()
  const mm = String(d.getMonth() + 1).padStart(2, "0")
  const dd = String(d.getDate()).padStart(2, "0")
  return `e.g. Outage - ${mm}/${dd}/${d.getFullYear()}`
})

function resetForm() {
  boardName.value = ""
  stages.value = [
    { name: "", color: PALETTE[0] },
    { name: "", color: PALETTE[1] },
    { name: "", color: PALETTE[2] },
  ]
  dragIndex.value = null
  dropTarget.value = null
}

resetForm()

watch(() => props.visible, (val) => {
  if (val) {
    resetForm()
    nextTick(() => nameInput.value?.focus())
  }
})

// ── Drag reorder ────────────────────────────────────────

function onStageDragStart(e: DragEvent, i: number) {
  dragIndex.value = i
  e.dataTransfer!.effectAllowed = "move"
}

function onStageDragEnd() {
  dragIndex.value = null
  dropTarget.value = null
}

function onStageDragOver(i: number) {
  if (dragIndex.value === null || dragIndex.value === i) return
  dropTarget.value = i
}

function onStageDragLeave(i: number) {
  if (dropTarget.value === i) dropTarget.value = null
}

function onStageDrop(i: number) {
  if (dragIndex.value === null || dragIndex.value === i) return
  const moved = stages.value.splice(dragIndex.value, 1)[0]
  stages.value.splice(i > dragIndex.value ? i : i, 0, moved)
  dragIndex.value = null
  dropTarget.value = null
}

function cycleColor(index: number) {
  const current = PALETTE.indexOf(stages.value[index].color)
  stages.value[index].color = PALETTE[(current + 1) % PALETTE.length]
}

function addStage() {
  const usedColors = new Set(stages.value.map((s) => s.color))
  const next = PALETTE.find((c) => !usedColors.has(c)) ?? PALETTE[stages.value.length % PALETTE.length]
  stages.value.push({ name: "", color: next })
}

const canCreate = computed(() => {
  return boardName.value.trim() && stages.value.some((s) => s.name.trim())
})

function handleCreate() {
  if (!canCreate.value) return
  const validStages = stages.value.filter((s) => s.name.trim())
  const board = createBoard(boardName.value.trim(), validStages)
  emit("created", board.id)
}
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
  padding: 16px;
}

.modal-panel {
  width: 100%;
  max-width: 440px;
  background: rgba(15, 23, 42, 0.97);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 25px 60px rgba(0, 0, 0, 0.5);
  animation: slideUp 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(16px) scale(0.97); }
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.modal-title {
  font-size: 16px;
  font-weight: 700;
  color: #f1f5f9;
  margin: 0;
}

.modal-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 7px;
  background: rgba(255, 255, 255, 0.04);
  color: #64748b;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.modal-close:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #94a3b8;
}

.modal-body {
  padding: 18px 20px;
}

.field-label {
  display: block;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: rgba(148, 163, 184, 0.4);
  margin-bottom: 8px;
}

.field-label--stages {
  margin-top: 18px;
}

.field-input {
  width: 100%;
  padding: 10px 14px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 10px;
  color: #e2e8f0;
  font-size: 14px;
  font-family: inherit;
  outline: none;
  transition: border-color 0.15s;
}

.field-input::placeholder { color: rgba(148, 163, 184, 0.3); }
.field-input:focus { border-color: rgba(99, 102, 241, 0.4); }

.stages-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.stage-row {
  display: flex;
  align-items: center;
  gap: 8px;
  border-radius: 8px;
  padding: 2px 0;
  transition: opacity 0.15s, background 0.15s;
}

.stage-row--dragging {
  opacity: 0.35;
}

.stage-row--drop-above {
  box-shadow: 0 -2px 0 0 rgba(99, 102, 241, 0.5);
}

.stage-row--drop-below {
  box-shadow: 0 2px 0 0 rgba(99, 102, 241, 0.5);
}

.drag-handle {
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(148, 163, 184, 0.2);
  cursor: grab;
  flex-shrink: 0;
  transition: color 0.15s;
}

.drag-handle:active {
  cursor: grabbing;
}

.stage-row:hover .drag-handle {
  color: rgba(148, 163, 184, 0.5);
}

.color-dot {
  width: 22px;
  height: 22px;
  border-radius: 7px;
  border: 2px solid rgba(255, 255, 255, 0.1);
  cursor: pointer;
  flex-shrink: 0;
  transition: transform 0.15s, border-color 0.15s;
}

.color-dot:hover {
  transform: scale(1.1);
  border-color: rgba(255, 255, 255, 0.25);
}

.stage-input {
  flex: 1;
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 8px;
  color: #e2e8f0;
  font-size: 13px;
  font-family: inherit;
  outline: none;
  transition: border-color 0.15s;
}

.stage-input::placeholder { color: rgba(148, 163, 184, 0.3); }
.stage-input:focus { border-color: rgba(99, 102, 241, 0.3); }

.stage-remove {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: rgba(148, 163, 184, 0.3);
  cursor: pointer;
  flex-shrink: 0;
  transition: background 0.15s, color 0.15s;
}

.stage-remove:hover {
  background: rgba(239, 68, 68, 0.12);
  color: #fca5a5;
}

.add-stage-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  padding: 6px 12px;
  border: 1px dashed rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  background: transparent;
  color: rgba(148, 163, 184, 0.5);
  font-size: 13px;
  font-family: inherit;
  cursor: pointer;
  transition: border-color 0.15s, color 0.15s, background 0.15s;
}

.add-stage-btn:hover {
  border-color: rgba(99, 102, 241, 0.3);
  color: #a5b4fc;
  background: rgba(99, 102, 241, 0.04);
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 14px 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 9px 18px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  border: none;
  transition: all 0.15s;
}

.btn--ghost {
  background: rgba(255, 255, 255, 0.04);
  color: #64748b;
}

.btn--ghost:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #94a3b8;
}

.btn--primary {
  background: linear-gradient(135deg, #6366f1, #a855f7);
  color: #fff;
  box-shadow: 0 4px 16px rgba(99, 102, 241, 0.25);
}

.btn--primary:hover:not(:disabled) {
  box-shadow: 0 4px 24px rgba(99, 102, 241, 0.4);
}

.btn--primary:disabled {
  opacity: 0.4;
  cursor: default;
}

.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.2s ease;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}
</style>
