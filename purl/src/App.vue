<template>
  <RouterView v-if="route.meta.public" />
  <AppLayout v-else />
</template>

<script setup lang="ts">
import { onMounted } from "vue"
import { useRoute } from "vue-router"
import AppLayout from "./layouts/AppLayout.vue"
import { useUserStore } from "./stores/useUserStore"

const route = useRoute()
const userStore = useUserStore()

onMounted(() => {
  userStore.loadUser()
})
</script>
