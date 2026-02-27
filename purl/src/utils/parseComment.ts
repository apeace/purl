// Zendesk comment body classifier and parser
// Ported from Python script — classifies raw comment bodies into message types
// and extracts structured data from calls, voicemails, merges, etc.

export type MessageType =
  | "regular_message"
  | "outbound_call"
  | "inbound_call"
  | "call_summary"
  | "voicemail"
  | "merge_notice"
  | "web_chat"

export type CommChannel =
  | "email_inbound"
  | "email_outbound"
  | "sms_inbound"
  | "call_outbound"
  | "call_inbound"
  | "call_summary"
  | "voicemail"
  | "web_chat"
  | "web_form"
  | "public_reply"
  | "internal_note"
  | "ticket_merge"

export interface CallData {
  direction: "inbound" | "outbound"
  customerPhone?: string
  callFrom?: string
  callTo?: string
  timeOfCall?: string
  location?: string
  agentName?: string
  duration?: string
  recordingUrl?: string
}

export interface VoicemailData {
  customerPhone?: string
  callFrom?: string
  callTo?: string
  timeOfCall?: string
  location?: string
  duration?: string
  recordingUrl?: string
}

export interface MergeData {
  mergedRequestNumbers?: number[]
  mergedRequestSubject?: string
  mergedIntoRequestNumber?: number
  mergedIntoSubject?: string
  raw?: string
}

export interface ParsedComment {
  messageType: MessageType
  commChannel: CommChannel
  cleanBody: string
  call?: CallData
  voicemail?: VoicemailData
  merge?: MergeData
  webChatUserId?: string
}

// ── Classification ──────────────────────────────────────

export function classifyBody(body: string): MessageType {
  if (body.startsWith("Outbound call")) return "outbound_call"
  if (body.startsWith("Inbound call")) return "inbound_call"
  if (body.startsWith("Voicemail from")) return "voicemail"
  if (/^Call (to|from):/.test(body)) return "call_summary"
  if (/^Request #\d+/.test(body) || /\bmerged\b/i.test(body)) return "merge_notice"
  if (body.includes("Conversation with Web User")) return "web_chat"
  // Chat transcripts with (HH:MM:SS) timestamps and a bot or Web User speaker
  if (/\(\d{1,2}:\d{2}:\d{2}\)\s/.test(body) && (/\bWeb User\b/.test(body) || /\bbot\b/i.test(body))) return "web_chat"
  return "regular_message"
}

export function classifyCommChannel(channel: string, messageType: MessageType, role: string): CommChannel {
  // Structured message types take precedence — body content is the strongest signal
  if (messageType === "outbound_call") return "call_outbound"
  if (messageType === "inbound_call") return "call_inbound"
  if (messageType === "call_summary") return "call_summary"
  if (messageType === "voicemail") return "voicemail"
  if (messageType === "web_chat") return "web_chat"
  if (messageType === "merge_notice") return "ticket_merge"

  // Individual web chat sub-messages stored with channel='chat' by the backend
  if (channel === "chat") return "web_chat"

  // Use channel + role to determine direction
  if (channel === "email") return role === "agent" ? "email_outbound" : "email_inbound"
  if (channel === "web") return role === "agent" ? "public_reply" : "web_form"
  if (channel === "sms") return "sms_inbound"
  if (channel === "internal") return "internal_note"

  return "internal_note"
}

// ── Field extraction helper ─────────────────────────────

function extractField(body: string, label: string): string | undefined {
  const re = new RegExp(`^${label.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}:\\s*(.+)$`, "m")
  const m = body.match(re)
  return m ? m[1].trim() : undefined
}

// ── Body parsers ────────────────────────────────────────

function parseOutboundCall(body: string): CallData {
  const header = body.match(/^Outbound call to (.+)/m)
  const recording = body.match(/Listen to the recording:\s*(https?:\/\/\S+)/)
  return {
    direction: "outbound",
    customerPhone: header?.[1]?.trim(),
    callFrom: extractField(body, "Call from"),
    callTo: extractField(body, "Call to"),
    timeOfCall: extractField(body, "Time of call"),
    agentName: extractField(body, "Called by"),
    duration: extractField(body, "Length of phone call"),
    recordingUrl: recording?.[1],
  }
}

function parseInboundCall(body: string): CallData {
  const header = body.match(/^Inbound call from (.+)/m)
  const recording = body.match(/Listen to the recording:\s*(https?:\/\/\S+)/)
  return {
    direction: "inbound",
    customerPhone: header?.[1]?.trim(),
    callFrom: extractField(body, "Call from"),
    callTo: extractField(body, "Call to"),
    timeOfCall: extractField(body, "Time of call"),
    location: extractField(body, "Location"),
    agentName: extractField(body, "Answered by"),
    duration: extractField(body, "Length of phone call"),
    recordingUrl: recording?.[1],
  }
}

function parseVoicemail(body: string): VoicemailData {
  const header = body.match(/^Voicemail from (.+)/m)
  const recording = body.match(/Listen to the (?:voicemail|recording):\s*(https?:\/\/\S+)/)
  return {
    customerPhone: header?.[1]?.trim(),
    callFrom: extractField(body, "Call from"),
    callTo: extractField(body, "Call to"),
    timeOfCall: extractField(body, "Time of call"),
    location: extractField(body, "Location"),
    duration: extractField(body, "Length of phone call"),
    recordingUrl: recording?.[1],
  }
}

function parseCallSummary(body: string): CallData {
  const direction = body.startsWith("Call to:") ? "outbound" : "inbound"
  return {
    direction,
    callTo: extractField(body, "Call to"),
    callFrom: extractField(body, "Call from"),
    timeOfCall: extractField(body, "Time of call"),
    agentName: extractField(body, "Called by") ?? extractField(body, "Answered by"),
  }
}

function parseMergeNotice(body: string): MergeData {
  // Single: Request #34478 "Subject" was closed and merged into this request.
  const single = body.match(/^Request #(\d+)\s+"([^"]+)"\s+was closed and merged/)
  if (single) {
    return { mergedRequestNumbers: [parseInt(single[1])], mergedRequestSubject: single[2] }
  }
  // Multi: Requests #34528, #34530 were closed and merged into this request.
  const multi = body.match(/^Requests? (#[\d,\s#]+) were? closed and merged/)
  if (multi) {
    const numbers = [...multi[1].matchAll(/\d+/g)].map((m) => parseInt(m[0]))
    return { mergedRequestNumbers: numbers }
  }
  // Reverse: This request was closed and merged into request #34548 "Subject".
  const reverse = body.match(/merged into request #(\d+)\s+"([^"]+)"/)
  if (reverse) {
    return { mergedIntoRequestNumber: parseInt(reverse[1]), mergedIntoSubject: reverse[2] }
  }
  return { raw: body }
}

function parseWebChat(body: string): string | undefined {
  const m = body.match(/Conversation with Web User (\S+)/)
  return m?.[1]
}

// ── HTML stripping ──────────────────────────────────────

export function stripHtml(html: string): string {
  return html
    .replace(/<br\s*\/?>/gi, "\n")
    .replace(/<\/p>/gi, "\n")
    .replace(/<[^>]+>/g, "")
    .replace(/&amp;/g, "&")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&quot;/g, "\"")
    .replace(/&#39;/g, "'")
    .replace(/&nbsp;/g, " ")
    .replace(/\n{3,}/g, "\n\n")
    .trim()
}

// Whitelist-based HTML sanitiser — preserves structural elements (tables,
// paragraphs, lists, blockquotes) while stripping scripts, styles, and
// event handlers. Only used for rich email rendering.
export function sanitizeHtml(html: string): string {
  return html
    // Remove dangerous elements and their content
    .replace(/<(script|style|iframe|object|embed|form)\b[\s\S]*?<\/\1>/gi, "")
    .replace(/<(link|meta|input|base)\b[^>]*\/?>/gi, "")
    // Remove images (tracking pixels, logos — they won't load from our context anyway)
    .replace(/<img\b[^>]*\/?>/gi, "")
    // Strip event handlers (onclick, onerror, etc.)
    .replace(/\s+on\w+\s*=\s*(?:"[^"]*"|'[^']*'|[^\s>]+)/gi, "")
    // Strip javascript: and data: URLs in href/src attributes
    .replace(/(href|src)\s*=\s*"(?:javascript|data):[^"]*"/gi, "$1=\"\"")
    .replace(/(href|src)\s*=\s*'(?:javascript|data):[^']*'/gi, "$1=''")
    // Strip style attributes (expression() XSS vector)
    .replace(/\s+style\s*=\s*(?:"[^"]*"|'[^']*'|[^\s>]+)/gi, "")
    // Strip class attributes (avoid style injection)
    .replace(/\s+class\s*=\s*(?:"[^"]*"|'[^']*'|[^\s>]+)/gi, "")
    // Strip email layout attributes that interfere with our styling
    .replace(/\s+(?:width|height|cellpadding|cellspacing|bgcolor|align|valign|border)\s*=\s*(?:"[^"]*"|'[^']*'|[^\s>]+)/gi, "")
    // Collapse cells that contain only whitespace/&nbsp; into empty cells
    .replace(/<(td|th)([^>]*)>\s*(?:&nbsp;\s*)*<\/\1>/gi, "<$1$2></$1>")
    .trim()
}

// ── Top-level parser ────────────────────────────────────

export function parseComment(body: string, channel: string, role: string): ParsedComment {
  const cleanBody = stripHtml(body)
  const messageType = classifyBody(cleanBody)
  const commChannel = classifyCommChannel(channel, messageType, role)

  const result: ParsedComment = { messageType, commChannel, cleanBody }

  switch (messageType) {
    case "outbound_call":
      result.call = parseOutboundCall(cleanBody)
      break
    case "inbound_call":
      result.call = parseInboundCall(cleanBody)
      break
    case "call_summary":
      result.call = parseCallSummary(cleanBody)
      break
    case "voicemail":
      result.voicemail = parseVoicemail(cleanBody)
      break
    case "merge_notice":
      result.merge = parseMergeNotice(cleanBody)
      break
    case "web_chat":
      result.webChatUserId = parseWebChat(cleanBody)
      break
  }

  return result
}
