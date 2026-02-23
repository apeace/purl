<template>
  <div class="stat-card" :class="{ 'stat-card--visible': visible }">
    <span class="stat-label">{{ label }}</span>
    <div class="stat-body">
      <div class="stat-values">
        <span class="stat-value">{{ value }}</span>
        <span class="stat-change" :class="change > 0 ? 'stat-change--up' : 'stat-change--down'">
          {{ change > 0 ? '↑' : '↓' }} {{ Math.abs(change) }}%
        </span>
      </div>
      <Sparkline :color="color" :data="data" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue"
import Sparkline from "./Sparkline.vue"

const props = withDefaults(defineProps<{
  label: string
  value: string
  change: number
  color?: string
  data: number[]
  delay?: number
}>(), {
  color: "#818cf8",
  delay: 0,
})

const visible = ref(false)

onMounted(() => {
  setTimeout(() => { visible.value = true }, props.delay)
})
</script>

<style scoped>
.stat-card {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 16px;
  padding: 20px 20px 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  opacity: 0;
  transform: translateY(16px);
  transition: opacity 0.5s cubic-bezier(0.16, 1, 0.3, 1), transform 0.5s cubic-bezier(0.16, 1, 0.3, 1);
  backdrop-filter: blur(12px);
}

.stat-card--visible {
  opacity: 1;
  transform: translateY(0);
}

.stat-label {
  font-size: 12px;
  color: rgba(148, 163, 184, 0.8);
  font-weight: 500;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.stat-body {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
}

.stat-values {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #f1f5f9;
  letter-spacing: -0.02em;
  line-height: 1;
}

.stat-change {
  font-size: 12px;
  font-weight: 600;
}

.stat-change--up {
  color: #34d399;
}

.stat-change--down {
  color: #f87171;
}
</style>
