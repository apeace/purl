<template>
  <svg :width="w" :height="h" :viewBox="`0 0 ${w} ${h}`" style="display: block">
    <defs>
      <linearGradient :id="gradId" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" :stop-color="color" stop-opacity="0.3" />
        <stop offset="100%" :stop-color="color" stop-opacity="0" />
      </linearGradient>
    </defs>
    <path :d="areaPath" :fill="`url(#${gradId})`" />
    <path
      :d="linePath"
      fill="none"
      :stroke="color"
      stroke-width="1.5"
      stroke-linecap="round"
      stroke-linejoin="round"
    />
  </svg>
</template>

<script setup>
import { computed } from "vue"

const props = defineProps({
  color: { type: String, default: "#818cf8" },
  data: { type: Array, default: () => [4, 7, 5, 9, 6, 8, 3, 7, 10, 6, 8, 9] },
})

const w = 80
const h = 28

const gradId = computed(() => `sp-${props.color.replace("#", "")}`)

const linePath = computed(() => {
  const max = Math.max(...props.data)
  return props.data
    .map((v, i) => {
      const x = (i / (props.data.length - 1)) * w
      const y = h - (v / max) * h
      return `${i === 0 ? "M" : "L"}${x},${y}`
    })
    .join(" ")
})

const areaPath = computed(() => `${linePath.value} L${w},${h} L0,${h} Z`)
</script>
