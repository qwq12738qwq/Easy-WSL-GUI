<template>
  <div class="view-card">
    <div class="setting-header">
      <h3>⚙️ 系统设置</h3>
      <p>自定义您的 WSL 管理器外观与偏好</p>
    </div>
    
    <div class="setting-list">
      <div class="setting-item">
        <div class="item-info">
          <span class="item-title">外观模式</span>
          <span class="item-desc">切换浅色或深色主题外观</span>
        </div>
        <button class="theme-toggle" @click="toggleTheme">
          {{ isDark ? '🌙 深色模式' : '☀️ 亮色模式' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const isDark = ref(true)

const toggleTheme = () => {
  isDark.value = !isDark.value
  const theme = isDark.value ? 'dark' : 'light'
  document.documentElement.setAttribute('data-theme', theme)
  localStorage.setItem('wsl-theme', theme)
}

onMounted(() => {
  const savedTheme = localStorage.getItem('wsl-theme') || 'dark'
  isDark.value = savedTheme === 'dark'
  document.documentElement.setAttribute('data-theme', savedTheme)
})
</script>

<style scoped>
.setting-header { margin-bottom: 24px; }
.setting-header h3 { margin: 0; color: var(--text-color); }
.setting-header p { font-size: 13px; color: #888; margin-top: 4px; }

.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background: var(--sidebar-hover);
  border-radius: 8px;
  border: 1px solid rgba(255,255,255,0.05);
}

.item-info { display: flex; flex-direction: column; gap: 4px; }
.item-title { font-weight: 600; font-size: 14px; color: var(--text-color); }
.item-desc { font-size: 12px; color: #888; }

.theme-toggle {
  padding: 8px 16px;
  background: var(--accent-color);
  color: var(--brand-color);
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: opacity 0.2s;
}
.theme-toggle:hover { opacity: 0.9; }
</style>