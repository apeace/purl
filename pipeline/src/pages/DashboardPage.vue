<template>
  <div>
    <div class="page-header">
      <h1 class="page-title">Good {{ greeting }}, Alex</h1>
      <p class="page-subtitle">Here's your support queue overview.</p>
    </div>

    <div class="stats-grid">
      <StatCard label="Open Tickets" :value="String(hudOpen)" :change="openChange" color="#818cf8" :data="openSparkline" :delay="100" />
      <StatCard label="Resolved Today" :value="String(hudResolvedToday)" :change="18" color="#34d399" :data="[3,5,4,6,5,7,8,6,7,9,8,10]" :delay="200" />
      <StatCard label="Avg Response" value="24m" :change="12" color="#f59e0b" :data="[40,38,35,32,30,28,30,27,26,25,24,24]" :delay="300" />
      <StatCard label="CSAT Score" value="94%" :change="3" color="#ec4899" :data="[88,89,90,89,91,92,91,93,92,93,94,94]" :delay="400" />
    </div>

    <div class="panels-grid">
      <div class="panel">
        <div class="panel-header">
          <span class="panel-title">Recent Activity</span>
          <span class="panel-link">View all</span>
        </div>
        <ActivityRow name="Alex Chen" action="Resolved #A3F2 — Config 404 bug" time="2m ago" color="#34d399" :delay="500" />
        <ActivityRow name="Sarah Lin" action="Escalated #B71C — Billing dispute" time="18m ago" color="#f97316" :delay="600" />
        <ActivityRow name="Mike Torres" action="Replied to #C904 — CSV export issue" time="45m ago" color="#6366f1" :delay="700" />
        <ActivityRow name="Priya Patel" action="Assigned #D5E8 to Alex Chen" time="1h ago" color="#a855f7" :delay="800" />
        <ActivityRow name="James Kim" action="Merged duplicate tickets #E1A0, #E1A3" time="2h ago" color="#ec4899" :delay="900" />
      </div>

      <div class="panel">
        <div class="panel-header">
          <span class="panel-title">Tickets by Priority</span>
          <span class="panel-link">View inbox</span>
        </div>
        <div
          v-for="(row, i) in priorityBreakdown"
          :key="row.label"
          class="bar-row"
          :class="{ 'bar-row--last': i === priorityBreakdown.length - 1 }"
        >
          <div class="bar-row-top">
            <div class="bar-label-wrap">
              <span class="bar-dot" :style="{ background: row.color }" />
              <span class="bar-name">{{ row.label }}</span>
            </div>
            <span class="bar-val" :style="{ color: row.color }">{{ row.count }}</span>
          </div>
          <div class="bar-wrap">
            <div class="bar-track">
              <div
                class="bar-fill"
                :style="{ width: `${row.pct}%`, background: `linear-gradient(90deg, ${row.color}, ${row.color}88)` }"
              />
            </div>
            <span class="bar-pct">{{ row.pct }}%</span>
          </div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <span class="panel-title">Team Leaderboard</span>
          <span class="panel-link">This week</span>
        </div>
        <div
          v-for="(agent, i) in teamStats"
          :key="agent.name"
          class="agent-row"
          :class="{ 'agent-row--last': i === teamStats.length - 1 }"
        >
          <div class="agent-avatar" :style="{ background: agent.color }">{{ agent.name[0] }}</div>
          <div class="agent-info">
            <span class="agent-name">{{ agent.name }}</span>
            <span class="agent-meta">{{ agent.resolved }} resolved · {{ agent.avgTime }} avg</span>
          </div>
          <div class="agent-score">
            <span class="agent-csat">{{ agent.csat }}%</span>
            <span class="agent-csat-label">CSAT</span>
          </div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <span class="panel-title">SLA Status</span>
          <span class="panel-link">View details</span>
        </div>
        <div
          v-for="(sla, i) in slaMetrics"
          :key="sla.label"
          class="bar-row"
          :class="{ 'bar-row--last': i === slaMetrics.length - 1 }"
        >
          <div class="bar-row-top">
            <span class="bar-name">{{ sla.label }}</span>
            <span class="bar-val" :style="{ color: sla.color }">{{ sla.value }}</span>
          </div>
          <div class="bar-wrap">
            <div class="bar-track">
              <div
                class="bar-fill"
                :style="{ width: `${sla.pct}%`, background: `linear-gradient(90deg, ${sla.color}, ${sla.color}88)` }"
              />
            </div>
            <span class="bar-pct">{{ sla.pct }}%</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from "vue"
import ActivityRow from "../components/ActivityRow.vue"
import StatCard from "../components/StatCard.vue"
import { useTickets } from "../composables/useTickets.js"

const { hudOpen, hudResolvedToday, tickets } = useTickets()

const hour = new Date().getHours()
const greeting = hour < 12 ? "morning" : hour < 18 ? "afternoon" : "evening"

// Sparkline from live open ticket counts (simulated trailing data + current)
const openSparkline = computed(() => {
  const current = hudOpen.value
  return [current + 4, current + 3, current + 5, current + 2, current + 3, current + 1, current + 2, current, current + 1, current - 1, current, current]
})

const openChange = computed(() => {
  const current = hudOpen.value
  if (!current) return 0
  return -Math.round(((current + 4 - current) / (current + 4)) * 100)
})

const priorityBreakdown = computed(() => {
  const open = tickets.value.filter((t) => t.status !== "closed" && t.status !== "solved")
  const total = open.length || 1
  const counts = { urgent: 0, high: 0, medium: 0, low: 0 }
  for (const t of open) {
    if (counts[t.priority] !== undefined) counts[t.priority]++
  }
  return [
    { label: "Urgent", count: counts.urgent, pct: Math.round((counts.urgent / total) * 100), color: "#ef4444" },
    { label: "High", count: counts.high, pct: Math.round((counts.high / total) * 100), color: "#f97316" },
    { label: "Medium", count: counts.medium, pct: Math.round((counts.medium / total) * 100), color: "#f59e0b" },
    { label: "Low", count: counts.low, pct: Math.round((counts.low / total) * 100), color: "#34d399" },
  ]
})

const teamStats = [
  { name: "Alex Chen", resolved: 12, avgTime: "18m", csat: 97, color: "#6366f1" },
  { name: "Sarah Lin", resolved: 9, avgTime: "22m", csat: 95, color: "#ec4899" },
  { name: "Mike Torres", resolved: 7, avgTime: "31m", csat: 92, color: "#34d399" },
  { name: "Priya Patel", resolved: 6, avgTime: "26m", csat: 94, color: "#f59e0b" },
]

const slaMetrics = [
  { label: "First Response (<15m)", value: "92%", pct: 92, color: "#34d399" },
  { label: "Resolution (<4h)", value: "87%", pct: 87, color: "#818cf8" },
  { label: "Escalation Rate", value: "8%", pct: 8, color: "#f97316" },
  { label: "Reopen Rate", value: "3%", pct: 3, color: "#ef4444" },
]
</script>

<style scoped>
.page-header {
  margin-bottom: 24px;
}

.page-title {
  font-size: 22px;
  font-weight: 700;
  color: #f1f5f9;
  letter-spacing: -0.02em;
}

.page-subtitle {
  font-size: 14px;
  color: rgba(148, 163, 184, 0.6);
  margin-top: 4px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 14px;
  margin-bottom: 28px;
}

.panels-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
}

.panel {
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 16px;
  padding: 20px;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.panel-title {
  font-size: 14px;
  font-weight: 600;
  color: #e2e8f0;
}

.panel-link {
  font-size: 12px;
  color: #818cf8;
  cursor: pointer;
  font-weight: 500;
}

/* ── Bar rows (priority breakdown + SLA) ─────────────────── */

.bar-row {
  padding: 12px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
}

.bar-row--last {
  border-bottom: none;
}

.bar-row-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.bar-label-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}

.bar-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.bar-name {
  font-size: 13px;
  font-weight: 600;
  color: #e2e8f0;
}

.bar-val {
  font-size: 13px;
  font-weight: 700;
}

.bar-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
}

.bar-track {
  flex: 1;
  height: 4px;
  border-radius: 2px;
  background: rgba(255, 255, 255, 0.04);
}

.bar-fill {
  height: 100%;
  border-radius: 2px;
  transition: width 1s cubic-bezier(0.16, 1, 0.3, 1);
}

.bar-pct {
  font-size: 11px;
  color: rgba(148, 163, 184, 0.5);
  min-width: 28px;
}

/* ── Agent rows (team leaderboard) ───────────────────────── */

.agent-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
}

.agent-row--last {
  border-bottom: none;
}

.agent-avatar {
  width: 32px;
  height: 32px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
}

.agent-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.agent-name {
  font-size: 13px;
  font-weight: 600;
  color: #e2e8f0;
}

.agent-meta {
  font-size: 12px;
  color: rgba(148, 163, 184, 0.6);
}

.agent-score {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  flex-shrink: 0;
}

.agent-csat {
  font-size: 14px;
  font-weight: 700;
  color: #34d399;
}

.agent-csat-label {
  font-size: 10px;
  color: rgba(148, 163, 184, 0.4);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
</style>
