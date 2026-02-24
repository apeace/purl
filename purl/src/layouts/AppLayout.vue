<template>
  <div class="app-shell">
    <!-- Command palette -->
    <Transition name="cmd-fade">
      <div v-if="cmdOpen" class="cmd-backdrop" @click="cmdOpen = false">
        <div class="cmd-panel" @click.stop>
          <div class="cmd-search-row">
            <Search :size="16" class="cmd-search-icon" />
            <input ref="cmdInput" class="cmd-input" placeholder="Search everything..." />
            <kbd class="cmd-kbd">esc</kbd>
          </div>
          <div class="cmd-list">
            <button
              v-for="item in cmdItems"
              :key="item"
              class="cmd-item"
              @click="cmdOpen = false"
            >{{ item }}</button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Mobile overlay -->
    <Transition name="overlay">
      <div v-if="open" class="sidebar-overlay" @click="open = false" />
    </Transition>

    <!-- Sidebar -->
    <aside class="sidebar" :class="{ 'sidebar--open': open, 'sidebar--collapsed': collapsed }">
      <div class="glow-orb" />

      <div class="sidebar-header">
        <div class="logo">
          <div class="logo-badge">P</div>
          <span class="logo-text">Purl</span>
        </div>
        <button class="icon-btn sidebar-close" aria-label="Close navigation" @click="open = false">
          <X :size="18" />
        </button>
        <button
          class="icon-btn collapse-toggle"
          :aria-label="collapsed ? 'Expand sidebar' : 'Collapse sidebar'"
          @click="collapsed = !collapsed"
        >
          <ChevronRight :size="14" />
        </button>
      </div>

      <div class="sidebar-search-wrap">
        <button class="sidebar-search-btn" @click="openCmd">
          <Search :size="16" />
          <span>Search…</span>
          <kbd>⌘K</kbd>
        </button>
      </div>

      <nav class="sidebar-nav">
        <RouterLink
          v-for="item in navBefore"
          :key="item.path"
          :to="item.path"
          class="nav-item"
          active-class="nav-item--active"
          @click="onNavClick(item.path)"
        >
          <component :is="item.icon" :size="18" class="nav-icon" />
          <span class="nav-label">{{ item.label }}</span>
        </RouterLink>

        <!-- Kanban expandable group -->
        <div class="kanban-group">
          <RouterLink
            to="/kanban"
            class="nav-item kanban-nav-item"
            :class="{ 'nav-item--active': isKanbanRoute }"
            @click="onNavClick('/kanban')"
          >
            <Workflow :size="18" class="nav-icon" />
            <span class="nav-label">Kanban</span>
            <span
              role="button"
              class="kanban-add-btn"
              @click.prevent.stop="showCreateModal = true"
            >
              <Plus :size="14" />
            </span>
          </RouterLink>
          <div v-if="isKanbanRoute && !collapsed" class="kanban-subnav">
            <RouterLink
              v-for="board in boards"
              :key="board.id"
              :to="board.isDefault ? '/kanban' : `/kanban/${board.id}`"
              class="subnav-item"
              :class="{ 'subnav-item--active': board.isDefault ? route.path === '/kanban' : route.params.boardId === board.id }"
              @click="onNavClick(board.isDefault ? '/kanban' : `/kanban/${board.id}`)"
              @contextmenu.prevent="!board.isDefault && openContextMenu($event, board.id)"
              @dragover.prevent="onBoardDragOver($event, board.id)"
              @dragleave="onBoardDragLeave(board.id)"
              @drop="onBoardDrop($event, board.id)"
            >
              <template v-if="renamingBoardId === board.id">
                <input
                  ref="renameInputRef"
                  v-model="renameValue"
                  class="subnav-rename-input"
                  @blur="commitRename"
                  @keydown.enter="commitRename"
                  @keydown.escape="cancelRename"
                  @click.prevent.stop
                />
              </template>
              <template v-else>
                <span class="subnav-dot" :style="{ background: board.stages[0]?.color ?? '#94a3b8' }" />
                <span class="subnav-label" :class="{ 'subnav-label--dragover': dragOverBoardId === board.id }">{{ board.name }}</span>
              </template>
            </RouterLink>
          </div>
        </div>

        <RouterLink
          v-for="item in navAfter"
          :key="item.path"
          :to="item.path"
          class="nav-item"
          active-class="nav-item--active"
          @click="onNavClick(item.path)"
        >
          <component :is="item.icon" :size="18" class="nav-icon" />
          <span class="nav-label">{{ item.label }}</span>
        </RouterLink>
      </nav>

      <div class="sidebar-footer">
        <RouterLink
          to="/settings"
          class="nav-item"
          active-class="nav-item--active"
          @click="onNavClick('/settings')"
        >
          <Settings :size="18" class="nav-icon" />
          <span class="nav-label">Settings</span>
        </RouterLink>
        <div class="user-card">
          <div class="user-avatar">A</div>
          <div class="user-info">
            <div class="user-name">Alex Chen</div>
            <div class="user-email">alex@purl.io</div>
          </div>
        </div>
      </div>
    </aside>

    <!-- Mobile header -->
    <header class="mobile-header">
      <div class="mobile-header-left">
        <button class="icon-btn" aria-label="Open navigation" @click="open = true">
          <Menu :size="20" />
        </button>
        <div class="logo">
          <div class="logo-badge logo-badge--sm">P</div>
          <span class="logo-text">Purl</span>
        </div>
      </div>
      <div class="mobile-header-right">
        <button class="icon-btn" @click="openCmd"><Search :size="16" /></button>
        <div class="user-avatar user-avatar--sm">A</div>
      </div>
    </header>

    <!-- Main content -->
    <main class="main-content" :class="{ 'main-content--collapsed': collapsed }">
      <!-- Desktop topbar -->
      <div class="topbar">
        <div class="topbar-crumb">
          <span class="topbar-page">{{ pageLabel }}</span>
          <ChevronRight :size="14" class="topbar-chevron" />
          <span class="topbar-section">Overview</span>
        </div>
        <div class="topbar-right">
          <button class="topbar-search-btn" @click="openCmd">
            <Search :size="16" />
            <span>Search…</span>
            <kbd>⌘K</kbd>
          </button>
          <div class="user-avatar user-avatar--sm">A</div>
        </div>
      </div>

      <div class="page-wrap">
        <RouterView :key="navResetKey" />
      </div>
    </main>
    <CreateBoardModal
      :visible="showCreateModal"
      @close="showCreateModal = false"
      @created="onBoardCreated"
    />

    <StagePickerModal
      :visible="showStagePicker"
      :board-name="stagePickerBoard?.name ?? ''"
      :stages="stagePickerBoard?.stages ?? []"
      @close="closeStagePicker"
      @pick="onStagePicked"
    />

    <BoardContextMenu
      :visible="ctxMenuVisible"
      :x="ctxMenuX"
      :y="ctxMenuY"
      @close="ctxMenuVisible = false"
      @rename="startRename"
      @delete="handleDeleteBoard"
    />
  </div>
</template>

<script setup lang="ts">
import { BarChart3, ChevronRight, Inbox, LayoutDashboard, Menu, Plus, Search, Settings, Workflow, X, Zap } from "lucide-vue-next"
import { storeToRefs } from "pinia"
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue"
import { useRoute, useRouter } from "vue-router"
import BoardContextMenu from "../components/BoardContextMenu.vue"
import CreateBoardModal from "../components/CreateBoardModal.vue"
import StagePickerModal from "../components/StagePickerModal.vue"
import { useKanbanStore } from "../stores/useKanbanStore"
import type { BoardStage } from "../stores/useKanbanStore"

const open = ref(false)
const collapsed = ref(false)
const cmdOpen = ref(false)
const cmdInput = ref<HTMLInputElement | null>(null)

const kanbanStore = useKanbanStore()
const { boards } = storeToRefs(kanbanStore)
const { addCardToBoard, deleteBoard, renameBoard } = kanbanStore

const navBefore = [
  { path: "/go", label: "Go", icon: Zap },
]

const navAfter = [
  { path: "/inbox", label: "Inbox", icon: Inbox },
  { path: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { path: "/reporting", label: "Reporting", icon: BarChart3 },
]

const cmdItems = [
  "Dashboard Overview",
  "Kanban — Active Deals",
  "Inbox — Unread Messages",
  "Revenue Report Q4",
]

const route = useRoute()
const router = useRouter()

// ── Kanban sidebar state ────────────────────────────────
const showCreateModal = ref(false)
const isKanbanRoute = computed(() => route.path.startsWith("/kanban"))

// Board creation
function onBoardCreated(boardId: string) {
  showCreateModal.value = false
  router.push(`/kanban/${boardId}`)
}

// Context menu
const ctxMenuVisible = ref(false)
const ctxMenuX = ref(0)
const ctxMenuY = ref(0)
const ctxBoardId = ref<string | null>(null)

function openContextMenu(e: MouseEvent, boardId: string) {
  ctxBoardId.value = boardId
  ctxMenuX.value = e.clientX
  ctxMenuY.value = e.clientY
  ctxMenuVisible.value = true
}

function handleDeleteBoard() {
  ctxMenuVisible.value = false
  if (!ctxBoardId.value) return
  const wasViewing = route.params.boardId === ctxBoardId.value
  deleteBoard(ctxBoardId.value)
  if (wasViewing) router.push("/kanban")
}

// Inline rename
const renamingBoardId = ref<string | null>(null)
const renameValue = ref("")
const renameInputRef = ref<HTMLInputElement[] | null>(null)

function startRename() {
  ctxMenuVisible.value = false
  if (!ctxBoardId.value) return
  const board = boards.value.find((b) => b.id === ctxBoardId.value)
  if (!board) return
  renamingBoardId.value = ctxBoardId.value
  renameValue.value = board.name
  nextTick(() => {
    const inputs = renameInputRef.value
    if (inputs && inputs.length) {
      inputs[0].focus()
      inputs[0].select()
    }
  })
}

function commitRename() {
  if (renamingBoardId.value && renameValue.value.trim()) {
    renameBoard(renamingBoardId.value, renameValue.value.trim())
  }
  renamingBoardId.value = null
}

function cancelRename() {
  renamingBoardId.value = null
}

// Drag-to-board (stage picker)
const showStagePicker = ref(false)
const stagePickerBoard = ref<{ name: string; stages: BoardStage[] } | null>(null)
const stagePickerBoardId = ref<string | null>(null)
const stagePickerTicketId = ref<string | null>(null)
const dragOverBoardId = ref<string | null>(null)

function onBoardDragOver(_e: DragEvent, boardId: string) {
  dragOverBoardId.value = boardId
}

function onBoardDragLeave(boardId: string) {
  if (dragOverBoardId.value === boardId) dragOverBoardId.value = null
}

function onBoardDrop(e: DragEvent, boardId: string) {
  dragOverBoardId.value = null
  const ticketId = e.dataTransfer?.getData("text/plain")
  if (!ticketId) return
  const board = boards.value.find((b) => b.id === boardId)
  if (!board) return
  stagePickerBoardId.value = boardId
  stagePickerTicketId.value = ticketId
  stagePickerBoard.value = { name: board.name, stages: board.stages }
  showStagePicker.value = true
}

function onStagePicked(stageId: string) {
  if (stagePickerBoardId.value && stagePickerTicketId.value) {
    addCardToBoard(stagePickerBoardId.value, stagePickerTicketId.value, stageId)
  }
  closeStagePicker()
}

function closeStagePicker() {
  showStagePicker.value = false
  stagePickerBoard.value = null
  stagePickerBoardId.value = null
  stagePickerTicketId.value = null
}

const navResetKey = ref(0)

function onNavClick(path: string) {
  open.value = false
  if (route.path === path) {
    navResetKey.value++
  }
}

const pageLabel = computed(() => {
  const seg = route.path.split("/").filter(Boolean)[0] || "dashboard"
  return seg.charAt(0).toUpperCase() + seg.slice(1)
})

function openCmd() {
  cmdOpen.value = true
}

watch(cmdOpen, (val) => {
  if (val) nextTick(() => cmdInput.value?.focus())
})

watch(open, (val) => {
  document.body.style.overflow = val ? "hidden" : ""
})

function onKeyDown(e: KeyboardEvent) {
  if (e.key === "Escape") {
    open.value = false
    cmdOpen.value = false
  }
  if ((e.metaKey || e.ctrlKey) && e.key === "k") {
    e.preventDefault()
    cmdOpen.value = !cmdOpen.value
  }
}

onMounted(() => document.addEventListener("keydown", onKeyDown))
onBeforeUnmount(() => document.removeEventListener("keydown", onKeyDown))
</script>

<style scoped>
/* ---- Shell ---- */

.app-shell {
  display: flex;
  min-height: 100dvh;
}

/* ---- Command palette ---- */

.cmd-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(8px);
  z-index: 100;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 120px;
}

.cmd-panel {
  width: 100%;
  max-width: 520px;
  margin: 0 16px;
  background: rgba(15, 23, 42, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 25px 60px rgba(0, 0, 0, 0.5);
  animation: slideUp 0.2s ease-out;
}

.cmd-search-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.cmd-search-icon {
  color: rgba(148, 163, 184, 0.5);
  flex-shrink: 0;
}

.cmd-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  color: #e2e8f0;
  font-size: 15px;
  font-family: inherit;
}

.cmd-input::placeholder {
  color: rgba(148, 163, 184, 0.5);
}

.cmd-kbd {
  font-size: 11px;
  color: rgba(148, 163, 184, 0.4);
  background: rgba(255, 255, 255, 0.05);
  border-radius: 5px;
  padding: 2px 6px;
}

.cmd-list {
  padding: 8px;
}

.cmd-item {
  display: block;
  width: 100%;
  padding: 10px 12px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: rgba(148, 163, 184, 0.8);
  font-size: 13px;
  font-family: inherit;
  text-align: left;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.cmd-item:hover {
  background: rgba(255, 255, 255, 0.05);
  color: #e2e8f0;
}

/* ---- Mobile overlay ---- */

.sidebar-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(6px);
  z-index: 39;
}

/* ---- Sidebar ---- */

.sidebar {
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  width: 256px;
  background: linear-gradient(180deg, rgba(15, 23, 42, 0.98) 0%, rgba(10, 14, 26, 0.99) 100%);
  border-right: 1px solid rgba(255, 255, 255, 0.05);
  display: flex;
  flex-direction: column;
  z-index: 40;
  overflow: hidden;
  transform: translateX(-100%);
  transition: transform 0.3s cubic-bezier(0.16, 1, 0.3, 1), width 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.sidebar--open {
  transform: translateX(0);
}

.glow-orb {
  position: absolute;
  top: -60px;
  left: -60px;
  width: 200px;
  height: 200px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(99, 102, 241, 0.25) 0%, rgba(168, 85, 247, 0.12) 40%, transparent 70%);
  filter: blur(40px);
  pointer-events: none;
  animation: orbFloat 8s ease-in-out infinite;
}

/* ---- Sidebar header ---- */

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
  padding: 0 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
  flex-shrink: 0;
  position: relative;
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
}

.logo-badge {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  background: linear-gradient(135deg, #6366f1 0%, #a855f7 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 800;
  color: #fff;
  box-shadow: 0 0 20px rgba(99, 102, 241, 0.3);
  flex-shrink: 0;
}

.logo-badge--sm {
  width: 24px;
  height: 24px;
  border-radius: 7px;
  font-size: 11px;
  box-shadow: 0 0 16px rgba(99, 102, 241, 0.25);
}

.logo-text {
  font-size: 16px;
  font-weight: 700;
  color: #f1f5f9;
  letter-spacing: -0.02em;
}

.icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.04);
  color: #64748b;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
  flex-shrink: 0;
}

.icon-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #94a3b8;
}

/* Mobile: show close, hide collapse toggle */
.sidebar-close {
  display: flex;
}

.collapse-toggle {
  display: none;
}

/* ---- Search shortcut ---- */

.sidebar-search-wrap {
  padding: 12px 12px 4px;
}

.sidebar-search-btn {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(255, 255, 255, 0.02);
  border-radius: 10px;
  cursor: pointer;
  color: rgba(148, 163, 184, 0.5);
  font-size: 13px;
  font-family: inherit;
  transition: border-color 0.15s, background 0.15s;
}

.sidebar-search-btn span {
  flex: 1;
  text-align: left;
}

.sidebar-search-btn kbd {
  font-size: 10px;
  color: rgba(148, 163, 184, 0.3);
  background: rgba(255, 255, 255, 0.04);
  border-radius: 4px;
  padding: 1px 5px;
}

.sidebar-search-btn:hover {
  border-color: rgba(255, 255, 255, 0.12);
  background: rgba(255, 255, 255, 0.04);
}

/* ---- Nav ---- */

.sidebar-nav {
  flex: 1;
  padding: 12px;
  overflow-y: auto;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 9px 12px;
  border: none;
  border-radius: 10px;
  cursor: pointer;
  background: transparent;
  color: #64748b;
  font-size: 13px;
  font-weight: 500;
  font-family: inherit;
  transition: background 0.15s, color 0.15s;
  text-decoration: none;
  position: relative;
  margin-bottom: 2px;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.04);
  color: #e2e8f0;
}

.nav-item--active {
  background: rgba(99, 102, 241, 0.12);
  color: #c7d2fe;
}

.nav-item--active .nav-icon {
  color: #818cf8;
}

/* Left-edge active indicator */
.nav-item--active::before {
  content: "";
  position: absolute;
  left: -12px;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 60%;
  background: linear-gradient(180deg, #6366f1, #a855f7);
  border-radius: 2px;
}

.nav-icon {
  flex-shrink: 0;
  transition: color 0.15s;
}

.nav-label {
  flex: 1;
}

/* ---- Kanban group ---- */

.kanban-group {
  margin-bottom: 2px;
}

.kanban-add-btn {
  display: none;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: rgba(148, 163, 184, 0.35);
  cursor: pointer;
  flex-shrink: 0;
  margin-left: auto;
  transition: background 0.15s, color 0.15s;
}

.kanban-add-btn:hover {
  background: rgba(99, 102, 241, 0.15);
  color: #a5b4fc;
}

/* Show on hover */
.kanban-nav-item:hover .kanban-add-btn {
  display: flex;
}

/* Show when kanban is active */
.kanban-nav-item.nav-item--active .kanban-add-btn {
  display: flex;
}

.kanban-subnav {
  padding: 2px 0 4px 20px;
}

.subnav-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 12px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  color: rgba(148, 163, 184, 0.6);
  cursor: pointer;
  text-decoration: none;
  transition: background 0.15s, color 0.15s;
}

.subnav-item:hover {
  background: rgba(255, 255, 255, 0.04);
  color: #e2e8f0;
}

.subnav-item--active {
  background: rgba(99, 102, 241, 0.1);
  color: #c7d2fe;
}

.subnav-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.subnav-label {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  transition: color 0.15s;
}

.subnav-label--dragover {
  color: #a5b4fc;
}

.subnav-rename-input {
  flex: 1;
  min-width: 0;
  padding: 2px 8px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(99, 102, 241, 0.4);
  border-radius: 5px;
  color: #e2e8f0;
  font-size: 13px;
  font-family: inherit;
  outline: none;
}

/* ---- Sidebar footer ---- */

.sidebar-footer {
  border-top: 1px solid rgba(255, 255, 255, 0.04);
  padding: 8px 12px;
}

.user-card {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  margin-top: 4px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.02);
}

.user-avatar {
  width: 30px;
  height: 30px;
  border-radius: 10px;
  background: linear-gradient(135deg, #6366f1, #ec4899);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
}

.user-avatar--sm {
  width: 32px;
  height: 32px;
  cursor: pointer;
  box-shadow: 0 0 16px rgba(99, 102, 241, 0.2);
}

.user-info {
  flex: 1;
  min-width: 0;
}

.user-name {
  font-size: 12px;
  font-weight: 600;
  color: #e2e8f0;
}

.user-email {
  font-size: 11px;
  color: rgba(148, 163, 184, 0.5);
}

/* ---- Mobile header ---- */

.mobile-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 56px;
  background: rgba(10, 14, 26, 0.9);
  backdrop-filter: blur(20px) saturate(180%);
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  z-index: 30;
}

.mobile-header-left,
.mobile-header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* ---- Main content ---- */

.main-content {
  flex: 1;
  min-width: 0;
  padding-top: 56px;
  min-height: 100dvh;
  transition: margin-left 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

/* ---- Desktop topbar ---- */

.topbar {
  display: none;
}

.topbar-crumb {
  display: flex;
  align-items: center;
  gap: 6px;
}

.topbar-page {
  font-size: 14px;
  font-weight: 600;
  color: #e2e8f0;
  text-transform: capitalize;
}

.topbar-chevron {
  color: rgba(148, 163, 184, 0.3);
}

.topbar-section {
  font-size: 13px;
  color: rgba(148, 163, 184, 0.5);
}

.topbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.topbar-search-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(255, 255, 255, 0.02);
  border-radius: 8px;
  color: rgba(148, 163, 184, 0.5);
  font-size: 12px;
  font-family: inherit;
  cursor: pointer;
  transition: border-color 0.15s;
}

.topbar-search-btn kbd {
  font-size: 10px;
  color: rgba(148, 163, 184, 0.3);
  background: rgba(255, 255, 255, 0.04);
  border-radius: 4px;
  padding: 1px 5px;
  margin-left: 8px;
}

.topbar-search-btn:hover {
  border-color: rgba(255, 255, 255, 0.12);
}

/* ---- Page wrap ---- */

.page-wrap {
  padding: 16px;
}

/* ---- Transitions ---- */

.cmd-fade-enter-active,
.cmd-fade-leave-active {
  transition: opacity 0.2s ease;
}

.cmd-fade-enter-from,
.cmd-fade-leave-to {
  opacity: 0;
}

.overlay-enter-active,
.overlay-leave-active {
  transition: opacity 0.2s ease;
}

.overlay-enter-from,
.overlay-leave-to {
  opacity: 0;
}

/* ---- Desktop ---- */

@media (min-width: 768px) {
  .mobile-header {
    display: none;
  }

  .sidebar {
    transform: translateX(0);
  }

  .sidebar-close {
    display: none;
  }

  .collapse-toggle {
    display: flex;
    width: 24px;
    height: 24px;
    border-radius: 6px;
    /* Point left (←) in expanded state to signal "collapse" */
    transform: rotate(180deg);
    transition: transform 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  }

  /* Point right (→) in collapsed state to signal "expand" */
  .sidebar--collapsed .collapse-toggle {
    transform: rotate(0deg);
  }

  .sidebar--collapsed {
    width: 68px;
  }

  .sidebar--collapsed .sidebar-header {
    padding: 0 12px;
    justify-content: center;
  }

  .sidebar--collapsed .logo-text {
    display: none;
  }

  .sidebar--collapsed .sidebar-search-wrap {
    display: none;
  }

  .sidebar--collapsed .sidebar-nav {
    padding: 12px 8px;
  }

  .sidebar--collapsed .nav-item {
    justify-content: center;
    padding: 10px;
  }

  .sidebar--collapsed .nav-label {
    display: none;
  }

  /* In collapsed mode, active indicator shifts to bottom edge */
  .sidebar--collapsed .nav-item--active::before {
    left: 50%;
    top: auto;
    bottom: -2px;
    transform: translateX(-50%);
    width: 60%;
    height: 3px;
  }

  .sidebar--collapsed .kanban-subnav {
    display: none;
  }

  .sidebar--collapsed .kanban-add-btn {
    display: none !important;
  }

  .sidebar--collapsed .sidebar-footer {
    padding: 8px;
  }

  .sidebar--collapsed .sidebar-footer .nav-item {
    justify-content: center;
  }

  .sidebar--collapsed .user-card {
    display: none;
  }

  .main-content {
    padding-top: 0;
    margin-left: 256px;
  }

  .main-content--collapsed {
    margin-left: 68px;
  }

  .topbar {
    display: none;
  }

  .page-wrap {
    padding: 28px;
  }
}

/* ---- Intermediate: narrower sidebar ---- */

@media (min-width: 768px) and (max-width: 1099px) {
  .sidebar:not(.sidebar--collapsed) {
    width: 200px;
  }

  .sidebar:not(.sidebar--collapsed) .sidebar-search-btn kbd {
    display: none;
  }

  .main-content:not(.main-content--collapsed) {
    margin-left: 200px;
  }

  .page-wrap {
    padding: 20px;
  }
}
</style>
