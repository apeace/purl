<template>
  <div>
    <div class="pipeline-header">
      <h1 class="page-title">Pipeline</h1>
    </div>
    <div class="stages-scroll">
      <PipelineStage
        v-for="(stage, i) in stages"
        :key="stage.title"
        v-bind="stage"
        :delay="150 + i * 100"
        @select="selectedTicketId = $event"
      />
    </div>
    <TicketModal :ticket-id="selectedTicketId" @close="selectedTicketId = null" />
  </div>
</template>

<script setup>
import { computed, ref } from "vue"
import PipelineStage from "../components/PipelineStage.vue"
import TicketModal from "../components/TicketModal.vue"
import { useTickets } from "../composables/useTickets.js"

const { tickets } = useTickets()

const selectedTicketId = ref(null)

const stageDefs = [
  { status: "new", title: "New", color: "#38bdf8" },
  { status: "open", title: "Open", color: "#60a5fa" },
  { status: "pending", title: "Pending", color: "#a855f7" },
  { status: "escalated", title: "Technical Escalation", color: "#f97316" },
  { status: "solved", title: "Solved", color: "#34d399" },
  { status: "closed", title: "Closed", color: "#94a3b8" },
]

const stages = computed(() =>
  stageDefs.map((def) => {
    const items = tickets.value.filter((t) => t.status === def.status)
    return {
      title: def.title,
      count: items.length,
      color: def.color,
      items: items.map((t) => ({
        id: t.id,
        name: t.name,
        company: t.company,
        subject: t.subject,
        priority: t.priority,
        avatarColor: t.avatarColor,
      })),
    }
  })
)
</script>

<style scoped>
.pipeline-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 12px;
}

.page-title {
  font-size: 22px;
  font-weight: 700;
  color: #f1f5f9;
  letter-spacing: -0.02em;
}

.stages-scroll {
  display: flex;
  gap: 14px;
  overflow-x: auto;
  padding-bottom: 16px;
}
</style>
