import { createRouter, createWebHistory } from "vue-router"
import type { RouteRecordRaw } from "vue-router"
import DashboardPage from "../pages/DashboardPage.vue"
import GoPage from "../pages/GoPage.vue"
import InboxPage from "../pages/InboxPage.vue"
import LoginPage from "../pages/LoginPage.vue"
import KanbanPage from "../pages/KanbanPage.vue"
import ReportingPage from "../pages/ReportingPage.vue"
import SettingsPage from "../pages/SettingsPage.vue"

const routes: RouteRecordRaw[] = [
  { path: "/", redirect: "/go" },
  { path: "/go", component: GoPage },
  { path: "/dashboard", component: DashboardPage },
  { path: "/login", component: LoginPage, meta: { public: true } },
  { path: "/kanban", component: KanbanPage },
  { path: "/kanban/:boardId", component: KanbanPage },
  { path: "/inbox", component: InboxPage },
  { path: "/reporting", component: ReportingPage },
  { path: "/settings", component: SettingsPage },
]

export default createRouter({
  history: createWebHistory(),
  routes,
})
