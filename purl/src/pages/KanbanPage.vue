<template>
  <div class="kanban-page">
    <template v-if="!selectedTicketId">
    <div class="kanban-header">
      <h1 class="page-title">Kanban</h1>
      <div class="header-controls">
        <div class="kanban-search">
          <Search :size="14" class="kanban-search-icon" />
          <input
            v-model="searchQuery"
            class="kanban-search-input"
            placeholder="Search cards…"
          />
          <button v-if="searchQuery" class="kanban-search-clear" @click="searchQuery = ''">
            <X :size="12" />
          </button>
        </div>
        <FilterPanel />
      </div>
    </div>
    <div class="stages-scroll">
      <KanbanStage
        v-for="(stage, i) in stages"
        :key="stage.title"
        v-bind="stage"
        :dragging-id="draggingId"
        :delay="150 + i * 100"
        @select="selectedTicketId = $event"
        @drag-start="draggingId = $event"
        @drag-end="draggingId = null"
        @drop="onDrop"
      />
    </div>
    </template>

    <!-- Split view (ticket selected) -->
    <div v-else class="kanban-split">
      <div class="workspace">
        <div class="workspace-active">
          <div class="strategy-bar">
            <button class="strategy-header" @click="selectedTicketId = null">
              <span class="strategy-dot" :style="{ background: selectedStage?.color }" />
              <span class="strategy-header-label">{{ selectedStage?.title }}</span>
            </button>
            <div class="strategy-nav">
              <span class="strategy-nav-pos">{{ queueIndex + 1 }} / {{ stageQueue.length }}</span>
              <button class="strategy-nav-btn" :disabled="!canGoPrev" @click="goPrev">
                <ChevronLeft :size="18" />
              </button>
              <button class="strategy-nav-btn" :disabled="!canGoNext" @click="goNext">
                <ChevronRight :size="18" />
              </button>
            </div>
          </div>
          <TicketDetail :ticket-id="selectedTicketId" @resolve="handleResolve" />
        </div>
      </div>
      <div class="queue-panel">
        <div class="queue-list">
          <div class="queue-section-label">Up next</div>
          <button
            v-for="item in displayQueue"
            :key="item.id"
            class="queue-card"
            @click="selectedTicketId = item.id"
          >
            <div class="qcard-top">
              <div class="qcard-avatar" :style="{ background: item.avatarColor }">
                {{ item.name[0] }}
              </div>
              <div class="qcard-meta">
                <div class="qcard-name">{{ item.name }}
                  <span v-if="item.company" class="qcard-company">· {{ item.company }}</span>
                </div>
                <div class="qcard-subject">{{ item.subject }}</div>
              </div>
            </div>
            <div class="qcard-footer">
              <span class="qcard-wait">
                <Clock :size="11" /> {{ item.wait }}
              </span>
              <span class="qcard-priority" :class="`qcard-priority--${item.priority}`">
                {{ item.priority }}
              </span>
            </div>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ChevronLeft, ChevronRight, Clock, Search, X } from "lucide-vue-next"
import { computed, onBeforeUnmount, onMounted, ref } from "vue"
import FilterPanel from "../components/FilterPanel.vue"
import KanbanStage from "../components/KanbanStage.vue"
import TicketDetail from "../components/TicketDetail.vue"
import { useTickets } from "../composables/useTickets"
import { STATUS_COLORS } from "../utils/colors"

const { filteredTickets, resolveTicket, setStatus } = useTickets()

const selectedTicketId = ref<string | null>(null)
const draggingId = ref<string | null>(null)
const searchQuery = ref("")

function onDrop({ ticketId, status }: { ticketId: string; status: string }) {
  draggingId.value = null
  setStatus(ticketId, status)
}

const stageDefs = [
  { status: "new", title: "New", color: STATUS_COLORS.new },
  { status: "open", title: "Open", color: STATUS_COLORS.open },
  { status: "pending", title: "Pending", color: STATUS_COLORS.pending },
  { status: "escalated", title: "Technical Escalation", color: STATUS_COLORS.escalated },
  { status: "solved", title: "Solved", color: STATUS_COLORS.solved },
  { status: "closed", title: "Closed", color: STATUS_COLORS.closed },
]

const stages = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  return stageDefs.map((def) => {
    let items = filteredTickets.value.filter((t) => t.status === def.status)
    if (q) {
      items = items.filter((t) => {
        const desc = t.messages[0]?.text ?? ""
        return t.name.toLowerCase().includes(q)
          || t.company.toLowerCase().includes(q)
          || t.subject.toLowerCase().includes(q)
          || desc.toLowerCase().includes(q)
      })
    }
    return {
      title: def.title,
      count: items.length,
      color: def.color,
      status: def.status,
      items: items.map((t) => ({
        id: t.id,
        name: t.name,
        company: t.company,
        subject: t.subject,
        priority: t.priority,
        avatarColor: t.avatarColor,
      })),
    }
  })
})

// ── Split view state ─────────────────────────────────────

const selectedStage = computed(() => {
  if (!selectedTicketId.value) return null
  const ticket = filteredTickets.value.find((t) => t.id === selectedTicketId.value)
  if (!ticket) return null
  return stageDefs.find((s) => s.status === ticket.status) ?? null
})

const stageQueue = computed(() => {
  if (!selectedTicketId.value) return []
  const ticket = filteredTickets.value.find((t) => t.id === selectedTicketId.value)
  if (!ticket) return []
  const q = searchQuery.value.trim().toLowerCase()
  let items = filteredTickets.value.filter((t) => t.status === ticket.status)
  if (q) {
    items = items.filter((t) => {
      const desc = t.messages[0]?.text ?? ""
      return t.name.toLowerCase().includes(q)
        || t.company.toLowerCase().includes(q)
        || t.subject.toLowerCase().includes(q)
        || desc.toLowerCase().includes(q)
    })
  }
  return items
})

const queueIndex = computed(() => stageQueue.value.findIndex((t) => t.id === selectedTicketId.value))
const canGoPrev = computed(() => queueIndex.value > 0)
const canGoNext = computed(() => queueIndex.value < stageQueue.value.length - 1)
const displayQueue = computed(() => stageQueue.value.filter((t) => t.id !== selectedTicketId.value))

function goPrev() {
  if (canGoPrev.value) selectedTicketId.value = stageQueue.value[queueIndex.value - 1].id
}

function goNext() {
  if (canGoNext.value) selectedTicketId.value = stageQueue.value[queueIndex.value + 1].id
}

function handleResolve() {
  const next = displayQueue.value[0] ?? null
  resolveTicket(selectedTicketId.value!)
  selectedTicketId.value = next ? next.id : null
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === "Escape" && selectedTicketId.value) {
    selectedTicketId.value = null
  }
}

onMounted(() => document.addEventListener("keydown", onKeydown))
onBeforeUnmount(() => document.removeEventListener("keydown", onKeydown))
</script>

<style scoped>
.kanban-page {
  min-width: 0;
}

.kanban-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 12px;
}

.page-title {
  font-size: 22px;
  font-weight: 700;
  color: #f1f5f9;
  letter-spacing: -0.02em;
}

.header-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.kanban-search {
  display: flex;
  align-items: center;
  gap: 6px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  padding: 0 10px;
  transition: border-color 0.15s;
}

.kanban-search:focus-within {
  border-color: rgba(99, 102, 241, 0.4);
}

.kanban-search-icon {
  color: rgba(148, 163, 184, 0.4);
  flex-shrink: 0;
}

.kanban-search-input {
  border: none;
  background: transparent;
  padding: 6px 0;
  width: 160px;
  font-size: 13px;
  font-family: inherit;
  color: #e2e8f0;
  outline: none;
}

.kanban-search-input::placeholder {
  color: rgba(148, 163, 184, 0.3);
}

.kanban-search-clear {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border: none;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.06);
  color: rgba(148, 163, 184, 0.6);
  cursor: pointer;
  flex-shrink: 0;
  transition: background 0.15s, color 0.15s;
}

.kanban-search-clear:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #e2e8f0;
}

.stages-scroll {
  display: flex;
  gap: 14px;
  overflow-x: auto;
  padding-bottom: 16px;
}

/* ── Split view ─────────────────────────────────────────── */

.kanban-split {
  display: flex;
  height: calc(100dvh - 56px);
  margin: -28px;
}

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

/* ── Strategy bar ─────────────────────────────────────── */

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

/* ── Queue panel ──────────────────────────────────────── */

.queue-panel {
  flex: 3;
  min-width: 0;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

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
  font-size: 15px;
  font-weight: 600;
  color: #e2e8f0;
}

.qcard-company {
  font-weight: 400;
  color: rgba(148, 163, 184, 0.5);
}

.qcard-subject {
  font-size: 14px;
  color: rgba(148, 163, 184, 0.6);
  margin-top: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
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

.qcard-priority {
  font-size: 12px;
  font-weight: 600;
  padding: 3px 8px;
  border-radius: 5px;
  text-transform: capitalize;
}

.qcard-priority--urgent {
  background: rgba(239, 68, 68, 0.1);
  color: #fca5a5;
}

.qcard-priority--high {
  background: rgba(239, 68, 68, 0.1);
  color: #fca5a5;
}

.qcard-priority--medium {
  background: rgba(245, 158, 11, 0.1);
  color: #fcd34d;
}

.qcard-priority--low {
  background: rgba(52, 211, 153, 0.1);
  color: #6ee7b7;
}

/* ── Mobile ──────────────────────────────────────────────── */

@media (max-width: 767px) {
  .kanban-split {
    flex-direction: column;
    height: auto;
    margin: -16px;
  }

  .workspace-active {
    border-right: none;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    height: 65dvh;
  }

  .queue-panel {
    max-height: 60dvh;
  }
}
</style>
