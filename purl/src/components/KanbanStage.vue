<template>
  <div class="stage" :class="{ 'stage--visible': visible, 'stage--dragover': dragOver }">
    <div class="stage-header">
      <div
        v-if="canEdit"
        class="column-drag-handle"
        draggable="true"
        @dragstart.stop="onColumnDragStart"
        @dragend.stop="onColumnDragEnd"
      >
        <GripVertical :size="12" />
      </div>
      <div class="stage-dot" :style="{ background: color }" />
      <template v-if="editingTitle">
        <input
          ref="titleInputRef"
          v-model="editValue"
          class="stage-title-input"
          @blur="commitTitle"
          @keydown.enter.prevent="commitTitle"
          @keydown.escape="cancelTitle"
        />
      </template>
      <span v-else class="stage-title" @click="startEditTitle">{{ title }}</span>
      <span class="stage-count">{{ count }}</span>
      <button v-if="canEdit" class="col-menu-btn" @click.stop="openMenu">
        <MoreHorizontal :size="13" />
      </button>
    </div>

    <!-- Column context menu (teleported to avoid clipping) -->
    <Teleport to="body">
      <div v-if="menuVisible" class="col-ctx-backdrop" @click="menuVisible = false">
        <div
          class="col-ctx-menu"
          :class="{ 'col-ctx-menu--wide': menuShowColorPicker }"
          :style="{ top: `${menuY}px`, left: `${menuX}px` }"
          @click.stop
        >
          <button class="col-ctx-item" @click="onMenuRename">
            <Pencil :size="13" />
            <span>Rename</span>
          </button>
          <button class="col-ctx-item" @click="menuShowColorPicker = !menuShowColorPicker">
            <Paintbrush :size="13" />
            <span>Change color</span>
            <span class="col-ctx-color-preview" :style="{ background: color }" />
          </button>
          <div v-if="menuShowColorPicker" class="col-ctx-picker">
            <ColorPicker :model-value="color" @update:model-value="onMenuColorChange" />
          </div>
          <div class="col-ctx-divider" />
          <button class="col-ctx-item col-ctx-item--danger" @click="onMenuDelete">
            <Trash2 :size="13" />
            <span>Delete</span>
          </button>
        </div>
      </div>
    </Teleport>
    <div
      ref="stageCardsEl"
      class="stage-cards"
      @dragover.prevent="onDragOver"
      @dragenter.prevent="onDragEnter"
      @dragleave="onDragLeave"
      @drop="onDrop"
    >
      <div
        v-for="(item, i) in items"
        :key="item.id"
        class="card-slot"
        :class="{
          'card-slot--ghost': item.id === draggingId && isSource && dragOver,
          'card-slot--collapsed': item.id === draggingId && isSource && !dragOver,
          'card-slot--insert-before': draggingId && !isSource && dragOver && i === dropIndex,
        }"
      >
        <div
          class="card"
          :style="{ '--accent': color }"
          draggable="true"
          @dragstart="onDragStart($event, item.id)"
          @dragend="emit('dragEnd')"
          @click="emit('select', item.id)"
        >
          <div class="card-header">
            <div class="card-avatar" :style="{ background: item.avatarColor }">{{ item.name[0] }}</div>
            <div class="card-meta">
              <div class="card-name">{{ item.name }}</div>
              <div class="card-company">{{ item.company }}</div>
            </div>
          </div>
          <div class="card-subject">{{ item.subject }}</div>
        </div>
      </div>
      <div class="drop-placeholder" :class="{ 'drop-placeholder--active': draggingId && !isSource && dragOver && dropIndex >= items.length }" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { GripVertical, MoreHorizontal, Paintbrush, Pencil, Trash2 } from "lucide-vue-next"
import { computed, nextTick, onMounted, ref, watch } from "vue"
import ColorPicker from "./ColorPicker.vue"

interface KanbanItem {
  id: string
  name: string
  company: string
  subject: string
  avatarColor: string
}

const props = withDefaults(defineProps<{
  title: string
  count: number
  color: string
  items: KanbanItem[]
  status: string
  draggingId: string | null
  canEdit?: boolean
  delay?: number
}>(), {
  canEdit: false,
  delay: 0,
})

const emit = defineEmits<{
  select: [id: string]
  drop: [payload: { ticketId: string; status: string }]
  dragStart: [id: string]
  dragEnd: []
  columnDragStart: [stageId: string]
  columnDragEnd: []
  rename: [stageId: string, name: string]
  changeColor: [stageId: string, color: string]
  delete: [stageId: string]
}>()

const visible = ref(false)
const dragOver = ref(false)
const dropIndex = ref(-1)
const stageCardsEl = ref<HTMLElement | null>(null)
const editingTitle = ref(false)
const editValue = ref("")
const titleInputRef = ref<HTMLInputElement | null>(null)
const menuVisible = ref(false)
const menuX = ref(0)
const menuY = ref(0)
const menuShowColorPicker = ref(false)
// Counter prevents false dragleave when entering child elements
let enterCount = 0

const isSource = computed(() => props.items.some((i) => i.id === props.draggingId))

function isCardDrag(event: DragEvent): boolean {
  return event.dataTransfer?.types.includes("text/plain") ?? false
}

function onDragStart(event: DragEvent, id: string) {
  event.dataTransfer!.effectAllowed = "move"
  event.dataTransfer!.setData("text/plain", id)
  // Defer so the browser captures the drag image before we hide the card
  requestAnimationFrame(() => emit("dragStart", id))
}

function onDragOver(event: DragEvent) {
  if (!isCardDrag(event) || !props.canEdit) return
  if (!stageCardsEl.value || !props.draggingId || isSource.value) return
  const slots = stageCardsEl.value.querySelectorAll(".card-slot")
  let index = slots.length
  for (let i = 0; i < slots.length; i++) {
    const rect = slots[i].getBoundingClientRect()
    if (event.clientY < rect.top + rect.height / 2) {
      index = i
      break
    }
  }
  dropIndex.value = index
}

function onDragEnter(event: DragEvent) {
  if (!isCardDrag(event) || !props.canEdit) return
  enterCount++
  dragOver.value = true
}

function onDragLeave(event: DragEvent) {
  if (!isCardDrag(event) || !props.canEdit) return
  enterCount--
  if (enterCount <= 0) {
    enterCount = 0
    dragOver.value = false
  }
}

function onDrop(event: DragEvent) {
  if (!isCardDrag(event) || !props.canEdit) return
  enterCount = 0
  dragOver.value = false
  dropIndex.value = -1
  const ticketId = event.dataTransfer!.getData("text/plain")
  if (ticketId) emit("drop", { ticketId, status: props.status })
}

function onColumnDragStart(event: DragEvent) {
  event.dataTransfer!.effectAllowed = "move"
  event.dataTransfer!.setData("application/x-purl-column", props.status)
  emit("columnDragStart", props.status)
}

function onColumnDragEnd() {
  emit("columnDragEnd")
}

function startEditTitle() {
  if (!props.canEdit) return
  editValue.value = props.title
  editingTitle.value = true
  nextTick(() => titleInputRef.value?.select())
}

function commitTitle() {
  if (editValue.value.trim()) emit("rename", props.status, editValue.value.trim())
  editingTitle.value = false
}

function cancelTitle() {
  editingTitle.value = false
}

function openMenu(event: MouseEvent) {
  menuX.value = event.clientX
  menuY.value = event.clientY
  menuShowColorPicker.value = false
  menuVisible.value = true
}

function onMenuRename() {
  menuVisible.value = false
  startEditTitle()
}

function onMenuColorChange(newColor: string) {
  menuVisible.value = false
  menuShowColorPicker.value = false
  emit("changeColor", props.status, newColor)
}

function onMenuDelete() {
  menuVisible.value = false
  emit("delete", props.status)
}

// Reset drag state when dragging ends (covers Escape/cancel and cross-column cleanup)
watch(() => props.draggingId, (val) => {
  if (!val) {
    enterCount = 0
    dragOver.value = false
    dropIndex.value = -1
  }
})

onMounted(() => {
  setTimeout(() => { visible.value = true }, props.delay)
})
</script>

<style scoped>
.stage {
  min-width: 220px;
  flex: 1 0 220px;
  display: flex;
  flex-direction: column;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 14px;
  padding: 14px;
  opacity: 0;
  transform: translateY(20px);
  transition: opacity 0.5s cubic-bezier(0.16, 1, 0.3, 1), transform 0.5s cubic-bezier(0.16, 1, 0.3, 1), background 0.2s, border-color 0.2s;
}

.stage--visible {
  opacity: 1;
  transform: translateY(0);
}

.stage--dragover {
  background: rgba(255, 255, 255, 0.05);
  border-color: rgba(255, 255, 255, 0.15);
}

.stage-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 14px;
}

.column-drag-handle {
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(148, 163, 184, 0.4);
  cursor: grab;
  opacity: 0;
  transition: opacity 0.15s;
  flex-shrink: 0;
}

.stage-header:hover .column-drag-handle {
  opacity: 1;
}

.column-drag-handle:active {
  cursor: grabbing;
}

.stage-dot {
  width: 8px;
  height: 8px;
  border-radius: 4px;
  flex-shrink: 0;
}

.stage-title {
  font-size: 13px;
  font-weight: 600;
  color: #e2e8f0;
  flex: 1;
}

.stage-title-input {
  flex: 1;
  font-size: 13px;
  font-weight: 600;
  color: #e2e8f0;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(99, 102, 241, 0.5);
  border-radius: 5px;
  padding: 1px 5px;
  font-family: inherit;
  outline: none;
  min-width: 0;
}

.stage-count {
  font-size: 11px;
  font-weight: 700;
  color: rgba(148, 163, 184, 0.6);
  background: rgba(255, 255, 255, 0.05);
  border-radius: 6px;
  padding: 2px 7px;
}

.col-menu-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  flex-shrink: 0;
  border: none;
  border-radius: 5px;
  background: transparent;
  color: rgba(148, 163, 184, 0.4);
  cursor: pointer;
  padding: 0;
  opacity: 0;
  transition: opacity 0.15s, background 0.12s, color 0.12s;
}

.stage-header:hover .col-menu-btn {
  opacity: 1;
}

.col-menu-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #94a3b8;
}

.stage-cards {
  display: flex;
  flex-direction: column;
  flex: 1;
}

.card-slot {
  margin-bottom: 8px;
  max-height: 300px;
  overflow: hidden;
  padding-top: 0;
  position: relative;
  transition: max-height 0.3s cubic-bezier(0.16, 1, 0.3, 1), margin-bottom 0.3s cubic-bezier(0.16, 1, 0.3, 1), padding-top 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.card-slot--ghost .card {
  opacity: 0;
  transition: opacity 0.15s ease;
}

.card-slot--ghost::after {
  content: "";
  position: absolute;
  inset: 0;
  border: 2px dashed rgba(255, 255, 255, 0.08);
  border-radius: 10px;
  pointer-events: none;
}

.card-slot--collapsed {
  max-height: 0;
  margin-bottom: 0;
}

.card-slot--collapsed .card {
  opacity: 0;
}

.card-slot--insert-before {
  padding-top: 76px;
}

.card-slot--insert-before::before {
  content: "";
  position: absolute;
  top: 4px;
  left: 0;
  right: 0;
  height: 68px;
  border: 2px dashed rgba(255, 255, 255, 0.08);
  border-radius: 10px;
}

.card {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 10px;
  padding: 12px 14px;
  cursor: grab;
  transition: background 0.2s, border-color 0.2s, opacity 0.2s;
}

.drop-placeholder {
  border: 2px dashed rgba(255, 255, 255, 0.08);
  border-radius: 10px;
  max-height: 0;
  overflow: hidden;
  transition: max-height 0.3s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.25s ease;
  opacity: 0;
}

.drop-placeholder--active {
  max-height: 72px;
  opacity: 1;
}

.card:hover {
  background: rgba(255, 255, 255, 0.06);
  border-color: color-mix(in srgb, var(--accent) 30%, transparent);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.card-avatar {
  width: 24px;
  height: 24px;
  border-radius: 7px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
}

.card-meta {
  min-width: 0;
}

.card-name {
  font-size: 13px;
  font-weight: 600;
  color: #e2e8f0;
  line-height: 1.2;
}

.card-company {
  font-size: 11px;
  color: rgba(148, 163, 184, 0.5);
}

.card-subject {
  font-size: 12px;
  color: rgba(148, 163, 184, 0.7);
  line-height: 1.35;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

</style>

<style>
.col-ctx-backdrop {
  position: fixed;
  inset: 0;
  z-index: 120;
}

.col-ctx-menu {
  position: fixed;
  min-width: 148px;
  background: rgba(15, 23, 42, 0.97);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 10px;
  padding: 4px;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.5);
  animation: colCtxPop 0.12s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes colCtxPop {
  from { opacity: 0; transform: scale(0.95); }
}

.col-ctx-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 8px 12px;
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

.col-ctx-item:hover {
  background: rgba(255, 255, 255, 0.06);
}

.col-ctx-item--danger:hover {
  background: rgba(239, 68, 68, 0.12);
  color: #fca5a5;
}

.col-ctx-menu--wide {
  min-width: 184px;
}

.col-ctx-color-preview {
  width: 12px;
  height: 12px;
  border-radius: 3px;
  flex-shrink: 0;
  margin-left: auto;
}

.col-ctx-picker {
  padding: 6px 12px 8px;
}

.col-ctx-divider {
  height: 1px;
  background: rgba(255, 255, 255, 0.06);
  margin: 3px 4px;
}
</style>
