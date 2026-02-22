import { computed, reactive, ref } from "vue"

// ── Helpers ──────────────────────────────────────────────

const AVATAR_COLORS = [
  "#6366f1", "#ec4899", "#34d399", "#f59e0b",
  "#3b82f6", "#a855f7", "#ef4444", "#14b8a6",
]

function avatarColor(name) {
  let hash = 0
  for (const ch of name) hash = (hash * 31 + ch.charCodeAt(0)) & 0xffff
  return AVATAR_COLORS[hash % AVATAR_COLORS.length]
}

function formatWait(createdAt) {
  const mins = Math.floor((Date.now() - new Date(createdAt).getTime()) / 60000)
  if (mins < 60) return `${mins}m`
  const hrs = Math.floor(mins / 60)
  const rem = mins % 60
  return rem ? `${hrs}h ${rem}m` : `${hrs}h`
}

function formatTime(createdAt) {
  return new Date(createdAt).toLocaleTimeString([], { hour: "numeric", minute: "2-digit" })
}

function toTicket(t) {
  return {
    id: t.id,
    name: t.reporter_name,
    company: "",
    ticketId: `#${t.id.slice(0, 6).toUpperCase()}`,
    subject: t.title,
    priority: t.priority,
    wait: formatWait(t.created_at),
    avatarColor: avatarColor(t.reporter_name),
    status: t.status,
    read: false,
    starred: false,
    labels: [],
    time: formatTime(t.created_at),
    email: "",
    phone: "",
    subscription: { status: "active", id: "", plan: "" },
    tags: [],
    temperature: "warm",
    assignee: t.assignee_name ?? "Unassigned",
    notes: "",
    messages: t.description
      ? [{ id: 1, from: "customer", channel: "email", time: formatWait(t.created_at), text: t.description }]
      : [],
    ticketHistory: [
      { time: formatWait(t.created_at), event: "Ticket created" },
    ],
    subscriberHistory: [],
  }
}

// ── Module-level state (shared across all consumers) ────

const tickets = ref([])

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

const resolvedToday = ref(8)

// ── Filters ─────────────────────────────────────────────

const filterKeyword = ref("")
const filterPriorities = reactive(new Set())
const filterAssignees = reactive(new Set())
const filterStatuses = reactive(new Set())

const filteredTickets = computed(() => {
  let result = tickets.value

  const kw = filterKeyword.value.trim().toLowerCase()
  if (kw) {
    result = result.filter((t) => {
      const firstMsg = t.messages[0]?.text ?? ""
      return t.subject.toLowerCase().includes(kw) || firstMsg.toLowerCase().includes(kw)
    })
  }

  if (filterPriorities.size) {
    result = result.filter((t) => filterPriorities.has(t.priority))
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
  if (filterPriorities.size) n++
  if (filterAssignees.size) n++
  if (filterStatuses.size) n++
  return n
})

const uniqueAssignees = computed(() =>
  [...new Set(tickets.value.map((t) => t.assignee))].sort()
)

function clearFilters() {
  filterKeyword.value = ""
  filterPriorities.clear()
  filterAssignees.clear()
  filterStatuses.clear()
}

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

// ── Helpers ─────────────────────────────────────────────

function parseWait(str) {
  let mins = 0
  const d = str.match(/(\d+)d/)
  const h = str.match(/(\d+)h/)
  const m = str.match(/(\d+)m/)
  if (d) mins += parseInt(d[1]) * 1440
  if (h) mins += parseInt(h[1]) * 60
  if (m) mins += parseInt(m[1])
  return mins
}

// ── Mutations ───────────────────────────────────────────

function resolveTicket(id) {
  const ticket = tickets.value.find((t) => t.id === id)
  if (ticket) {
    ticket.status = "solved"
    ticket.read = true
    resolvedToday.value++
  }
}

function archiveTicket(id) {
  const ticket = tickets.value.find((t) => t.id === id)
  if (ticket) ticket.status = "closed"
}

function deleteTicket(id) {
  const ticket = tickets.value.find((t) => t.id === id)
  if (ticket) ticket.status = "closed"
}

function markRead(id) {
  const ticket = tickets.value.find((t) => t.id === id)
  if (ticket) ticket.read = true
}

function toggleStar(id) {
  const ticket = tickets.value.find((t) => t.id === id)
  if (ticket) ticket.starred = !ticket.starred
}

function sendReply(id, text) {
  const ticket = tickets.value.find((t) => t.id === id)
  if (!ticket) return
  ticket.messages.push({
    id: ticket.messages.length + 1,
    from: "agent",
    channel: "email",
    time: "just now",
    text,
  })
  ticket.read = true
}

function followAi(id) {
  const suggestion = aiSuggestions[id]
  if (!suggestion) return
  sendReply(id, suggestion.replyText)
}

function setStatus(id, status) {
  const ticket = tickets.value.find((t) => t.id === id)
  if (ticket) ticket.status = status
}

function setAssignee(id, assignee) {
  const ticket = tickets.value.find((t) => t.id === id)
  if (ticket) ticket.assignee = assignee
}

function setTemperature(id, temperature) {
  const ticket = tickets.value.find((t) => t.id === id)
  if (ticket) ticket.temperature = temperature
}

function addTag(id, tag) {
  const ticket = tickets.value.find((t) => t.id === id)
  if (ticket && !ticket.tags.includes(tag)) ticket.tags.push(tag)
}

function removeTag(id, tag) {
  const ticket = tickets.value.find((t) => t.id === id)
  if (ticket) ticket.tags = ticket.tags.filter((t) => t !== tag)
}

function updateNotes(id, text) {
  const ticket = tickets.value.find((t) => t.id === id)
  if (ticket) ticket.notes = text
}

// ── Data fetching ────────────────────────────────────────

async function loadTickets() {
  const res = await fetch("http://localhost:8080/tickets")
  const data = await res.json()
  tickets.value = data.map(toTicket)
}

// ── Public composable ───────────────────────────────────

export function useTickets() {
  return {
    activeFilterCount,
    addTag,
    aiSuggestions,
    archiveTicket,
    clearFilters,
    deleteTicket,
    filterAssignees,
    filterKeyword,
    filterPriorities,
    filterStatuses,
    filteredTickets,
    followAi,
    hudLongestWait,
    hudOpen,
    hudResolvedToday,
    loadTickets,
    markRead,
    openTickets,
    parseWait,
    removeTag,
    resolveTicket,
    resolvedToday,
    sendReply,
    setAssignee,
    setStatus,
    setTemperature,
    tickets,
    toggleStar,
    uniqueAssignees,
    updateNotes,
  }
}
