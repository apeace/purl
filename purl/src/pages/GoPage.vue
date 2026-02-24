<template>
  <div class="go-page" :class="{ 'go-page--lobby': !activeThread }">
    <div class="go-split">

      <!-- ── Left: AI Workspace (always present) ─────────── -->
      <div class="workspace">

        <!-- Lobby: AI guides you to a task -->
        <div v-if="!activeThread" class="workspace-lobby">
          <!-- Mobile-only: HUD + health at top of lobby -->
          <div class="mobile-lobby-hud">
            <div class="hud">
              <div class="hud-stat">
                <span class="hud-value">{{ hudOpen }}</span>
                <span class="hud-label">open</span>
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
          </div>

          <div class="lobby-content">
            <div class="ai-lobby-icon">
              <Sparkles :size="28" />
            </div>

            <div class="priority-grid">
              <button
                v-for="opt in priorityOptions"
                :key="opt.id"
                class="priority-card"
                :class="{ 'priority-card--recommended': opt.id === recommendedStrategy }"
                @click="choosePriority(opt)"
              >
                <div v-if="opt.id === recommendedStrategy" class="priority-rec-badge">
                  <Sparkles :size="10" /> Recommended
                </div>
                <ComingSoon v-if="opt.id === 'urgent' || opt.id === 'quick'" />
                <component :is="opt.icon" :size="36" class="priority-icon" :style="{ color: opt.color }" />
                <div class="priority-label">{{ opt.label }}</div>
                <div class="priority-stats">
                  <span class="priority-stat-value">{{ cardStats[opt.id].stat }}</span>
                  <span class="priority-stat-detail">{{ cardStats[opt.id].detail }}</span>
                </div>
                <div v-if="cardPreviews[opt.id]" class="priority-preview">
                  <div class="preview-label">Up next</div>
                  <div class="preview-ticket">
                    <div class="preview-ticket-top">
                      <div class="preview-avatar" :style="{ background: cardPreviews[opt.id]!.avatarColor }">{{ cardPreviews[opt.id]!.name[0] }}</div>
                      <span class="preview-name">{{ cardPreviews[opt.id]!.name }}</span>
                    </div>
                    <div class="preview-subject">{{ cardPreviews[opt.id]!.subject }}</div>
                    <div class="preview-summary">{{ cardPreviews[opt.id]!.messages[0]?.text }}</div>
                  </div>
                </div>
              </button>
            </div>
          </div>
        </div>

        <!-- Active: AI assist + thread conversation -->
        <div v-else class="workspace-active">
          <!-- Strategy bar — back to lobby + queue nav -->
          <div v-if="chosenOption" class="strategy-bar">
            <button class="strategy-header" @click="activeId = null">
              <component :is="chosenOption.icon" :size="16" :style="{ color: chosenOption.color }" />
              <span class="strategy-header-label">{{ chosenOption.label }}</span>
            </button>
            <div class="strategy-nav">
              <span class="strategy-nav-pos">{{ queueIndex + 1 }} / {{ sortedQueue.length }}</span>
              <button class="strategy-nav-btn" :disabled="!canGoPrev" @click="goPrev">
                <ChevronLeft :size="18" />
              </button>
              <button class="strategy-nav-btn" :disabled="!canGoNext" @click="goNext">
                <ChevronRight :size="18" />
              </button>
            </div>
          </div>

          <TicketDetail :ticket-id="activeId!" @resolve="resolve" />
        </div>
      </div>

      <!-- ── Right: Dashboard (always visible) ───────────── -->
      <div class="queue-panel">

        <!-- HUD -->
        <div class="hud">
          <div class="hud-stat">
            <span class="hud-value">{{ hudOpen }}</span>
            <span class="hud-label">open</span>
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

        <!-- Queue (always visible) -->
        <div class="queue-list">
          <div class="queue-section-label">{{ activeThread ? "Up next" : "Your queue" }}</div>
          <button
            v-for="thread in displayQueue"
            :key="thread.id"
            class="queue-card"
            @click="activeId = thread.id"
          >
            <div class="qcard-top">
              <div class="qcard-avatar" :style="{ background: thread.avatarColor }">
                {{ thread.name[0] }}
              </div>
              <div class="qcard-meta">
                <div class="qcard-name">{{ thread.name }}
                  <span class="qcard-company">· {{ thread.company }}</span>
                </div>
                <div class="qcard-subject">{{ thread.subject }}</div>
              </div>
            </div>
            <div class="qcard-footer">
              <span class="qcard-wait">
                <Clock :size="11" /> {{ thread.wait }}
              </span>
            </div>
          </button>
        </div>

      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ChevronLeft, ChevronRight, Clock, Flame, Hourglass, ListOrdered, Sparkles, Zap } from "lucide-vue-next"
import { storeToRefs } from "pinia"
import { computed, ref, watch } from "vue"
import ComingSoon from "../components/ComingSoon.vue"
import ShiftHealth from "../components/ShiftHealth.vue"
import TicketDetail from "../components/TicketDetail.vue"
import { useAiStore } from "../stores/useAiStore"
import { parseWait, useTicketStore } from "../stores/useTicketStore"
import type { Ticket } from "../stores/useTicketStore"

const ticketStore = useTicketStore()
const {
  hudLongestWait,
  hudOpen,
  openTickets: threads,
  resolvedToday,
} = storeToRefs(ticketStore)
const { resolveTicket } = ticketStore

const aiStore = useAiStore()
const { suggestions: aiSuggestions } = storeToRefs(aiStore)

const priorityOptions = [
  { id: "urgent", label: "Urgent first", description: "Tackle high-priority tickets before they escalate", icon: Flame, color: "#f87171" },
  { id: "waiting", label: "Longest waiting", description: "Help the customers who've been waiting the most", icon: Hourglass, color: "#fbbf24" },
  { id: "quick", label: "Quick wins", description: "Clear straightforward tickets to build momentum", icon: Zap, color: "#34d399" },
  { id: "queue", label: "Work the queue", description: "Go through tickets in the order they came in", icon: ListOrdered, color: "#818cf8" },
]

// ── State ────────────────────────────────────────────────

const activeId = ref<string | null>(null)
const chosenPriority = ref<string | null>(null)

const activeThread = computed(() => activeId.value != null ? threads.value.find((t) => t.id === activeId.value) : null)
const queue = computed(() => threads.value.filter((t) => t.id !== activeId.value))
const displayQueue = computed(() => activeThread.value ? queue.value : threads.value)

const queueDetail = computed(() => {
  const counts: Record<string, number> = {}
  for (const t of tickets.value) {
    if (t.status === "new" || t.status === "open") continue
    counts[t.status] = (counts[t.status] ?? 0) + 1
  }
  return Object.entries(counts).map(([s, n]) => `${n} ${statusLabel(s)}`).join(" · ") || "0 resolved"
})

const cardStats = computed<Record<string, { stat: string; detail: string }>>(() => {
  const readyCount = threads.value.filter((t) => aiSuggestions.value[t.id]).length
  return {
    urgent: { stat: `Longest: ${hudLongestWait.value}`, detail: `${threads.value.length} in queue` },
    waiting: { stat: hudLongestWait.value, detail: `${threads.value.length} in queue` },
    quick: { stat: `${readyCount} AI solutions ready`, detail: `${threads.value.length} in queue` },
    queue: { stat: `${hudOpen.value} open`, detail: `${hudResolvedToday.value}/${DAILY_GOAL} resolved` },
  }
})

const cardPreviews = computed<Record<string, Ticket | null>>(() => {
  const all = [...threads.value]

  const byWait = [...all].sort((a, b) => parseWait(b.wait) - parseWait(a.wait))
  const byQuick = [...all].sort((a, b) => (aiSuggestions.value[b.id] ? 1 : 0) - (aiSuggestions.value[a.id] ? 1 : 0) || parseWait(a.wait) - parseWait(b.wait))

  return {
    urgent: byWait[0] ?? null,
    waiting: byWait[0] ?? null,
    quick: byQuick[0] ?? null,
    queue: all[0] ?? null,
  }
})

const chosenOption = computed(() => priorityOptions.find((o) => o.id === chosenPriority.value) ?? null)

const sortedQueue = computed(() => {
  const all = [...threads.value]
  const id = chosenPriority.value
  if (id === "urgent" || id === "waiting") {
    all.sort((a, b) => parseWait(b.wait) - parseWait(a.wait))
  } else if (id === "quick") {
    all.sort((a, b) => (aiSuggestions.value[b.id] ? 1 : 0) - (aiSuggestions.value[a.id] ? 1 : 0) || parseWait(a.wait) - parseWait(b.wait))
  }
  return all
})

const queueIndex = computed(() => sortedQueue.value.findIndex((t) => t.id === activeId.value))
const canGoPrev = computed(() => queueIndex.value > 0)
const canGoNext = computed(() => queueIndex.value < sortedQueue.value.length - 1)

const recommendedStrategy = computed(() => {
  const maxWait = Math.max(...threads.value.map((t) => parseWait(t.wait)))
  if (maxWait >= 120) return "waiting"
  return "quick"
})

// ── Actions ──────────────────────────────────────────────

function goPrev() {
  if (canGoPrev.value) activeId.value = sortedQueue.value[queueIndex.value - 1].id
}

function goNext() {
  if (canGoNext.value) activeId.value = sortedQueue.value[queueIndex.value + 1].id
}

function choosePriority(opt: typeof priorityOptions[number]) {
  chosenPriority.value = opt.id
  const first = cardPreviews.value[opt.id]
  if (first) {
    activeId.value = first.id
  }
}

function resolve() {
  const currentId = activeId.value
  const nextThread = queue.value[0]
  resolveTicket(currentId!)
  activeId.value = nextThread ? nextThread.id : null
}

watch(activeId, (val) => {
  if (val == null) {
    chosenPriority.value = null
  }
})
</script>

<style scoped>
/* ── Layout ─────────────────────────────────────────────── */

.go-page {
  margin: -28px;
}

.go-split {
  display: flex;
  height: 100dvh;
}

/* ── Workspace (left panel — always present) ───────────── */

.workspace {
  flex: 7;
  min-width: 0;
  display: flex;
  flex-direction: column;
  border-right: 1px solid rgba(255, 255, 255, 0.05);
}

/* ── Lobby state ───────────────────────────────────────── */

.workspace-lobby {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  animation: content-up 0.5s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes content-up {
  from { opacity: 0; transform: translateY(12px); }
}

.lobby-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  padding: 24px 32px;
}

.ai-lobby-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: 18px;
  background: rgba(168, 85, 247, 0.1);
  color: #c084fc;
  margin-bottom: 16px;
  box-shadow: 0 0 48px rgba(168, 85, 247, 0.14);
}

/* ── Priority cards ────────────────────────────────────── */

.priority-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  width: 100%;
  flex: 1;
}

.priority-card {
  position: relative;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 20px;
  padding: 28px 28px 24px;
  text-align: left;
  cursor: pointer;
  font-family: inherit;
  transition: all 0.2s ease;
  display: flex;
  flex-direction: column;
}

.priority-card:hover,
.priority-card:active {
  background: rgba(255, 255, 255, 0.04);
  border-color: rgba(255, 255, 255, 0.12);
  box-shadow: 0 0 32px rgba(255, 255, 255, 0.03);
  transform: translateY(-2px);
}

.priority-card--recommended {
  background: rgba(168, 85, 247, 0.05);
  border-color: rgba(168, 85, 247, 0.2);
  box-shadow: 0 0 32px rgba(168, 85, 247, 0.06);
}

.priority-card--recommended:hover,
.priority-card--recommended:active {
  background: rgba(168, 85, 247, 0.08);
  border-color: rgba(168, 85, 247, 0.3);
  box-shadow: 0 0 40px rgba(168, 85, 247, 0.1);
}

.priority-rec-badge {
  position: absolute;
  top: 16px;
  right: 18px;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  font-weight: 700;
  color: #c084fc;
  background: rgba(168, 85, 247, 0.12);
  border-radius: 8px;
  padding: 5px 10px;
  letter-spacing: 0.03em;
}

.priority-icon {
  margin-bottom: 10px;
}

.priority-label {
  font-size: 26px;
  font-weight: 700;
  color: #f1f5f9;
  margin-bottom: 6px;
  letter-spacing: -0.02em;
}

.priority-stats {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin-bottom: 16px;
}

.priority-stat-value {
  font-size: 20px;
  font-weight: 600;
  color: #e2e8f0;
}

.priority-stat-detail {
  font-size: 15px;
  color: rgba(148, 163, 184, 0.45);
}

.priority-preview {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.preview-label {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: rgba(148, 163, 184, 0.3);
  margin-bottom: 8px;
}

.preview-ticket {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 14px 16px;
  background: rgba(255, 255, 255, 0.025);
  border-radius: 12px;
  min-height: 0;
  overflow: hidden;
  transition: background 0.15s;
}

.priority-card:hover .preview-ticket {
  background: rgba(255, 255, 255, 0.04);
}

.preview-ticket-top {
  display: flex;
  align-items: center;
  gap: 8px;
}

.preview-avatar {
  width: 24px;
  height: 24px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
}

.preview-name {
  font-size: 15px;
  font-weight: 600;
  color: rgba(226, 232, 240, 0.8);
  flex: 1;
}

.preview-subject {
  font-size: 15px;
  font-weight: 600;
  color: rgba(226, 232, 240, 0.65);
  line-height: 1.35;
}

.preview-summary {
  font-size: 14px;
  color: rgba(148, 163, 184, 0.45);
  line-height: 1.55;
  flex: 1;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
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

/* ── Active state ──────────────────────────────────────── */

.workspace-active {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  animation: content-up 0.35s cubic-bezier(0.16, 1, 0.3, 1);
}

/* ── Queue panel (right, always visible) ───────────────── */

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

/* ── Intermediate screens (tablets / small laptops) ────── */

@media (min-width: 768px) and (max-width: 1099px) {
  .workspace {
    flex: 6;
  }

  .queue-panel {
    flex: 4;
  }

  .priority-grid {
    grid-template-columns: 1fr;
  }

  .priority-card {
    padding: 16px 20px;
  }

  .priority-label {
    font-size: 18px;
  }

  .priority-stat-value {
    font-size: 15px;
  }

  .priority-stat-detail {
    font-size: 13px;
  }
}

/* ── Large screens — restore generous sizing ───────────── */

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

/* ── Mobile lobby HUD (hidden on desktop) ──────────────── */

.mobile-lobby-hud {
  display: none;
}

/* ── Mobile ─────────────────────────────────────────────── */

@media (max-width: 767px) {
  .go-page {
    margin: -16px;
  }

  .go-split {
    flex-direction: column;
    height: auto;
  }

  .workspace-active {
    border-right: none;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    height: 65dvh;
  }

  .workspace-lobby {
    min-height: auto;
    flex-direction: column;
    align-items: stretch;
    justify-content: flex-start;
  }

  /* Show the inline HUD at top of lobby */
  .mobile-lobby-hud {
    display: block;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }

  /* Hide queue panel when in lobby */
  .go-page--lobby .queue-panel {
    display: none;
  }

  .lobby-content {
    padding: 16px 16px 20px;
  }

  .ai-lobby-icon {
    width: 44px;
    height: 44px;
    border-radius: 14px;
    margin-bottom: 12px;
  }

  /* Compact 2×2 tile grid */
  .priority-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 10px;
  }

  .priority-card {
    padding: 14px 16px;
    border-radius: 14px;
  }

  .priority-icon {
    width: 24px;
    height: 24px;
    margin-bottom: 6px;
  }

  .priority-label {
    font-size: 15px;
    margin-bottom: 2px;
  }

  .priority-stats {
    margin-bottom: 0;
  }

  .priority-stat-value {
    font-size: 13px;
  }

  .priority-stat-detail {
    font-size: 11px;
  }

  .priority-rec-badge {
    top: 8px;
    right: 10px;
    font-size: 10px;
    padding: 3px 7px;
  }

  /* Hide previews on mobile lobby */
  .priority-preview {
    display: none;
  }

  /* Queue panel visible in active state */
  .queue-panel {
    min-width: 0;
    max-height: 60dvh;
  }
}
</style>
