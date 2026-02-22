<template>
  <div class="ticket-detail">
    <!-- AI assist banner -->
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
          <div class="thread-avatar" :style="{ background: ticket.avatarColor }">
            {{ ticket.name[0] }}
          </div>
          <div>
            <div class="thread-name">{{ ticket.name }}
              <span class="thread-company">· {{ ticket.company }}</span>
            </div>
            <div class="thread-id">{{ ticket.ticketId }}</div>
          </div>
        </div>
        <div class="thread-subject">{{ ticket.subject }}</div>
      </div>
      <div class="thread-badges">
        <span class="badge" :class="`badge--${ticket.status}`">{{ ticket.status }}</span>
        <span class="badge" :class="`badge--${ticket.priority}`">{{ ticket.priority }}</span>
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

    <!-- Tab: Communications -->
    <template v-if="activeTab === 'comms'">
      <div ref="messagesEl" class="thread-messages">
        <TransitionGroup name="msg">
          <div
            v-for="msg in ticket.messages"
            :key="msg.id"
            class="message"
            :class="msg.from === 'agent' ? 'message--agent' : 'message--customer'"
          >
            <div v-if="msg.from === 'customer'" class="msg-avatar" :style="{ background: ticket.avatarColor }">
              {{ ticket.name[0] }}
            </div>
            <div class="msg-bubble">
              <div class="msg-header">
                <span class="msg-sender">{{ msg.from === 'agent' ? 'You' : ticket.name }}</span>
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
          <button class="btn btn--ghost" @click="handleResolve">Resolve</button>
          <button class="btn btn--primary" :disabled="!replyText.trim()" @click="sendReply">
            <Send :size="14" /> Send Reply
          </button>
        </div>
      </div>
    </template>

    <!-- Tab: Contact Info -->
    <div v-else-if="activeTab === 'contact'" class="tab-panel">
      <div class="contact-card">
        <div class="contact-avatar" :style="{ background: ticket.avatarColor }">
          {{ ticket.name[0] }}
        </div>
        <div class="contact-name">{{ ticket.name }}</div>
        <div class="contact-company">{{ ticket.company }}</div>
      </div>

      <div class="detail-grid">
        <div class="detail-row">
          <Mail :size="14" class="detail-icon" />
          <div class="detail-content">
            <div class="detail-label">Email</div>
            <div class="detail-value">{{ ticket.email }}</div>
          </div>
        </div>
        <div class="detail-row">
          <Phone :size="14" class="detail-icon" />
          <div class="detail-content">
            <div class="detail-label">Phone</div>
            <div class="detail-value">{{ ticket.phone }}</div>
          </div>
        </div>
        <div class="detail-row">
          <Zap :size="14" class="detail-icon" />
          <div class="detail-content">
            <div class="detail-label">Subscription</div>
            <div class="detail-value">
              {{ ticket.subscription.plan }}
              <span class="sub-status" :class="`sub-status--${ticket.subscription.status}`">{{ ticket.subscription.status }}</span>
            </div>
            <div class="detail-sub">{{ ticket.subscription.id }}</div>
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
          :value="ticket.notes"
          @input="updateNotes(ticketId, $event.target.value)"
        />
      </div>
    </div>

    <!-- Tab: Ticket History -->
    <div v-else-if="activeTab === 'history'" class="tab-panel">
      <div class="panel-title">Activity on {{ ticket.ticketId }}</div>
      <div class="timeline">
        <div
          v-for="(entry, i) in ticket.ticketHistory"
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

    <!-- Tab: Subscriber History -->
    <div v-else-if="activeTab === 'subscriber'" class="tab-panel">
      <div class="panel-title">{{ ticket.name }}'s ticket history</div>
      <div class="panel-subtitle">Customer since {{ ticket.subscription.id }}</div>

      <!-- Current ticket -->
      <div class="history-card history-card--current">
        <div class="history-card-top">
          <span class="history-tid">{{ ticket.ticketId }}</span>
          <span class="history-status" :class="`history-status--${ticket.status}`">{{ ticket.status }}</span>
        </div>
        <div class="history-subject">{{ ticket.subject }}</div>
        <div class="history-date">Current</div>
      </div>

      <!-- Past tickets -->
      <div
        v-for="(past, i) in ticket.subscriberHistory"
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

      <div v-if="!ticket.subscriberHistory.length" class="panel-empty">
        No previous tickets on record.
      </div>
    </div>

    <!-- Tab: Settings -->
    <div v-else-if="activeTab === 'settings'" class="tab-panel">
      <!-- Tags -->
      <div class="settings-section">
        <div class="settings-label">Tags</div>
        <div class="tags-wrap">
          <span
            v-for="tag in ticket.tags"
            :key="tag"
            class="tag"
          >
            {{ tag }}
            <button class="tag-remove" @click="removeTag(ticketId, tag)">&times;</button>
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
            :class="{ 'option-btn--active': ticket.temperature === opt, [`option-btn--${opt}`]: true }"
            @click="setTemperature(ticketId, opt)"
          >{{ opt }}</button>
        </div>
      </div>

      <!-- Assignee -->
      <div class="settings-section">
        <div class="settings-label">Assignee</div>
        <select
          class="settings-select"
          :value="ticket.assignee"
          @change="setAssignee(ticketId, $event.target.value)"
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
            :class="{ 'option-btn--active': ticket.status === s }"
            @click="setStatus(ticketId, s)"
          >{{ s }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ChevronRight, Cog, History, Mail, MessageSquare, Phone, Send, Sparkles, User, Users, Zap } from "lucide-vue-next"
import { computed, nextTick, ref, watch } from "vue"
import { useTickets } from "../composables/useTickets.js"

const props = defineProps({
  ticketId: { type: Number, required: true },
})

const emit = defineEmits(["resolve"])

const {
  addTag,
  aiSuggestions,
  removeTag,
  resolveTicket,
  sendReply: sharedSendReply,
  setAssignee,
  setStatus,
  setTemperature,
  tickets,
  updateNotes,
} = useTickets()

const activeTab = ref("comms")
const replyText = ref("")
const newTag = ref("")
const messagesEl = ref(null)

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

const ticket = computed(() => tickets.value.find((t) => t.id === props.ticketId))
const currentAi = computed(() => aiSuggestions[props.ticketId] ?? null)

function scrollToBottom() {
  nextTick(() => {
    if (messagesEl.value) {
      messagesEl.value.scrollTop = messagesEl.value.scrollHeight
    }
  })
}

function sendReply() {
  const text = replyText.value.trim()
  if (!text || !props.ticketId) return
  sharedSendReply(props.ticketId, text)
  replyText.value = ""
  scrollToBottom()
}

function followAi() {
  const suggestion = currentAi.value
  if (!suggestion || !props.ticketId) return
  replyText.value = suggestion.replyText
  nextTick(() => sendReply())
}

function handleResolve() {
  resolveTicket(props.ticketId)
  emit("resolve")
}

function handleAddTag() {
  const tag = newTag.value.trim().toLowerCase()
  if (!tag || !props.ticketId) return
  addTag(props.ticketId, tag)
  newTag.value = ""
}

watch(() => props.ticketId, () => {
  replyText.value = ""
  activeTab.value = "comms"
  scrollToBottom()
})
</script>

<style scoped>
/* ── AI assist banner ──────────────────────────────────── */

.ticket-detail {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
}

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
  padding-right: 36px;
}

.badge {
  font-size: 13px;
  font-weight: 600;
  padding: 4px 10px;
  border-radius: 7px;
  text-transform: capitalize;
}

.badge--new { background: rgba(56, 189, 248, 0.15); color: #7dd3fc; }
.badge--open { background: rgba(99, 102, 241, 0.15); color: #a5b4fc; }
.badge--pending { background: rgba(168, 85, 247, 0.15); color: #d8b4fe; }
.badge--escalated { background: rgba(249, 115, 22, 0.15); color: #fdba74; }
.badge--solved { background: rgba(52, 211, 153, 0.15); color: #6ee7b7; }
.badge--closed { background: rgba(148, 163, 184, 0.15); color: #94a3b8; }
.badge--high { background: rgba(239, 68, 68, 0.12); color: #fca5a5; }
.badge--medium { background: rgba(245, 158, 11, 0.12); color: #fcd34d; }
.badge--low { background: rgba(52, 211, 153, 0.12); color: #6ee7b7; }

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

.msg-channel--email { background: rgba(99, 102, 241, 0.1); color: #a5b4fc; }
.msg-channel--sms { background: rgba(52, 211, 153, 0.1); color: #6ee7b7; }
.msg-channel--phone { background: rgba(245, 158, 11, 0.1); color: #fcd34d; }

/* ── Tab panels (shared) ───────────────────────────────── */

.tab-panel {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  animation: content-up 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes content-up {
  from { opacity: 0; transform: translateY(12px); }
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

.sub-status--active { background: rgba(52, 211, 153, 0.1); color: #6ee7b7; }
.sub-status--trial { background: rgba(245, 158, 11, 0.1); color: #fcd34d; }
.sub-status--churned { background: rgba(239, 68, 68, 0.1); color: #fca5a5; }

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

.notes-input::placeholder { color: rgba(148, 163, 184, 0.3); }
.notes-input:focus { border-color: rgba(99, 102, 241, 0.4); }

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

.history-card:hover { background: rgba(255, 255, 255, 0.04); }

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
.history-status--new { background: rgba(99, 102, 241, 0.12); color: #a5b4fc; }
.history-status--pending { background: rgba(168, 85, 247, 0.12); color: #d8b4fe; }
.history-status--escalated { background: rgba(249, 115, 22, 0.12); color: #fdba74; }
.history-status--solved { background: rgba(52, 211, 153, 0.12); color: #6ee7b7; }
.history-status--closed { background: rgba(148, 163, 184, 0.12); color: #94a3b8; }

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

.settings-section { margin-bottom: 24px; }

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

.tag-add { display: inline-flex; }

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

.tag-input::placeholder { color: rgba(148, 163, 184, 0.3); }

.tag-input:focus {
  border-color: rgba(99, 102, 241, 0.4);
  border-style: solid;
}

.option-row {
  display: flex;
  gap: 6px;
}

.option-row--wrap { flex-wrap: wrap; }

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

.settings-select:focus { border-color: rgba(99, 102, 241, 0.4); }
.settings-select option { background: #0f172a; color: #e2e8f0; }

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

.message--agent { flex-direction: row-reverse; }

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

.compose-input::placeholder { color: rgba(148, 163, 184, 0.3); }
.compose-input:focus { border-color: rgba(99, 102, 241, 0.4); }

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

/* ── Message transitions ────────────────────────────────── */

.msg-enter-active {
  transition: opacity 0.3s ease, transform 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.msg-enter-from {
  opacity: 0;
  transform: translateY(8px);
}
</style>
