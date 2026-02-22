<template>
  <div class="health-wrap">
    <div class="health-top">
      <span class="health-title">Shift Health</span>
      <span class="health-trend" :class="trendUp ? 'health-trend--up' : 'health-trend--down'">
        <TrendingUp v-if="trendUp" :size="14" />
        <TrendingDown v-else :size="14" />
      </span>
    </div>
    <div class="health-bar-track">
      <div class="health-bar-fill" :style="{ width: `${healthScore}%`, background: healthGradient }">
        <div class="health-bar-shimmer" />
      </div>
    </div>
    <button class="health-expand" @click="expanded = !expanded">
      <ChevronDown :size="12" class="health-expand-icon" :class="{ 'health-expand-icon--open': expanded }" />
      <span>Details</span>
    </button>
    <div class="health-components" :class="{ 'health-components--open': expanded }">
      <div class="health-components-inner">
        <div v-for="c in healthComponents" :key="c.label" class="health-comp">
          <div class="health-comp-top">
            <span class="health-comp-label">{{ c.label }}</span>
          </div>
          <div class="health-comp-track">
            <div class="health-comp-fill" :style="{ width: expanded ? `${c.score}%` : '0%', background: c.color }" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
// ── Shift Health ────────────────────────────────────────
//
// Composite score = (0.30 × Backlog) + (0.30 × Clearance) + (0.25 × Response) + (0.15 × Efficiency)
//
// Backlog Score:    max(0, 100 × (1 − openTickets / (threshold × 2)))
//                   threshold = agentsOnShift × acceptablePerAgent
//                   Measures how full the queue is relative to capacity.
//
// Clearance Score:  min(100, (resolvedToday / ticketsOpenedToday) × 100)
//                   Are we resolving at least as fast as tickets come in?
//
// Response Score:   If avgFRT ≤ goal → 100
//                   Else max(0, 100 × (1 − (avgFRT − goal) / goal))
//                   Linear decay once first-response time exceeds the goal.
//
// Efficiency Score: FCR rate (first-contact resolution %)
//                   Higher = fewer follow-ups needed.
//
// Status tiers: 90–100 Excellent, 75–89 Good, 60–74 Fair, 40–59 At Risk, 0–39 Critical

import { ChevronDown, TrendingDown, TrendingUp } from "lucide-vue-next"
import { computed, ref, watch } from "vue"
import { useTickets } from "../composables/useTickets.js"

const { hudOpen, hudResolvedToday } = useTickets()

const AGENTS_ON_SHIFT = 2
const ACCEPTABLE_PER_AGENT = 15
const FRT_GOAL_MINS = 60
// Simulated shift inputs (in production these come from live telemetry)
const AVG_FRT_MINS = 45
const TICKETS_OPENED_TODAY = 25
const FCR_RATE = 78

const backlogScore = computed(() => {
  const threshold = AGENTS_ON_SHIFT * ACCEPTABLE_PER_AGENT
  return Math.round(Math.max(0, 100 * (1 - hudOpen.value / (threshold * 2))))
})

const clearanceScore = computed(() =>
  Math.round(Math.min(100, (hudResolvedToday.value / TICKETS_OPENED_TODAY) * 100))
)

const responseScore = computed(() => {
  if (AVG_FRT_MINS <= FRT_GOAL_MINS) return 100
  return Math.round(Math.max(0, 100 * (1 - (AVG_FRT_MINS - FRT_GOAL_MINS) / FRT_GOAL_MINS)))
})

const efficiencyScore = computed(() => FCR_RATE)

const healthScore = computed(() =>
  Math.round(
    0.30 * backlogScore.value +
    0.30 * clearanceScore.value +
    0.25 * responseScore.value +
    0.15 * efficiencyScore.value
  )
)

const expanded = ref(false)

// Track trend by comparing current score to the previous value
const prevScore = ref(null)
const trendUp = ref(true)

watch(healthScore, (curr, old) => {
  if (old != null) {
    prevScore.value = old
    trendUp.value = curr >= old
  }
}, { immediate: true })

// Main bar color — uses red/amber/green spectrum
const healthColor = computed(() => {
  const s = healthScore.value
  if (s >= 90) return "#34d399"
  if (s >= 75) return "#22c55e"
  if (s >= 60) return "#f59e0b"
  if (s >= 40) return "#f97316"
  return "#ef4444"
})

const healthGradient = computed(() => {
  const c = healthColor.value
  return `linear-gradient(90deg, ${c}88, ${c})`
})

// Component bar colors — distinct from the main bar's red/amber/green palette
const healthComponents = computed(() => [
  { label: "Backlog", score: backlogScore.value, color: "#818cf8" },
  { label: "Clearance", score: clearanceScore.value, color: "#38bdf8" },
  { label: "Response", score: responseScore.value, color: "#c084fc" },
  { label: "Efficiency", score: efficiencyScore.value, color: "#e879f9" },
])
</script>

<style scoped>
.health-wrap {
  padding: 14px 20px 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  flex-shrink: 0;
}

.health-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.health-title {
  font-size: 12px;
  font-weight: 600;
  color: rgba(148, 163, 184, 0.4);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.health-trend {
  display: flex;
  align-items: center;
}

.health-trend--up {
  color: #34d399;
}

.health-trend--down {
  color: #ef4444;
}

.health-bar-track {
  height: 7px;
  border-radius: 99px;
  background: rgba(255, 255, 255, 0.05);
  overflow: hidden;
  margin-bottom: 10px;
}

.health-bar-fill {
  position: relative;
  height: 100%;
  border-radius: 99px;
  transition: width 0.8s cubic-bezier(0.16, 1, 0.3, 1);
  overflow: hidden;
}

.health-bar-shimmer {
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, transparent 20%, rgba(255, 255, 255, 0.22) 50%, transparent 80%);
  background-size: 50% 100%;
  background-repeat: no-repeat;
  animation: bar-shimmer 2s ease-in-out infinite alternate;
}

@keyframes bar-shimmer {
  from { background-position: -50% 0; }
  to   { background-position: 150% 0; }
}

.health-expand {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: none;
  border: none;
  padding: 0;
  font-family: inherit;
  font-size: 11px;
  font-weight: 500;
  color: rgba(148, 163, 184, 0.35);
  cursor: pointer;
  transition: color 0.15s;
}

.health-expand:hover {
  color: rgba(148, 163, 184, 0.6);
}

.health-expand-icon {
  transition: transform 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}

.health-expand-icon--open {
  transform: rotate(180deg);
}

.health-components {
  display: grid;
  grid-template-rows: 0fr;
  transition: grid-template-rows 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.health-components--open {
  grid-template-rows: 1fr;
}

.health-components-inner {
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-top: 10px;
}

.health-comp-top {
  margin-bottom: 3px;
}

.health-comp-label {
  font-size: 11px;
  font-weight: 500;
  color: rgba(148, 163, 184, 0.5);
}

.health-comp-track {
  height: 3px;
  border-radius: 99px;
  background: rgba(255, 255, 255, 0.04);
  overflow: hidden;
}

.health-comp-fill {
  height: 100%;
  border-radius: 99px;
  transition: width 0.8s cubic-bezier(0.16, 1, 0.3, 1);
}
</style>
