<template>
  <div class="view-card">
    <div class="setting-header">
      <h3>⚙️ 系统设置</h3>
      <p>自定义您的管理器外观与偏好</p>
    </div>
    
    <div class="setting-list">
      <div class="setting-item">
        <div class="item-info">
          <span class="item-title">外观模式</span>
          <span class="item-desc">切换浅色或深色主题外观</span>
        </div>
        <div class="theme-switch">
          <label class="switch">
            <input type="checkbox" v-model="isDark" @change="handleThemeChange">
            <span class="slider round"></span>
          </label>
        </div>
      </div>

      <div class="setting-group">
        <div class="setting-item" :class="{ 'expanded': isDetailExpanded }">
          <div class="item-info">
            <span class="item-title">WSL 版本</span>
            <span class="item-desc">当前版本: <span class="version-tag">{{ wslVersion }}</span></span>
          </div>
          <div class="button-group">
            <button class="btn btn-secondary" @click="toggleDetail">
              {{ isDetailExpanded ? '🔼 收起信息' : 'ℹ️ 详细信息' }}
            </button>
            <!-- 暂时禁用检查更新功能
            <button class="btn" :class="updateBtnClass" @click="checkUpdate" :disabled="isChecking || updateStatus === 'success' || updateStatus === 'no-update'">
              <span v-if="isChecking" class="spinner-sm"></span>
              {{ updateBtnText }}
            </button>
            -->
          </div>
        </div>
        
        <Transition name="slide-fade">
          <div v-if="isDetailExpanded" class="setting-detail">
            <div v-if="loadingDetail" class="loading-state">
              正在获取详细信息...
            </div>
            <div v-else class="detail-content">
              <pre>{{ wslDetailInfo }}</pre>
            </div>
          </div>
        </Transition>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { THEME_KEY, setTheme } from '../utils/theme'
// Import backend functions (mocked if running in browser without wails)
import { GetWSLVersion, ShowWSLInfo } from '../../wailsjs/go/main/App'

const isDark = ref(true)
const wslVersion = ref('正在获取...')
const isDetailExpanded = ref(false)
const loadingDetail = ref(false)
const wslDetailInfo = ref('')

// Update logic
const isChecking = ref(false)
const updateStatus = ref('idle') // idle, success, no-update

const updateBtnText = computed(() => {
    if (isChecking.value) return '正在检查...'
    if (updateStatus.value === 'success') return '✅ 更新成功'
    if (updateStatus.value === 'no-update') return '✨ 无需更新'
    return '🔄 检查并更新'
})

const updateBtnClass = computed(() => {
    if (updateStatus.value === 'success') return 'btn-success'
    if (updateStatus.value === 'no-update') return 'btn-secondary' // or a specific disabled look
    return 'btn-primary'
})

const handleThemeChange = () => {
  const theme = isDark.value ? 'dark' : 'light'
  setTheme(theme)
}

const toggleDetail = async () => {
  isDetailExpanded.value = !isDetailExpanded.value
  
  if (isDetailExpanded.value && !wslDetailInfo.value) {
    loadingDetail.value = true
    try {
        const info = await ShowWSLInfo()
        wslDetailInfo.value = info
    } catch (e) {
        console.error("Failed to get WSL info:", e)
        wslDetailInfo.value = "获取详细信息失败: " + e
    } finally {
        loadingDetail.value = false
    }
  }
}

const checkUpdate = async () => {
  if (isChecking.value) return
  isChecking.value = true
  updateStatus.value = 'idle'
  
  try {
      // 模拟异步检查过程
      // 如果有后端函数: await CheckAndUpdateWSL()
      await new Promise(resolve => setTimeout(resolve, 2000))
      
      // 这里模拟一个随机结果，或者总是成功
      // 实际逻辑应根据后端返回决定
      const hasUpdate = Math.random() > 0.7 // 30% 概率有更新
      
      if (hasUpdate) {
          // 模拟更新过程
          await new Promise(resolve => setTimeout(resolve, 1500))
          updateStatus.value = 'success'
          // 更新版本显示
          wslVersion.value = "Latest"
      } else {
          updateStatus.value = 'no-update'
      }
  } catch (e) {
      console.error("Update check failed", e)
      alert("检查更新失败: " + e)
      updateStatus.value = 'idle'
  } finally {
      isChecking.value = false
      
      // 3秒后重置按钮状态，允许再次检查
      setTimeout(() => {
          updateStatus.value = 'idle'
      }, 5000)
  }
}

onMounted(async () => {
  const savedTheme = localStorage.getItem(THEME_KEY)
  // 如果没有保存的主题，读取当前的 data-theme 属性（由 App.vue 初始化）
  const currentTheme = savedTheme || document.documentElement.getAttribute('data-theme') || 'dark'
  isDark.value = currentTheme === 'dark'
  // 确保 DOM 状态同步 (双重保险)
  if (!document.documentElement.getAttribute('data-theme')) {
      document.documentElement.setAttribute('data-theme', currentTheme)
  }

  // 获取 WSL 版本
  try {
      const version = await GetWSLVersion()
      wslVersion.value = version
  } catch (e) {
      console.error("Failed to get WSL version:", e)
      wslVersion.value = "未知"
  }
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

.version-tag {
  font-family: monospace;
  background: var(--color-bg-tertiary);
  padding: 2px 6px;
  border-radius: 4px;
}

.button-group {
  display: flex;
  gap: var(--spacing-sm);
}

/* Toggle Switch - Copied from PerformanceConfig.vue for consistency */
.switch {
  position: relative;
  display: inline-block;
  width: 40px;
  height: 22px;
  flex-shrink: 0;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #d9d9d9;
  transition: .3s cubic-bezier(0.4, 0, 0.2, 1);
}

.slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 2px;
  bottom: 2px;
  background-color: white;
  transition: .3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.slider.round {
  border-radius: 22px;
}

.slider.round:before {
  border-radius: 50%;
}

input:checked + .slider {
  background-color: var(--color-brand);
}

input:checked + .slider:before {
  transform: translateX(18px);
}

/* Dark mode specific for toggle */
[data-theme='dark'] .slider {
    background-color: #4a4a4a;
}
[data-theme='dark'] input:checked + .slider {
    background-color: var(--color-brand);
}

.theme-switch {
  display: flex;
  align-items: center;
}

/* WSL Version Detail Styles */
.setting-group {
  display: flex;
  flex-direction: column;
}

.setting-item.expanded {
  border-bottom-left-radius: 0;
  border-bottom-right-radius: 0;
  background-color: var(--color-bg-hover); /* Keep hover color or slightly darker */
}

.setting-detail {
  background: var(--color-bg-secondary); /* Slightly different from item background */
  padding: var(--spacing-md);
  border-bottom-left-radius: var(--radius-md);
  border-bottom-right-radius: var(--radius-md);
  margin-top: 0;
  border-top: 1px solid var(--color-border); /* Optional separator */
  overflow: hidden;
}

.loading-state {
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  text-align: center;
  padding: var(--spacing-sm);
}

.detail-content pre {
  margin: 0;
  font-family: 'Consolas', 'Monaco', monospace;
  white-space: pre-wrap;
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  line-height: 1.6;
}

.spinner-sm {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: white;
  border-radius: 50%;
  display: inline-block;
  animation: spin 1s linear infinite;
  margin-right: 6px;
  vertical-align: middle;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.btn-success {
    background-color: var(--color-success);
    color: white;
    border: none;
    cursor: default;
}

/* Animations */
.slide-fade-enter-active,
.slide-fade-leave-active {
  transition: all 0.3s ease-out;
  max-height: 500px;
  opacity: 1;
}

.slide-fade-enter-from,
.slide-fade-leave-to {
  max-height: 0;
  opacity: 0;
  padding-top: 0;
  padding-bottom: 0;
}
</style>