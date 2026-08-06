import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// Proxy /api → backend Go supaya dev tanpa urusan CORS; di production, admin
// panel disajikan di balik reverse proxy yang sama dengan API.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': { target: 'http://127.0.0.1:8081', changeOrigin: true },
    },
  },
});
