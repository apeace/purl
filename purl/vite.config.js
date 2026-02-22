import { execSync } from "child_process"
import path from "path"
import { fileURLToPath } from "url"
import vue from "@vitejs/plugin-vue"
import { defineConfig } from "vite"

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const SPEC_PATH = path.resolve(__dirname, "../api/docs/swagger.json")

function apiClientPlugin() {
  return {
    name: "api-client-generator",
    buildStart() {
      execSync("npm run generate -w @purl/lib", { stdio: "inherit" })
    },
    configureServer(server) {
      server.watcher.add(SPEC_PATH)
      server.watcher.on("change", (file) => {
        if (file !== SPEC_PATH) return
        execSync("npm run generate -w @purl/lib", { stdio: "inherit" })
        server.ws.send({ type: "full-reload" })
      })
    },
  }
}

export default defineConfig({
  plugins: [vue(), apiClientPlugin()],
  server: {
    host: true,
    port: 9091,
    // Polling is needed for live-reload inside Docker containers,
    // where native FS events from mounted volumes may not fire.
    watch: {
      usePolling: true,
    },
  },
})
