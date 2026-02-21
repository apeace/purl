<template>
  <div>
    <div class="page-header">
      <h1 class="page-title">Good evening, Alex</h1>
      <p class="page-subtitle">Here's what's happening with your pipeline.</p>
    </div>

    <div class="stats-grid">
      <StatCard label="Revenue" value="$142K" :change="12" color="#818cf8" :data="[4,5,6,5,7,8,9,7,8,10,9,11]" :delay="100" />
      <StatCard label="Deals Won" value="24" :change="8" color="#34d399" :data="[3,4,3,5,6,5,7,6,8,7,8,9]" :delay="200" />
      <StatCard label="Pipeline" value="$380K" :change="-3" color="#f59e0b" :data="[9,8,7,8,7,6,7,6,5,6,5,6]" :delay="300" />
      <StatCard label="Avg Deal" value="$18.2K" :change="5" color="#ec4899" :data="[5,6,5,7,6,8,7,8,9,8,9,10]" :delay="400" />
    </div>

    <div class="panels-grid">
      <div class="panel">
        <div class="panel-header">
          <span class="panel-title">Recent Activity</span>
          <span class="panel-link">View all</span>
        </div>
        <ActivityRow name="Sarah Lin" action="Moved Acme Corp to Negotiation" time="2m ago" color="#6366f1" :delay="500" />
        <ActivityRow name="Mike Torres" action="Added note on Bolt deal" time="15m ago" color="#ec4899" :delay="600" />
        <ActivityRow name="Priya Patel" action="Won TechFlow — $24,000" time="1h ago" color="#34d399" :delay="700" />
        <ActivityRow name="James Kim" action="Created new lead: NexGen" time="3h ago" color="#f59e0b" :delay="800" />
        <ActivityRow name="Alex Chen" action="Updated Q4 forecast" time="5h ago" color="#6366f1" :delay="900" />
      </div>

      <div class="panel">
        <div class="panel-header">
          <span class="panel-title">Top Deals</span>
          <span class="panel-link">View pipeline</span>
        </div>
        <div
          v-for="(deal, i) in topDeals"
          :key="deal.name"
          class="deal-row"
          :class="{ 'deal-row--last': i === topDeals.length - 1 }"
        >
          <div class="deal-row-top">
            <span class="deal-name">{{ deal.name }}</span>
            <span class="deal-val" :style="{ color: deal.color }">{{ deal.val }}</span>
          </div>
          <div class="deal-bar-wrap">
            <div class="deal-track">
              <div
                class="deal-fill"
                :style="{ width: `${deal.pct}%`, background: `linear-gradient(90deg, ${deal.color}, ${deal.color}88)` }"
              />
            </div>
            <span class="deal-pct">{{ deal.pct }}%</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import ActivityRow from "../components/ActivityRow.vue"
import StatCard from "../components/StatCard.vue"

const topDeals = [
  { name: "Acme Corp", val: "$52,000", pct: 85, color: "#6366f1" },
  { name: "Quantum AI", val: "$38,000", pct: 60, color: "#a855f7" },
  { name: "CloudBase", val: "$28,000", pct: 35, color: "#ec4899" },
  { name: "NexGen Inc", val: "$45,000", pct: 45, color: "#f59e0b" },
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

.deal-row {
  padding: 12px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
}

.deal-row--last {
  border-bottom: none;
}

.deal-row-top {
  display: flex;
  justify-content: space-between;
  margin-bottom: 6px;
}

.deal-name {
  font-size: 13px;
  font-weight: 600;
  color: #e2e8f0;
}

.deal-val {
  font-size: 13px;
  font-weight: 700;
}

.deal-bar-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
}

.deal-track {
  flex: 1;
  height: 4px;
  border-radius: 2px;
  background: rgba(255, 255, 255, 0.04);
}

.deal-fill {
  height: 100%;
  border-radius: 2px;
  transition: width 1s cubic-bezier(0.16, 1, 0.3, 1);
}

.deal-pct {
  font-size: 11px;
  color: rgba(148, 163, 184, 0.5);
  min-width: 28px;
}
</style>
