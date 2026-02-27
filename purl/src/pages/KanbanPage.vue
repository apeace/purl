<template>
  <div class="kanban-page">
    <div class="kanban-header">
      <template v-if="editingBoardTitle">
        <input
          ref="boardTitleInputRef"
          v-model="boardTitleValue"
          class="page-title-input"
          @blur="commitBoardTitle"
          @keydown.enter.prevent="commitBoardTitle"
          @keydown.escape="cancelBoardTitle"
        />
      </template>
      <h1
        v-else
        class="page-title"
        :class="{ 'page-title--editable': canEditBoard }"
        @click="startEditBoardTitle"
      >{{ pageTitle }}</h1>
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
        <FilterPanel :custom-stages="currentBoard?.stages" />
      </div>
    </div>
    <div
      ref="stagesScrollEl"
      class="stages-scroll"
      @dragover.prevent="onColumnDragOver"
      @dragleave="onColumnDragLeave"
      @drop="onColumnDrop"
    >
      <KanbanStage
        v-for="(stage, i) in stages"
        :key="stage.status"
        v-bind="stage"
        :dragging-id="draggingId"
        :can-edit="canEditBoard"
        :class="{ 'stage--col-dragging': columnDraggingId === stage.status }"
        :delay="150 + i * 100"
        @select="onCardSelect"
        @drag-start="draggingId = $event"
        @drag-end="draggingId = null"
        @drop="onDrop"
        @column-drag-start="columnDraggingId = $event"
        @column-drag-end="columnDraggingId = null; columnDropIndex = -1"
        @rename="onColumnRename"
        @change-color="onColumnColorChange"
        @delete="onColumnDelete"
      />
      <div
        v-if="columnDropIndex >= 0 && columnDropX !== null"
        class="col-drop-indicator"
        :style="{ left: `${columnDropX}px` }"
      />

      <!-- Add column -->
      <template v-if="canEditBoard">
        <div v-if="addingColumn" class="add-column-form">
          <input
            ref="newColumnInputRef"
            v-model="newColumnName"
            class="add-column-input"
            placeholder="Column name…"
            @blur="commitAddColumn"
            @keydown.enter.prevent="commitAddColumn"
            @keydown.escape="cancelAddColumn"
          />
        </div>
        <button v-else class="add-column-btn" @click="startAddColumn">
          <Plus :size="14" />
          <span>Add column</span>
        </button>
      </template>
    </div>

    <!-- Board picker for "Add to Board" flow -->
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

    <ConfirmModal
      :visible="confirmDeleteColumnVisible"
      title="Delete this column?"
      message="Cards in this column will be removed from the board. Tickets will not be deleted."
      @confirm="confirmDeleteColumn"
      @cancel="confirmDeleteColumnVisible = false"
    />
  </div>
</template>

<script setup lang="ts">
import { Plus, Search, X } from "lucide-vue-next"
import { storeToRefs } from "pinia"
import { computed, nextTick, ref, watch } from "vue"
import { useRoute, useRouter } from "vue-router"
import { COLUMN_COLORS } from "../components/ColorPicker.vue"
import ConfirmModal from "../components/ConfirmModal.vue"
import FilterPanel from "../components/FilterPanel.vue"
import KanbanStage from "../components/KanbanStage.vue"
import StagePickerModal from "../components/StagePickerModal.vue"
import { useKanbanStore } from "../stores/useKanbanStore"
import type { BoardStage } from "../stores/useKanbanStore"
import { useTicketStore } from "../stores/useTicketStore"

const route = useRoute()
const router = useRouter()
const kanbanStore = useKanbanStore()
const { boards } = storeToRefs(kanbanStore)
const { addCardToBoard, addColumn, changeColumnColor, deleteColumn, getBoardById, moveCard, renameBoard, renameColumn, reorderColumns } = kanbanStore
const ticketStore = useTicketStore()
const { filterKeyword, tickets } = storeToRefs(ticketStore)
const { filterAssignees, filterStatuses } = ticketStore

const draggingId = ref<string | null>(null)
const searchQuery = ref("")
const stagesScrollEl = ref<HTMLElement | null>(null)

// ── Column drag-and-drop ─────────────────────────────────

const columnDraggingId = ref<string | null>(null)
const columnDropIndex = ref(-1)
const columnDropX = ref<number | null>(null)

// ── Board title rename ───────────────────────────────────

const editingBoardTitle = ref(false)
const boardTitleValue = ref("")
const boardTitleInputRef = ref<HTMLInputElement | null>(null)

// ── Board awareness ──────────────────────────────────────

const boardId = computed(() => (route.params.boardId as string) ?? null)
const currentBoard = computed(() =>
  boardId.value ? getBoardById(boardId.value) : (boards.value.find((b) => b.isDefault) ?? null)
)

const pageTitle = computed(() => currentBoard.value?.name ?? "Kanban")
const canEditBoard = computed(() => !!currentBoard.value && !currentBoard.value.isDefault)

// Reset local state on board switch
watch(boardId, () => {
  searchQuery.value = ""
  filterStatuses.clear()
  editingBoardTitle.value = false
})

// ── Stage definitions ────────────────────────────────────

const stageDefs = computed(() => {
  if (!currentBoard.value) return []
  return currentBoard.value.stages.map((s) => ({
    status: s.id,
    title: s.name,
    color: s.color,
  }))
})

// ── Search helper ────────────────────────────────────────

function matchesSearch(t: { name: string; company: string; subject: string; messages: { text: string }[] }, q: string): boolean {
  if (!q) return true
  const desc = t.messages[0]?.text ?? ""
  return t.name.toLowerCase().includes(q)
    || t.company.toLowerCase().includes(q)
    || t.subject.toLowerCase().includes(q)
    || desc.toLowerCase().includes(q)
}

// ── Stages computed ──────────────────────────────────────

const stages = computed(() => {
  if (!currentBoard.value) return []
  const board = currentBoard.value
  const q = searchQuery.value.trim().toLowerCase()
  const stageFilter = filterStatuses.size > 0
  return stageDefs.value
    .filter((def) => !stageFilter || filterStatuses.has(def.status))
    .map((def) => {
      const assignedIds = Object.entries(board.cardAssignments)
        .filter(([, stageId]) => stageId === def.status)
        .map(([ticketId]) => ticketId)
      let items = tickets.value.filter((t) => assignedIds.includes(t.id))
      const kw = filterKeyword.value.trim().toLowerCase()
      if (kw) items = items.filter((t) => t.subject.toLowerCase().includes(kw) || (t.messages[0]?.text ?? "").toLowerCase().includes(kw))
      if (filterAssignees.size) items = items.filter((t) => filterAssignees.has(t.assignee))
      if (q) items = items.filter((t) => matchesSearch(t, q))
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
          avatarColor: t.avatarColor,
        })),
      }
    })
})

// ── Drop handler ─────────────────────────────────────────

function onDrop({ ticketId, status }: { ticketId: string; status: string }) {
  draggingId.value = null
  if (!currentBoard.value) return
  moveCard(currentBoard.value.id, ticketId, status)
}

// ── Column rename / delete ───────────────────────────────

function onColumnRename(stageId: string, name: string) {
  if (!currentBoard.value) return
  renameColumn(currentBoard.value.id, stageId, name)
}

function onColumnColorChange(stageId: string, color: string) {
  if (!currentBoard.value) return
  changeColumnColor(currentBoard.value.id, stageId, color)
}

const confirmDeleteColumnVisible = ref(false)
const pendingDeleteColumnId = ref<string | null>(null)

function onColumnDelete(stageId: string) {
  pendingDeleteColumnId.value = stageId
  confirmDeleteColumnVisible.value = true
}

function confirmDeleteColumn() {
  confirmDeleteColumnVisible.value = false
  if (!pendingDeleteColumnId.value || !currentBoard.value) return
  deleteColumn(currentBoard.value.id, pendingDeleteColumnId.value)
  pendingDeleteColumnId.value = null
}

// ── Add column ───────────────────────────────────────────

const COLUMN_PALETTE = COLUMN_COLORS

const addingColumn = ref(false)
const newColumnName = ref("")
const newColumnInputRef = ref<HTMLInputElement | null>(null)

function nextColumnColor(): string {
  if (!currentBoard.value) return COLUMN_PALETTE[0]
  const used = new Set(currentBoard.value.stages.map((s) => s.color))
  return COLUMN_PALETTE.find((c) => !used.has(c)) ?? COLUMN_PALETTE[currentBoard.value.stages.length % COLUMN_PALETTE.length]
}

function startAddColumn() {
  addingColumn.value = true
  newColumnName.value = ""
  nextTick(() => newColumnInputRef.value?.focus())
}

function commitAddColumn() {
  const name = newColumnName.value.trim()
  addingColumn.value = false
  if (!name || !currentBoard.value) return
  addColumn(currentBoard.value.id, name, nextColumnColor())
}

function cancelAddColumn() {
  addingColumn.value = false
}

// ── Column drag-and-drop reorder ─────────────────────────

function onColumnDragOver(event: DragEvent) {
  if (!event.dataTransfer?.types.includes("application/x-purl-column")) return
  const scroll = stagesScrollEl.value
  if (!scroll) return
  const stageEls = scroll.querySelectorAll(".stage")
  let idx = stageEls.length
  for (let i = 0; i < stageEls.length; i++) {
    const rect = stageEls[i].getBoundingClientRect()
    if (event.clientX < rect.left + rect.width / 2) { idx = i; break }
  }
  columnDropIndex.value = idx

  // Calculate pixel X for the drop indicator, relative to scroll container
  const scrollRect = scroll.getBoundingClientRect()
  if (stageEls.length === 0) {
    columnDropX.value = null
    return
  }
  if (idx === 0) {
    const rect = stageEls[0].getBoundingClientRect()
    // Clamp to 2px minimum so the indicator is never clipped at the left edge
    columnDropX.value = Math.max(2, rect.left - scrollRect.left + scroll.scrollLeft - 8)
  } else if (idx >= stageEls.length) {
    const rect = stageEls[stageEls.length - 1].getBoundingClientRect()
    columnDropX.value = rect.right - scrollRect.left + scroll.scrollLeft + 6
  } else {
    const prevRect = stageEls[idx - 1].getBoundingClientRect()
    const nextRect = stageEls[idx].getBoundingClientRect()
    columnDropX.value = (prevRect.right + nextRect.left) / 2 - scrollRect.left + scroll.scrollLeft
  }
}

function onColumnDragLeave(event: DragEvent) {
  if (!stagesScrollEl.value?.contains(event.relatedTarget as Node)) {
    columnDropIndex.value = -1
    columnDropX.value = null
  }
}

function onColumnDrop(event: DragEvent) {
  const stageId = event.dataTransfer?.getData("application/x-purl-column")
  if (!stageId || !currentBoard.value) return
  const stages = currentBoard.value.stages
  const fromIdx = stages.findIndex((s) => s.id === stageId)
  let toIdx = columnDropIndex.value
  columnDropIndex.value = -1
  columnDropX.value = null
  if (toIdx < 0 || fromIdx === -1 || toIdx === fromIdx || toIdx === fromIdx + 1) return
  const newOrder = stages.map((s) => s.id)
  newOrder.splice(fromIdx, 1)
  if (toIdx > fromIdx) toIdx--
  newOrder.splice(toIdx, 0, stageId)
  reorderColumns(currentBoard.value.id, newOrder)
}

// ── Board title rename ───────────────────────────────────

function startEditBoardTitle() {
  if (!canEditBoard.value || !currentBoard.value) return
  boardTitleValue.value = currentBoard.value.name
  editingBoardTitle.value = true
  nextTick(() => { boardTitleInputRef.value?.focus(); boardTitleInputRef.value?.select() })
}

function commitBoardTitle() {
  if (boardTitleValue.value.trim() && currentBoard.value) {
    renameBoard(currentBoard.value.id, boardTitleValue.value.trim())
  }
  editingBoardTitle.value = false
}

function cancelBoardTitle() {
  editingBoardTitle.value = false
}

// ── Card navigation ───────────────────────────────────────

function onCardSelect(ticketId: string) {
  if (!currentBoard.value) return
  const stageId = currentBoard.value.cardAssignments[ticketId]
  router.push({
    path: `/ticket/${ticketId}`,
    query: { board: currentBoard.value.id, stage: stageId },
  })
}

// ── "Add to Board" flow ──────────────────────────────────

const showBoardPicker = ref(false)
const addToBoardTicketId = ref<string | null>(null)
const showStagePicker = ref(false)
const stagePickerBoardName = ref("")
const stagePickerStages = ref<BoardStage[]>([])
const stagePickerBoardId = ref<string | null>(null)

const availableBoards = computed(() => boards.value.filter((b) => !b.isDefault))

function handleAddToBoard(ticketId: string) {
  addToBoardTicketId.value = ticketId
  if (boards.value.length === 0) return
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

.page-title--editable {
  cursor: text;
}

.page-title--editable:hover {
  text-decoration: underline;
  text-decoration-color: rgba(99, 102, 241, 0.4);
  text-underline-offset: 3px;
}

.page-title-input {
  font-size: 22px;
  font-weight: 700;
  color: #f1f5f9;
  letter-spacing: -0.02em;
  background: transparent;
  border: none;
  border-bottom: 2px solid rgba(99, 102, 241, 0.6);
  outline: none;
  font-family: inherit;
  padding: 0 2px 2px;
  min-width: 120px;
  width: auto;
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
  padding: 0 0 16px 4px;
  position: relative;
}

.stage--col-dragging {
  opacity: 0.35;
  transition: opacity 0.15s;
}

.add-column-btn {
  display: inline-flex;
  flex-shrink: 0;
  align-self: flex-start;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border: 1px dashed rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  background: transparent;
  color: rgba(148, 163, 184, 0.4);
  font-size: 13px;
  font-family: inherit;
  cursor: pointer;
  white-space: nowrap;
  transition: border-color 0.15s, color 0.15s, background 0.15s;
}

.add-column-btn:hover {
  border-color: rgba(99, 102, 241, 0.35);
  color: #a5b4fc;
  background: rgba(99, 102, 241, 0.05);
}

.add-column-form {
  flex-shrink: 0;
  align-self: flex-start;
  min-width: 180px;
}

.add-column-input {
  width: 100%;
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(99, 102, 241, 0.45);
  border-radius: 10px;
  color: #e2e8f0;
  font-size: 13px;
  font-family: inherit;
  outline: none;
}

.add-column-input::placeholder {
  color: rgba(148, 163, 184, 0.3);
}

.col-drop-indicator {
  position: absolute;
  top: 0;
  bottom: 16px;
  width: 2px;
  background: rgba(99, 102, 241, 0.7);
  border-radius: 1px;
  pointer-events: none;
  z-index: 10;
  transform: translateX(-50%);
}

/* ── Board / Stage picker (inline in KanbanPage) ─────────── */

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

/* ── Intermediate screens ────────────────────────────────── */

@media (min-width: 768px) and (max-width: 1099px) {
  .kanban-search-input {
    width: 120px;
  }
}
</style>
