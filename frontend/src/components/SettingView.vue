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
        <button class="btn btn-secondary theme-toggle" @click="toggleTheme">
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
.setting-header { 
  margin-bottom: var(--spacing-lg); 
}
.setting-header h3 { 
  margin: 0; 
  color: var(--color-text-primary); 
}
.setting-header p { 
  font-size: var(--font-size-sm); 
  color: var(--color-text-secondary); 
  margin-top: var(--spacing-xs); 
}

.setting-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-md);
  background: var(--color-bg-hover);
  border-radius: var(--radius-md);
  transition: background-color var(--transition-fast);
}

.item-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.item-title {
  font-weight: 500;
  color: var(--color-text-primary);
  font-size: var(--font-size-md);
}

.item-desc {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
}

.theme-toggle {
  min-width: 120px;
}
</style>