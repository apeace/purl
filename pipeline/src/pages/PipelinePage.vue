<template>
  <div>
    <div class="pipeline-header">
      <h1 class="page-title">Pipeline</h1>
      <button class="add-btn">
        <Plus :size="16" />
        Add Deal
      </button>
    </div>
    <div class="stages-scroll">
      <PipelineStage
        v-for="(stage, i) in stages"
        :key="stage.title"
        v-bind="stage"
        :delay="150 + i * 100"
      />
    </div>
  </div>
</template>

<script setup>
import { Plus } from "lucide-vue-next"
import PipelineStage from "../components/PipelineStage.vue"

const stages = [
  {
    title: "Qualified",
    count: 4,
    color: "#818cf8",
    items: [
      { name: "NexGen Inc", company: "Enterprise", value: "$45K", owner: "J", avatarColor: "#f59e0b" },
      { name: "DataFlow", company: "Mid-market", value: "$22K", owner: "S", avatarColor: "#6366f1" },
    ],
  },
  {
    title: "Discovery",
    count: 3,
    color: "#a855f7",
    items: [
      { name: "CloudBase", company: "Startup", value: "$28K", owner: "P", avatarColor: "#ec4899" },
      { name: "Orion Labs", company: "Enterprise", value: "$67K", owner: "A", avatarColor: "#34d399" },
    ],
  },
  {
    title: "Proposal",
    count: 2,
    color: "#ec4899",
    items: [
      { name: "Quantum AI", company: "Mid-market", value: "$38K", owner: "M", avatarColor: "#f59e0b" },
    ],
  },
  {
    title: "Negotiation",
    count: 2,
    color: "#f59e0b",
    items: [
      { name: "Acme Corp", company: "Enterprise", value: "$52K", owner: "S", avatarColor: "#6366f1" },
      { name: "TerraForm", company: "Startup", value: "$18K", owner: "J", avatarColor: "#ec4899" },
    ],
  },
]
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

.add-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: linear-gradient(135deg, #6366f1, #a855f7);
  border: none;
  border-radius: 10px;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  font-family: inherit;
  box-shadow: 0 4px 20px rgba(99, 102, 241, 0.3);
  transition: box-shadow 0.2s;
}

.add-btn:hover {
  box-shadow: 0 4px 28px rgba(99, 102, 241, 0.45);
}

.stages-scroll {
  display: flex;
  gap: 14px;
  overflow-x: auto;
  padding-bottom: 16px;
}
</style>
