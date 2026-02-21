<template>
  <div class="activity-row" :class="{ 'activity-row--visible': visible }">
    <div class="activity-avatar" :style="{ background: color }">
      {{ name[0] }}
    </div>
    <div class="activity-text">
      <div class="activity-name">{{ name }}</div>
      <div class="activity-action">{{ action }}</div>
    </div>
    <span class="activity-time">{{ time }}</span>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue"

const props = defineProps({
  name: String,
  action: String,
  time: String,
  color: String,
  delay: { type: Number, default: 0 },
})

const visible = ref(false)

onMounted(() => {
  setTimeout(() => { visible.value = true }, props.delay)
})
</script>

<style scoped>
.activity-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
  opacity: 0;
  transform: translateX(-12px);
  transition: opacity 0.4s cubic-bezier(0.16, 1, 0.3, 1), transform 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

.activity-row:last-child {
  border-bottom: none;
}

.activity-row--visible {
  opacity: 1;
  transform: translateX(0);
}

.activity-avatar {
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

.activity-text {
  flex: 1;
  min-width: 0;
}

.activity-name {
  font-size: 13px;
  font-weight: 600;
  color: #e2e8f0;
}

.activity-action {
  font-size: 12px;
  color: rgba(148, 163, 184, 0.7);
  margin-top: 1px;
}

.activity-time {
  font-size: 11px;
  color: rgba(148, 163, 184, 0.5);
  white-space: nowrap;
  flex-shrink: 0;
}
</style>
