<template>
  <div class="pipeline-page">
    <div class="pipeline-header">
      <h1 class="page-title">Pipeline</h1>
      <div class="header-controls">
        <div class="pipeline-search">
          <Search :size="14" class="pipeline-search-icon" />
          <input
            v-model="searchQuery"
            class="pipeline-search-input"
            placeholder="Search cards…"
          />
          <button v-if="searchQuery" class="pipeline-search-clear" @click="searchQuery = ''">
            <X :size="12" />
          </button>
        </div>
        <FilterPanel />
      </div>
    </div>
    <div class="stages-scroll">
      <PipelineStage
        v-for="(stage, i) in stages"
        :key="stage.title"
        v-bind="stage"
        :dragging-id="draggingId"
        :delay="150 + i * 100"
        @select="selectedTicketId = $event"
        @drag-start="draggingId = $event"
        @drag-end="draggingId = null"
        @drop="onDrop"
      />
    </div>
    <TicketModal :ticket-id="selectedTicketId" @close="selectedTicketId = null" />
  </div>
</template>

<script setup>
import { Search, X } from "lucide-vue-next"
import { computed, ref } from "vue"
import FilterPanel from "../components/FilterPanel.vue"
import PipelineStage from "../components/PipelineStage.vue"
import TicketModal from "../components/TicketModal.vue"
import { useTickets } from "../composables/useTickets.js"
import { STATUS_COLORS } from "../utils/colors.js"

const { filteredTickets, setStatus } = useTickets()

const selectedTicketId = ref(null)
const draggingId = ref(null)
const searchQuery = ref("")

function onDrop({ ticketId, status }) {
  draggingId.value = null
  setStatus(ticketId, status)
}

const stageDefs = [
  { status: "new", title: "New", color: STATUS_COLORS.new },
  { status: "open", title: "Open", color: STATUS_COLORS.open },
  { status: "pending", title: "Pending", color: STATUS_COLORS.pending },
  { status: "escalated", title: "Technical Escalation", color: STATUS_COLORS.escalated },
  { status: "solved", title: "Solved", color: STATUS_COLORS.solved },
  { status: "closed", title: "Closed", color: STATUS_COLORS.closed },
]

const stages = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  return stageDefs.map((def) => {
    let items = filteredTickets.value.filter((t) => t.status === def.status)
    if (q) {
      items = items.filter((t) => {
        const desc = t.messages[0]?.text ?? ""
        return t.name.toLowerCase().includes(q)
          || t.company.toLowerCase().includes(q)
          || t.subject.toLowerCase().includes(q)
          || desc.toLowerCase().includes(q)
      })
    }
    return {
      title: def.title,
      count: items.length,
      color: def.color,
      status: def.status,
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
})
</script>

<style scoped>
.pipeline-page {
  min-width: 0;
}

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

.header-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pipeline-search {
  display: flex;
  align-items: center;
  gap: 6px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  padding: 0 10px;
  transition: border-color 0.15s;
}

.pipeline-search:focus-within {
  border-color: rgba(99, 102, 241, 0.4);
}

.pipeline-search-icon {
  color: rgba(148, 163, 184, 0.4);
  flex-shrink: 0;
}

.pipeline-search-input {
  border: none;
  background: transparent;
  padding: 6px 0;
  width: 160px;
  font-size: 13px;
  font-family: inherit;
  color: #e2e8f0;
  outline: none;
}

.pipeline-search-input::placeholder {
  color: rgba(148, 163, 184, 0.3);
}

.pipeline-search-clear {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border: none;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.06);
  color: rgba(148, 163, 184, 0.6);
  cursor: pointer;
  flex-shrink: 0;
  transition: background 0.15s, color 0.15s;
}

.pipeline-search-clear:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #e2e8f0;
}

.stages-scroll {
  display: flex;
  gap: 14px;
  overflow-x: auto;
  padding-bottom: 16px;
}
</style>
