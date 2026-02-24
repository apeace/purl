import { ref, watch } from "vue"
import { defineStore } from "pinia"

// ── Types ───────────────────────────────────────────────

export interface BoardStage {
  id: string
  name: string
  color: string
}

export interface KanbanBoard {
  id: string
  name: string
  stages: BoardStage[]
  cardAssignments: Record<string, string> // ticketId → stageId
}

// ── Persistence ─────────────────────────────────────────

const STORAGE_KEY = "purl_kanban_boards"

function loadFromStorage(): KanbanBoard[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    return raw ? JSON.parse(raw) : []
  } catch {
    return []
  }
}

function saveToStorage(data: KanbanBoard[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(data))
}

// ── Helpers ─────────────────────────────────────────────

function uid(): string {
  return Date.now().toString(36) + Math.random().toString(36).slice(2, 8)
}

// ── Store ────────────────────────────────────────────────

export const useKanbanStore = defineStore("kanban", () => {
  const boards = ref<KanbanBoard[]>(loadFromStorage())

  watch(boards, (val) => saveToStorage(val), { deep: true })

  // ── Mutations ───────────────────────────────────────────

  function createBoard(name: string, stages: { name: string; color: string }[]): KanbanBoard {
    const board: KanbanBoard = {
      id: uid(),
      name,
      stages: stages.map((s) => ({ id: uid(), name: s.name, color: s.color })),
      cardAssignments: {},
    }
    boards.value.push(board)
    return board
  }

  function deleteBoard(boardId: string) {
    const idx = boards.value.findIndex((b) => b.id === boardId)
    if (idx !== -1) boards.value.splice(idx, 1)
  }

  function renameBoard(boardId: string, name: string) {
    const board = boards.value.find((b) => b.id === boardId)
    if (board) board.name = name
  }

  function addCardToBoard(boardId: string, ticketId: string, stageId: string) {
    const board = boards.value.find((b) => b.id === boardId)
    if (board) board.cardAssignments[ticketId] = stageId
  }

  function removeCardFromBoard(boardId: string, ticketId: string) {
    const board = boards.value.find((b) => b.id === boardId)
    if (board) delete board.cardAssignments[ticketId]
  }

  function moveCard(boardId: string, ticketId: string, stageId: string) {
    const board = boards.value.find((b) => b.id === boardId)
    if (board) board.cardAssignments[ticketId] = stageId
  }

  function getBoardById(boardId: string): KanbanBoard | undefined {
    return boards.value.find((b) => b.id === boardId)
  }

  return {
    addCardToBoard,
    boards,
    createBoard,
    deleteBoard,
    getBoardById,
    moveCard,
    removeCardFromBoard,
    renameBoard,
  }
})
