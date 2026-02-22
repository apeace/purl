<template>
  <div class="stage" :class="{ 'stage--visible': visible, 'stage--dragover': dragOver }">
    <div class="stage-header">
      <div class="stage-dot" :style="{ background: color }" />
      <span class="stage-title">{{ title }}</span>
      <span class="stage-count">{{ count }}</span>
    </div>
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
          <div v-if="item.priority" class="card-footer">
            <span class="card-priority" :class="`card-priority--${item.priority}`">{{ item.priority }}</span>
          </div>
        </div>
      </div>
      <div class="drop-placeholder" :class="{ 'drop-placeholder--active': draggingId && !isSource && dragOver && dropIndex >= items.length }" />
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue"

const props = defineProps({
  title: String,
  count: Number,
  color: String,
  items: Array,
  status: String,
  draggingId: String,
  delay: { type: Number, default: 0 },
})

const emit = defineEmits(["select", "drop", "dragStart", "dragEnd"])

const visible = ref(false)
const dragOver = ref(false)
const dropIndex = ref(-1)
const stageCardsEl = ref(null)
// Counter prevents false dragleave when entering child elements
let enterCount = 0

const isSource = computed(() => props.items.some((i) => i.id === props.draggingId))

function onDragStart(event, id) {
  event.dataTransfer.effectAllowed = "move"
  event.dataTransfer.setData("text/plain", id)
  // Defer so the browser captures the drag image before we hide the card
  requestAnimationFrame(() => emit("dragStart", id))
}

function onDragOver(event) {
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

function onDragEnter() {
  enterCount++
  dragOver.value = true
}

function onDragLeave() {
  enterCount--
  if (enterCount <= 0) {
    enterCount = 0
    dragOver.value = false
  }
}

function onDrop(event) {
  enterCount = 0
  dragOver.value = false
  dropIndex.value = -1
  const ticketId = event.dataTransfer.getData("text/plain")
  if (ticketId) emit("drop", { ticketId, status: props.status })
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

.stage-count {
  font-size: 11px;
  font-weight: 700;
  color: rgba(148, 163, 184, 0.6);
  background: rgba(255, 255, 255, 0.05);
  border-radius: 6px;
  padding: 2px 7px;
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

.card-footer {
  display: flex;
  align-items: center;
  margin-top: 8px;
}

.card-priority {
  font-size: 10px;
  font-weight: 600;
  padding: 2px 7px;
  border-radius: 5px;
  text-transform: capitalize;
}

.card-priority--high {
  background: rgba(239, 68, 68, 0.1);
  color: #fca5a5;
}

.card-priority--medium {
  background: rgba(245, 158, 11, 0.1);
  color: #fcd34d;
}

.card-priority--low {
  background: rgba(52, 211, 153, 0.1);
  color: #6ee7b7;
}
</style>
