import { createRouter, createWebHistory } from "vue-router"
import DashboardPage from "../pages/DashboardPage.vue"
import InboxPage from "../pages/InboxPage.vue"
import PipelinePage from "../pages/PipelinePage.vue"
import ReportingPage from "../pages/ReportingPage.vue"
import SettingsPage from "../pages/SettingsPage.vue"

const routes = [
  { path: "/", redirect: "/dashboard" },
  { path: "/dashboard", component: DashboardPage },
  { path: "/pipeline", component: PipelinePage },
  { path: "/inbox", component: InboxPage },
  { path: "/reporting", component: ReportingPage },
  { path: "/settings", component: SettingsPage },
]

export default createRouter({
  history: createWebHistory(),
  routes,
})
