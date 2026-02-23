<template>
  <div class="filter-wrap" ref="wrapEl">
    <button
      class="filter-trigger"
      :class="{ 'filter-trigger--active': open || activeFilterCount }"
      @click="open = !open"
    >
      <SlidersHorizontal :size="15" />
      <span class="filter-trigger-label">Filter</span>
      <span v-if="activeFilterCount" class="filter-badge">{{ activeFilterCount }}</span>
    </button>

    <Transition name="panel">
      <div v-if="open" class="filter-panel">
        <!-- Search -->
        <div class="filter-section">
          <div class="search-row">
            <Search :size="14" class="search-icon" />
            <input
              ref="searchEl"
              v-model="filterKeyword"
              class="search-input"
              placeholder="Search tickets…"
              @keydown.escape="open = false"
            />
            <button
              v-if="filterKeyword"
              class="search-clear"
              @click="filterKeyword = ''"
            >
              <X :size="12" />
            </button>
          </div>
        </div>

        <!-- Priority -->
        <div class="filter-section">
          <div class="filter-label">Priority</div>
          <div class="chip-row">
            <button
              v-for="p in priorities"
              :key="p"
              class="chip"
              :class="[
                `chip--${p}`,
                { 'chip--on': filterPriorities.has(p) },
              ]"
              @click="toggleSet(filterPriorities, p)"
            >{{ p }}</button>
          </div>
        </div>

        <!-- Assignee -->
        <div v-if="uniqueAssignees.length" class="filter-section">
          <div class="filter-label">Member</div>
          <div class="chip-row">
            <button
              v-for="a in uniqueAssignees"
              :key="a"
              class="chip chip--member"
              :class="{ 'chip--on': filterAssignees.has(a) }"
              @click="toggleSet(filterAssignees, a)"
            >{{ a }}</button>
          </div>
        </div>

        <!-- Status / Stages -->
        <div class="filter-section">
          <div class="filter-label">{{ customStages?.length ? 'Column' : 'Status' }}</div>
          <div class="chip-row">
            <button
              v-for="s in statusItems"
              :key="s.value"
              class="chip chip--status"
              :class="[
                `chip--${s.value}`,
                { 'chip--on': filterStatuses.has(s.value) },
              ]"
              @click="toggleSet(filterStatuses, s.value)"
            >{{ s.label }}</button>
          </div>
        </div>

        <!-- Clear all -->
        <button
          v-if="activeFilterCount"
          class="clear-btn"
          @click="clearFilters"
        >Clear all filters</button>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { Search, SlidersHorizontal, X } from "lucide-vue-next"
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue"
import { useTickets } from "../composables/useTickets"

const props = withDefaults(defineProps<{
  customStages?: { id: string; name: string; color: string }[]
}>(), {
  customStages: undefined,
})

const {
  activeFilterCount,
  clearFilters,
  filterAssignees,
  filterKeyword,
  filterPriorities,
  filterStatuses,
  uniqueAssignees,
} = useTickets()

const priorities = ["low", "medium", "high", "urgent"]
const defaultStatuses = ["new", "open", "pending", "escalated", "solved", "closed"]
const statusItems = computed(() => {
  if (props.customStages?.length) {
    return props.customStages.map((s) => ({ value: s.id, label: s.name }))
  }
  return defaultStatuses.map((s) => ({ value: s, label: s }))
})

const open = ref(false)
const wrapEl = ref<HTMLElement | null>(null)
const searchEl = ref<HTMLInputElement | null>(null)

function toggleSet(set: Set<string>, val: string) {
  if (set.has(val)) set.delete(val)
  else set.add(val)
}

// Focus search input when panel opens
watch(open, (isOpen) => {
  if (isOpen) nextTick(() => searchEl.value?.focus())
})

// Click outside to close
function onClickOutside(e: PointerEvent) {
  if (open.value && wrapEl.value && !wrapEl.value.contains(e.target as Node)) {
    open.value = false
  }
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === "Escape" && open.value) open.value = false
}

onMounted(() => {
  document.addEventListener("pointerdown", onClickOutside)
  document.addEventListener("keydown", onKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener("pointerdown", onClickOutside)
  document.removeEventListener("keydown", onKeydown)
})
</script>

<style scoped>
.filter-wrap {
  position: relative;
}

/* ── Trigger ──────────────────────────────────────────── */

.filter-trigger {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.03);
  color: #94a3b8;
  font-size: 13px;
  font-weight: 500;
  font-family: inherit;
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
}

.filter-trigger:hover {
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(255, 255, 255, 0.12);
  color: #e2e8f0;
}

.filter-trigger--active {
  background: rgba(99, 102, 241, 0.1);
  border-color: rgba(99, 102, 241, 0.25);
  color: #a5b4fc;
}

.filter-trigger-label {
  display: none;
}

@media (min-width: 480px) {
  .filter-trigger-label {
    display: inline;
  }
}

.filter-badge {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  border-radius: 9px;
  background: #6366f1;
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  padding: 0 5px;
}

/* ── Panel ────────────────────────────────────────────── */

.filter-panel {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  z-index: 50;
  min-width: 280px;
  max-width: 340px;
  background: #0f172a;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 12px;
  padding: 12px;
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.5);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* ── Section ──────────────────────────────────────────── */

.filter-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.filter-label {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: rgba(148, 163, 184, 0.4);
}

/* ── Search ───────────────────────────────────────────── */

.search-row {
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  padding: 0 10px;
  transition: border-color 0.15s;
}

.search-row:focus-within {
  border-color: rgba(99, 102, 241, 0.4);
}

.search-icon {
  color: rgba(148, 163, 184, 0.4);
  flex-shrink: 0;
}

.search-input {
  flex: 1;
  border: none;
  background: transparent;
  padding: 8px 0;
  font-size: 13px;
  font-family: inherit;
  color: #e2e8f0;
  outline: none;
}

.search-input::placeholder {
  color: rgba(148, 163, 184, 0.3);
}

.search-clear {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border: none;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.06);
  color: rgba(148, 163, 184, 0.6);
  cursor: pointer;
  flex-shrink: 0;
  transition: background 0.15s, color 0.15s;
}

.search-clear:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #e2e8f0;
}

/* ── Chips ────────────────────────────────────────────── */

.chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 5px;
}

.chip {
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 500;
  font-family: inherit;
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(255, 255, 255, 0.02);
  color: rgba(148, 163, 184, 0.6);
  cursor: pointer;
  text-transform: capitalize;
  transition: all 0.15s;
  white-space: nowrap;
}

.chip:hover {
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(255, 255, 255, 0.12);
  color: #e2e8f0;
}

/* Priority chip active states */
.chip--low.chip--on {
  background: rgba(52, 211, 153, 0.12);
  border-color: rgba(52, 211, 153, 0.25);
  color: #6ee7b7;
}

.chip--medium.chip--on {
  background: rgba(245, 158, 11, 0.12);
  border-color: rgba(245, 158, 11, 0.25);
  color: #fcd34d;
}

.chip--high.chip--on {
  background: rgba(239, 68, 68, 0.12);
  border-color: rgba(239, 68, 68, 0.25);
  color: #fca5a5;
}

.chip--urgent.chip--on {
  background: rgba(239, 68, 68, 0.15);
  border-color: rgba(239, 68, 68, 0.3);
  color: #fca5a5;
}

/* Member chip active */
.chip--member.chip--on {
  background: rgba(99, 102, 241, 0.12);
  border-color: rgba(99, 102, 241, 0.25);
  color: #a5b4fc;
}

/* Status chip active states */
.chip--status.chip--on {
  background: rgba(99, 102, 241, 0.12);
  border-color: rgba(99, 102, 241, 0.25);
  color: #a5b4fc;
}

.chip--new.chip--on {
  background: rgba(56, 189, 248, 0.12);
  border-color: rgba(56, 189, 248, 0.25);
  color: #7dd3fc;
}

.chip--open.chip--on {
  background: rgba(99, 102, 241, 0.12);
  border-color: rgba(99, 102, 241, 0.25);
  color: #a5b4fc;
}

.chip--pending.chip--on {
  background: rgba(168, 85, 247, 0.12);
  border-color: rgba(168, 85, 247, 0.25);
  color: #d8b4fe;
}

.chip--escalated.chip--on {
  background: rgba(249, 115, 22, 0.12);
  border-color: rgba(249, 115, 22, 0.25);
  color: #fdba74;
}

.chip--solved.chip--on {
  background: rgba(52, 211, 153, 0.12);
  border-color: rgba(52, 211, 153, 0.25);
  color: #6ee7b7;
}

.chip--closed.chip--on {
  background: rgba(148, 163, 184, 0.12);
  border-color: rgba(148, 163, 184, 0.25);
  color: #94a3b8;
}

/* ── Clear button ─────────────────────────────────────── */

.clear-btn {
  width: 100%;
  padding: 7px;
  font-size: 12px;
  font-weight: 600;
  font-family: inherit;
  border: none;
  border-radius: 7px;
  background: rgba(239, 68, 68, 0.08);
  color: #fca5a5;
  cursor: pointer;
  transition: background 0.15s;
}

.clear-btn:hover {
  background: rgba(239, 68, 68, 0.15);
}

/* ── Transitions ──────────────────────────────────────── */

.panel-enter-active,
.panel-leave-active {
  transition: opacity 0.15s, transform 0.15s cubic-bezier(0.16, 1, 0.3, 1);
}

.panel-enter-from,
.panel-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

/* ── Mobile ───────────────────────────────────────────── */

@media (max-width: 767px) {
  .filter-panel {
    left: 0;
    right: 0;
    min-width: 0;
    max-width: none;
    width: calc(100vw - 32px);
  }
}
</style>
