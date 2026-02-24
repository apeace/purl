export const STATUS_COLORS: Record<string, string> = {
  new: "#38bdf8",
  open: "#60a5fa",
  pending: "#a855f7",
  escalated: "#f97316",
  solved: "#34d399",
  closed: "#94a3b8",
}

// { bg: "rgba(..., 0.12)", text: "#hex" } for each status
export const STATUS_PILL: Record<string, { bg: string; text: string }> = Object.fromEntries(
  Object.entries(STATUS_COLORS).map(([k, hex]) => {
    const r = parseInt(hex.slice(1, 3), 16)
    const g = parseInt(hex.slice(3, 5), 16)
    const b = parseInt(hex.slice(5, 7), 16)
    return [k, { bg: `rgba(${r}, ${g}, ${b}, 0.12)`, text: hex }]
  })
)

export const STATUS_LIST = [
  { value: "new", label: "New", color: STATUS_COLORS.new },
  { value: "open", label: "Open", color: STATUS_COLORS.open },
  { value: "pending", label: "Pending", color: STATUS_COLORS.pending },
  { value: "escalated", label: "Escalated", color: STATUS_COLORS.escalated },
  { value: "solved", label: "Solved", color: STATUS_COLORS.solved },
  { value: "closed", label: "Closed", color: STATUS_COLORS.closed },
]
