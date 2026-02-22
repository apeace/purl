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
    email: "sarah.lin@acmecorp.com",
    phone: "+1 (415) 555-0134",
    subscription: { status: "active", id: "SUB-4820", plan: "Enterprise" },
    tags: ["v2.4", "deploy-issue"],
    temperature: "hot",
    assignee: "Alex Chen",
    notes: "Key contact at Acme — manages their whole engineering org. Very responsive, usually prefers quick email replies.",
    messages: [
      { id: 1, from: "customer", channel: "email", time: "2h ago", text: "Hi, after the latest update we can't access the admin panel. It just shows a blank screen. We've tried clearing the cache and restarting the server but nothing works. This is blocking our entire team." },
      { id: 2, from: "agent", channel: "email", time: "1h 45m ago", text: "Hi Sarah, thanks for reaching out — sorry to hear you're blocked. Could you share which browser and OS version you're on? And do you see any errors in the browser console?" },
      { id: 3, from: "customer", channel: "email", time: "45m ago", text: "Chrome 121 on Windows 11. The console shows: 'Error: Failed to load config.json — 404'. We have 15 people waiting on this, please help ASAP." },
    ],
    ticketHistory: [
      { time: "2h 15m ago", event: "Ticket created via email" },
      { time: "2h ago", event: "Auto-assigned to Alex Chen" },
      { time: "1h 45m ago", event: "Agent replied via email" },
      { time: "45m ago", event: "Customer replied" },
      { time: "30m ago", event: "Priority escalated to high" },
    ],
    subscriberHistory: [
      { ticketId: "#1780", subject: "SSO login failing intermittently", status: "solved", date: "Feb 10" },
      { ticketId: "#1623", subject: "Bulk import timing out", status: "solved", date: "Jan 18" },
      { ticketId: "#1501", subject: "Onboarding help — team setup", status: "closed", date: "Dec 3" },
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
    email: "mike.t@dataflow.io",
    phone: "+1 (628) 555-0198",
    subscription: { status: "active", id: "SUB-3911", plan: "Business" },
    tags: ["csv", "export-bug"],
    temperature: "warm",
    assignee: "Alex Chen",
    notes: "",
    messages: [
      { id: 1, from: "customer", channel: "email", time: "1h 20m ago", text: "The CSV export button on the Reports page doesn't do anything when I click it. No download, no error — just nothing. Started happening this morning." },
    ],
    ticketHistory: [
      { time: "1h 20m ago", event: "Ticket created via email" },
      { time: "1h ago", event: "Auto-assigned to Alex Chen" },
      { time: "40m ago", event: "Escalated to engineering" },
    ],
    subscriberHistory: [
      { ticketId: "#1710", subject: "Dashboard widget not loading", status: "solved", date: "Feb 1" },
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
    email: "priya@orionlabs.co",
    phone: "+1 (510) 555-0276",
    subscription: { status: "active", id: "SUB-5102", plan: "Enterprise" },
    tags: ["billing", "invoice"],
    temperature: "warm",
    assignee: "Alex Chen",
    notes: "Priya is the billing admin. Always CC finance@orionlabs.co on billing threads.",
    messages: [
      { id: 1, from: "customer", channel: "email", time: "45m ago", text: "Our March invoice shows $2,400 but we're on the $1,800/mo plan. We haven't added any seats or changed our plan. Can you investigate?" },
    ],
    ticketHistory: [
      { time: "45m ago", event: "Ticket created via email" },
      { time: "40m ago", event: "Auto-assigned to Alex Chen" },
      { time: "35m ago", event: "Tagged as billing" },
    ],
    subscriberHistory: [
      { ticketId: "#1690", subject: "Need invoice for tax filing", status: "solved", date: "Jan 28" },
      { ticketId: "#1445", subject: "Upgrade to Enterprise plan", status: "closed", date: "Nov 15" },
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
    email: "jkim@nexgen.com",
    phone: "+1 (650) 555-0312",
    subscription: { status: "active", id: "SUB-4455", plan: "Business" },
    tags: ["api", "rate-limit"],
    temperature: "hot",
    assignee: "Alex Chen",
    notes: "",
    messages: [
      { id: 1, from: "customer", channel: "sms", time: "30m ago", text: "We're getting 429 errors on the /events endpoint even though our dashboard shows we're only at 40% of our rate limit quota. This started about an hour ago." },
    ],
    ticketHistory: [
      { time: "30m ago", event: "Ticket created via SMS" },
      { time: "28m ago", event: "Auto-assigned to Alex Chen" },
      { time: "25m ago", event: "Priority set to high" },
    ],
    subscriberHistory: [
      { ticketId: "#1802", subject: "API key rotation question", status: "solved", date: "Feb 14" },
      { ticketId: "#1598", subject: "Webhook delivery delays", status: "solved", date: "Jan 7" },
      { ticketId: "#1390", subject: "Sandbox environment setup", status: "closed", date: "Oct 20" },
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
    email: "aisha.j@vertexco.com",
    phone: "+1 (212) 555-0187",
    subscription: { status: "active", id: "SUB-3680", plan: "Pro" },
    tags: ["feature-request"],
    temperature: "cool",
    assignee: "Unassigned",
    notes: "",
    messages: [
      { id: 1, from: "customer", channel: "email", time: "3h ago", text: "Hi team, we'd love to be able to bulk export all our data in one click. Currently we have to go page by page which takes forever for us." },
    ],
    ticketHistory: [
      { time: "3h 10m ago", event: "Ticket created via email" },
      { time: "3h ago", event: "Tagged as feature-request" },
      { time: "2h ago", event: "Status set to pending" },
    ],
    subscriberHistory: [],
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
    email: "carlos.m@helixltd.com",
    phone: "+1 (305) 555-0245",
    subscription: { status: "active", id: "SUB-2910", plan: "Business" },
    tags: ["auth", "password-reset"],
    temperature: "cool",
    assignee: "Alex Chen",
    notes: "",
    messages: [
      { id: 1, from: "customer", channel: "email", time: "5h ago", text: "The password reset email never arrives. We've checked spam folders." },
      { id: 2, from: "agent", channel: "email", time: "4h ago", text: "Hi Carlos, I've fixed the email routing issue on your account. Please try again now." },
      { id: 3, from: "customer", channel: "email", time: "4h ago", text: "Thanks for the quick fix! The password reset is working perfectly now. I'll let the rest of my team know. Really appreciate the fast response." },
    ],
    ticketHistory: [
      { time: "5h ago", event: "Ticket created via email" },
      { time: "4h 30m ago", event: "Assigned to Alex Chen" },
      { time: "4h ago", event: "Agent replied" },
      { time: "4h ago", event: "Customer confirmed fix" },
      { time: "3h 30m ago", event: "Resolved by Alex Chen" },
    ],
    subscriberHistory: [
      { ticketId: "#1560", subject: "Two-factor authentication help", status: "solved", date: "Jan 2" },
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
    email: "emma.w@stratosinc.com",
    phone: "+1 (312) 555-0167",
    subscription: { status: "active", id: "SUB-5230", plan: "Pro" },
    tags: ["integration", "salesforce"],
    temperature: "warm",
    assignee: "Unassigned",
    notes: "",
    messages: [
      { id: 1, from: "customer", channel: "email", time: "1d ago", text: "We're looking to integrate Pipeline with our Salesforce setup. Is there a native connector available or would we need to use the API directly?" },
    ],
    ticketHistory: [
      { time: "1d ago", event: "Ticket created via email" },
      { time: "23h ago", event: "Status set to pending" },
    ],
    subscriberHistory: [
      { ticketId: "#1480", subject: "Custom field configuration", status: "solved", date: "Nov 22" },
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
    email: "d.park@nimbusco.io",
    phone: "+1 (408) 555-0293",
    subscription: { status: "active", id: "SUB-4100", plan: "Enterprise" },
    tags: ["webhook", "api"],
    temperature: "hot",
    assignee: "Unassigned",
    notes: "Enterprise account — high value. David is their CTO, prefers phone calls.",
    messages: [
      { id: 1, from: "customer", channel: "phone", time: "1d 2h ago", text: "Called in about webhook endpoint — registered and URL is correct, but events aren't arriving. No errors in logs. Broke overnight. Wants urgent callback." },
      { id: 2, from: "customer", channel: "email", time: "1d ago", text: "Our webhook endpoint is registered and the URL is correct, but events just aren't arriving. No errors in logs either. This broke overnight." },
    ],
    ticketHistory: [
      { time: "1d 2h ago", event: "Ticket created via phone call" },
      { time: "1d ago", event: "Customer followed up via email" },
    ],
    subscriberHistory: [
      { ticketId: "#1720", subject: "Webhook payload format change request", status: "solved", date: "Feb 5" },
      { ticketId: "#1550", subject: "API authentication errors", status: "solved", date: "Dec 28" },
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
    email: "mei.zhang@luminacorp.com",
    phone: "+1 (617) 555-0321",
    subscription: { status: "active", id: "SUB-3450", plan: "Enterprise" },
    tags: ["gdpr", "compliance"],
    temperature: "cool",
    assignee: "Unassigned",
    notes: "",
    messages: [
      { id: 1, from: "customer", channel: "email", time: "2d ago", text: "Hi, one of our customers has submitted a formal data export request under GDPR Article 20. What's the process for fulfilling this through Pipeline?" },
    ],
    ticketHistory: [
      { time: "2d ago", event: "Ticket created via email" },
      { time: "2d ago", event: "Status set to pending" },
    ],
    subscriberHistory: [
      { ticketId: "#1380", subject: "Data retention policy question", status: "solved", date: "Oct 12" },
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
    email: "raj@cobaltsystems.com",
    phone: "+1 (503) 555-0188",
    subscription: { status: "active", id: "SUB-4780", plan: "Pro" },
    tags: ["ui", "branding"],
    temperature: "cool",
    assignee: "Unassigned",
    notes: "",
    messages: [
      { id: 1, from: "customer", channel: "email", time: "2d ago", text: "Since the last update, the login screen is displaying our old company logo instead of the one we uploaded in brand settings last month." },
    ],
    ticketHistory: [
      { time: "2d ago", event: "Ticket created via email" },
    ],
    subscriberHistory: [
      { ticketId: "#1650", subject: "Brand color customization", status: "solved", date: "Jan 20" },
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
    email: "s.turner@apexglobal.com",
    phone: "+1 (720) 555-0144",
    subscription: { status: "active", id: "SUB-5500", plan: "Business" },
    tags: ["upsell"],
    temperature: "warm",
    assignee: "Unassigned",
    notes: "",
    messages: [
      { id: 1, from: "customer", channel: "email", time: "3d ago", text: "We've been really happy with Pipeline and want to discuss upgrading to the Enterprise plan. Can someone from your team reach out to us?" },
    ],
    ticketHistory: [
      { time: "3d ago", event: "Ticket created via email" },
      { time: "3d ago", event: "Status set to pending" },
    ],
    subscriberHistory: [
      { ticketId: "#1520", subject: "Team seat addition", status: "solved", date: "Dec 15" },
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
    email: "a.rivera@forgelabs.dev",
    phone: "+1 (206) 555-0257",
    subscription: { status: "trial", id: "SUB-5801", plan: "Enterprise Trial" },
    tags: ["sso", "saml", "config"],
    temperature: "warm",
    assignee: "Alex Chen",
    notes: "",
    messages: [
      { id: 1, from: "customer", channel: "email", time: "4d ago", text: "We configured our SAML SSO settings but they don't seem to save. Every time we navigate away and come back the fields are blank again." },
      { id: 2, from: "agent", channel: "email", time: "3d ago", text: "Hi Alex, thanks for reporting this. I'm looking into the SSO persistence issue now — can you confirm which SAML provider you're using?" },
      { id: 3, from: "customer", channel: "sms", time: "3d ago", text: "We're using Okta. Happy to do a screenshare if that helps debug it." },
    ],
    ticketHistory: [
      { time: "4d ago", event: "Ticket created via email" },
      { time: "4d ago", event: "Assigned to Alex Chen" },
      { time: "3d ago", event: "Agent replied via email" },
      { time: "3d ago", event: "Customer replied via SMS" },
    ],
    subscriberHistory: [],
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
    setStatus,
    setAssignee,
    setTemperature,
    addTag,
    removeTag,
    updateNotes,
  }
}
