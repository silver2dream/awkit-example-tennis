import { defineConfig } from 'vite'

export default defineConfig({
  cacheDir: '.vite',
  build: {
    target: 'es2020',
  },
})
