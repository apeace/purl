<template>
  <div class="ticket-page">
    <div class="ticket-split">

      <!-- ── Left: strategy bar + ticket detail ──────────── -->
      <div class="workspace">
        <div class="workspace-active">
          <div v-if="context" class="strategy-bar">
            <button class="strategy-header" @click="goBack">
              <component :is="context.icon" v-if="context.icon" :size="16" :style="{ color: context.color }" />
              <span v-else class="strategy-dot" :style="{ background: context.color }" />
              <span class="strategy-header-label">{{ context.label }}</span>
            </button>
            <div class="strategy-nav">
              <span class="strategy-nav-pos">{{ queueIndex + 1 }} / {{ queue.length }}</span>
              <button class="strategy-nav-btn" :disabled="!canGoPrev" @click="goPrev">
                <ChevronLeft :size="18" />
              </button>
              <button class="strategy-nav-btn" :disabled="!canGoNext" @click="goNext">
                <ChevronRight :size="18" />
              </button>
            </div>
          </div>
          <TicketDetail
            v-if="ticket"
            :ticket-id="ticketId"
            show-add-to-board
            @resolve="handleResolve"
            @add-to-board="handleAddToBoard"
          />
          <div v-else class="ticket-not-found">Ticket not found.</div>
        </div>
      </div>

      <!-- ── Right: HUD + health + queue ─────────────────── -->
      <div class="queue-panel">
        <div class="hud">
          <div class="hud-stat">
            <span class="hud-value">{{ hudWaiting }}</span>
            <span class="hud-label">waiting</span>
          </div>
          <div class="hud-divider" />
          <div class="hud-stat">
            <span class="hud-value">{{ hudLongestWait }}</span>
            <span class="hud-label">longest wait</span>
          </div>
          <div class="hud-divider" />
          <div class="hud-stat">
            <span class="hud-value">{{ hudResolvedToday }}</span>
            <span class="hud-label">resolved today</span>
          </div>
        </div>
        <ShiftHealth />
        <div v-if="queue.length" class="queue-list">
          <div class="queue-section-label">{{ context?.label ?? "Queue" }}</div>
          <button
            v-for="item in queue"
            :key="item.id"
            :ref="el => setQueueCardRef(el, item.id)"
            class="queue-card"
            :class="{ 'queue-card--active': item.id === ticketId }"
            @click="navigateTo(item.id)"
          >
            <div class="qcard-top">
              <div class="qcard-avatar" :style="{ background: item.avatarColor }">
                {{ item.name[0] }}
              </div>
              <div class="qcard-meta">
                <div class="qcard-name">{{ item.name }}
                  <span v-if="item.company" class="qcard-company">· {{ item.company }}</span>
                </div>
                <div class="qcard-title">{{ item.aiTitle ?? item.subject }}</div>
              </div>
              <div
                v-if="item.aiTemperature"
                class="qcard-temp"
                :style="{ background: tempColor(item.aiTemperature) }"
                :title="`Temperature: ${item.aiTemperature}/10`"
              />
            </div>
            <div class="qcard-footer">
              <span class="qcard-wait">
                <Clock :size="11" /> {{ item.wait }}
              </span>
              <div
                v-if="item.aiSummary"
                class="qcard-info"
                @click.stop
              >
                <Info :size="11" />
                <div class="qcard-info-tooltip">{{ item.aiSummary }}</div>
              </div>
              <ChevronRight v-else-if="item.id === ticketId" :size="14" class="qcard-active-arrow" />
            </div>
          </button>
        </div>
      </div>

    </div>

    <!-- ── Add-to-board pickers ─────────────────────────── -->
    <Transition name="picker-fade">
      <div v-if="showBoardPicker" class="picker-backdrop" @click.self="closeBoardPicker">
        <div class="picker-panel">
          <div class="picker-header">
            <h3 class="picker-title">Add to Board</h3>
            <button class="picker-close" @click="closeBoardPicker">
              <X :size="14" />
            </button>
          </div>
          <div class="picker-subtitle">Choose a board</div>
          <div class="picker-list">
            <button
              v-for="board in availableBoards"
              :key="board.id"
              class="picker-option"
              @click="pickBoard(board.id)"
            >
              <span class="picker-dot" :style="{ background: board.stages[0]?.color ?? '#94a3b8' }" />
              <span class="picker-name">{{ board.name }}</span>
            </button>
            <div v-if="!availableBoards.length" class="picker-empty">No custom boards yet</div>
          </div>
        </div>
      </div>
    </Transition>

    <StagePickerModal
      :visible="showStagePicker"
      :board-name="stagePickerBoardName"
      :stages="stagePickerStages"
      @close="closeStagePicker"
      @pick="onStagePicked"
    />
  </div>
</template>

<script setup lang="ts">
import { ChevronLeft, ChevronRight, Clock, Flame, Hourglass, Info, ListOrdered, X } from "lucide-vue-next"
import { storeToRefs } from "pinia"
import { type Component, computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue"
import { useRoute, useRouter } from "vue-router"
import ShiftHealth from "../components/ShiftHealth.vue"
import StagePickerModal from "../components/StagePickerModal.vue"
import TicketDetail from "../components/TicketDetail.vue"
import { useKanbanStore } from "../stores/useKanbanStore"
import type { BoardStage } from "../stores/useKanbanStore"
import { lastCustomerReplyMs, waitingMinutes, useTicketStore } from "../stores/useTicketStore"

const route = useRoute()
const router = useRouter()

const ticketStore = useTicketStore()
const { hudLongestWait, hudResolvedToday, hudWaiting, openTickets, tickets } = storeToRefs(ticketStore)
const { resolveTicket } = ticketStore

const kanbanStore = useKanbanStore()
const { boards } = storeToRefs(kanbanStore)
const { addCardToBoard, getBoardById } = kanbanStore

// ── Current ticket ────────────────────────────────────────

const ticketId = computed(() => route.params.id as string)
const ticket = computed(() => tickets.value.find((t) => t.id === ticketId.value) ?? null)

// ── Context from query params ─────────────────────────────

type Context = { label: string; color: string; icon?: Component }

const context = computed((): Context | null => {
  const q = route.query
  if (q.queue === "longest") return { label: "Longest waiting", color: "#fbbf24", icon: Hourglass }
  if (q.queue === "work") return { label: "Work the queue", color: "#818cf8", icon: ListOrdered }
  if (q.queue === "urgent") return { label: "Urgent first", color: "#f87171", icon: Flame }
  if (q.board && q.stage) {
    const board = boards.value.find((b) => b.id === q.board)
    const stage = board?.stages.find((s) => s.id === q.stage)
    if (stage) return { label: stage.name, color: stage.color }
  }
  return null
})

// ── Queue built from context ──────────────────────────────

const queue = computed(() => {
  const q = route.query
  if (q.queue === "longest") {
    return [...openTickets.value].sort((a, b) => waitingMinutes(b) - waitingMinutes(a))
  }
  if (q.queue === "work") {
    return [...openTickets.value].sort((a, b) => lastCustomerReplyMs(b) - lastCustomerReplyMs(a))
  }
  if (q.queue === "urgent") {
    return [...openTickets.value].sort((a, b) => (b.aiTemperature ?? 0) - (a.aiTemperature ?? 0))
  }
  if (q.board && q.stage) {
    const board = boards.value.find((b) => b.id === q.board)
    if (!board) return []
    const stageId = q.stage as string
    const assignedIds = Object.entries(board.cardAssignments)
      .filter(([, sid]) => sid === stageId)
      .map(([tid]) => tid)
    return tickets.value.filter((t) => assignedIds.includes(t.id))
  }
  return []
})

function tempColor(n: number): string {
  if (n <= 3) return "#34d399"
  if (n <= 6) return "#fbbf24"
  if (n <= 8) return "#f97316"
  return "#ef4444"
}

// ── Queue navigation ──────────────────────────────────────

const queueIndex = computed(() => queue.value.findIndex((t) => t.id === ticketId.value))
const canGoPrev = computed(() => queueIndex.value > 0)
const canGoNext = computed(() => queueIndex.value < queue.value.length - 1)

function navigateTo(id: string) {
  router.replace({ path: `/ticket/${id}`, query: route.query })
}

function goPrev() {
  if (canGoPrev.value) navigateTo(queue.value[queueIndex.value - 1].id)
}

function goNext() {
  if (canGoNext.value) navigateTo(queue.value[queueIndex.value + 1].id)
}

function goBack() {
  if (window.history.state?.back) {
    router.back()
  } else {
    router.push("/go")
  }
}

function handleResolve() {
  const idx = queueIndex.value
  const next = queue.value[idx + 1] ?? queue.value[idx - 1] ?? null
  resolveTicket(ticketId.value)
  if (next) {
    navigateTo(next.id)
  } else {
    goBack()
  }
}

// ── Queue card scroll-into-view ───────────────────────────

const queueCardRefs = ref<Record<string, HTMLElement>>({})

function setQueueCardRef(el: unknown, id: string) {
  if (el instanceof HTMLElement) queueCardRefs.value[id] = el
}

watch(ticketId, (id) => {
  nextTick(() => queueCardRefs.value[id]?.scrollIntoView({ block: "nearest" }))
})

// ── Add-to-board flow ─────────────────────────────────────

const showBoardPicker = ref(false)
const addToBoardTicketId = ref<string | null>(null)
const showStagePicker = ref(false)
const stagePickerBoardName = ref("")
const stagePickerStages = ref<BoardStage[]>([])
const stagePickerBoardId = ref<string | null>(null)

const availableBoards = computed(() => boards.value.filter((b) => !b.isDefault))

function handleAddToBoard(id: string) {
  addToBoardTicketId.value = id
  showBoardPicker.value = true
}

function closeBoardPicker() {
  showBoardPicker.value = false
  addToBoardTicketId.value = null
}

function pickBoard(pickedBoardId: string) {
  showBoardPicker.value = false
  const board = getBoardById(pickedBoardId)
  if (!board) return
  stagePickerBoardId.value = pickedBoardId
  stagePickerBoardName.value = board.name
  stagePickerStages.value = board.stages
  showStagePicker.value = true
}

function onStagePicked(stageId: string) {
  if (stagePickerBoardId.value && addToBoardTicketId.value) {
    addCardToBoard(stagePickerBoardId.value, addToBoardTicketId.value, stageId)
  }
  closeStagePicker()
}

function closeStagePicker() {
  showStagePicker.value = false
  stagePickerBoardId.value = null
  stagePickerBoardName.value = ""
  stagePickerStages.value = []
  addToBoardTicketId.value = null
}

// ── Keyboard ──────────────────────────────────────────────

function onKeydown(e: KeyboardEvent) {
  if (e.key === "Escape") goBack()
}

onMounted(() => document.addEventListener("keydown", onKeydown))
onBeforeUnmount(() => document.removeEventListener("keydown", onKeydown))
</script>

<style scoped>
.ticket-page {
  margin: -28px;
}

.ticket-split {
  display: flex;
  height: 100dvh;
}

/* ── Workspace (left) ───────────────────────────────────── */

.workspace {
  flex: 7;
  min-width: 0;
  display: flex;
  flex-direction: column;
  border-right: 1px solid rgba(255, 255, 255, 0.05);
}

.workspace-active {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  animation: content-up 0.35s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes content-up {
  from { opacity: 0; transform: translateY(12px); }
}

/* ── Strategy bar ───────────────────────────────────────── */

.strategy-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  flex-shrink: 0;
}

.strategy-header {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  background: none;
  border: none;
  padding: 12px 20px;
  cursor: pointer;
  font-family: inherit;
  transition: background 0.15s;
  border-radius: 0;
}

.strategy-header:hover {
  background: rgba(255, 255, 255, 0.04);
}

.strategy-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.strategy-header-label {
  font-size: 18px;
  font-weight: 600;
  color: #e2e8f0;
}

.strategy-nav {
  display: flex;
  align-items: center;
  gap: 4px;
  padding-right: 12px;
}

.strategy-nav-pos {
  font-size: 15px;
  font-weight: 600;
  color: rgba(148, 163, 184, 0.45);
  margin-right: 8px;
}

.strategy-nav-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 9px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(255, 255, 255, 0.03);
  color: #e2e8f0;
  cursor: pointer;
  font-family: inherit;
  transition: all 0.15s;
}

.strategy-nav-btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.12);
}

.strategy-nav-btn:disabled {
  opacity: 0.25;
  cursor: default;
}

/* ── Not-found state ────────────────────────────────────── */

.ticket-not-found {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
  color: rgba(148, 163, 184, 0.4);
}

/* ── Queue panel (right) ────────────────────────────────── */

.queue-panel {
  flex: 3;
  min-width: 0;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

/* ── HUD ────────────────────────────────────────────────── */

.hud {
  display: flex;
  align-items: center;
  gap: 0;
  padding: 14px 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  flex-shrink: 0;
  min-width: 0;
}

.hud-stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
  gap: 2px;
  min-width: 0;
  padding: 0 4px;
}

.hud-value {
  font-size: 18px;
  font-weight: 700;
  color: #f1f5f9;
  letter-spacing: -0.02em;
  line-height: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
}

.hud-label {
  font-size: 11px;
  color: rgba(148, 163, 184, 0.45);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  font-weight: 500;
  text-align: center;
  line-height: 1.3;
}

.hud-divider {
  width: 1px;
  height: 24px;
  background: rgba(255, 255, 255, 0.06);
  flex-shrink: 0;
}

/* ── Queue cards ────────────────────────────────────────── */

.queue-list {
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex: 1;
}

.queue-section-label {
  font-size: 13px;
  font-weight: 600;
  color: rgba(148, 163, 184, 0.35);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  margin-bottom: 4px;
}

.queue-card {
  width: 100%;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  padding: 14px;
  cursor: pointer;
  text-align: left;
  font-family: inherit;
  transition: background 0.15s, border-color 0.15s;
}

.queue-card:hover,
.queue-card:active {
  background: rgba(255, 255, 255, 0.05);
  border-color: rgba(255, 255, 255, 0.1);
}

.queue-card--active {
  background: rgba(99, 102, 241, 0.06);
  border-color: rgba(99, 102, 241, 0.22);
}

.queue-card--active:hover,
.queue-card--active:active {
  background: rgba(99, 102, 241, 0.1);
  border-color: rgba(99, 102, 241, 0.35);
}

.qcard-top {
  display: flex;
  align-items: flex-start;
  gap: 9px;
  margin-bottom: 10px;
}

.qcard-avatar {
  width: 30px;
  height: 30px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
  margin-top: 1px;
}

.qcard-meta {
  min-width: 0;
}

.qcard-name {
  font-size: 12px;
  font-weight: 500;
  color: rgba(148, 163, 184, 0.6);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.qcard-company {
  font-weight: 400;
  color: rgba(148, 163, 184, 0.4);
}

.qcard-title {
  font-size: 14px;
  font-weight: 500;
  color: #e2e8f0;
  margin-top: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.qcard-temp {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
  margin-left: auto;
  align-self: flex-start;
  margin-top: 4px;
  transition: transform 0.15s;
}

.queue-card:hover .qcard-temp {
  transform: scale(1.3);
}

.qcard-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.qcard-wait {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: rgba(148, 163, 184, 0.4);
}

.qcard-info {
  display: flex;
  align-items: center;
  color: rgba(148, 163, 184, 0.3);
  cursor: default;
  position: relative;
  transition: color 0.15s;
}

.qcard-info:hover,
.qcard-info:focus-within {
  color: rgba(148, 163, 184, 0.7);
}

.qcard-info-tooltip {
  position: absolute;
  bottom: calc(100% + 6px);
  right: -4px;
  width: 220px;
  background: rgba(15, 23, 42, 0.97);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  padding: 8px 10px;
  font-size: 11px;
  line-height: 1.5;
  color: #94a3b8;
  white-space: normal;
  text-align: left;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.15s;
  z-index: 50;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
}

.qcard-info:hover .qcard-info-tooltip,
.qcard-info:focus-within .qcard-info-tooltip {
  opacity: 1;
}

.qcard-active-arrow {
  color: rgba(129, 140, 248, 0.6);
  flex-shrink: 0;
}

/* ── Board picker ───────────────────────────────────────── */

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
  animation: pickerUp 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes pickerUp {
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

.picker-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.picker-name {
  flex: 1;
}

.picker-empty {
  font-size: 13px;
  color: rgba(148, 163, 184, 0.35);
  padding: 12px;
  text-align: center;
}

.picker-fade-enter-active,
.picker-fade-leave-active {
  transition: opacity 0.15s ease;
}

.picker-fade-enter-from,
.picker-fade-leave-to {
  opacity: 0;
}

/* ── Intermediate screens ───────────────────────────────── */

@media (min-width: 768px) and (max-width: 1099px) {
  .workspace {
    flex: 6;
  }

  .queue-panel {
    flex: 4;
  }

  .strategy-header-label {
    font-size: 15px;
  }
}

/* ── Large screens ──────────────────────────────────────── */

@media (min-width: 1200px) {
  .hud {
    padding: 18px 20px;
  }

  .hud-stat {
    gap: 3px;
    padding: 0 6px;
  }

  .hud-value {
    font-size: 24px;
  }

  .hud-label {
    font-size: 12px;
  }

  .hud-divider {
    height: 28px;
  }
}

/* ── Mobile ─────────────────────────────────────────────── */

@media (max-width: 767px) {
  .ticket-page {
    margin: -16px;
  }

  .ticket-split {
    flex-direction: column;
    height: auto;
  }

  .workspace-active {
    border-right: none;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    height: 65dvh;
  }

  .queue-panel {
    min-width: 0;
    max-height: 60dvh;
  }
}
</style>
