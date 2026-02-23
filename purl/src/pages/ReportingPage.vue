<template>
  <div class="reporting">

    <!-- Header -->
    <div class="rpt-header">
      <h1 class="page-title">Reporting</h1>
      <div class="range-tabs">
        <button
          v-for="r in ranges"
          :key="r"
          class="range-tab"
          :class="{ 'range-tab--active': activeRange === r }"
          @click="activeRange = r"
        >{{ r }}</button>
      </div>
    </div>

    <!-- Charts row -->
    <div class="charts-row">

      <!-- Volume trend chart -->
      <div class="chart-panel">
        <div class="chart-header">
          <span class="chart-title">Ticket Volume</span>
          <span class="chart-subtitle">{{ volumeLabels.length }}-day trend</span>
        </div>
        <div class="chart-body">
          <svg class="area-chart" :viewBox="`0 0 ${chartW} ${chartH}`" preserveAspectRatio="none">
            <defs>
              <linearGradient id="vol-grad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stop-color="#6366f1" stop-opacity="0.25" />
                <stop offset="100%" stop-color="#6366f1" stop-opacity="0" />
              </linearGradient>
            </defs>
            <!-- Grid lines -->
            <line v-for="i in 4" :key="i" :x1="0" :y1="chartH * (i / 4)" :x2="chartW" :y2="chartH * (i / 4)" class="grid-line" />
            <!-- Area + line -->
            <path :d="volumeAreaPath" fill="url(#vol-grad)" />
            <path :d="volumeLinePath" fill="none" stroke="#6366f1" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
            <!-- Dots -->
            <circle
              v-for="(pt, i) in volumePoints"
              :key="i"
              :cx="pt.x"
              :cy="pt.y"
              r="3"
              fill="#0a0e1a"
              stroke="#6366f1"
              stroke-width="2"
            />
          </svg>
          <div class="chart-x-labels">
            <span v-for="label in volumeLabels" :key="label">{{ label }}</span>
          </div>
        </div>
      </div>

      <!-- Status donut -->
      <div class="chart-panel chart-panel--narrow">
        <div class="chart-header">
          <span class="chart-title">By Status</span>
          <span class="chart-subtitle">Current open</span>
        </div>
        <div class="donut-body">
          <svg class="donut-chart" viewBox="0 0 120 120">
            <circle
              v-for="seg in donutSegments"
              :key="seg.label"
              cx="60" cy="60" r="46"
              fill="none"
              :stroke="seg.color"
              stroke-width="12"
              :stroke-dasharray="`${seg.arc} ${circumference - seg.arc}`"
              :stroke-dashoffset="-seg.offset"
              stroke-linecap="round"
              class="donut-seg"
            />
            <text x="60" y="56" text-anchor="middle" class="donut-total">{{ donutTotal }}</text>
            <text x="60" y="70" text-anchor="middle" class="donut-total-label">open</text>
          </svg>
          <div class="donut-legend">
            <div v-for="seg in donutSegments" :key="seg.label" class="legend-item">
              <span class="legend-dot" :style="{ background: seg.color }" />
              <span class="legend-label">{{ seg.label }}</span>
              <span class="legend-value">{{ seg.count }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Channel breakdown -->
      <div class="chart-panel chart-panel--narrow">
        <div class="chart-header">
          <span class="chart-title">By Channel</span>
          <span class="chart-subtitle">Ticket source</span>
        </div>
        <div class="hbar-body">
          <div v-for="ch in channels" :key="ch.label" class="hbar-row">
            <div class="hbar-label-row">
              <span class="hbar-label">{{ ch.label }}</span>
              <span class="hbar-value">{{ ch.pct }}%</span>
            </div>
            <div class="hbar-track">
              <div class="hbar-fill" :style="{ width: `${ch.pct}%`, background: ch.color }" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Response time mini charts -->
    <div class="mini-charts-row">
      <div v-for="mc in miniCharts" :key="mc.label" class="mini-chart-card">
        <div class="mini-chart-top">
          <span class="mini-chart-label">{{ mc.label }}</span>
          <span class="mini-chart-change" :class="mc.good ? 'change--good' : 'change--bad'">
            {{ mc.change > 0 ? '↑' : '↓' }} {{ Math.abs(mc.change) }}%
          </span>
        </div>
        <span class="mini-chart-value">{{ mc.value }}</span>
        <svg class="mini-spark" :viewBox="`0 0 ${miniW} ${miniH}`" preserveAspectRatio="none">
          <defs>
            <linearGradient :id="`mc-${mc.label.replace(/\s/g, '')}`" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" :stop-color="mc.color" stop-opacity="0.2" />
              <stop offset="100%" :stop-color="mc.color" stop-opacity="0" />
            </linearGradient>
          </defs>
          <path :d="sparkArea(mc.data)" :fill="`url(#mc-${mc.label.replace(/\\s/g, '')})`" />
          <path :d="sparkLine(mc.data)" fill="none" :stroke="mc.color" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
      </div>
    </div>

    <!-- Metric sections -->
    <div v-for="section in sections" :key="section.id" class="section">
      <div class="section-header">
        <component :is="section.icon" :size="13" class="section-icon" />
        <span class="section-title">{{ section.title }}</span>
        <div class="section-rule" />
      </div>
      <div class="metrics-grid">
        <div
          v-for="metric in section.metrics"
          :key="metric.name"
          class="metric-card"
        >
          <span class="metric-name">{{ metric.name }}</span>
          <span class="metric-value">{{ metric.value }}</span>
          <span class="metric-change" :class="isGood(metric) ? 'change--good' : 'change--bad'">
            {{ metric.change > 0 ? '↑' : '↓' }} {{ Math.abs(metric.change) }}%
          </span>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { Activity, AlertTriangle, BookOpen, Clock, Heart, ShieldCheck, Users } from "lucide-vue-next"
import { computed, ref } from "vue"
import { useTickets } from "../composables/useTickets"
import { STATUS_COLORS } from "../utils/colors"

const { tickets } = useTickets()

const ranges = ["7D", "30D", "90D", "1Y"]
const activeRange = ref("30D")

function isGood({ change, positiveDir }: { change: number; positiveDir: string }) {
  return (positiveDir === "down" && change < 0) || (positiveDir === "up" && change > 0)
}

// ── Volume trend chart ──────────────────────────────────

const chartW = 400
const chartH = 160

const volumeData = computed(() => {
  if (activeRange.value === "7D") return [42, 38, 55, 47, 61, 44, 52]
  if (activeRange.value === "90D") return [31, 38, 42, 35, 48, 52, 45, 58, 62, 47, 55, 51]
  if (activeRange.value === "1Y") return [28, 32, 35, 38, 42, 45, 41, 48, 52, 55, 51, 49]
  return [34, 42, 38, 55, 47, 61, 44, 52, 48, 58]
})

const volumeLabels = computed(() => {
  if (activeRange.value === "7D") return ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]
  if (activeRange.value === "90D") return ["W1", "W2", "W3", "W4", "W5", "W6", "W7", "W8", "W9", "W10", "W11", "W12"]
  if (activeRange.value === "1Y") return ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"]
  return ["D1", "D4", "D7", "D10", "D13", "D16", "D19", "D22", "D25", "D30"]
})

const volumePoints = computed(() => {
  const d = volumeData.value
  const max = Math.max(...d) * 1.1
  return d.map((v, i) => ({
    x: (i / (d.length - 1)) * chartW,
    y: chartH - (v / max) * chartH,
  }))
})

const volumeLinePath = computed(() =>
  volumePoints.value.map((p, i) => `${i === 0 ? "M" : "L"}${p.x},${p.y}`).join(" ")
)

const volumeAreaPath = computed(() =>
  `${volumeLinePath.value} L${chartW},${chartH} L0,${chartH} Z`
)

// ── Status donut ────────────────────────────────────────

const circumference = 2 * Math.PI * 46

const statusCounts = computed(() => {
  const counts: Record<string, number> = { new: 0, open: 0, pending: 0, escalated: 0 }
  for (const t of tickets.value) {
    if (counts[t.status] !== undefined) counts[t.status]++
  }
  return counts
})

const donutTotal = computed(() =>
  Object.values(statusCounts.value).reduce((a, b) => a + b, 0)
)

const donutSegments = computed(() => {
  const total = donutTotal.value || 1
  const items = [
    { label: "New", count: statusCounts.value.new, color: STATUS_COLORS.new },
    { label: "Open", count: statusCounts.value.open, color: STATUS_COLORS.open },
    { label: "Pending", count: statusCounts.value.pending, color: STATUS_COLORS.pending },
    { label: "Escalated", count: statusCounts.value.escalated, color: STATUS_COLORS.escalated },
  ]
  let offset = 0
  return items.map((item) => {
    const arc = (item.count / total) * circumference
    const seg = { ...item, arc, offset }
    offset += arc
    return seg
  })
})

// ── Channel breakdown ───────────────────────────────────

const channels = [
  { label: "Email", pct: 45, color: "#6366f1" },
  { label: "Chat", pct: 28, color: "#34d399" },
  { label: "Phone", pct: 15, color: "#f59e0b" },
  { label: "Social", pct: 8, color: "#ec4899" },
  { label: "API", pct: 4, color: "#94a3b8" },
]

// ── Mini sparkline charts ───────────────────────────────

const miniW = 120
const miniH = 32

const miniCharts = [
  { label: "First Response", value: "23m", change: -12, good: true, color: "#34d399", data: [40, 35, 38, 32, 30, 28, 26, 25, 24, 23] },
  { label: "Resolution Time", value: "4h 12m", change: -8, good: true, color: "#818cf8", data: [6.5, 6, 5.8, 5.5, 5.2, 5, 4.8, 4.5, 4.3, 4.2] },
  { label: "CSAT Score", value: "94%", change: 3, good: true, color: "#ec4899", data: [88, 89, 90, 89, 91, 92, 91, 93, 92, 94] },
  { label: "FCR Rate", value: "67%", change: 5, good: true, color: "#f59e0b", data: [58, 59, 61, 60, 63, 62, 64, 65, 66, 67] },
]

function sparkLine(data: number[]) {
  const max = Math.max(...data) * 1.1
  const min = Math.min(...data) * 0.9
  const range = max - min || 1
  return data.map((v, i) => {
    const x = (i / (data.length - 1)) * miniW
    const y = miniH - ((v - min) / range) * miniH
    return `${i === 0 ? "M" : "L"}${x},${y}`
  }).join(" ")
}

function sparkArea(data: number[]) {
  return `${sparkLine(data)} L${miniW},${miniH} L0,${miniH} Z`
}

// ── Metric sections ─────────────────────────────────────

const sections = [
  {
    id: "speed",
    title: "Speed & SLA",
    icon: Clock,
    metrics: [
      { name: "MTTR",                   value: "4h 12m",  change: -8,   positiveDir: "down" },
      { name: "FRT",                    value: "23m",     change: -12,  positiveDir: "down" },
      { name: "ART — Response",         value: "1h 45m",  change: +3,   positiveDir: "down" },
      { name: "ART — Resolution",       value: "5h 30m",  change: -5,   positiveDir: "down" },
      { name: "AHT",                    value: "18m",     change: -2,   positiveDir: "down" },
      { name: "TTFV",                   value: "45m",     change: -15,  positiveDir: "down" },
      { name: "Queue Time",             value: "8m",      change: +6,   positiveDir: "down" },
      { name: "Callback Time",          value: "12m",     change: -4,   positiveDir: "down" },
      { name: "SLA Compliance",         value: "94.2%",   change: -1.8, positiveDir: "up"   },
    ],
  },
  {
    id: "quality",
    title: "Quality & Resolution",
    icon: ShieldCheck,
    metrics: [
      { name: "FCR",                    value: "67%",     change: +3,   positiveDir: "up"   },
      { name: "Re-opened Rate",         value: "4.2%",    change: -0.8, positiveDir: "down" },
      { name: "Escalation Rate",        value: "8.1%",    change: +1.2, positiveDir: "down" },
      { name: "Transfer Rate",          value: "11.3%",   change: -2.4, positiveDir: "down" },
      { name: "Backlog Volume",         value: "127",     change: +12,  positiveDir: "down" },
      { name: "Backlog Age",            value: "2d 14h",  change: +8,   positiveDir: "down" },
      { name: "One-touch Resolution",   value: "41%",     change: +4,   positiveDir: "up"   },
      { name: "Replies / Resolution",   value: "2.8",     change: -11,  positiveDir: "down" },
    ],
  },
  {
    id: "agent",
    title: "Agent Performance",
    icon: Users,
    metrics: [
      { name: "Tickets / Agent / Day",  value: "22.4",    change: +3,   positiveDir: "up"   },
      { name: "Agent Utilization",      value: "78%",     change: +2,   positiveDir: "up"   },
      { name: "Agent CSAT",             value: "4.6 / 5", change: +4,   positiveDir: "up"   },
      { name: "Agent FCR",              value: "71%",     change: +5,   positiveDir: "up"   },
      { name: "Response Time",          value: "1h 12m",  change: -8,   positiveDir: "down" },
      { name: "Q/A Score",              value: "88.4%",   change: +1.6, positiveDir: "up"   },
      { name: "Schedule Adherence",     value: "92%",     change: -3,   positiveDir: "up"   },
    ],
  },
  {
    id: "ops",
    title: "Operational",
    icon: Activity,
    metrics: [
      { name: "Volume",                 value: "1,847",   change: +12,  positiveDir: "down" },
      { name: "Volume / Hour",          value: "7.8",     change: +9,   positiveDir: "down" },
      { name: "Peak Hour",              value: "14 / hr", change: +6,   positiveDir: "down" },
      { name: "Cost per Ticket",        value: "$8.20",   change: -4,   positiveDir: "down" },
      { name: "Capacity Utilization",   value: "83%",     change: +5,   positiveDir: "down" },
    ],
  },
  {
    id: "outage",
    title: "Incidents",
    icon: AlertTriangle,
    metrics: [
      { name: "Time to Acknowledge",       value: "4m 30s",  change: -15,  positiveDir: "down" },
      { name: "Time to Notify",            value: "12m",     change: -8,   positiveDir: "down" },
      { name: "Outage Volume",             value: "243",     change: -22,  positiveDir: "down" },
      { name: "Outage CSAT",               value: "3.2 / 5", change: +14,  positiveDir: "up"   },
      { name: "MTTR (Outage)",             value: "1h 47m",  change: -22,  positiveDir: "down" },
    ],
  },
  {
    id: "impact",
    title: "Customer Impact",
    icon: Heart,
    metrics: [
      { name: "Churn After Contact",          value: "2.1%",    change: -19,  positiveDir: "down" },
      { name: "Retention After Negative",     value: "68%",     change: +3,   positiveDir: "up"   },
      { name: "Support-influenced Churn",     value: "0.8%",    change: -20,  positiveDir: "down" },
      { name: "Expansion After Support",      value: "12%",     change: +2,   positiveDir: "up"   },
      { name: "Review Sentiment",             value: "4.4 / 5", change: +7,   positiveDir: "up"   },
      { name: "Detractor Recovery",           value: "34%",     change: +5,   positiveDir: "up"   },
    ],
  },
  {
    id: "selfservice",
    title: "Self-service & Deflection",
    icon: BookOpen,
    metrics: [
      { name: "KB Views",                value: "12,847",  change: +18,  positiveDir: "up"   },
      { name: "Article Helpfulness",     value: "76%",     change: +4,   positiveDir: "up"   },
      { name: "Deflection Rate",         value: "31%",     change: +6,   positiveDir: "up"   },
      { name: "Bot Resolution Rate",     value: "24%",     change: +8,   positiveDir: "up"   },
      { name: "Search Success Rate",     value: "68%",     change: +3,   positiveDir: "up"   },
    ],
  },
]
</script>

<style scoped>
/* ── Shell ───────────────────────────────────────────────── */

.reporting {
  max-width: 1200px;
}

/* ── Header ──────────────────────────────────────────────── */

.rpt-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 28px;
}

.page-title {
  font-size: 22px;
  font-weight: 700;
  color: #f1f5f9;
  letter-spacing: -0.02em;
}

.range-tabs {
  display: flex;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 10px;
  padding: 3px;
  gap: 2px;
}

.range-tab {
  padding: 5px 14px;
  border: none;
  border-radius: 7px;
  background: transparent;
  color: rgba(148, 163, 184, 0.5);
  font-size: 12px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.range-tab:hover {
  color: #94a3b8;
}

.range-tab--active {
  background: rgba(99, 102, 241, 0.2);
  color: #c7d2fe;
}

/* ── Charts row ──────────────────────────────────────────── */

.charts-row {
  display: grid;
  grid-template-columns: 1fr;
  gap: 14px;
  margin-bottom: 20px;
}

@media (min-width: 768px) {
  .charts-row {
    grid-template-columns: 1.4fr 1fr 1fr;
  }
}

.chart-panel {
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 16px;
  padding: 20px;
}

.chart-header {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 16px;
}

.chart-title {
  font-size: 14px;
  font-weight: 600;
  color: #e2e8f0;
}

.chart-subtitle {
  font-size: 11px;
  color: rgba(148, 163, 184, 0.4);
}

/* ── Area chart ──────────────────────────────────────────── */

.chart-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.area-chart {
  width: 100%;
  height: 140px;
}

.grid-line {
  stroke: rgba(255, 255, 255, 0.04);
  stroke-width: 1;
}

.chart-x-labels {
  display: flex;
  justify-content: space-between;
  font-size: 10px;
  color: rgba(148, 163, 184, 0.35);
}

/* ── Donut chart ─────────────────────────────────────────── */

.donut-body {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
}

.donut-chart {
  width: 120px;
  height: 120px;
}

.donut-seg {
  transition: stroke-dasharray 0.6s cubic-bezier(0.16, 1, 0.3, 1);
}

.donut-total {
  font-size: 22px;
  font-weight: 700;
  fill: #f1f5f9;
}

.donut-total-label {
  font-size: 10px;
  fill: rgba(148, 163, 184, 0.5);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.donut-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 14px;
  justify-content: center;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 5px;
}

.legend-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.legend-label {
  font-size: 11px;
  color: rgba(148, 163, 184, 0.6);
}

.legend-value {
  font-size: 11px;
  font-weight: 700;
  color: #e2e8f0;
}

/* ── Horizontal bar chart ────────────────────────────────── */

.hbar-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.hbar-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.hbar-label-row {
  display: flex;
  justify-content: space-between;
}

.hbar-label {
  font-size: 12px;
  font-weight: 500;
  color: #94a3b8;
}

.hbar-value {
  font-size: 12px;
  font-weight: 700;
  color: #e2e8f0;
}

.hbar-track {
  height: 6px;
  border-radius: 3px;
  background: rgba(255, 255, 255, 0.04);
}

.hbar-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 0.8s cubic-bezier(0.16, 1, 0.3, 1);
}

/* ── Mini chart cards ────────────────────────────────────── */

.mini-charts-row {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 14px;
  margin-bottom: 32px;
}

.mini-chart-card {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 14px;
  padding: 16px 18px 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  transition: background 0.15s, border-color 0.15s;
}

.mini-chart-card:hover {
  background: rgba(255, 255, 255, 0.05);
  border-color: rgba(255, 255, 255, 0.1);
}

.mini-chart-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.mini-chart-label {
  font-size: 11px;
  font-weight: 500;
  color: rgba(148, 163, 184, 0.6);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.mini-chart-change {
  font-size: 11px;
  font-weight: 600;
}

.mini-chart-value {
  font-size: 24px;
  font-weight: 700;
  color: #f1f5f9;
  letter-spacing: -0.02em;
  line-height: 1;
}

.mini-spark {
  width: 100%;
  height: 32px;
  margin-top: 8px;
}

/* ── Shared change colors ────────────────────────────────── */

.change--good {
  color: #34d399;
}

.change--bad {
  color: #f87171;
}

/* ── Section ─────────────────────────────────────────────── */

.section {
  margin-bottom: 36px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 14px;
}

.section-icon {
  color: rgba(148, 163, 184, 0.4);
  flex-shrink: 0;
}

.section-title {
  font-size: 11px;
  font-weight: 700;
  color: rgba(148, 163, 184, 0.5);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  white-space: nowrap;
}

.section-rule {
  flex: 1;
  height: 1px;
  background: rgba(255, 255, 255, 0.05);
}

/* ── Metrics grid ────────────────────────────────────────── */

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 10px;
}

.metric-card {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 12px;
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  transition: background 0.15s, border-color 0.15s;
}

.metric-card:hover {
  background: rgba(255, 255, 255, 0.05);
  border-color: rgba(255, 255, 255, 0.1);
}

.metric-name {
  font-size: 11px;
  font-weight: 500;
  color: rgba(148, 163, 184, 0.55);
  line-height: 1.3;
}

.metric-value {
  font-size: 22px;
  font-weight: 700;
  color: #f1f5f9;
  letter-spacing: -0.02em;
  line-height: 1;
  font-variant-numeric: tabular-nums;
}

.metric-change {
  font-size: 11px;
  font-weight: 600;
}

/* ── Mobile ──────────────────────────────────────────────── */

@media (max-width: 767px) {
  .metrics-grid {
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 8px;
  }

  .metric-value {
    font-size: 18px;
  }

  .mini-charts-row {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
