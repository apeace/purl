import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

export default defineConfig({
  plugins: [vue()],
  server: {
    host: true,
    port: 9091,
    // Polling is needed for live-reload inside Docker containers,
    // where native FS events from mounted volumes may not fire.
    watch: {
      usePolling: true,
    },
  },
});
