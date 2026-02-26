<template>
  <div class="page">
    <ComingSoon page />
    <h1 class="page-title">Settings</h1>
  </div>
</template>

<script setup lang="ts">
import ComingSoon from "../components/ComingSoon.vue"
</script>

<style scoped>
.page {
  position: relative;
  min-height: 400px;
  padding: 32px 24px;
}

.page-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: #0f172a;
  letter-spacing: -0.02em;
}
</style>
