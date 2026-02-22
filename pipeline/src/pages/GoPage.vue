<template>
  <div class="go-page">
    <div class="go-split">

      <!-- ── Left: AI Workspace (always present) ─────────── -->
      <div class="workspace">

        <!-- Lobby: AI guides you to a task -->
        <div v-if="!activeThread" class="workspace-lobby">
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
                      <div class="preview-avatar" :style="{ background: cardPreviews[opt.id].avatarColor }">{{ cardPreviews[opt.id].name[0] }}</div>
                      <span class="preview-name">{{ cardPreviews[opt.id].name }}</span>
                      <span class="preview-badge" :class="`preview-badge--${cardPreviews[opt.id].priority}`">{{ cardPreviews[opt.id].priority }}</span>
                    </div>
                    <div class="preview-subject">{{ cardPreviews[opt.id].subject }}</div>
                    <div class="preview-summary">{{ cardPreviews[opt.id].messages[0].text }}</div>
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
          <!-- AI assist banner (continuous AI presence) -->
          <div v-if="currentAi" class="ai-assist">
            <div class="ai-assist-top">
              <div class="ai-badge">
                <Sparkles :size="11" /> AI
              </div>
              <span class="ai-assist-headline">{{ currentAi.headline }}</span>
            </div>
            <p class="ai-assist-body">{{ currentAi.body }}</p>
            <button class="btn btn--ai" @click="followAi">
              {{ currentAi.action }} <ChevronRight :size="14" />
            </button>
          </div>

          <!-- Thread header -->
          <div class="thread-header">
            <div class="thread-meta">
              <div class="thread-customer-row">
                <div class="thread-avatar" :style="{ background: activeThread.avatarColor }">
                  {{ activeThread.name[0] }}
                </div>
                <div>
                  <div class="thread-name">{{ activeThread.name }}
                    <span class="thread-company">· {{ activeThread.company }}</span>
                  </div>
                  <div class="thread-id">{{ activeThread.ticketId }}</div>
                </div>
              </div>
              <div class="thread-subject">{{ activeThread.subject }}</div>
            </div>
            <div class="thread-badges">
              <span class="badge" :class="`badge--${activeThread.status}`">{{ activeThread.status }}</span>
              <span class="badge" :class="`badge--${activeThread.priority}`">{{ activeThread.priority }}</span>
            </div>
          </div>

          <!-- Tab bar -->
          <div class="tab-bar">
            <button
              v-for="tab in tabs"
              :key="tab.id"
              class="tab-btn"
              :class="{ 'tab-btn--active': activeTab === tab.id }"
              :title="tab.label"
              @click="activeTab = tab.id"
            >
              <component :is="tab.icon" :size="16" />
            </button>
          </div>

          <!-- ── Tab: Communications ──────────────────── -->
          <template v-if="activeTab === 'comms'">
            <div ref="messagesEl" class="thread-messages">
              <TransitionGroup name="msg">
                <div
                  v-for="msg in activeThread.messages"
                  :key="msg.id"
                  class="message"
                  :class="msg.from === 'agent' ? 'message--agent' : 'message--customer'"
                >
                  <div v-if="msg.from === 'customer'" class="msg-avatar" :style="{ background: activeThread.avatarColor }">
                    {{ activeThread.name[0] }}
                  </div>
                  <div class="msg-bubble">
                    <div class="msg-header">
                      <span class="msg-sender">{{ msg.from === 'agent' ? 'You' : activeThread.name }}</span>
                      <span class="msg-channel" :class="`msg-channel--${msg.channel}`">
                        <Mail v-if="msg.channel === 'email'" :size="10" />
                        <MessageSquare v-else-if="msg.channel === 'sms'" :size="10" />
                        <Phone v-else-if="msg.channel === 'phone'" :size="10" />
                        {{ msg.channel }}
                      </span>
                      <span class="msg-time">{{ msg.time }}</span>
                    </div>
                    <div class="msg-body">{{ msg.text }}</div>
                  </div>
                  <div v-if="msg.from === 'agent'" class="msg-avatar msg-avatar--agent">Y</div>
                </div>
              </TransitionGroup>
            </div>

            <div class="thread-compose">
              <textarea
                v-model="replyText"
                class="compose-input"
                placeholder="Write a reply…"
                rows="3"
                @keydown.meta.enter="sendReply"
              />
              <div class="compose-actions">
                <button class="btn btn--ghost" @click="resolve">Resolve</button>
                <button class="btn btn--primary" :disabled="!replyText.trim()" @click="sendReply">
                  <Send :size="14" /> Send Reply
                </button>
              </div>
            </div>
          </template>

          <!-- ── Tab: Contact Info ────────────────────── -->
          <div v-else-if="activeTab === 'contact'" class="tab-panel">
            <div class="contact-card">
              <div class="contact-avatar" :style="{ background: activeThread.avatarColor }">
                {{ activeThread.name[0] }}
              </div>
              <div class="contact-name">{{ activeThread.name }}</div>
              <div class="contact-company">{{ activeThread.company }}</div>
            </div>

            <div class="detail-grid">
              <div class="detail-row">
                <Mail :size="14" class="detail-icon" />
                <div class="detail-content">
                  <div class="detail-label">Email</div>
                  <div class="detail-value">{{ activeThread.email }}</div>
                </div>
              </div>
              <div class="detail-row">
                <Phone :size="14" class="detail-icon" />
                <div class="detail-content">
                  <div class="detail-label">Phone</div>
                  <div class="detail-value">{{ activeThread.phone }}</div>
                </div>
              </div>
              <div class="detail-row">
                <Zap :size="14" class="detail-icon" />
                <div class="detail-content">
                  <div class="detail-label">Subscription</div>
                  <div class="detail-value">
                    {{ activeThread.subscription.plan }}
                    <span class="sub-status" :class="`sub-status--${activeThread.subscription.status}`">{{ activeThread.subscription.status }}</span>
                  </div>
                  <div class="detail-sub">{{ activeThread.subscription.id }}</div>
                </div>
              </div>
            </div>

            <!-- Internal notes -->
            <div class="notes-section">
              <div class="notes-label">Internal Notes</div>
              <textarea
                class="notes-input"
                placeholder="Add internal notes about this subscriber…"
                rows="4"
                :value="activeThread.notes"
                @input="updateNotes(activeId, $event.target.value)"
              />
            </div>
          </div>

          <!-- ── Tab: Ticket History ──────────────────── -->
          <div v-else-if="activeTab === 'history'" class="tab-panel">
            <div class="panel-title">Activity on {{ activeThread.ticketId }}</div>
            <div class="timeline">
              <div
                v-for="(entry, i) in activeThread.ticketHistory"
                :key="i"
                class="timeline-item"
              >
                <div class="timeline-dot" />
                <div class="timeline-content">
                  <div class="timeline-event">{{ entry.event }}</div>
                  <div class="timeline-time">{{ entry.time }}</div>
                </div>
              </div>
            </div>
          </div>

          <!-- ── Tab: Subscriber History ──────────────── -->
          <div v-else-if="activeTab === 'subscriber'" class="tab-panel">
            <div class="panel-title">{{ activeThread.name }}'s ticket history</div>
            <div class="panel-subtitle">Customer since {{ activeThread.subscription.id }}</div>

            <!-- Current ticket -->
            <div class="history-card history-card--current">
              <div class="history-card-top">
                <span class="history-tid">{{ activeThread.ticketId }}</span>
                <span class="history-status" :class="`history-status--${activeThread.status}`">{{ activeThread.status }}</span>
              </div>
              <div class="history-subject">{{ activeThread.subject }}</div>
              <div class="history-date">Current</div>
            </div>

            <!-- Past tickets -->
            <div
              v-for="(past, i) in activeThread.subscriberHistory"
              :key="i"
              class="history-card"
            >
              <div class="history-card-top">
                <span class="history-tid">{{ past.ticketId }}</span>
                <span class="history-status" :class="`history-status--${past.status}`">{{ past.status }}</span>
              </div>
              <div class="history-subject">{{ past.subject }}</div>
              <div class="history-date">{{ past.date }}</div>
            </div>

            <div v-if="!activeThread.subscriberHistory.length" class="panel-empty">
              No previous tickets on record.
            </div>
          </div>

          <!-- ── Tab: Settings ────────────────────────── -->
          <div v-else-if="activeTab === 'settings'" class="tab-panel">
            <!-- Tags -->
            <div class="settings-section">
              <div class="settings-label">Tags</div>
              <div class="tags-wrap">
                <span
                  v-for="tag in activeThread.tags"
                  :key="tag"
                  class="tag"
                >
                  {{ tag }}
                  <button class="tag-remove" @click="removeTag(activeId, tag)">&times;</button>
                </span>
                <div class="tag-add">
                  <input
                    v-model="newTag"
                    class="tag-input"
                    placeholder="Add tag…"
                    @keydown.enter="handleAddTag"
                  />
                </div>
              </div>
            </div>

            <!-- Temperature -->
            <div class="settings-section">
              <div class="settings-label">Temperature</div>
              <div class="option-row">
                <button
                  v-for="opt in tempOptions"
                  :key="opt"
                  class="option-btn"
                  :class="{ 'option-btn--active': activeThread.temperature === opt, [`option-btn--${opt}`]: true }"
                  @click="setTemperature(activeId, opt)"
                >{{ opt }}</button>
              </div>
            </div>

            <!-- Assignee -->
            <div class="settings-section">
              <div class="settings-label">Assignee</div>
              <select
                class="settings-select"
                :value="activeThread.assignee"
                @change="setAssignee(activeId, $event.target.value)"
              >
                <option v-for="a in assigneeOptions" :key="a" :value="a">{{ a }}</option>
              </select>
            </div>

            <!-- Status -->
            <div class="settings-section">
              <div class="settings-label">Status</div>
              <div class="option-row option-row--wrap">
                <button
                  v-for="s in statusOptions"
                  :key="s"
                  class="option-btn"
                  :class="{ 'option-btn--active': activeThread.status === s }"
                  @click="setStatus(activeId, s)"
                >{{ s }}</button>
              </div>
            </div>
          </div>
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

        <!-- Health bar -->
        <div class="health-bar-wrap">
          <div class="health-bar-labels">
            <span class="health-bar-title">Shift Health</span>
          </div>
          <div class="health-bar-track">
            <div class="health-bar-bg" />
          </div>
        </div>

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
              <span class="qcard-priority" :class="`qcard-priority--${thread.priority}`">
                {{ thread.priority }}
              </span>
            </div>
          </button>
        </div>

      </div>
    </div>
  </div>
</template>

<script setup>
import { ChevronLeft, ChevronRight, Clock, Cog, Flame, History, Hourglass, ListOrdered, Mail, MessageSquare, Phone, Send, Sparkles, User, Users, Zap } from "lucide-vue-next"
import { computed, nextTick, ref, watch } from "vue"
import { useTickets } from "../composables/useTickets.js"

const {
  addTag,
  aiSuggestions,
  followAi: sharedFollowAi,
  hudLongestWait,
  hudOpen,
  hudResolvedToday,
  openTickets: threads,
  parseWait,
  removeTag,
  resolveTicket,
  resolvedToday,
  sendReply: sharedSendReply,
  setAssignee,
  setStatus,
  setTemperature,
  updateNotes,
} = useTickets()

const DAILY_GOAL = 20

const priorityOptions = [
  { id: "urgent", label: "Urgent first", description: "Tackle high-priority tickets before they escalate", icon: Flame, color: "#f87171" },
  { id: "waiting", label: "Longest waiting", description: "Help the customers who've been waiting the most", icon: Hourglass, color: "#fbbf24" },
  { id: "quick", label: "Quick wins", description: "Clear straightforward tickets to build momentum", icon: Zap, color: "#34d399" },
  { id: "queue", label: "Work the queue", description: "Go through tickets in the order they came in", icon: ListOrdered, color: "#818cf8" },
]

// ── State ────────────────────────────────────────────────

const activeId = ref(null)
const chosenPriority = ref(null)
const activeTab = ref("comms")
const newTag = ref("")

const tabs = [
  { id: "comms", icon: MessageSquare, label: "Communications" },
  { id: "contact", icon: User, label: "Contact" },
  { id: "history", icon: History, label: "Ticket History" },
  { id: "subscriber", icon: Users, label: "Subscriber History" },
  { id: "settings", icon: Cog, label: "Settings" },
]

const statusOptions = ["new", "open", "pending", "escalated", "solved", "closed"]
const tempOptions = ["hot", "warm", "cool"]
const assigneeOptions = ["Alex Chen", "Sarah Kim", "Jordan Lee", "Unassigned"]

const barClipRight = computed(() => `${Math.max(0, 100 - (hudResolvedToday.value / DAILY_GOAL) * 100)}%`)
const replyText = ref("")
const messagesEl = ref(null)

const activeThread = computed(() => activeId.value != null ? threads.value.find((t) => t.id === activeId.value) : null)
const queue = computed(() => threads.value.filter((t) => t.id !== activeId.value))
const displayQueue = computed(() => activeThread.value ? queue.value : threads.value)
const currentAi = computed(() => aiSuggestions[activeId.value] ?? null)

const cardStats = computed(() => {
  const highCount = threads.value.filter((t) => t.priority === "high").length
  const readyCount = threads.value.filter((t) => aiSuggestions[t.id]).length
  return {
    urgent: { stat: `${highCount} high priority`, detail: `Longest: ${hudLongestWait.value}` },
    waiting: { stat: hudLongestWait.value, detail: `${threads.value.length} in queue` },
    quick: { stat: `${readyCount} AI solutions ready`, detail: `${threads.value.length - highCount} medium/low` },
    queue: { stat: `${hudOpen.value} open`, detail: `${hudResolvedToday.value}/${DAILY_GOAL} resolved` },
  }
})

const cardPreviews = computed(() => {
  const priorityRank = { high: 0, medium: 1, low: 2 }
  const all = [...threads.value]

  const urgent = [...all].sort((a, b) => priorityRank[a.priority] - priorityRank[b.priority] || parseWait(b.wait) - parseWait(a.wait))
  const waiting = [...all].sort((a, b) => parseWait(b.wait) - parseWait(a.wait))
  const quick = [...all].sort((a, b) => (aiSuggestions[b.id] ? 1 : 0) - (aiSuggestions[a.id] ? 1 : 0) || priorityRank[b.priority] - priorityRank[a.priority] || parseWait(a.wait) - parseWait(b.wait))
  const queueSorted = [...all]

  return {
    urgent: urgent[0] ?? null,
    waiting: waiting[0] ?? null,
    quick: quick[0] ?? null,
    queue: queueSorted[0] ?? null,
  }
})

const chosenOption = computed(() => priorityOptions.find((o) => o.id === chosenPriority.value) ?? null)

const sortedQueue = computed(() => {
  const priorityRank = { high: 0, medium: 1, low: 2 }
  const all = [...threads.value]
  const id = chosenPriority.value
  if (id === "urgent") {
    all.sort((a, b) => priorityRank[a.priority] - priorityRank[b.priority] || parseWait(b.wait) - parseWait(a.wait))
  } else if (id === "waiting") {
    all.sort((a, b) => parseWait(b.wait) - parseWait(a.wait))
  } else if (id === "quick") {
    all.sort((a, b) => (aiSuggestions[b.id] ? 1 : 0) - (aiSuggestions[a.id] ? 1 : 0) || priorityRank[b.priority] - priorityRank[a.priority] || parseWait(a.wait) - parseWait(b.wait))
  }
  return all
})

const queueIndex = computed(() => sortedQueue.value.findIndex((t) => t.id === activeId.value))
const canGoPrev = computed(() => queueIndex.value > 0)
const canGoNext = computed(() => queueIndex.value < sortedQueue.value.length - 1)

const recommendedStrategy = computed(() => {
  const highCount = threads.value.filter((t) => t.priority === "high").length
  if (highCount >= 2) return "urgent"
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

function choosePriority(opt) {
  chosenPriority.value = opt.id
  const first = cardPreviews.value[opt.id]
  if (first) {
    activeId.value = first.id
  }
}

function sendReply() {
  const text = replyText.value.trim()
  if (!text || !activeId.value) return
  sharedSendReply(activeId.value, text)
  replyText.value = ""
  scrollToBottom()
}

function followAi() {
  const suggestion = currentAi.value
  if (!suggestion || !activeId.value) return
  replyText.value = suggestion.replyText
  nextTick(() => sendReply())
}

function resolve() {
  const currentId = activeId.value
  const nextThread = queue.value[0]
  resolveTicket(currentId)
  activeId.value = nextThread ? nextThread.id : null
  replyText.value = ""
}

function handleAddTag() {
  const tag = newTag.value.trim().toLowerCase()
  if (!tag || !activeId.value) return
  addTag(activeId.value, tag)
  newTag.value = ""
}

function scrollToBottom() {
  nextTick(() => {
    if (messagesEl.value) {
      messagesEl.value.scrollTop = messagesEl.value.scrollHeight
    }
  })
}

watch(activeId, (val) => {
  replyText.value = ""
  activeTab.value = "comms"
  scrollToBottom()
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

.priority-card:hover {
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

.priority-card--recommended:hover {
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

.preview-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 7px;
  border-radius: 5px;
  text-transform: capitalize;
  flex-shrink: 0;
}

.preview-badge--high {
  background: rgba(239, 68, 68, 0.1);
  color: #fca5a5;
}

.preview-badge--medium {
  background: rgba(245, 158, 11, 0.1);
  color: #fcd34d;
}

.preview-badge--low {
  background: rgba(52, 211, 153, 0.1);
  color: #6ee7b7;
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

/* ── AI assist banner ──────────────────────────────────── */

.ai-assist {
  padding: 18px 24px;
  background: rgba(168, 85, 247, 0.04);
  border-bottom: 1px solid rgba(168, 85, 247, 0.12);
  flex-shrink: 0;
}

.ai-assist-top {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.ai-assist-headline {
  font-size: 17px;
  font-weight: 600;
  color: #e2e8f0;
  line-height: 1.3;
}

.ai-assist-body {
  font-size: 15px;
  color: rgba(148, 163, 184, 0.65);
  line-height: 1.6;
  margin: 0 0 4px;
  min-height: 1.6em;
}

/* ── Thread header ─────────────────────────────────────── */

.thread-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 18px 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  flex-shrink: 0;
}

.thread-customer-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.thread-avatar {
  width: 40px;
  height: 40px;
  border-radius: 11px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
}

.thread-name {
  font-size: 17px;
  font-weight: 600;
  color: #e2e8f0;
}

.thread-company {
  font-weight: 400;
  color: rgba(148, 163, 184, 0.6);
}

.thread-id {
  font-size: 13px;
  color: rgba(148, 163, 184, 0.4);
  margin-top: 2px;
}

.thread-subject {
  font-size: 18px;
  font-weight: 600;
  color: #f1f5f9;
  letter-spacing: -0.01em;
}

.thread-badges {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.badge {
  font-size: 13px;
  font-weight: 600;
  padding: 4px 10px;
  border-radius: 7px;
  text-transform: capitalize;
}

.badge--new {
  background: rgba(56, 189, 248, 0.15);
  color: #7dd3fc;
}

.badge--open {
  background: rgba(99, 102, 241, 0.15);
  color: #a5b4fc;
}

.badge--pending {
  background: rgba(168, 85, 247, 0.15);
  color: #d8b4fe;
}

.badge--escalated {
  background: rgba(249, 115, 22, 0.15);
  color: #fdba74;
}

.badge--solved {
  background: rgba(52, 211, 153, 0.15);
  color: #6ee7b7;
}

.badge--closed {
  background: rgba(148, 163, 184, 0.15);
  color: #94a3b8;
}

.badge--high {
  background: rgba(239, 68, 68, 0.12);
  color: #fca5a5;
}

.badge--medium {
  background: rgba(245, 158, 11, 0.12);
  color: #fcd34d;
}

.badge--low {
  background: rgba(52, 211, 153, 0.12);
  color: #6ee7b7;
}

/* ── Tab bar ────────────────────────────────────────────── */

.tab-bar {
  display: flex;
  gap: 2px;
  padding: 6px 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  flex-shrink: 0;
}

.tab-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 32px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: rgba(148, 163, 184, 0.4);
  cursor: pointer;
  font-family: inherit;
  transition: all 0.15s;
}

.tab-btn:hover {
  background: rgba(255, 255, 255, 0.05);
  color: #94a3b8;
}

.tab-btn--active {
  background: rgba(99, 102, 241, 0.12);
  color: #a5b4fc;
}

/* ── Channel badges ────────────────────────────────────── */

.msg-channel {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  font-size: 10px;
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 4px;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.msg-channel--email {
  background: rgba(99, 102, 241, 0.1);
  color: #a5b4fc;
}

.msg-channel--sms {
  background: rgba(52, 211, 153, 0.1);
  color: #6ee7b7;
}

.msg-channel--phone {
  background: rgba(245, 158, 11, 0.1);
  color: #fcd34d;
}

/* ── Tab panels (shared) ───────────────────────────────── */

.tab-panel {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  animation: content-up 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}

.panel-title {
  font-size: 16px;
  font-weight: 600;
  color: #e2e8f0;
  margin-bottom: 4px;
}

.panel-subtitle {
  font-size: 13px;
  color: rgba(148, 163, 184, 0.45);
  margin-bottom: 20px;
}

.panel-empty {
  font-size: 14px;
  color: rgba(148, 163, 184, 0.35);
  padding: 24px 0;
  text-align: center;
}

/* ── Contact tab ───────────────────────────────────────── */

.contact-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px 0 28px;
}

.contact-avatar {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  font-weight: 700;
  color: #fff;
  margin-bottom: 12px;
}

.contact-name {
  font-size: 18px;
  font-weight: 700;
  color: #f1f5f9;
}

.contact-company {
  font-size: 14px;
  color: rgba(148, 163, 184, 0.5);
  margin-top: 2px;
}

.detail-grid {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.detail-row {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  padding: 14px 16px;
  border-radius: 10px;
  transition: background 0.15s;
}

.detail-row:hover {
  background: rgba(255, 255, 255, 0.03);
}

.detail-icon {
  color: rgba(148, 163, 184, 0.4);
  margin-top: 2px;
  flex-shrink: 0;
}

.detail-label {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: rgba(148, 163, 184, 0.35);
  margin-bottom: 2px;
}

.detail-value {
  font-size: 15px;
  font-weight: 500;
  color: #e2e8f0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.detail-sub {
  font-size: 12px;
  color: rgba(148, 163, 184, 0.4);
  margin-top: 2px;
}

.sub-status {
  font-size: 10px;
  font-weight: 700;
  padding: 2px 7px;
  border-radius: 5px;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.sub-status--active {
  background: rgba(52, 211, 153, 0.1);
  color: #6ee7b7;
}

.sub-status--trial {
  background: rgba(245, 158, 11, 0.1);
  color: #fcd34d;
}

.sub-status--churned {
  background: rgba(239, 68, 68, 0.1);
  color: #fca5a5;
}

/* ── Notes ──────────────────────────────────────────────── */

.notes-section {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
}

.notes-label {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: rgba(148, 163, 184, 0.4);
  margin-bottom: 10px;
}

.notes-input {
  width: 100%;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 10px;
  padding: 12px 16px;
  color: #e2e8f0;
  font-size: 14px;
  font-family: inherit;
  line-height: 1.6;
  resize: vertical;
  outline: none;
  transition: border-color 0.15s;
  min-height: 80px;
}

.notes-input::placeholder {
  color: rgba(148, 163, 184, 0.3);
}

.notes-input:focus {
  border-color: rgba(99, 102, 241, 0.4);
}

/* ── Timeline (ticket history) ─────────────────────────── */

.timeline {
  display: flex;
  flex-direction: column;
  padding-top: 8px;
}

.timeline-item {
  display: flex;
  gap: 14px;
  padding: 10px 0;
  position: relative;
}

/* Vertical connector line */
.timeline-item:not(:last-child)::before {
  content: "";
  position: absolute;
  left: 5px;
  top: 28px;
  bottom: -2px;
  width: 1px;
  background: rgba(255, 255, 255, 0.06);
}

.timeline-dot {
  width: 11px;
  height: 11px;
  border-radius: 50%;
  background: rgba(99, 102, 241, 0.3);
  border: 2px solid rgba(99, 102, 241, 0.5);
  flex-shrink: 0;
  margin-top: 3px;
}

.timeline-event {
  font-size: 14px;
  color: #e2e8f0;
  line-height: 1.4;
}

.timeline-time {
  font-size: 12px;
  color: rgba(148, 163, 184, 0.35);
  margin-top: 2px;
}

/* ── History cards (subscriber history) ────────────────── */

.history-card {
  padding: 14px 16px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.05);
  margin-bottom: 8px;
  transition: background 0.15s;
}

.history-card:hover {
  background: rgba(255, 255, 255, 0.04);
}

.history-card--current {
  border-color: rgba(99, 102, 241, 0.2);
  background: rgba(99, 102, 241, 0.04);
}

.history-card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}

.history-tid {
  font-size: 12px;
  font-weight: 600;
  color: rgba(148, 163, 184, 0.5);
}

.history-status {
  font-size: 10px;
  font-weight: 700;
  padding: 2px 7px;
  border-radius: 5px;
  text-transform: capitalize;
}

.history-status--open,
.history-status--new {
  background: rgba(99, 102, 241, 0.12);
  color: #a5b4fc;
}

.history-status--pending {
  background: rgba(168, 85, 247, 0.12);
  color: #d8b4fe;
}

.history-status--escalated {
  background: rgba(249, 115, 22, 0.12);
  color: #fdba74;
}

.history-status--solved {
  background: rgba(52, 211, 153, 0.12);
  color: #6ee7b7;
}

.history-status--closed {
  background: rgba(148, 163, 184, 0.12);
  color: #94a3b8;
}

.history-subject {
  font-size: 14px;
  font-weight: 500;
  color: #e2e8f0;
  line-height: 1.35;
}

.history-date {
  font-size: 12px;
  color: rgba(148, 163, 184, 0.35);
  margin-top: 4px;
}

/* ── Settings tab ──────────────────────────────────────── */

.settings-section {
  margin-bottom: 24px;
}

.settings-label {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: rgba(148, 163, 184, 0.4);
  margin-bottom: 10px;
}

.tags-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  font-weight: 500;
  color: #c7d2fe;
  background: rgba(99, 102, 241, 0.1);
  border: 1px solid rgba(99, 102, 241, 0.15);
  border-radius: 6px;
  padding: 4px 8px;
}

.tag-remove {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  border: none;
  background: transparent;
  color: rgba(199, 210, 254, 0.5);
  cursor: pointer;
  font-size: 14px;
  line-height: 1;
  padding: 0;
  border-radius: 3px;
  transition: color 0.15s, background 0.15s;
}

.tag-remove:hover {
  color: #fca5a5;
  background: rgba(239, 68, 68, 0.15);
}

.tag-add {
  display: inline-flex;
}

.tag-input {
  width: 80px;
  padding: 4px 8px;
  font-size: 12px;
  font-family: inherit;
  background: rgba(255, 255, 255, 0.03);
  border: 1px dashed rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  color: #e2e8f0;
  outline: none;
  transition: border-color 0.15s;
}

.tag-input::placeholder {
  color: rgba(148, 163, 184, 0.3);
}

.tag-input:focus {
  border-color: rgba(99, 102, 241, 0.4);
  border-style: solid;
}

.option-row {
  display: flex;
  gap: 6px;
}

.option-row--wrap {
  flex-wrap: wrap;
}

.option-btn {
  padding: 6px 14px;
  font-size: 13px;
  font-weight: 500;
  font-family: inherit;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(255, 255, 255, 0.02);
  color: rgba(148, 163, 184, 0.6);
  cursor: pointer;
  text-transform: capitalize;
  transition: all 0.15s;
}

.option-btn:hover {
  background: rgba(255, 255, 255, 0.05);
  border-color: rgba(255, 255, 255, 0.12);
  color: #e2e8f0;
}

.option-btn--active {
  background: rgba(99, 102, 241, 0.12);
  border-color: rgba(99, 102, 241, 0.25);
  color: #a5b4fc;
}

.option-btn--hot.option-btn--active {
  background: rgba(239, 68, 68, 0.12);
  border-color: rgba(239, 68, 68, 0.25);
  color: #fca5a5;
}

.option-btn--warm.option-btn--active {
  background: rgba(245, 158, 11, 0.12);
  border-color: rgba(245, 158, 11, 0.25);
  color: #fcd34d;
}

.option-btn--cool.option-btn--active {
  background: rgba(52, 211, 153, 0.12);
  border-color: rgba(52, 211, 153, 0.25);
  color: #6ee7b7;
}

.settings-select {
  width: 100%;
  padding: 9px 14px;
  font-size: 14px;
  font-family: inherit;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 9px;
  color: #e2e8f0;
  outline: none;
  cursor: pointer;
  appearance: none;
  transition: border-color 0.15s;
}

.settings-select:focus {
  border-color: rgba(99, 102, 241, 0.4);
}

.settings-select option {
  background: #0f172a;
  color: #e2e8f0;
}

/* ── Messages ───────────────────────────────────────────── */

.thread-messages {
  flex: 1;
  overflow-y: auto;
  padding: 20px 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.message {
  display: flex;
  gap: 12px;
  align-items: flex-end;
}

.message--agent {
  flex-direction: row-reverse;
}

.msg-avatar {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
}

.msg-avatar--agent {
  background: linear-gradient(135deg, #6366f1, #ec4899);
}

.msg-bubble {
  max-width: 72%;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 14px;
  padding: 14px 18px;
}

.message--agent .msg-bubble {
  background: rgba(99, 102, 241, 0.1);
  border-color: rgba(99, 102, 241, 0.2);
}

.msg-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.msg-sender {
  font-size: 14px;
  font-weight: 600;
  color: #94a3b8;
}

.msg-time {
  font-size: 13px;
  color: rgba(148, 163, 184, 0.4);
}

.msg-body {
  font-size: 16px;
  color: #e2e8f0;
  line-height: 1.6;
}

/* ── Compose ────────────────────────────────────────────── */

.thread-compose {
  border-top: 1px solid rgba(255, 255, 255, 0.05);
  padding: 16px 24px;
  flex-shrink: 0;
}

.compose-input {
  width: 100%;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 12px;
  padding: 14px 18px;
  color: #e2e8f0;
  font-size: 16px;
  font-family: inherit;
  resize: none;
  outline: none;
  transition: border-color 0.15s;
}

.compose-input::placeholder {
  color: rgba(148, 163, 184, 0.3);
}

.compose-input:focus {
  border-color: rgba(99, 102, 241, 0.4);
}

.compose-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 10px;
}

/* ── Buttons ────────────────────────────────────────────── */

.btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 18px;
  border-radius: 10px;
  font-size: 16px;
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

.btn--ai {
  background: rgba(168, 85, 247, 0.12);
  color: #c084fc;
  border: 1px solid rgba(168, 85, 247, 0.2);
  width: 100%;
  justify-content: center;
  margin-top: 8px;
}

.btn--ai:hover {
  background: rgba(168, 85, 247, 0.2);
  border-color: rgba(168, 85, 247, 0.35);
}

.ai-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 13px;
  font-weight: 700;
  color: #c084fc;
  background: rgba(168, 85, 247, 0.15);
  border-radius: 6px;
  padding: 3px 8px;
  letter-spacing: 0.04em;
  flex-shrink: 0;
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
  padding: 18px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  flex-shrink: 0;
}

.hud-stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
  gap: 3px;
}

.hud-value {
  font-size: 24px;
  font-weight: 700;
  color: #f1f5f9;
  letter-spacing: -0.02em;
  line-height: 1;
}

.hud-label {
  font-size: 12px;
  color: rgba(148, 163, 184, 0.45);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  font-weight: 500;
}

.hud-divider {
  width: 1px;
  height: 28px;
  background: rgba(255, 255, 255, 0.06);
}

/* ── Health bar ─────────────────────────────────────────── */

.health-bar-wrap {
  padding: 10px 20px 14px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  flex-shrink: 0;
}

.health-bar-labels {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 7px;
}

.health-bar-title {
  font-size: 12px;
  font-weight: 600;
  color: rgba(148, 163, 184, 0.4);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.health-bar-track {
  position: relative;
  height: 7px;
  border-radius: 99px;
  background: rgba(255, 255, 255, 0.05);
  overflow: hidden;
}

.health-bar-bg {
  position: absolute;
  inset: 0;
  border-radius: 99px;
  background: linear-gradient(90deg, #ef4444 0%, #f59e0b 45%, #22c55e 100%);
  clip-path: inset(0 v-bind(barClipRight) 0 0 round 99px);
  transition: clip-path 0.6s cubic-bezier(0.16, 1, 0.3, 1);
  overflow: hidden;
}

.health-bar-bg::after {
  content: "";
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, transparent 20%, rgba(255, 255, 255, 0.22) 50%, transparent 80%);
  background-size: 50% 100%;
  background-repeat: no-repeat;
  animation: bar-shimmer 2s ease-in-out infinite alternate;
}

@keyframes bar-shimmer {
  from { background-position: -50% 0; }
  to   { background-position: 150% 0; }
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

.queue-card:hover {
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

/* ── Message transitions ────────────────────────────────── */

.msg-enter-active {
  transition: opacity 0.3s ease, transform 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.msg-enter-from {
  opacity: 0;
  transform: translateY(8px);
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
    min-height: 50dvh;
  }

  .lobby-content {
    padding: 24px 20px;
  }

  .priority-grid {
    grid-template-columns: 1fr;
  }

  .queue-panel {
    max-height: 60dvh;
  }
}
</style>
