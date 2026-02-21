<template>
  <div class="go-page">
    <div class="go-split">

      <!-- ── Left: Active thread ─────────────────────────── -->
      <div class="thread-panel">
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
            <span class="badge badge--open">Open</span>
            <span class="badge" :class="`badge--${activeThread.priority}`">{{ activeThread.priority }}</span>
          </div>
        </div>

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
      </div>

      <!-- ── Right: HUD + queue ──────────────────────────── -->
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

        <!-- AI suggestion -->
        <div v-if="currentAi" class="ai-card">
          <div class="ai-card-header">
            <div class="ai-badge">
              <Sparkles :size="11" /> AI
            </div>
            <span class="ai-headline">{{ currentAi.headline }}</span>
          </div>
          <p class="ai-body">{{ currentAi.body }}</p>
          <button class="btn btn--ai" @click="followAi">
            {{ currentAi.action }} <ChevronRight :size="14" />
          </button>
        </div>

        <!-- Waiting tickets -->
        <div class="queue-list">
          <div class="queue-section-label">Up next</div>
          <button
            v-for="thread in queue"
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
import { ChevronRight, Clock, Send, Sparkles } from "lucide-vue-next"
import { computed, nextTick, ref, watch } from "vue"

// ── Data ────────────────────────────────────────────────

const threads = ref([
  {
    id: 1,
    name: "Sarah Lin",
    company: "Acme Corp",
    ticketId: "#1842",
    subject: "Setup not working after update",
    priority: "high",
    wait: "2h 15m",
    avatarColor: "#6366f1",
    messages: [
      { id: 1, from: "customer", time: "2h ago", text: "Hi, after the latest update we can't access the admin panel. It just shows a blank screen. We've tried clearing the cache and restarting the server but nothing works. This is blocking our entire team." },
      { id: 2, from: "agent", time: "1h 45m ago", text: "Hi Sarah, thanks for reaching out — sorry to hear you're blocked. Could you share which browser and OS version you're on? And do you see any errors in the browser console?" },
      { id: 3, from: "customer", time: "45m ago", text: "Chrome 121 on Windows 11. The console shows: 'Error: Failed to load config.json — 404'. We have 15 people waiting on this, please help ASAP." },
    ],
  },
  {
    id: 2,
    name: "Mike Torres",
    company: "DataFlow",
    ticketId: "#1839",
    subject: "Can't export CSV report",
    priority: "medium",
    wait: "1h 20m",
    avatarColor: "#ec4899",
    messages: [
      { id: 1, from: "customer", time: "1h 20m ago", text: "The CSV export button on the Reports page doesn't do anything when I click it. No download, no error — just nothing. Started happening this morning." },
    ],
  },
  {
    id: 3,
    name: "Priya Patel",
    company: "Orion Labs",
    ticketId: "#1847",
    subject: "Billing discrepancy on March invoice",
    priority: "medium",
    wait: "45m",
    avatarColor: "#34d399",
    messages: [
      { id: 1, from: "customer", time: "45m ago", text: "Our March invoice shows $2,400 but we're on the $1,800/mo plan. We haven't added any seats or changed our plan. Can you investigate?" },
    ],
  },
  {
    id: 4,
    name: "James Kim",
    company: "NexGen Inc",
    ticketId: "#1851",
    subject: "API rate limit hit unexpectedly",
    priority: "high",
    wait: "30m",
    avatarColor: "#f59e0b",
    messages: [
      { id: 1, from: "customer", time: "30m ago", text: "We're getting 429 errors on the /events endpoint even though our dashboard shows we're only at 40% of our rate limit quota. This started about an hour ago." },
    ],
  },
])

const aiSuggestions = {
  1: {
    headline: "Config 404 — matches known post-update bug",
    body: "Sarah's error is identical to 3 tickets resolved last week after the v2.4 rollout. A missing config.json is caused by the new deploy script skipping static asset copy. Send the one-line fix and mark resolved.",
    action: "Send fix & resolve",
    replyText: "Hi Sarah! This is a known issue with the v2.4 update — the deploy script misses copying config.json. Run this in your project root: `cp node_modules/@pipeline/defaults/config.json public/`. That should fix it immediately. Let me know if you need anything else!",
  },
  2: {
    headline: "CSV export bug — patch ships in 24h",
    body: "This is a confirmed bug in v2.4.1 affecting all accounts on Chrome. Engineering has a fix merging today, deploying tomorrow morning. Recommend acknowledging and setting the expectation.",
    action: "Acknowledge & set ETA",
    replyText: "Hi Mike! This is a confirmed bug in v2.4.1 — our team already has a fix and it's deploying tomorrow morning. I'll follow up as soon as it's live. Sorry for the inconvenience!",
  },
  3: {
    headline: "Billing overage — likely proration edge case",
    body: "Orion Labs upgraded mid-cycle on Mar 3rd. The $600 difference matches a prorated annual add-on. Check their billing history to confirm, then share the breakdown.",
    action: "Pull billing history",
    replyText: "Hi Priya! I looked into this — the $600 difference is a prorated charge for the annual add-on activated on March 3rd. I've attached the itemized breakdown. Let me know if anything looks off and I'm happy to escalate to billing.",
  },
  4: {
    headline: "Rate limit counter bug — known issue",
    body: "The dashboard quota display has a caching lag of ~2h, so real usage can exceed what's shown. Check their actual usage in the admin panel and consider a temporary limit increase.",
    action: "Check usage & offer increase",
    replyText: "Hi James! There's a known 2-hour caching lag in the dashboard quota display, so your real-time usage can exceed what's shown there. I checked your account directly and you've hit 94% of your limit. I've bumped your limit by 20% for the next 48 hours while we sort out a permanent solution.",
  },
}

const DAILY_GOAL = 20

// ── State ────────────────────────────────────────────────

const activeId = ref(1)
const hudOpen = ref(14)
const hudLongestWait = ref("2h 15m")
const hudResolvedToday = ref(8)

// Clips the gradient from the right so the visible color reflects current health level
const barClipRight = computed(() => `${Math.max(0, 100 - (hudResolvedToday.value / DAILY_GOAL) * 100)}%`)
const replyText = ref("")
const messagesEl = ref(null)

const activeThread = computed(() => threads.value.find((t) => t.id === activeId.value))
const queue = computed(() => threads.value.filter((t) => t.id !== activeId.value))
const currentAi = computed(() => aiSuggestions[activeId.value] ?? null)

// ── Actions ──────────────────────────────────────────────

function sendReply() {
  const text = replyText.value.trim()
  if (!text) return
  const msgs = activeThread.value.messages
  msgs.push({ id: msgs.length + 1, from: "agent", time: "just now", text })
  replyText.value = ""
  scrollToBottom()
}

function followAi() {
  const suggestion = currentAi.value
  if (!suggestion) return
  replyText.value = suggestion.replyText
  nextTick(() => sendReply())
}

function resolve() {
  // Remove the resolved ticket and advance to the next one
  const nextThread = queue.value[0]
  threads.value = threads.value.filter((t) => t.id !== activeId.value)
  if (nextThread) activeId.value = nextThread.id
  replyText.value = ""
  hudResolvedToday.value++
  hudOpen.value = Math.max(0, hudOpen.value - 1)
}

function scrollToBottom() {
  nextTick(() => {
    if (messagesEl.value) {
      messagesEl.value.scrollTop = messagesEl.value.scrollHeight
    }
  })
}

// Reset reply draft when switching tickets
watch(activeId, () => {
  replyText.value = ""
  scrollToBottom()
})
</script>

<style scoped>
/* ── Layout ─────────────────────────────────────────────── */

.go-page {
  /* Negative margin to stretch past the page-wrap padding, filling the viewport */
  margin: -28px;
}

.go-split {
  display: flex;
  height: calc(100dvh - 56px); /* full height minus desktop topbar */
}

/* ── Thread panel (left ~70%) ───────────────────────────── */

.thread-panel {
  flex: 7;
  min-width: 0;
  display: flex;
  flex-direction: column;
  border-right: 1px solid rgba(255, 255, 255, 0.05);
}

.thread-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 20px 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  flex-shrink: 0;
}

.thread-customer-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.thread-avatar {
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

.thread-name {
  font-size: 14px;
  font-weight: 600;
  color: #e2e8f0;
}

.thread-company {
  font-weight: 400;
  color: rgba(148, 163, 184, 0.6);
}

.thread-id {
  font-size: 11px;
  color: rgba(148, 163, 184, 0.4);
  margin-top: 1px;
}

.thread-subject {
  font-size: 15px;
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
  font-size: 11px;
  font-weight: 600;
  padding: 3px 8px;
  border-radius: 6px;
  text-transform: capitalize;
}

.badge--open {
  background: rgba(99, 102, 241, 0.15);
  color: #a5b4fc;
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
  gap: 10px;
  align-items: flex-end;
}

.message--agent {
  flex-direction: row-reverse;
}

.msg-avatar {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
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
  border-radius: 12px;
  padding: 10px 14px;
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
  font-size: 12px;
  font-weight: 600;
  color: #94a3b8;
}

.msg-time {
  font-size: 11px;
  color: rgba(148, 163, 184, 0.4);
}

.msg-body {
  font-size: 13px;
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
  border-radius: 10px;
  padding: 10px 14px;
  color: #e2e8f0;
  font-size: 13px;
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
  gap: 6px;
  padding: 7px 14px;
  border-radius: 8px;
  font-size: 13px;
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
  margin-top: 10px;
}

.btn--ai:hover {
  background: rgba(168, 85, 247, 0.2);
  border-color: rgba(168, 85, 247, 0.35);
}

/* ── Queue panel (right ~30%) ───────────────────────────── */

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
  padding: 14px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  flex-shrink: 0;
}

.hud-stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
  gap: 2px;
}

.hud-value {
  font-size: 18px;
  font-weight: 700;
  color: #f1f5f9;
  letter-spacing: -0.02em;
  line-height: 1;
}

.hud-label {
  font-size: 10px;
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
  font-size: 10px;
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
  /* Red = low health (left), green = full health (right) */
  background: linear-gradient(90deg, #ef4444 0%, #f59e0b 45%, #22c55e 100%);
  /* Clip from the right to reveal only the filled portion */
  clip-path: inset(0 v-bind(barClipRight) 0 0 round 99px);
  transition: clip-path 0.6s cubic-bezier(0.16, 1, 0.3, 1);
  overflow: hidden;
}

/* Glint that sweeps back and forth across the filled portion */
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

/* ── AI card ────────────────────────────────────────────── */

.ai-card {
  margin: 14px 16px 0;
  background: rgba(168, 85, 247, 0.06);
  border: 1px solid rgba(168, 85, 247, 0.2);
  border-radius: 12px;
  padding: 14px;
  flex-shrink: 0;
  /* Subtle animated glow */
  box-shadow: 0 0 24px rgba(168, 85, 247, 0.07);
}

.ai-card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.ai-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 10px;
  font-weight: 700;
  color: #c084fc;
  background: rgba(168, 85, 247, 0.15);
  border-radius: 5px;
  padding: 2px 6px;
  letter-spacing: 0.04em;
  flex-shrink: 0;
}

.ai-headline {
  font-size: 13px;
  font-weight: 600;
  color: #e2e8f0;
  line-height: 1.3;
}

.ai-body {
  font-size: 12px;
  color: rgba(148, 163, 184, 0.7);
  line-height: 1.6;
}

/* ── Queue cards ────────────────────────────────────────── */

.queue-list {
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.queue-section-label {
  font-size: 11px;
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
  border-radius: 10px;
  padding: 12px;
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
  width: 26px;
  height: 26px;
  border-radius: 7px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
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
  font-weight: 600;
  color: #e2e8f0;
}

.qcard-company {
  font-weight: 400;
  color: rgba(148, 163, 184, 0.5);
}

.qcard-subject {
  font-size: 12px;
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
  font-size: 11px;
  color: rgba(148, 163, 184, 0.4);
}

.qcard-priority {
  font-size: 10px;
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 4px;
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

  .thread-panel {
    border-right: none;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    /* Give the thread panel a fixed height on mobile so it doesn't push queue off screen */
    height: 65dvh;
  }

  .queue-panel {
    max-height: 60dvh;
  }
}
</style>
