import { defineStore } from "pinia"
import { ref } from "vue"
import {
  deleteKanbansByBoardId,
  getKanbans,
  getKanbansByBoardIdTickets,
  patchKanbansByBoardId,
  postKanbans,
  putKanbansByBoardIdColumns,
  putKanbansByBoardIdColumnsByColumnIdTickets,
} from "@purl/lib"
import type { AppKanbanBoard } from "@purl/lib"

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

// ── Helpers ─────────────────────────────────────────────

function toKanbanBoard(apiBoard: AppKanbanBoard): KanbanBoard {
  return {
    id: apiBoard.id ?? "",
    name: apiBoard.name ?? "",
    stages: (apiBoard.columns ?? [])
      .sort((a, b) => (a.position ?? 0) - (b.position ?? 0))
      .map((c) => ({ id: c.id ?? "", name: c.name ?? "", color: c.color ?? "#94a3b8" })),
    cardAssignments: {},
  }
}

// ── Store ────────────────────────────────────────────────

export const useKanbanStore = defineStore("kanban", () => {
  const boards = ref<KanbanBoard[]>([])

  // ── Data fetching ────────────────────────────────────────

  async function loadBoardTickets(boardId: string) {
    const { data } = await getKanbansByBoardIdTickets({ path: { boardID: boardId } })
    if (!data) return
    const board = boards.value.find((b) => b.id === boardId)
    if (!board) return
    board.cardAssignments = {}
    for (const t of data) {
      if (t.id && t.column_id) board.cardAssignments[t.id] = t.column_id
    }
  }

  async function loadBoards() {
    const { data } = await getKanbans()
    if (!data) return
    // The default board is the built-in Service Flow; exclude it from custom boards
    boards.value = data.filter((b) => !b.is_default).map(toKanbanBoard)
    await Promise.all(boards.value.map((b) => loadBoardTickets(b.id)))
  }

  // ── Mutations ───────────────────────────────────────────

  async function createBoard(name: string, stages: { name: string; color: string }[]): Promise<KanbanBoard> {
    const { data: created } = await postKanbans({ body: { name } })
    if (!created?.id) throw new Error("Failed to create board")

    const { data: columns } = await putKanbansByBoardIdColumns({
      path: { boardID: created.id },
      body: stages.map((s, i) => ({ color: s.color, name: s.name, position: i })),
    })

    const board: KanbanBoard = {
      id: created.id,
      name: created.name ?? name,
      stages: (columns ?? []).map((c) => ({
        color: c.color ?? "#94a3b8",
        id: c.id ?? "",
        name: c.name ?? "",
      })),
      cardAssignments: {},
    }
    boards.value.push(board)
    return board
  }

  async function deleteBoard(boardId: string) {
    const idx = boards.value.findIndex((b) => b.id === boardId)
    if (idx !== -1) boards.value.splice(idx, 1) // optimistic update
    await deleteKanbansByBoardId({ path: { boardID: boardId } })
  }

  async function renameBoard(boardId: string, name: string) {
    const board = boards.value.find((b) => b.id === boardId)
    if (board) board.name = name // optimistic update
    await patchKanbansByBoardId({ path: { boardID: boardId }, body: { name } })
  }

  async function addCardToBoard(boardId: string, ticketId: string, stageId: string) {
    const board = boards.value.find((b) => b.id === boardId)
    if (!board) return
    board.cardAssignments[ticketId] = stageId // optimistic update
    const ticketIds = Object.entries(board.cardAssignments)
      .filter(([, sid]) => sid === stageId)
      .map(([tid]) => tid)
    await putKanbansByBoardIdColumnsByColumnIdTickets({
      path: { boardID: boardId, columnID: stageId },
      body: ticketIds,
    })
  }

  async function moveCard(boardId: string, ticketId: string, stageId: string) {
    const board = boards.value.find((b) => b.id === boardId)
    if (!board) return
    board.cardAssignments[ticketId] = stageId // optimistic update
    // The API automatically moves the ticket out of its previous column when
    // it appears in a PUT for a different column, so we only need to update the target.
    const ticketIds = Object.entries(board.cardAssignments)
      .filter(([, sid]) => sid === stageId)
      .map(([tid]) => tid)
    await putKanbansByBoardIdColumnsByColumnIdTickets({
      path: { boardID: boardId, columnID: stageId },
      body: ticketIds,
    })
  }

  function getBoardById(boardId: string): KanbanBoard | undefined {
    return boards.value.find((b) => b.id === boardId)
  }

  // Auto-load on first store access
  loadBoards()

  return {
    addCardToBoard,
    boards,
    createBoard,
    deleteBoard,
    getBoardById,
    moveCard,
    renameBoard,
  }
})
