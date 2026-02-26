import type { AppTicketCommentRow, AppTicketRow } from "@purl/lib"
import { getTickets, getTicketsByTicketIdComments } from "@purl/lib"
import { defineStore } from "pinia"
import { computed, reactive, ref } from "vue"
import type { CallData, CommChannel, MergeData, MessageType, VoicemailData } from "../utils/parseComment"
import { parseComment, stripHtml } from "../utils/parseComment"

// ── Types ───────────────────────────────────────────────

export interface Message {
  id: number
  from: string
  channel: string
  time: string
  text: string
  authorName?: string
  messageType?: MessageType
  commChannel?: CommChannel
  call?: CallData
  voicemail?: VoicemailData
  merge?: MergeData
  transcript?: string
  hasRecording?: boolean
  commentId?: string
  type?: string
  recording?: {
    duration: number
    waveform: number[]
    transcript: { speaker: string; time: string; text: string }[]
  }
}

export interface Ticket {
  id: string
  name: string
  company: string
  ticketId: string
  subject: string
  description: string
  createdAt: string
  wait: string
  avatarColor: string
  status: string
  read: boolean
  starred: boolean
  labels: string[]
  time: string
  email: string
  phone: string
  subscription: { status: string; id: string; plan: string }
  tags: string[]
  temperature: string
  assignee: string
  notes: string
  messages: Message[]
  ticketHistory: { time: string; event: string }[]
  subscriberHistory: { ticketId: string; status: string; subject: string; date: string }[]
}

// ── Module-level helpers (exported for direct import at call sites) ──

const STATUS_ORDER: Record<string, number> = { escalated: 0, new: 1, open: 2, pending: 3, solved: 4, closed: 5 }

const AVATAR_COLORS = [
  "#6366f1", "#ec4899", "#34d399", "#f59e0b",
  "#3b82f6", "#a855f7", "#ef4444", "#14b8a6",
]

export function avatarColor(name: string): string {
  let hash = 0
  for (const ch of name) hash = (hash * 31 + ch.charCodeAt(0)) & 0xffff
  return AVATAR_COLORS[hash % AVATAR_COLORS.length]
}

export function parseWait(str: string): number {
  let mins = 0
  const d = str.match(/(\d+)d/)
  const h = str.match(/(\d+)h/)
  const m = str.match(/(\d+)m/)
  if (d) mins += parseInt(d[1]) * 1440
  if (h) mins += parseInt(h[1]) * 60
  if (m) mins += parseInt(m[1])
  return mins
}

function formatWait(createdAt: string): string {
  const mins = Math.floor((Date.now() - new Date(createdAt).getTime()) / 60000)
  if (mins < 60) return `${mins}m`
  const hrs = Math.floor(mins / 60)
  const rem = mins % 60
  return rem ? `${hrs}h ${rem}m` : `${hrs}h`
}

function formatTime(createdAt: string): string {
  return new Date(createdAt).toLocaleTimeString([], { hour: "numeric", minute: "2-digit" })
}

// Extended type until the OpenAPI client regenerates with new fields
type TicketRowExt = AppTicketRow & {
  zendesk_ticket_id?: number
  assignee_name?: string
  reporter_email?: string
}

function toTicket(raw: AppTicketRow): Ticket {
  const t = raw as TicketRowExt
  const id = t.id ?? ""
  const reporterName = t.reporter_name ?? ""
  const createdAt = t.created_at ?? ""
  const zendeskId = t.zendesk_ticket_id
  return {
    id,
    name: reporterName,
    company: "",
    ticketId: zendeskId ? `#${zendeskId}` : `#${id.slice(0, 6).toUpperCase()}`,
    subject: t.title ?? "",
    description: stripHtml(t.description ?? ""),
    createdAt,
    wait: formatWait(createdAt),
    avatarColor: avatarColor(reporterName),
    status: t.zendesk_status ?? "",
    read: false,
    starred: false,
    labels: [],
    time: formatTime(createdAt),
    email: t.reporter_email ?? "",
    phone: "",
    subscription: { status: "active", id: "", plan: "" },
    tags: [],
    temperature: "warm",
    assignee: t.assignee_name ?? "Unassigned",
    notes: "",
    messages: [],
    ticketHistory: [
      { time: formatWait(createdAt), event: "Ticket created" },
    ],
    subscriberHistory: [],
  }
}

// ── Store ────────────────────────────────────────────────

export const useTicketStore = defineStore("tickets", () => {
  // ── State ─────────────────────────────────────────────

  const tickets = ref<Ticket[]>([])

  // TODO: derive resolvedToday from real ticket data instead of hardcoding a starting value
  const resolvedToday = ref(8)

  // ── Filters ─────────────────────────────────────────────

  const filterKeyword = ref("")
  const filterAssignees = reactive(new Set<string>())
  const filterStatuses = reactive(new Set<string>())

  const filteredTickets = computed(() => {
    let result = tickets.value

    const kw = filterKeyword.value.trim().toLowerCase()
    if (kw) {
      result = result.filter((t) => {
        const firstMsg = t.messages[0]?.text ?? ""
        return t.subject.toLowerCase().includes(kw) || firstMsg.toLowerCase().includes(kw)
      })
    }

    if (filterAssignees.size) {
      result = result.filter((t) => filterAssignees.has(t.assignee))
    }
    if (filterStatuses.size) {
      result = result.filter((t) => filterStatuses.has(t.status))
    }

    return result
  })

  const activeFilterCount = computed(() => {
    let n = 0
    if (filterKeyword.value.trim()) n++
    if (filterAssignees.size) n++
    if (filterStatuses.size) n++
    return n
  })

  const uniqueAssignees = computed(() =>
    [...new Set(tickets.value.map((t) => t.assignee))].sort()
  )

  function clearFilters() {
    filterKeyword.value = ""
    filterAssignees.clear()
    filterStatuses.clear()
  }

  // ── Sorting ─────────────────────────────────────────────

  const sortBy = ref("time")

  const sortedTickets = computed(() => {
    const list = [...filteredTickets.value]
    if (sortBy.value === "status") {
      list.sort((a, b) => (STATUS_ORDER[a.status] ?? 9) - (STATUS_ORDER[b.status] ?? 9))
    } else if (sortBy.value === "assignee") {
      list.sort((a, b) => a.assignee.localeCompare(b.assignee))
    }
    return list
  })

  // ── Derived state ───────────────────────────────────────

  const openTickets = computed(() => tickets.value.filter((t) => t.status === "new" || t.status === "open"))

  const hudOpen = computed(() => openTickets.value.length)

  const hudLongestWait = computed(() => {
    if (!openTickets.value.length) return "0m"
    let max = 0
    let maxStr = "0m"
    for (const t of openTickets.value) {
      const mins = parseWait(t.wait)
      if (mins > max) {
        max = mins
        maxStr = t.wait
      }
    }
    return maxStr
  })

  const hudResolvedToday = computed(() => resolvedToday.value)

  // ── Mutations ───────────────────────────────────────────

  // TODO: persist status change via API
  function resolveTicket(id: string) {
    const ticket = tickets.value.find((t) => t.id === id)
    if (!ticket || ticket.status === "solved" || ticket.status === "closed") return
    ticket.status = "solved"
    ticket.read = true
    resolvedToday.value++
  }

  // TODO: persist status change via API
  function archiveTicket(id: string) {
    const ticket = tickets.value.find((t) => t.id === id)
    if (ticket) ticket.status = "closed"
  }

  // TODO: persist deletion via API
  function deleteTicket(id: string) {
    const ticket = tickets.value.find((t) => t.id === id)
    if (ticket) ticket.status = "closed"
  }

  // TODO: persist read state via API
  function markRead(id: string) {
    const ticket = tickets.value.find((t) => t.id === id)
    if (ticket) ticket.read = true
  }

  // TODO: persist starred state via API
  function toggleStar(id: string) {
    const ticket = tickets.value.find((t) => t.id === id)
    if (ticket) ticket.starred = !ticket.starred
  }

  // TODO: send reply via API
  function sendReply(id: string, text: string, channel = "email") {
    const ticket = tickets.value.find((t) => t.id === id)
    if (!ticket) return
    ticket.messages.push({
      id: ticket.messages.length + 1,
      from: "agent",
      channel,
      time: "just now",
      text,
    })
    ticket.read = true
  }

  // TODO: persist status change via API
  function setStatus(id: string, status: string) {
    const ticket = tickets.value.find((t) => t.id === id)
    if (ticket) ticket.status = status
  }

  // TODO: persist assignee change via API
  function setAssignee(id: string, assignee: string) {
    const ticket = tickets.value.find((t) => t.id === id)
    if (ticket) ticket.assignee = assignee
  }

  // TODO: persist temperature change via API
  function setTemperature(id: string, temperature: string) {
    const ticket = tickets.value.find((t) => t.id === id)
    if (ticket) ticket.temperature = temperature
  }

  // TODO: persist tag addition via API
  function addTag(id: string, tag: string) {
    const ticket = tickets.value.find((t) => t.id === id)
    if (ticket && !ticket.tags.includes(tag)) ticket.tags.push(tag)
  }

  // TODO: persist tag removal via API
  function removeTag(id: string, tag: string) {
    const ticket = tickets.value.find((t) => t.id === id)
    if (ticket) ticket.tags = ticket.tags.filter((t) => t !== tag)
  }

  // TODO: persist notes via API
  function updateNotes(id: string, text: string) {
    const ticket = tickets.value.find((t) => t.id === id)
    if (ticket) ticket.notes = text
  }

  // ── Data fetching ────────────────────────────────────────

  let loadPromise: Promise<void> | null = null

  function loadTickets() {
    if (!loadPromise) {
      loadPromise = getTickets().then(({ data }) => {
        if (data) tickets.value = data.map(toTicket)
      })
    }
    return loadPromise
  }

  function reloadTickets() {
    loadPromise = null
    return loadTickets()
  }

  // Track which ticket IDs have had comments loaded to avoid duplicate fetches
  const loadedCommentTickets = new Set<string>()

  // Extended type until the OpenAPI client regenerates with new fields
  type CommentRowExt = AppTicketCommentRow & {
    author_name?: string
    call_id?: number
    has_recording?: boolean
    transcription_text?: string
    transcription_status?: string
    call_duration?: number
    call_from?: string
    call_to?: string
    answered_by_name?: string
    call_location?: string
    call_started_at?: string
  }

  function commentToMessage(raw: AppTicketCommentRow, index: number): Message {
    const c = raw as CommentRowExt
    const body = c.body ?? ""
    const channel = c.channel ?? "email"
    const role = c.role ?? "customer"
    const parsed = parseComment(body, channel, role)

    const msg: Message = {
      id: index,
      from: role === "agent" ? "agent" : "customer",
      channel,
      time: formatWait(c.created_at ?? ""),
      text: parsed.cleanBody,
      authorName: c.author_name || undefined,
      messageType: parsed.messageType,
      commChannel: parsed.commChannel,
      call: parsed.call,
      voicemail: parsed.voicemail,
      merge: parsed.merge,
    }

    // Prefer structured voice data from DB when available
    if (c.call_id) {
      msg.commentId = c.id ?? undefined
      msg.hasRecording = c.has_recording ?? false
      msg.transcript = c.transcription_text || undefined

      // Overlay structured fields onto regex-parsed call/voicemail data
      if (msg.call) {
        if (c.call_from) msg.call.callFrom = c.call_from
        if (c.call_to) msg.call.callTo = c.call_to
        if (c.call_duration) msg.call.duration = formatCallDuration(c.call_duration)
        if (c.answered_by_name) msg.call.agentName = c.answered_by_name
        if (c.call_location) msg.call.location = c.call_location
      } else if (msg.voicemail) {
        if (c.call_from) msg.voicemail.callFrom = c.call_from
        if (c.call_to) msg.voicemail.callTo = c.call_to
        if (c.call_duration) msg.voicemail.duration = formatCallDuration(c.call_duration)
        if (c.call_location) msg.voicemail.location = c.call_location
      }
    }

    return msg
  }

  function formatCallDuration(seconds: number): string {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    if (mins === 0) return `${secs}s`
    return secs ? `${mins}m ${secs}s` : `${mins}m`
  }

  async function loadComments(ticketId: string) {
    if (loadedCommentTickets.has(ticketId)) return
    loadedCommentTickets.add(ticketId)

    const { data } = await getTicketsByTicketIdComments({ path: { ticketID: ticketId } })
    if (!data) return

    const ticket = tickets.value.find((t) => t.id === ticketId)
    if (!ticket) return

    // Deduplicate comments that share the same call_id. Zendesk often creates
    // both a detailed "Inbound/Outbound call" comment and a shorter "Call to/from"
    // summary for the same call. Keep the richer one.
    const deduped = deduplicateByCallId(data)
    ticket.messages = deduped.map((c, i) => commentToMessage(c, i + 1))
  }

  function deduplicateByCallId(comments: AppTicketCommentRow[]): AppTicketCommentRow[] {
    const seen = new Map<number, number>() // call_id → index in result
    const result: AppTicketCommentRow[] = []

    for (const c of comments) {
      const ext = c as CommentRowExt
      if (!ext.call_id) {
        result.push(c)
        continue
      }

      const existing = seen.get(ext.call_id)
      if (existing === undefined) {
        seen.set(ext.call_id, result.length)
        result.push(c)
      } else {
        // Keep whichever comment is richer (has recording, or is a full
        // inbound/outbound call rather than a summary)
        const prev = result[existing] as CommentRowExt
        const prevIsDetailed = /^(Inbound|Outbound) call/.test(prev.body ?? "")
        const curIsDetailed = /^(Inbound|Outbound) call/.test(ext.body ?? "")

        if (!prevIsDetailed && curIsDetailed) {
          // Current is richer — replace, carrying over recording/transcript if missing
          if (!ext.has_recording && prev.has_recording) ext.has_recording = prev.has_recording
          if (!ext.transcription_text && prev.transcription_text) ext.transcription_text = prev.transcription_text
          result[existing] = c
        } else {
          // Previous is richer — carry over any data the prev lacks
          if (!prev.has_recording && ext.has_recording) prev.has_recording = ext.has_recording
          if (!prev.transcription_text && ext.transcription_text) prev.transcription_text = ext.transcription_text
        }
      }
    }

    return result
  }

  // Auto-load on first store access
  loadTickets()

  return {
    activeFilterCount,
    addTag,
    archiveTicket,
    clearFilters,
    deleteTicket,
    filterAssignees,
    filterKeyword,
    filterStatuses,
    filteredTickets,
    hudLongestWait,
    hudOpen,
    hudResolvedToday,
    loadComments,
    loadTickets,
    markRead,
    openTickets,
    reloadTickets,
    removeTag,
    resolveTicket,
    resolvedToday,
    sendReply,
    setAssignee,
    setStatus,
    setTemperature,
    sortBy,
    sortedTickets,
    tickets,
    toggleStar,
    uniqueAssignees,
    updateNotes,
  }
})
