<template>
  <div class="coming-soon" :class="{ 'coming-soon--page': page }" @click.stop @pointerdown.stop>
    <div class="coming-soon-badge">
      <Rocket :size="14" />
      <span>Coming Soon</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Rocket } from "lucide-vue-next"

withDefaults(defineProps<{ page?: boolean }>(), { page: false })
</script>

<style scoped>
.coming-soon {
  position: absolute;
  inset: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(10, 14, 26, 0.6);
  backdrop-filter: blur(6px);
  border-radius: inherit;
  cursor: default;
  animation: comingSoonFade 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

.coming-soon--page {
  align-items: flex-start;
  padding-top: 120px;
}

@keyframes comingSoonFade {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.coming-soon-badge {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 8px 18px;
  background: rgba(99, 102, 241, 0.12);
  border: 1px solid rgba(99, 102, 241, 0.2);
  border-radius: 20px;
  color: #a5b4fc;
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.02em;
  white-space: nowrap;
  box-shadow: 0 0 24px rgba(99, 102, 241, 0.08);
  transition: background 0.2s, border-color 0.2s;
}
</style>
