<script setup>
import { ref, onMounted } from 'vue'
import HomeView from './components/HomeView.vue'
import InstallView from './components/InstallView.vue'
import SettingView from './components/SettingView.vue'
import StartupChecks from './components/StartupChecks.vue'
import PerformanceConfig from './components/PerformanceConfig.vue'
import { initTheme } from './utils/theme'
import { EventsOn, BrowserOpenURL } from './wailsjs/runtime/runtime'
import { Home, Download, Settings, Tag, X, Info, Cpu } from 'lucide-vue-next'

// 初始化主题 (修复 1.1 - 1.3)
initTheme()

// 启动检查状态
const checksComplete = ref(false)

// 当前选中的标签
const currentTab = ref('install')

// 建立 ID 到 组件的映射
const views = {
  home: HomeView,
  install: InstallView,
  performance: PerformanceConfig, // 注册为独立视图
  setting: SettingView
}

// --- Version Control Logic ---
const appVersion = ref('v1.0.0') // Default version
const hasNewVersion = ref(false)
const newVersionInfo = ref({
  version: '',
  updateLog: '',
  releaseDate: '',
  url: ''
})
const showUpdateModal = ref(false)

// Expose bindings for backend to call directly (Requirement 5)
// Note: While EventsOn is preferred, we expose these as requested.
window.WindowSetNewVersionAvailable = (version, updateLog, releaseDate) => {
  hasNewVersion.value = true
  newVersionInfo.value = {
    version,
    updateLog,
    releaseDate,
    url: '' // Will be populated if available or handled via event
  }
}

window.WindowClearNewVersionFlag = () => {
  hasNewVersion.value = false
  showUpdateModal.value = false
}

// Handle opening the modal
const handleVersionClick = () => {
  if (hasNewVersion.value) {
    showUpdateModal.value = true
  }
}

const closeUpdateModal = () => {
  showUpdateModal.value = false
  // Note: Requirement says "Do not close red dot status", so we don't set hasNewVersion to false
}

const goToUpdate = () => {
  if (newVersionInfo.value.url) {
    BrowserOpenURL(newVersionInfo.value.url)
  } else {
    // Fallback or handle if URL is not provided in payload
    // Ideally backend provides it.
    // For now, we can log or try a default.
    console.log("No update URL provided")
  }
}

onMounted(() => {
  // Requirement 6: Listen to 'new-version' event
  EventsOn("new-version", (data) => {
    console.log("New version event received:", data)
    if (data) {
      hasNewVersion.value = true
      newVersionInfo.value = {
        version: data.version,
        updateLog: data.updateLog,
        releaseDate: data.releaseDate,
        url: data.url || '' // Assume URL might be passed
      }
    }
  })
})

</script>

<template>
  <div class="app-wrapper">
    <!-- 启动检查覆盖层 -->
    <Transition name="fade">
      <StartupChecks v-if="!checksComplete" @complete="checksComplete = true" />
    </Transition>

    <aside class="sidebar">
      <div class="brand">
        <span class="brand-text">WSL-Manager</span>
      </div>
      
      <nav class="menu">
        <div 
          :class="['menu-item', { active: currentTab === 'home' }]" 
          @click="currentTab = 'home'"
        >
          <Home class="menu-icon" :size="20" />
          <span class="menu-text">首页</span>
        </div>
        <div 
          :class="['menu-item', { active: currentTab === 'install' }]" 
          @click="currentTab = 'install'"
        >
          <Download class="menu-icon" :size="20" />
          <span class="menu-text">安装</span>
        </div>
        
        <!-- Performance Config Entry -->
        <div 
          :class="['menu-item', { active: currentTab === 'performance' }]" 
          @click="currentTab = 'performance'"
        >
          <Cpu class="menu-icon" :size="20" />
          <span class="menu-text">性能配置</span>
        </div>

        <div 
          :class="['menu-item', { active: currentTab === 'setting' }]" 
          @click="currentTab = 'setting'"
        >
          <Settings class="menu-icon" :size="20" />
          <span class="menu-text">设置</span>
        </div>
      </nav>

      <!-- Version Info Area -->
      <div class="version-area" @click="handleVersionClick" :class="{ 'clickable': hasNewVersion }">
        <div class="version-content">
          <Tag class="version-icon" :size="16" />
          <span class="version-text">{{ appVersion }}</span>
        </div>
        <div v-if="hasNewVersion" class="version-badge"></div>
      </div>
    </aside>

    <main class="main-body">
      <section class="content-area">
        <Transition name="page" mode="out-in">
          <KeepAlive>
            <component :is="views[currentTab]" :key="currentTab" />
          </KeepAlive>
        </Transition>
      </section>
    </main>

    <!-- Update Modal -->
    <Transition name="modal">
      <div v-if="showUpdateModal" class="modal-overlay" @click.self="closeUpdateModal">
        <div class="modal-window update-modal">
          <div class="modal-header">
            <h3>版本更新</h3>
            <button class="close-btn" @click="closeUpdateModal">
              <X :size="20" />
            </button>
          </div>
          <div class="modal-body">
            <div class="update-meta">
              <span class="new-version">新版本: {{ newVersionInfo.version }}</span>
              <span class="release-date">{{ newVersionInfo.releaseDate }}</span>
            </div>
            <div class="update-log">
              <p v-for="(line, index) in newVersionInfo.updateLog.split('\n')" :key="index">
                {{ line }}
              </p>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="closeUpdateModal">关闭</button>
            <button class="btn btn-primary" @click="goToUpdate">前往更新</button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style>
/* 引入全局设计系统 */
@import './assets/styles/main.css';

/* 兼容性适配 */
.modal-window {
  background: var(--color-bg-modal);
  color: var(--color-text-primary);
}

body {
  font-family: var(--font-family-base);
  margin: 0;
  background-color: var(--color-bg-body);
}

.app-wrapper {
  display: flex;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
}

/* --- 侧边栏优化 --- */
.sidebar {
  width: 240px;
  background: var(--color-bg-sidebar);
  color: var(--color-text-primary);
  display: flex;
  flex-direction: column;
  z-index: 100;
  border-right: 1px solid var(--color-border);
  transition: all var(--transition-normal);
  flex-shrink: 0;
}

.brand {
  height: 80px; /* Fixed height for consistency */
  display: flex;
  align-items: center;
  padding: 0 24px;
  border-bottom: 1px solid transparent; /* Placeholder for divider if needed */
}

.brand-text {
  font-size: 1.2rem;
  font-weight: 700;
  letter-spacing: 0.5px;
  color: var(--color-text-primary);
}

.menu { 
  padding: 16px; 
  flex: 1; /* Pushes version info to bottom */
  display: flex;
  flex-direction: column;
  gap: 8px; /* Spacing between items */
}

.menu-item {
  height: 48px; /* >= 44px clickable area */
  padding: 0 16px;
  border-radius: var(--radius-md);
  cursor: pointer;
  font-size: var(--font-size-md);
  color: var(--color-text-secondary);
  transition: all var(--transition-fast);
  display: flex;
  align-items: center;
  gap: 12px; /* Icon and text spacing */
}

.menu-item:hover {
  background-color: var(--color-bg-hover);
  color: var(--color-text-primary);
}

.menu-item.active {
  background-color: var(--color-brand);
  color: #fff;
  font-weight: 500;
  box-shadow: 0 4px 12px rgba(24, 144, 255, 0.3);
}

.menu-icon {
  opacity: 0.8;
}

.menu-item.active .menu-icon {
  opacity: 1;
}

/* --- Version Info Area --- */
.version-area {
  padding: 16px 24px;
  border-top: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: default;
  transition: background-color var(--transition-fast);
  height: 60px; /* Consistent height */
}

.version-area.clickable {
  cursor: pointer;
}

.version-area.clickable:hover {
  background-color: var(--color-bg-hover);
}

.version-content {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}

.version-badge {
  width: 6px;
  height: 6px;
  background-color: #ff4d4f; /* Red dot */
  border-radius: 50%;
  box-shadow: 0 0 4px rgba(255, 77, 79, 0.5);
}

/* --- 主内容区 --- */
.main-body {
  background-color: var(--color-bg-body);
  color: var(--color-text-primary);
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.content-area {
  flex: 1;
  padding: 32px; /* Unified spacing */
  overflow-y: auto;
  position: relative;
}

/* Scrollbar Styling */
.content-area::-webkit-scrollbar {
  width: 6px;
}
.content-area::-webkit-scrollbar-thumb {
  background: var(--color-border-hover);
  border-radius: 10px;
}

/* --- Modal Styles --- */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.update-modal {
  width: 480px;
  max-width: 90%;
  display: flex;
  flex-direction: column;
  background: var(--color-bg-card);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  overflow: hidden;
  border: 1px solid var(--color-border);
}

.modal-header {
  padding: 20px 24px;
  border-bottom: 1px solid var(--color-border);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.2rem;
  color: var(--color-text-primary);
}

.close-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-text-secondary);
  padding: 4px;
  border-radius: 4px;
  transition: background var(--transition-fast);
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-btn:hover {
  background: var(--color-bg-hover);
  color: var(--color-text-primary);
}

.modal-body {
  padding: 24px;
  max-height: 400px;
  overflow-y: auto;
}

.update-meta {
  display: flex;
  justify-content: space-between;
  margin-bottom: 16px;
  font-weight: 600;
  color: var(--color-brand);
}

.release-date {
  color: var(--color-text-secondary);
  font-weight: normal;
}

.update-log {
  color: var(--color-text-primary);
  line-height: 1.6;
  white-space: pre-wrap; /* Preserve newlines */
  font-size: 0.95rem;
}

.modal-footer {
  padding: 16px 24px;
  border-top: 1px solid var(--color-border);
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  background: var(--color-bg-body); /* Slightly different bg for footer */
}

.btn {
  padding: 8px 20px;
  border-radius: var(--radius-md);
  border: none;
  cursor: pointer;
  font-size: 0.95rem;
  transition: all var(--transition-fast);
  font-weight: 500;
}

.btn-secondary {
  background: var(--color-bg-hover);
  color: var(--color-text-primary);
}

.btn-secondary:hover {
  background: var(--color-border-hover);
}

.btn-primary {
  background: var(--color-brand);
  color: white;
}

.btn-primary:hover {
  opacity: 0.9;
  transform: translateY(-1px);
}

/* Animations */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
  transform: scale(0.95);
}
</style>
