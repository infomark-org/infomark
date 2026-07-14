/// <reference types="vitest/config" />
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  build: {
    rollupOptions: {
      output: {
        // Split the large, rarely-changing vendor libraries into their own
        // long-term-cacheable chunks so a page change does not invalidate the
        // whole bundle. Route-level React.lazy already code-splits the pages.
        // Split React and i18n into their own long-term-cacheable chunks.
        // We deliberately do NOT force antd into a manual chunk: naming the
        // barrel package here defeats Rollup's tree-shaking and inflates the
        // bundle. Route-level React.lazy already splits antd usage per page.
        manualChunks: {
          'react-vendor': ['react', 'react-dom', 'react-router-dom'],
          i18n: ['i18next', 'react-i18next'],
        },
      },
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:2020',
        changeOrigin: true,
      }
    }
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    css: false,
  },
})
