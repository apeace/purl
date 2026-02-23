<template>
  <div class="inbox">

    <template v-if="!selectedTicketId">
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
    </template>

    <!-- Split view (ticket selected) -->
    <div v-else class="inbox-split">
      <div class="workspace">
        <div class="workspace-active">
          <div class="strategy-bar">
            <button class="strategy-header" @click="selectedTicketId = null">
              <Inbox :size="16" style="color: #60a5fa" />
              <span class="strategy-header-label">Inbox</span>
            </button>
            <div class="strategy-nav">
              <span class="strategy-nav-pos">{{ queueIndex + 1 }} / {{ emails.length }}</span>
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
            @click="openEmail(item)"
          >
            <div class="qcard-top">
              <div class="qcard-avatar" :style="{ background: item.sender.color }">
                {{ item.sender.name[0] }}
              </div>
              <div class="qcard-meta">
                <div class="qcard-name">{{ item.sender.name }}
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
import { Archive, ArrowUpDown, ChevronDown, ChevronLeft, ChevronRight, Clock, Inbox, MailOpen, RefreshCw, Star, Trash2 } from "lucide-vue-next"
import { computed, onBeforeUnmount, onMounted, reactive, ref } from "vue"
import InboxTabs from "../components/InboxTabs.vue"
import TicketDetail from "../components/TicketDetail.vue"
import { useTickets } from "../composables/useTickets"
import { PRIORITY_COLORS, PRIORITY_LIST, STATUS_LIST, STATUS_PILL } from "../utils/colors"

const {
  archiveTicket,
  avatarColor,
  CURRENT_USER,
  deleteTicket,
  filterAssignees,
  filterPriorities,
  filterStatuses,
  markRead,
  resolveTicket,
  sortBy,
  sortedTickets,
  toggleStar,
  uniqueAssignees,
} = useTickets()

const activeTab = ref("all")

// ── Filter / sort data ───────────────────────────────────

const priorities = PRIORITY_LIST
const statuses = STATUS_LIST

const sortOptions = [
  { value: "time", label: "Last Updated" },
  { value: "priority", label: "Priority" },
  { value: "status", label: "Status" },
  { value: "assignee", label: "Assignee" },
]

const sortLabels: Record<string, string> = { time: "Last Updated", priority: "Priority", status: "Status", assignee: "Assignee" }

function toggleSet(set: Set<string>, val: string) {
  if (set.has(val)) set.delete(val)
  else set.add(val)
}

// ── Dropdowns ────────────────────────────────────────────

const openDropdown = ref<string | null>(null)
const toolbarEl = ref<HTMLElement | null>(null)
const priorityWrapEl = ref<HTMLElement | null>(null)
const statusWrapEl = ref<HTMLElement | null>(null)
const assigneeWrapEl = ref<HTMLElement | null>(null)
const sortWrapEl = ref<HTMLElement | null>(null)

function onPointerDown(e: PointerEvent) {
  if (!openDropdown.value) return
  const wraps = [priorityWrapEl.value, statusWrapEl.value, assigneeWrapEl.value, sortWrapEl.value]
  if (!wraps.some((w) => w?.contains(e.target as Node))) {
    openDropdown.value = null
  }
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === "Escape") {
    if (selectedTicketId.value) {
      selectedTicketId.value = null
    } else if (openDropdown.value) {
      openDropdown.value = null
    }
  }
}

onMounted(() => {
  document.addEventListener("pointerdown", onPointerDown)
  document.addEventListener("keydown", onKeydown)
})
onBeforeUnmount(() => {
  document.removeEventListener("pointerdown", onPointerDown)
  document.removeEventListener("keydown", onKeydown)
})

const priorityColors = PRIORITY_COLORS
const statusColors = STATUS_PILL

function timeColor(createdAt: string) {
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
    assigneeColor: t.assignee !== "Unassigned" ? avatarColor(t.assignee) : undefined,
    company: t.company,
    createdAt: t.createdAt,
    wait: t.wait,
  }))
})

// ── Selection ────────────────────────────────────────────

const selected = reactive(new Set<string>())

const anySelected = computed(() => selected.size > 0)
const allSelected = computed(() => selected.size === emails.value.length && emails.value.length > 0)
const someSelected = computed(() => selected.size > 0 && selected.size < emails.value.length)

function toggleSelect(id: string) {
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

const selectedTicketId = ref<string | null>(null)

function openEmail(email: { id: string }) {
  markRead(email.id)
  selectedTicketId.value = email.id
}

function handleToggleStar(email: { id: string }) {
  toggleStar(email.id)
}

// ── Queue navigation ─────────────────────────────────

const queueIndex = computed(() => emails.value.findIndex((e) => e.id === selectedTicketId.value))
const canGoPrev = computed(() => queueIndex.value > 0)
const canGoNext = computed(() => queueIndex.value < emails.value.length - 1)
const displayQueue = computed(() => emails.value.filter((e) => e.id !== selectedTicketId.value))

function goPrev() {
  if (canGoPrev.value) {
    const prev = emails.value[queueIndex.value - 1]
    markRead(prev.id)
    selectedTicketId.value = prev.id
  }
}

function goNext() {
  if (canGoNext.value) {
    const next = emails.value[queueIndex.value + 1]
    markRead(next.id)
    selectedTicketId.value = next.id
  }
}

function handleResolve() {
  const next = displayQueue.value[0] ?? null
  resolveTicket(selectedTicketId.value!)
  selectedTicketId.value = next ? next.id : null
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

.toolbar-btn:hover:not(:disabled),
.toolbar-btn:active:not(:disabled) {
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

.email-row:hover,
.email-row:active {
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
  width: 36px;
  height: 36px;
  border: none;
  background: transparent;
  color: rgba(148, 163, 184, 0.25);
  cursor: pointer;
  transition: color 0.15s, transform 0.15s;
  border-radius: 4px;
  flex-shrink: 0;
  margin: -6px;
}

.star-btn:hover,
.star-btn:active {
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

/* ── Split view ─────────────────────────────────────────── */

.inbox-split {
  display: flex;
  height: 100dvh;
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

/* ── Intermediate screens ────────────────────────────────── */

@media (min-width: 768px) and (max-width: 1099px) {
  .row-sender {
    width: 130px;
  }

  .row-labels {
    display: none;
  }

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

  .inbox-split {
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
