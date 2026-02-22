<template>
  <div class="stage" :class="{ 'stage--visible': visible }">
    <div class="stage-header">
      <div class="stage-dot" :style="{ background: color }" />
      <span class="stage-title">{{ title }}</span>
      <span class="stage-count">{{ count }}</span>
    </div>
    <div class="stage-cards">
      <div
        v-for="(item, i) in items"
        :key="i"
        class="deal-card"
        :style="{ '--accent': color }"
        @click="emit('select', item.id)"
      >
        <div class="deal-top">
          <div class="deal-avatar" :style="{ background: item.avatarColor }">{{ item.name[0] }}</div>
          <div class="deal-meta">
            <div class="deal-name">{{ item.name }}</div>
            <div class="deal-company">{{ item.company }}</div>
          </div>
        </div>
        <div class="deal-subject">{{ item.subject }}</div>
        <div v-if="item.priority" class="deal-footer">
          <span class="deal-priority" :class="`deal-priority--${item.priority}`">{{ item.priority }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue"

const props = defineProps({
  title: String,
  count: Number,
  color: String,
  items: Array,
  delay: { type: Number, default: 0 },
})

const emit = defineEmits(["select"])

const visible = ref(false)

onMounted(() => {
  setTimeout(() => { visible.value = true }, props.delay)
})
</script>

<style scoped>
.stage {
  min-width: 220px;
  flex: 1 0 220px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 14px;
  padding: 14px;
  opacity: 0;
  transform: translateY(20px);
  transition: opacity 0.5s cubic-bezier(0.16, 1, 0.3, 1), transform 0.5s cubic-bezier(0.16, 1, 0.3, 1);
}

.stage--visible {
  opacity: 1;
  transform: translateY(0);
}

.stage-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 14px;
}

.stage-dot {
  width: 8px;
  height: 8px;
  border-radius: 4px;
  flex-shrink: 0;
}

.stage-title {
  font-size: 13px;
  font-weight: 600;
  color: #e2e8f0;
  flex: 1;
}

.stage-count {
  font-size: 11px;
  font-weight: 700;
  color: rgba(148, 163, 184, 0.6);
  background: rgba(255, 255, 255, 0.05);
  border-radius: 6px;
  padding: 2px 7px;
}

.stage-cards {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.deal-card {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 10px;
  padding: 12px 14px;
  cursor: pointer;
  transition: background 0.2s, border-color 0.2s;
}

.deal-card:hover {
  background: rgba(255, 255, 255, 0.06);
  border-color: color-mix(in srgb, var(--accent) 30%, transparent);
}

.deal-top {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.deal-avatar {
  width: 24px;
  height: 24px;
  border-radius: 7px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
}

.deal-meta {
  min-width: 0;
}

.deal-name {
  font-size: 13px;
  font-weight: 600;
  color: #e2e8f0;
  line-height: 1.2;
}

.deal-company {
  font-size: 11px;
  color: rgba(148, 163, 184, 0.5);
}

.deal-subject {
  font-size: 12px;
  color: rgba(148, 163, 184, 0.7);
  line-height: 1.35;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.deal-footer {
  display: flex;
  align-items: center;
  margin-top: 8px;
}

.deal-priority {
  font-size: 10px;
  font-weight: 600;
  padding: 2px 7px;
  border-radius: 5px;
  text-transform: capitalize;
}

.deal-priority--high {
  background: rgba(239, 68, 68, 0.1);
  color: #fca5a5;
}

.deal-priority--medium {
  background: rgba(245, 158, 11, 0.1);
  color: #fcd34d;
}

.deal-priority--low {
  background: rgba(52, 211, 153, 0.1);
  color: #6ee7b7;
}
</style>
