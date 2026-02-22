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

    <!-- Sections -->
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
          <span class="metric-change" :class="isGood(metric) ? 'metric-change--good' : 'metric-change--bad'">
            {{ metric.change > 0 ? '↑' : '↓' }} {{ Math.abs(metric.change) }}%
          </span>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import { Activity, AlertTriangle, BookOpen, Clock, Heart, ShieldCheck, Users } from "lucide-vue-next"
import { ref } from "vue"

const ranges = ["7D", "30D", "90D", "1Y"]
const activeRange = ref("30D")

// positiveDir: "down" = lower is better (response times, error rates)
//              "up"   = higher is better (FCR, CSAT, retention)
function isGood({ change, positiveDir }) {
  return (positiveDir === "down" && change < 0) || (positiveDir === "up" && change > 0)
}

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
      { name: "Re-opened Ticket Rate",  value: "4.2%",    change: -0.8, positiveDir: "down" },
      { name: "Escalation Rate",        value: "8.1%",    change: +1.2, positiveDir: "down" },
      { name: "Transfer Rate",          value: "11.3%",   change: -2.4, positiveDir: "down" },
      { name: "Backlog Volume",         value: "127",     change: +12,  positiveDir: "down" },
      { name: "Backlog Age",            value: "2d 14h",  change: +8,   positiveDir: "down" },
      { name: "One-touch Resolution",   value: "41%",     change: +4,   positiveDir: "up"   },
      { name: "Replies per Resolution", value: "2.8",     change: -11,  positiveDir: "down" },
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
      { name: "Agent Response Time",    value: "1h 12m",  change: -8,   positiveDir: "down" },
      { name: "Q/A Score",              value: "88.4%",   change: +1.6, positiveDir: "up"   },
      { name: "Schedule Adherence",     value: "92%",     change: -3,   positiveDir: "up"   },
      { name: "AFK Compliance",         value: "96%",     change: +1,   positiveDir: "up"   },
    ],
  },
  {
    id: "ops",
    title: "Operational / Capacity",
    icon: Activity,
    metrics: [
      { name: "Volume",                 value: "1,847",   change: +12,  positiveDir: "down" },
      { name: "Volume / Hour",          value: "7.8",     change: +9,   positiveDir: "down" },
      { name: "Peak Hour Volume",       value: "14 / hr", change: +6,   positiveDir: "down" },
      { name: "Tickets per Customer",   value: "1.4",     change: +14,  positiveDir: "down" },
      { name: "Cost per Ticket",        value: "$8.20",   change: -4,   positiveDir: "down" },
      { name: "Staffing Ratio",         value: "1 : 22",  change: +5,   positiveDir: "down" },
      { name: "Capacity Utilization",   value: "83%",     change: +5,   positiveDir: "down" },
    ],
  },
  {
    id: "outage",
    title: "Outage / Incident",
    icon: AlertTriangle,
    metrics: [
      { name: "TTA",                       value: "4m 30s",  change: -15,  positiveDir: "down" },
      { name: "Time to Notify Customers",  value: "12m",     change: -8,   positiveDir: "down" },
      { name: "Outage Ticket Volume",      value: "243",     change: -22,  positiveDir: "down" },
      { name: "Outage CSAT",               value: "3.2 / 5", change: +14,  positiveDir: "up"   },
      { name: "MTTR (Outage)",             value: "1h 47m",  change: -22,  positiveDir: "down" },
      { name: "Comms Updates / Incident",  value: "4.2",     change: +5,   positiveDir: "up"   },
      { name: "Post-outage Ticket Volume", value: "31",      change: -18,  positiveDir: "down" },
    ],
  },
  {
    id: "impact",
    title: "Customer Impact",
    icon: Heart,
    metrics: [
      { name: "Churn After Contact",          value: "2.1%",  change: -19,  positiveDir: "down" },
      { name: "Retention After Negative",     value: "68%",   change: +3,   positiveDir: "up"   },
      { name: "Support-influenced Churn",     value: "0.8%",  change: -20,  positiveDir: "down" },
      { name: "Expansion After Support",      value: "12%",   change: +2,   positiveDir: "up"   },
      { name: "Review Conversion Rate",       value: "8.3%",  change: +1.2, positiveDir: "up"   },
      { name: "Review Sentiment",             value: "4.4 / 5", change: +7, positiveDir: "up"   },
      { name: "Detractor Recovery Rate",      value: "34%",   change: +5,   positiveDir: "up"   },
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
  margin-bottom: 32px;
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

/* ── Metric card ─────────────────────────────────────────── */

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

.metric-change--good {
  color: #34d399;
}

.metric-change--bad {
  color: #f87171;
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
}
</style>
