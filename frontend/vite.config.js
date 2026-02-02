import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      // 别名必须指向具体的物理路径
      'wailsjs': path.resolve(__dirname, './wailsjs')
    }
  }
})