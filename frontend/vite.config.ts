import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import { tanstackRouter } from '@tanstack/router-plugin/vite'

export default defineConfig(({ mode }) => ({
  // The router plugin must come before the React plugin.
  plugins: [tanstackRouter({ target: 'react', autoCodeSplitting: true }), react()],
  server: {
    proxy: {
      '/api': loadEnv(mode, '.', 'MINIV2_').MINIV2_API_URL || 'http://localhost:8080',
    },
  },
}))
