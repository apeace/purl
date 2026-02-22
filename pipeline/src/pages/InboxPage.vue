<template>
  <div class="inbox">

    <!-- Sticky header: tabs + toolbar -->
    <div class="inbox-header">
    <InboxTabs v-model="activeTab" />

    <div class="toolbar" ref="toolbarEl">
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

          <!-- Filter chips -->
          <div class="chip-wrap" ref="priorityWrapEl">
            <button
              class="chip"
              :class="{ 'chip--active': filterPriorities.size }"
              @click="openDropdown = openDropdown === 'priority' ? null : 'priority'"
            >
              <span class="chip-dot" style="background: #f59e0b" />
              <span class="chip-text">Priority</span>
              <span v-if="filterPriorities.size" class="chip-count">{{ filterPriorities.size }}</span>
              <ChevronDown :size="12" class="chip-chevron" />
            </button>
            <Transition name="drop">
              <div v-if="openDropdown === 'priority'" class="dropdown">
                <button
                  v-for="p in priorities"
                  :key="p.value"
                  class="dropdown-item"
                  :class="{ 'dropdown-item--on': filterPriorities.has(p.value) }"
                  @click="toggleSet(filterPriorities, p.value)"
                >
                  <span class="filter-dot" :style="{ background: p.color }" />
                  <span>{{ p.label }}</span>
                  <span v-if="filterPriorities.has(p.value)" class="check-mark">&#10003;</span>
                </button>
              </div>
            </Transition>
          </div>

          <div class="chip-wrap" ref="statusWrapEl">
            <button
              class="chip"
              :class="{ 'chip--active': filterStatuses.size }"
              @click="openDropdown = openDropdown === 'status' ? null : 'status'"
            >
              <span class="chip-dot" style="background: #60a5fa" />
              <span class="chip-text">Status</span>
              <span v-if="filterStatuses.size" class="chip-count">{{ filterStatuses.size }}</span>
              <ChevronDown :size="12" class="chip-chevron" />
            </button>
            <Transition name="drop">
              <div v-if="openDropdown === 'status'" class="dropdown">
                <button
                  v-for="s in statuses"
                  :key="s.value"
                  class="dropdown-item"
                  :class="{ 'dropdown-item--on': filterStatuses.has(s.value) }"
                  @click="toggleSet(filterStatuses, s.value)"
                >
                  <span class="filter-dot" :style="{ background: s.color }" />
                  <span>{{ s.label }}</span>
                  <span v-if="filterStatuses.has(s.value)" class="check-mark">&#10003;</span>
                </button>
              </div>
            </Transition>
          </div>

          <div class="chip-wrap" ref="assigneeWrapEl">
            <button
              class="chip"
              :class="{ 'chip--active': filterAssignees.size }"
              @click="openDropdown = openDropdown === 'assignee' ? null : 'assignee'"
            >
              <span class="chip-dot" style="background: #a855f7" />
              <span class="chip-text">Assignee</span>
              <span v-if="filterAssignees.size" class="chip-count">{{ filterAssignees.size }}</span>
              <ChevronDown :size="12" class="chip-chevron" />
            </button>
            <Transition name="drop">
              <div v-if="openDropdown === 'assignee'" class="dropdown">
                <button
                  v-for="a in uniqueAssignees"
                  :key="a"
                  class="dropdown-item"
                  :class="{ 'dropdown-item--on': filterAssignees.has(a) }"
                  @click="toggleSet(filterAssignees, a)"
                >
                  <span>{{ a }}</span>
                  <span v-if="filterAssignees.has(a)" class="check-mark">&#10003;</span>
                </button>
              </div>
            </Transition>
          </div>

          <!-- Sort -->
          <div class="chip-wrap" ref="sortWrapEl">
            <button
              class="chip chip--sort"
              @click="openDropdown = openDropdown === 'sort' ? null : 'sort'"
            >
              <ArrowUpDown :size="13" />
              <span class="chip-text">{{ sortLabels[sortBy] }}</span>
            </button>
            <Transition name="drop">
              <div v-if="openDropdown === 'sort'" class="dropdown dropdown--right">
                <button
                  v-for="opt in sortOptions"
                  :key="opt.value"
                  class="dropdown-item"
                  :class="{ 'dropdown-item--on': sortBy === opt.value }"
                  @click="sortBy = opt.value; openDropdown = null"
                >
                  <span>{{ opt.label }}</span>
                  <span v-if="sortBy === opt.value" class="check-mark">&#10003;</span>
                </button>
              </div>
            </Transition>
          </div>
        </div>
      </Transition>

      <div class="toolbar-spacer" />
      <span class="pagination-info">1–{{ emails.length }} of {{ emails.length }}</span>
      <button class="toolbar-btn" title="Newer" disabled><ChevronLeft :size="16" /></button>
      <button class="toolbar-btn" title="Older"><ChevronRight :size="16" /></button>
    </div>
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

          <!-- Priority dot -->
          <span
            class="priority-dot"
            :class="{ 'priority-dot--glow': email.priority === 'urgent' }"
            :style="{ background: priorityColors[email.priority] }"
            :title="email.priority"
          />

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

          <!-- Status pill -->
          <span
            class="status-pill"
            :style="{
              background: statusColors[email.status]?.bg,
              color: statusColors[email.status]?.text,
            }"
          >{{ email.status }}</span>

          <!-- Labels -->
          <div class="row-labels">
            <span
              v-for="label in email.labels"
              :key="label"
              class="label"
              :class="`label--${label}`"
            >{{ label }}</span>
          </div>

          <!-- Assignee avatar -->
          <div
            v-if="email.assignee !== 'Unassigned'"
            class="assignee-avatar"
            :style="{ background: email.assigneeColor }"
            :title="email.assignee"
          >{{ email.assignee[0] }}</div>
          <div v-else class="assignee-spacer" />

          <!-- Time -->
          <span class="row-time" :style="{ color: timeColor(email.createdAt) }">
            {{ email.time }}
          </span>
        </div>
      </TransitionGroup>
    </div>

    <TicketModal :ticket-id="selectedTicketId" @close="selectedTicketId = null" />
  </div>
</template>

<script setup>
import { Archive, ArrowUpDown, ChevronDown, ChevronLeft, ChevronRight, MailOpen, RefreshCw, Star, Trash2 } from "lucide-vue-next"
import { computed, onBeforeUnmount, onMounted, reactive, ref } from "vue"
import InboxTabs from "../components/InboxTabs.vue"
import TicketModal from "../components/TicketModal.vue"
import { useTickets } from "../composables/useTickets.js"

const {
  archiveTicket,
  avatarColor,
  CURRENT_USER,
  deleteTicket,
  filterAssignees,
  filterPriorities,
  filterStatuses,
  markRead,
  sortBy,
  sortedTickets,
  toggleStar,
  uniqueAssignees,
} = useTickets()

const activeTab = ref("all")

// ── Filter / sort data ───────────────────────────────────

const priorities = [
  { value: "urgent", label: "Urgent", color: "#ef4444" },
  { value: "high", label: "High", color: "#f97316" },
  { value: "medium", label: "Medium", color: "#f59e0b" },
  { value: "low", label: "Low", color: "#34d399" },
]

const statuses = [
  { value: "new", label: "New", color: "#38bdf8" },
  { value: "open", label: "Open", color: "#60a5fa" },
  { value: "pending", label: "Pending", color: "#a855f7" },
  { value: "escalated", label: "Escalated", color: "#f97316" },
  { value: "solved", label: "Solved", color: "#34d399" },
  { value: "closed", label: "Closed", color: "#94a3b8" },
]

const sortOptions = [
  { value: "time", label: "Last Updated" },
  { value: "priority", label: "Priority" },
  { value: "status", label: "Status" },
  { value: "assignee", label: "Assignee" },
]

const sortLabels = { time: "Last Updated", priority: "Priority", status: "Status", assignee: "Assignee" }

function toggleSet(set, val) {
  if (set.has(val)) set.delete(val)
  else set.add(val)
}

// ── Dropdowns ────────────────────────────────────────────

const openDropdown = ref(null)
const toolbarEl = ref(null)
const priorityWrapEl = ref(null)
const statusWrapEl = ref(null)
const assigneeWrapEl = ref(null)
const sortWrapEl = ref(null)

function onPointerDown(e) {
  if (!openDropdown.value) return
  const wraps = [priorityWrapEl.value, statusWrapEl.value, assigneeWrapEl.value, sortWrapEl.value]
  if (!wraps.some((w) => w?.contains(e.target))) {
    openDropdown.value = null
  }
}

onMounted(() => document.addEventListener("pointerdown", onPointerDown))
onBeforeUnmount(() => document.removeEventListener("pointerdown", onPointerDown))

// ── Row color maps ───────────────────────────────────────

const priorityColors = {
  urgent: "#ef4444",
  high: "#f97316",
  medium: "#f59e0b",
  low: "#34d399",
}

const statusColors = {
  new: { bg: "rgba(56, 189, 248, 0.12)", text: "#38bdf8" },
  open: { bg: "rgba(96, 165, 250, 0.12)", text: "#60a5fa" },
  pending: { bg: "rgba(168, 85, 247, 0.12)", text: "#a855f7" },
  escalated: { bg: "rgba(249, 115, 22, 0.12)", text: "#f97316" },
  solved: { bg: "rgba(52, 211, 153, 0.12)", text: "#34d399" },
  closed: { bg: "rgba(148, 163, 184, 0.12)", text: "#94a3b8" },
}

function timeColor(createdAt) {
  const mins = (Date.now() - new Date(createdAt).getTime()) / 60000
  if (mins < 30) return "#fca5a5"
  if (mins < 120) return "#fcd34d"
  return "rgba(148, 163, 184, 0.4)"
}

// ── Email rows ───────────────────────────────────────────

const emails = computed(() => {
  let list = sortedTickets.value

  switch (activeTab.value) {
    case "mine":
      list = list.filter((t) => t.assignee === CURRENT_USER && t.status !== "closed")
      break
    case "unassigned":
      list = list.filter((t) => t.assignee === "Unassigned" && t.status !== "closed")
      break
    case "all":
      list = list.filter((t) => t.status !== "closed")
      break
    case "starred":
      list = list.filter((t) => t.starred)
      break
  }

  return list.map((t) => ({
    id: t.id,
    read: t.read,
    starred: t.starred,
    sender: { name: t.name, color: t.avatarColor },
    subject: t.subject,
    preview: t.messages[t.messages.length - 1]?.text ?? "",
    labels: t.labels,
    time: t.time,
    priority: t.priority,
    status: t.status,
    assignee: t.assignee,
    assigneeColor: t.assignee !== "Unassigned" ? avatarColor(t.assignee) : null,
    createdAt: t.createdAt,
  }))
})

// ── Selection ────────────────────────────────────────────

const selected = reactive(new Set())

const anySelected = computed(() => selected.size > 0)
const allSelected = computed(() => selected.size === emails.value.length && emails.value.length > 0)
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

// ── Row actions ──────────────────────────────────────────

const selectedTicketId = ref(null)

function openEmail(email) {
  markRead(email.id)
  selectedTicketId.value = email.id
}

function handleToggleStar(email) {
  toggleStar(email.id)
}

// ── Refresh ──────────────────────────────────────────────

const refreshing = ref(false)

function refresh() {
  refreshing.value = true
  setTimeout(() => { refreshing.value = false }, 700)
}
</script>

<style scoped>
/* ── Shell ───────────────────────────────────────────────── */

.inbox {
  margin: -28px;
}

/* ── Sticky header ───────────────────────────────────────── */

.inbox-header {
  position: sticky;
  top: 56px;
  z-index: 10;
  background: #0a0e1a;
}

/* ── Toolbar ─────────────────────────────────────────────── */

.toolbar {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 4px;
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

/* ── Filter chips ────────────────────────────────────────── */

.chip-wrap {
  position: relative;
  flex-shrink: 0;
}

.chip {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 10px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.02);
  color: #94a3b8;
  font-size: 12px;
  font-weight: 500;
  font-family: inherit;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s cubic-bezier(0.16, 1, 0.3, 1);
}

.chip:hover {
  background: rgba(255, 255, 255, 0.05);
  border-color: rgba(255, 255, 255, 0.12);
  color: #e2e8f0;
}

.chip--active {
  background: rgba(99, 102, 241, 0.1);
  border-color: rgba(99, 102, 241, 0.25);
  color: #a5b4fc;
  box-shadow: 0 1px 6px rgba(99, 102, 241, 0.12);
}

.chip--sort {
  color: #64748b;
}

.chip-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}

.chip-count {
  font-size: 10px;
  font-weight: 700;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  border-radius: 8px;
  background: rgba(99, 102, 241, 0.2);
  color: #a5b4fc;
  display: flex;
  align-items: center;
  justify-content: center;
}

.chip-chevron {
  opacity: 0.5;
  flex-shrink: 0;
}

/* ── Dropdowns ───────────────────────────────────────────── */

.dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  z-index: 50;
  min-width: 160px;
  background: #0f172a;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 10px;
  padding: 4px;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.5);
}

.dropdown--right {
  left: auto;
  right: 0;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 7px 10px;
  border: none;
  border-radius: 7px;
  background: transparent;
  color: #94a3b8;
  font-size: 12px;
  font-weight: 500;
  font-family: inherit;
  cursor: pointer;
  text-transform: capitalize;
  transition: background 0.12s;
}

.dropdown-item:hover {
  background: rgba(255, 255, 255, 0.05);
  color: #e2e8f0;
}

.dropdown-item--on {
  color: #c7d2fe;
}

.filter-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.check-mark {
  margin-left: auto;
  color: #6366f1;
  font-size: 12px;
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
  padding: 0 8px 0 4px;
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

/* ── Row: priority dot ───────────────────────────────────── */

.priority-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
  margin-right: 10px;
  transition: box-shadow 0.15s;
}

.priority-dot--glow {
  box-shadow: 0 0 6px rgba(239, 68, 68, 0.5);
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

/* ── Row: status pill ────────────────────────────────────── */

.status-pill {
  font-size: 10px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 10px;
  text-transform: capitalize;
  white-space: nowrap;
  flex-shrink: 0;
  margin-right: 8px;
}

/* ── Row: labels ─────────────────────────────────────────── */

.row-labels {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
  padding-right: 8px;
}

.label {
  font-size: 10px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 10px;
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

/* ── Row: assignee avatar ────────────────────────────────── */

.assignee-avatar {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
  margin-right: 10px;
}

.assignee-spacer {
  width: 22px;
  flex-shrink: 0;
  margin-right: 10px;
}

/* ── Row: time ───────────────────────────────────────────── */

.row-time {
  font-size: 12px;
  white-space: nowrap;
  flex-shrink: 0;
  width: 56px;
  text-align: right;
}

.email-row--unread .row-time {
  font-weight: 600;
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

.drop-enter-active,
.drop-leave-active {
  transition: opacity 0.12s, transform 0.12s cubic-bezier(0.16, 1, 0.3, 1);
}

.drop-enter-from,
.drop-leave-to {
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

/* ── Desktop ─────────────────────────────────────────────── */

@media (min-width: 768px) {
  .inbox-header {
    top: 0;
  }
}

/* ── Mobile ──────────────────────────────────────────────── */

@media (max-width: 767px) {
  .inbox {
    margin: -16px;
  }

  .chip-text {
    display: none;
  }

  .row-sender {
    width: 120px;
  }

  .row-sep,
  .row-preview {
    display: none;
  }

  .row-labels,
  .status-pill,
  .assignee-avatar,
  .assignee-spacer {
    display: none;
  }
}
</style>
