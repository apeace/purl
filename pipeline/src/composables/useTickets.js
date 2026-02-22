import { computed, ref } from "vue"

// ── Module-level state (shared across all consumers) ────

const tickets = ref([
  {
    id: 1,
    name: "Sarah Lin",
    company: "Acme Corp",
    ticketId: "#1842",
    subject: "Setup not working after update",
    priority: "high",
    wait: "2h 15m",
    avatarColor: "#6366f1",
    status: "open",
    read: false,
    starred: true,
    labels: ["urgent"],
    time: "2:14 PM",
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
    status: "escalated",
    read: false,
    starred: false,
    labels: ["bug"],
    time: "1:47 PM",
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
    status: "open",
    read: true,
    starred: false,
    labels: ["billing"],
    time: "12:30 PM",
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
    status: "open",
    read: false,
    starred: false,
    labels: ["urgent", "bug"],
    time: "11:15 AM",
    messages: [
      { id: 1, from: "customer", time: "30m ago", text: "We're getting 429 errors on the /events endpoint even though our dashboard shows we're only at 40% of our rate limit quota. This started about an hour ago." },
    ],
  },
  {
    id: 5,
    name: "Aisha Johnson",
    company: "Vertex Co",
    ticketId: "#1835",
    subject: "Feature request: bulk export",
    priority: "low",
    wait: "3h 10m",
    avatarColor: "#a855f7",
    status: "pending",
    read: true,
    starred: true,
    labels: ["feature"],
    time: "10:02 AM",
    messages: [
      { id: 1, from: "customer", time: "3h ago", text: "Hi team, we'd love to be able to bulk export all our data in one click. Currently we have to go page by page which takes forever for us." },
    ],
  },
  {
    id: 6,
    name: "Carlos Mendez",
    company: "Helix Ltd",
    ticketId: "#1828",
    subject: "Re: Password reset not working",
    priority: "medium",
    wait: "4h",
    avatarColor: "#06b6d4",
    status: "solved",
    read: true,
    starred: false,
    labels: [],
    time: "9:44 AM",
    messages: [
      { id: 1, from: "customer", time: "5h ago", text: "The password reset email never arrives. We've checked spam folders." },
      { id: 2, from: "agent", time: "4h ago", text: "Hi Carlos, I've fixed the email routing issue on your account. Please try again now." },
      { id: 3, from: "customer", time: "4h ago", text: "Thanks for the quick fix! The password reset is working perfectly now. I'll let the rest of my team know. Really appreciate the fast response." },
    ],
  },
  {
    id: 7,
    name: "Emma Wilson",
    company: "Stratos Inc",
    ticketId: "#1822",
    subject: "Salesforce integration question",
    priority: "low",
    wait: "1d",
    avatarColor: "#f43f5e",
    status: "pending",
    read: true,
    starred: false,
    labels: ["feature"],
    time: "Yesterday",
    messages: [
      { id: 1, from: "customer", time: "1d ago", text: "We're looking to integrate Pipeline with our Salesforce setup. Is there a native connector available or would we need to use the API directly?" },
    ],
  },
  {
    id: 8,
    name: "David Park",
    company: "Nimbus Co",
    ticketId: "#1819",
    subject: "Webhook failing silently",
    priority: "high",
    wait: "1d 2h",
    avatarColor: "#8b5cf6",
    status: "new",
    read: false,
    starred: false,
    labels: ["bug", "urgent"],
    time: "Yesterday",
    messages: [
      { id: 1, from: "customer", time: "1d ago", text: "Our webhook endpoint is registered and the URL is correct, but events just aren't arriving. No errors in logs either. This broke overnight." },
    ],
  },
  {
    id: 9,
    name: "Mei Zhang",
    company: "Lumina Corp",
    ticketId: "#1814",
    subject: "GDPR data export request",
    priority: "medium",
    wait: "2d",
    avatarColor: "#10b981",
    status: "pending",
    read: true,
    starred: false,
    labels: ["billing"],
    time: "Mon",
    messages: [
      { id: 1, from: "customer", time: "2d ago", text: "Hi, one of our customers has submitted a formal data export request under GDPR Article 20. What's the process for fulfilling this through Pipeline?" },
    ],
  },
  {
    id: 10,
    name: "Raj Patel",
    company: "Cobalt Systems",
    ticketId: "#1810",
    subject: "Login screen shows wrong logo",
    priority: "low",
    wait: "2d",
    avatarColor: "#f97316",
    status: "new",
    read: true,
    starred: false,
    labels: ["bug"],
    time: "Mon",
    messages: [
      { id: 1, from: "customer", time: "2d ago", text: "Since the last update, the login screen is displaying our old company logo instead of the one we uploaded in brand settings last month." },
    ],
  },
  {
    id: 11,
    name: "Sophie Turner",
    company: "Apex Global",
    ticketId: "#1805",
    subject: "Interested in upgrading to Enterprise",
    priority: "low",
    wait: "3d",
    avatarColor: "#6366f1",
    status: "pending",
    read: true,
    starred: true,
    labels: [],
    time: "Sun",
    messages: [
      { id: 1, from: "customer", time: "3d ago", text: "We've been really happy with Pipeline and want to discuss upgrading to the Enterprise plan. Can someone from your team reach out to us?" },
    ],
  },
  {
    id: 12,
    name: "Alex Rivera",
    company: "Forge Labs",
    ticketId: "#1800",
    subject: "SSO configuration not persisting",
    priority: "medium",
    wait: "4d",
    avatarColor: "#ec4899",
    status: "open",
    read: true,
    starred: false,
    labels: ["bug"],
    time: "Mar 18",
    messages: [
      { id: 1, from: "customer", time: "4d ago", text: "We configured our SAML SSO settings but they don't seem to save. Every time we navigate away and come back the fields are blank again." },
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

const resolvedToday = ref(8)

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

// ── Public composable ───────────────────────────────────

export function useTickets() {
  return {
    tickets,
    aiSuggestions,
    openTickets,
    hudOpen,
    hudLongestWait,
    hudResolvedToday,
    resolvedToday,
    parseWait,
    resolveTicket,
    archiveTicket,
    deleteTicket,
    markRead,
    toggleStar,
    sendReply,
    followAi,
  }
}
