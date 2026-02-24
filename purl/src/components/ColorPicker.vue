<template>
  <div class="color-picker">
    <button
      v-for="c in colors"
      :key="c"
      class="color-swatch"
      :class="{ 'color-swatch--selected': c === modelValue }"
      :style="{ background: c }"
      :title="c"
      @click="emit('update:modelValue', c)"
    />
  </div>
</template>

<script lang="ts">
// Curated palette for support ticket workflows.
// Ordered by semantic meaning: urgency → caution → success → info → neutral.
export const COLUMN_COLORS = [
  "#f87171",  // red-400    — Critical / Urgent
  "#fb923c",  // orange-400 — High Priority
  "#fbbf24",  // amber-400  — At Risk
  "#facc15",  // yellow-400 — Waiting for Info
  "#4ade80",  // green-400  — Resolved
  "#34d399",  // emerald-400 — Done / Closed
  "#2dd4bf",  // teal-400   — On Hold
  "#22d3ee",  // cyan-400   — In Progress
  "#38bdf8",  // sky-400    — Open
  "#60a5fa",  // blue-400   — New
  "#818cf8",  // indigo-400 — Assigned
  "#a78bfa",  // violet-400 — Scheduled
  "#c084fc",  // purple-400 — VIP / Special
  "#f472b6",  // pink-400   — Escalated
  "#fb7185",  // rose-400   — Urgent / Breached
  "#94a3b8",  // slate-400  — Neutral / Default
]
</script>

<script setup lang="ts">
// Expose palette to template via a local binding (COLUMN_COLORS is in module scope above)
const colors = COLUMN_COLORS

defineProps<{ modelValue: string }>()

const emit = defineEmits<{ "update:modelValue": [color: string] }>()
</script>

<style scoped>
.color-picker {
  display: grid;
  grid-template-columns: repeat(4, 22px);
  gap: 6px;
}

.color-swatch {
  width: 22px;
  height: 22px;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  padding: 0;
  transition: transform 0.12s, outline-color 0.12s;
  outline: 2px solid transparent;
  outline-offset: 2px;
}

.color-swatch:hover {
  transform: scale(1.15);
}

.color-swatch--selected {
  outline-color: rgba(255, 255, 255, 0.85);
}
</style>
