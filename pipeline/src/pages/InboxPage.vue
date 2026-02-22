<template>
  <div class="inbox">

    <!-- Toolbar -->
    <div class="toolbar">
      <label class="cb-wrap" title="Select all">
        <input
          type="checkbox"
          class="cb"
          :checked="allSelected"
          :indeterminate="someSelected"
          @change="toggleAll"
        />
      </label>

      <Transition name="toolbar-swap">
        <div v-if="anySelected" class="toolbar-actions" key="bulk">
          <button class="toolbar-btn" title="Archive" @click="archiveSelected">
            <Archive :size="15" />
          </button>
          <button class="toolbar-btn" title="Delete" @click="deleteSelected">
            <Trash2 :size="15" />
          </button>
          <button class="toolbar-btn" title="Mark as read" @click="markReadSelected">
            <MailOpen :size="15" />
          </button>
          <span class="selection-count">{{ selected.size }} selected</span>
        </div>
        <div v-else class="toolbar-actions" key="default">
          <button class="toolbar-btn" title="Refresh" @click="refresh">
            <RefreshCw :size="15" :class="{ 'spinning': refreshing }" />
          </button>
        </div>
      </Transition>

      <div class="toolbar-spacer" />
      <span class="pagination-info">1–{{ emails.length }} of 247</span>
      <button class="toolbar-btn" title="Newer" disabled><ChevronLeft :size="16" /></button>
      <button class="toolbar-btn" title="Older"><ChevronRight :size="16" /></button>
    </div>

    <!-- Email list -->
    <div class="email-list">
      <TransitionGroup name="row">
        <div
          v-for="email in emails"
          :key="email.id"
          class="email-row"
          :class="{
            'email-row--unread': !email.read,
            'email-row--selected': selected.has(email.id),
          }"
          @click="openEmail(email)"
        >
          <!-- Checkbox + star -->
          <div class="row-left">
            <label class="cb-wrap" @click.stop>
              <input
                type="checkbox"
                class="cb"
                :checked="selected.has(email.id)"
                @change="toggleSelect(email.id)"
              />
            </label>
            <button
              class="star-btn"
              :class="{ 'star-btn--on': email.starred }"
              title="Star"
              @click.stop="handleToggleStar(email)"
            >
              <Star :size="14" />
            </button>
          </div>

          <!-- Sender -->
          <div class="row-sender">
            <div class="sender-avatar" :style="{ background: email.sender.color }">
              {{ email.sender.name[0] }}
            </div>
            <span class="sender-name">{{ email.sender.name }}</span>
          </div>

          <!-- Subject + preview -->
          <div class="row-content">
            <span class="row-subject">{{ email.subject }}</span>
            <span class="row-sep"> — </span>
            <span class="row-preview">{{ email.preview }}</span>
          </div>

          <!-- Labels -->
          <div class="row-labels">
            <span
              v-for="label in email.labels"
              :key="label"
              class="label"
              :class="`label--${label}`"
            >{{ label }}</span>
          </div>

          <!-- Time -->
          <span class="row-time">{{ email.time }}</span>
        </div>
      </TransitionGroup>
    </div>

    <TicketModal :ticket-id="selectedTicketId" @close="selectedTicketId = null" />
  </div>
</template>

<script setup>
import { Archive, ChevronLeft, ChevronRight, MailOpen, RefreshCw, Star, Trash2 } from "lucide-vue-next"
import { computed, reactive, ref } from "vue"
import TicketModal from "../components/TicketModal.vue"
import { useTickets } from "../composables/useTickets.js"

const {
  archiveTicket,
  deleteTicket,
  markRead,
  tickets,
  toggleStar,
} = useTickets()

// Derive inbox rows from shared tickets (exclude closed)
const emails = computed(() =>
  tickets.value
    .filter((t) => t.status !== "closed")
    .map((t) => ({
      id: t.id,
      read: t.read,
      starred: t.starred,
      sender: { name: t.name, color: t.avatarColor },
      subject: t.subject,
      preview: t.messages[t.messages.length - 1]?.text ?? "",
      labels: t.labels,
      time: t.time,
    }))
)

// ── Selection ────────────────────────────────────────────

const selected = reactive(new Set())

const anySelected = computed(() => selected.size > 0)
const allSelected = computed(() => selected.size === emails.value.length)
const someSelected = computed(() => selected.size > 0 && selected.size < emails.value.length)

function toggleSelect(id) {
  if (selected.has(id)) selected.delete(id)
  else selected.add(id)
}

function toggleAll() {
  if (allSelected.value) {
    selected.clear()
  } else {
    emails.value.forEach((e) => selected.add(e.id))
  }
}

// ── Bulk actions ─────────────────────────────────────────

function archiveSelected() {
  selected.forEach((id) => archiveTicket(id))
  selected.clear()
}

function deleteSelected() {
  selected.forEach((id) => deleteTicket(id))
  selected.clear()
}

function markReadSelected() {
  selected.forEach((id) => markRead(id))
  selected.clear()
}

// ── Row actions ───────────────────────────────────────────

const selectedTicketId = ref(null)

function openEmail(email) {
  markRead(email.id)
  selectedTicketId.value = email.id
}

function handleToggleStar(email) {
  toggleStar(email.id)
}

// ── Refresh ───────────────────────────────────────────────

const refreshing = ref(false)

function refresh() {
  refreshing.value = true
  setTimeout(() => { refreshing.value = false }, 700)
}
</script>

<style scoped>
/* ── Shell ───────────────────────────────────────────────── */

.inbox {
  /* Stretch past page-wrap padding to fill the content area edge-to-edge */
  margin: -28px;
}

/* ── Toolbar ─────────────────────────────────────────────── */

.toolbar {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  position: sticky;
  top: 56px; /* sits below the app topbar */
  z-index: 10;
  background: rgba(10, 14, 26, 0.95);
  backdrop-filter: blur(12px);
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 2px;
}

.toolbar-spacer {
  flex: 1;
}

.toolbar-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: #64748b;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.toolbar-btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.06);
  color: #94a3b8;
}

.toolbar-btn:disabled {
  opacity: 0.3;
  cursor: default;
}

.selection-count {
  font-size: 12px;
  color: rgba(148, 163, 184, 0.6);
  padding: 0 8px;
}

.pagination-info {
  font-size: 12px;
  color: rgba(148, 163, 184, 0.4);
  padding: 0 8px;
  white-space: nowrap;
}

/* ── Checkbox ────────────────────────────────────────────── */

.cb-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  cursor: pointer;
}

.cb {
  width: 15px;
  height: 15px;
  accent-color: #6366f1;
  cursor: pointer;
  flex-shrink: 0;
}

/* ── Email list ──────────────────────────────────────────── */

.email-list {
  display: flex;
  flex-direction: column;
}

.email-row {
  display: flex;
  align-items: center;
  gap: 0;
  padding: 0 8px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
  cursor: pointer;
  transition: background 0.12s;
  min-height: 52px;
  position: relative;
}

.email-row:hover {
  background: rgba(255, 255, 255, 0.03);
}

.email-row--selected {
  background: rgba(99, 102, 241, 0.08);
}

.email-row--selected:hover {
  background: rgba(99, 102, 241, 0.12);
}

/* Unread accent line on left edge */
.email-row--unread::before {
  content: "";
  position: absolute;
  left: 0;
  top: 20%;
  height: 60%;
  width: 2px;
  border-radius: 0 2px 2px 0;
  background: #6366f1;
}

/* ── Row: left controls ──────────────────────────────────── */

.row-left {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 10px 0 4px;
  flex-shrink: 0;
}

.star-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  color: rgba(148, 163, 184, 0.25);
  cursor: pointer;
  transition: color 0.15s, transform 0.15s;
  border-radius: 4px;
  flex-shrink: 0;
}

.star-btn:hover {
  color: #f59e0b;
  transform: scale(1.15);
}

.star-btn--on {
  color: #f59e0b;
}

/* ── Row: sender ─────────────────────────────────────────── */

.row-sender {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 160px;
  flex-shrink: 0;
  padding-right: 12px;
}

.sender-avatar {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
}

.sender-name {
  font-size: 13px;
  font-weight: 400;
  color: #94a3b8;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.email-row--unread .sender-name {
  font-weight: 700;
  color: #f1f5f9;
}

/* ── Row: content ────────────────────────────────────────── */

.row-content {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: baseline;
  gap: 0;
  overflow: hidden;
  padding-right: 12px;
}

.row-subject {
  font-size: 13px;
  font-weight: 400;
  color: #94a3b8;
  white-space: nowrap;
  flex-shrink: 0;
}

.email-row--unread .row-subject {
  font-weight: 600;
  color: #e2e8f0;
}

.row-sep {
  font-size: 13px;
  color: rgba(148, 163, 184, 0.3);
  flex-shrink: 0;
}

.row-preview {
  font-size: 13px;
  color: rgba(148, 163, 184, 0.45);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* ── Row: labels ─────────────────────────────────────────── */

.row-labels {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
  padding-right: 12px;
}

.label {
  font-size: 10px;
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 4px;
  text-transform: capitalize;
  white-space: nowrap;
}

.label--urgent {
  background: rgba(239, 68, 68, 0.12);
  color: #fca5a5;
}

.label--bug {
  background: rgba(245, 158, 11, 0.12);
  color: #fcd34d;
}

.label--billing {
  background: rgba(99, 102, 241, 0.12);
  color: #a5b4fc;
}

.label--feature {
  background: rgba(168, 85, 247, 0.12);
  color: #d8b4fe;
}

/* ── Row: time ───────────────────────────────────────────── */

.row-time {
  font-size: 12px;
  color: rgba(148, 163, 184, 0.4);
  white-space: nowrap;
  flex-shrink: 0;
  width: 56px;
  text-align: right;
}

.email-row--unread .row-time {
  font-weight: 600;
  color: rgba(148, 163, 184, 0.7);
}

/* ── Refresh spin ────────────────────────────────────────── */

.spinning {
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* ── Transitions ─────────────────────────────────────────── */

.toolbar-swap-enter-active,
.toolbar-swap-leave-active {
  transition: opacity 0.15s, transform 0.15s;
  position: absolute;
}

.toolbar-swap-enter-from,
.toolbar-swap-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

.row-leave-active {
  transition: opacity 0.2s, transform 0.2s;
}

.row-leave-to {
  opacity: 0;
  transform: translateX(-12px);
}

/* ── Mobile ──────────────────────────────────────────────── */

@media (max-width: 767px) {
  .inbox {
    margin: -16px;
  }

  .toolbar {
    top: 0;
  }

  .row-sender {
    width: 120px;
  }

  /* Hide preview snippet on small screens */
  .row-sep,
  .row-preview {
    display: none;
  }

  /* Hide labels on small screens */
  .row-labels {
    display: none;
  }
}
</style>
