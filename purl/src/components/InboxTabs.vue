<template>
  <div class="inbox-tabs">
    <button
      v-for="tab in tabs"
      :key="tab.key"
      class="tab"
      :class="{ 'tab--active': modelValue === tab.key }"
      @click="$emit('update:modelValue', tab.key)"
    >
      <component :is="tab.icon" :size="16" class="tab-icon" />
      <span class="tab-label">{{ tab.label }}</span>
      <span v-if="tab.count > 0" class="tab-badge">{{ tab.count }}</span>
      <div v-if="modelValue === tab.key" class="tab-indicator" />
    </button>
  </div>
</template>

<script setup lang="ts">
import { Inbox, Star, User, UserX } from "lucide-vue-next"
import { storeToRefs } from "pinia"
import { computed } from "vue"
import { useTicketStore } from "../stores/useTicketStore"
import { useUserStore } from "../stores/useUserStore"

withDefaults(defineProps<{
  modelValue?: string
}>(), {
  modelValue: "all",
})

defineEmits<{
  "update:modelValue": [value: string]
}>()

const ticketStore = useTicketStore()
const { tickets } = storeToRefs(ticketStore)
const { CURRENT_USER } = useUserStore()

const tabs = computed(() => [
  {
    key: "all",
    label: "All",
    icon: Inbox,
    count: tickets.value.filter((t) => t.status !== "closed").length,
  },
  {
    key: "mine",
    label: "Mine",
    icon: User,
    count: tickets.value.filter((t) => t.assignee === CURRENT_USER && t.status !== "closed").length,
  },
  {
    key: "unassigned",
    label: "Unassigned",
    icon: UserX,
    count: tickets.value.filter((t) => t.assignee === "Unassigned" && t.status !== "closed").length,
  },
  {
    key: "starred",
    label: "Starred",
    icon: Star,
    count: tickets.value.filter((t) => t.starred).length,
  },
])
</script>

<style scoped>
.inbox-tabs {
  display: flex;
  gap: 0;
  padding: 0 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  overflow-x: auto;
  scrollbar-width: none;
}

.inbox-tabs::-webkit-scrollbar {
  display: none;
}

.tab {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 12px 16px;
  border: none;
  background: transparent;
  color: #64748b;
  font-size: 13px;
  font-weight: 500;
  font-family: inherit;
  cursor: pointer;
  white-space: nowrap;
  position: relative;
  transition: color 0.15s cubic-bezier(0.16, 1, 0.3, 1);
  flex-shrink: 0;
}

.tab:hover {
  color: #94a3b8;
}

.tab--active {
  color: #e2e8f0;
}

.tab-icon {
  flex-shrink: 0;
}

.tab-badge {
  font-size: 11px;
  font-weight: 600;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: 9px;
  background: rgba(255, 255, 255, 0.06);
  color: #94a3b8;
  display: flex;
  align-items: center;
  justify-content: center;
}

.tab--active .tab-badge {
  background: rgba(99, 102, 241, 0.15);
  color: #a5b4fc;
}

.tab-indicator {
  position: absolute;
  bottom: 0;
  left: 16px;
  right: 16px;
  height: 2px;
  background: #6366f1;
  border-radius: 2px 2px 0 0;
  animation: indicator-in 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes indicator-in {
  from { transform: scaleX(0); }
  to { transform: scaleX(1); }
}

@media (max-width: 479px) {
  .tab-label {
    display: none;
  }

  .tab {
    padding: 12px 14px;
  }
}
</style>
