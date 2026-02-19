<script setup>
import { ref, onMounted } from 'vue'
import HomeView from './components/HomeView.vue'
import InstallView from './components/InstallView.vue'
import SettingView from './components/SettingView.vue'
import StartupChecks from './components/StartupChecks.vue'
import PerformanceConfig from './components/PerformanceConfig.vue'
import ConfigView from './components/ConfigView.vue'
import EnvironmentView from './components/EnvironmentView.vue'
import { initTheme } from './utils/theme'
import { GetAppVersion } from './wailsjs/go/main/App'
import { EventsOn, BrowserOpenURL } from './wailsjs/runtime/runtime'
import { Home, Download, Settings, Tag, X, Info, Cpu, Sliders, LayoutGrid, Github } from 'lucide-vue-next'
import { useAppStore } from './stores/app'

// 初始化主题 (修复 1.1 - 1.3)
initTheme()

const appStore = useAppStore()

// 启动检查状态
const checksComplete = ref(false)

// 建立 ID 到 组件的映射
const views = {
  home: HomeView,
  install: InstallView,
  performance: PerformanceConfig, // 注册为独立视图
  config: ConfigView,
  environment: EnvironmentView,
  setting: SettingView
}

// --- Version Control Logic ---
const appVersion = ref('v1.0.0 Beta')
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

const openGithub = (e) => {
  e.stopPropagation()
  BrowserOpenURL("https://github.com/qwq12738qwq/Easy-WSL-GUI")
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

onMounted(async () => {
  // 获取当前版本号
  try {
    const ver = await GetAppVersion()
    if (ver) {
      appVersion.value = ver
    }
  } catch (e) {
    console.error("获取版本信息失败:", e)
  }
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
        <span class="brand-text">Easy-WSL-GUI</span>
      </div>
      
      <nav class="menu">
        <div 
          :class="['menu-item', { active: appStore.currentTab === 'home' }]" 
          @click="appStore.setCurrentTab('home')"
        >
          <Home class="menu-icon" :size="20" />
          <span class="menu-text">首页</span>
        </div>
        <div 
          :class="['menu-item', { active: appStore.currentTab === 'install' }]" 
          @click="appStore.setCurrentTab('install')"
        >
          <Download class="menu-icon" :size="20" />
          <span class="menu-text">安装</span>
        </div>
        
        <!-- Performance Config Entry -->
        <div 
          :class="['menu-item', { active: appStore.currentTab === 'performance' }]" 
          @click="appStore.setCurrentTab('performance')"
        >
          <Cpu class="menu-icon" :size="20" />
          <span class="menu-text">性能配置</span>
        </div>

        <!-- Distro Config Entry -->
        <div 
          :class="['menu-item', { active: appStore.currentTab === 'config' }]" 
          @click="appStore.selectDistro(''); appStore.setCurrentTab('config')"
        >
          <Sliders class="menu-icon" :size="20" />
          <span class="menu-text">发行版配置</span>
        </div>

        <!-- Environment Market Entry -->
        <div 
          :class="['menu-item', { active: appStore.currentTab === 'environment' }]" 
          @click="appStore.setCurrentTab('environment')"
        >
          <LayoutGrid class="menu-icon" :size="20" />
          <span class="menu-text">环境市场</span>
        </div>

        <div 
          :class="['menu-item', { active: appStore.currentTab === 'setting' }]" 
          @click="appStore.setCurrentTab('setting')"
        >
          <Settings class="menu-icon" :size="20" />
          <span class="menu-text">设置</span>
        </div>
      </nav>

      <!-- Version Info Area -->
      <div class="version-area">
        <div class="version-left" @click="handleVersionClick" :class="{ 'clickable': hasNewVersion }">
          <span class="version-text">
            {{ appVersion }}
            <div v-if="hasNewVersion" class="version-badge"></div>
          </span>
        </div>
        <button class="github-btn" @click="openGithub" title="GitHub Repository">
          <Github :size="18" />
        </button>
      </div>
    </aside>

    <main class="main-body">
      <section class="content-area">
        <Transition name="page" mode="out-in">
          <KeepAlive>
            <component :is="views[appStore.currentTab]" :key="appStore.currentTab" />
          </KeepAlive>
        </Transition>
      </section>
    </main>

    <!-- Update Modal -->
    <Transition name="modal">
      <div v-if="showUpdateModal" class="modal-overlay" @click.self="closeUpdateModal">
        <div class="modal-window update-modal">
          <div class="modal-header">
            <h3>{{ newVersionInfo.version }}</h3>
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
  width: 250px;
  background: var(--color-bg-sidebar);
  color: var(--color-text-primary);
  display: flex;
  flex-direction: column;
  z-index: 100;
  border-right: 1px solid var(--color-border);
  transition: all var(--transition-normal);
  flex-shrink: 0;
  box-shadow: 4px 0 24px rgba(0, 0, 0, 0.02);
}

.brand {
  height: 80px; /* Fixed height for consistency */
  display: flex;
  align-items: center;
  padding: 0 24px;
  /* border-bottom: 1px solid var(--color-border); */
}

.brand-text {
  font-size: 1.25rem;
  font-weight: 700;
  letter-spacing: -0.5px;
  background: linear-gradient(135deg, var(--color-brand), var(--color-brand-active));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.menu { 
  padding: 24px 16px; 
  flex: 1; /* Pushes version info to bottom */
  display: flex;
  flex-direction: column;
  gap: 6px; /* Spacing between items */
}

.menu-item {
  height: 46px; /* >= 44px clickable area */
  padding: 0 16px;
  border-radius: var(--radius-md);
  cursor: pointer;
  font-size: var(--font-size-md);
  color: var(--color-text-secondary);
  transition: all var(--transition-fast);
  display: flex;
  align-items: center;
  gap: 12px; /* Icon and text spacing */
  position: relative;
  overflow: hidden;
}

.menu-item:hover {
  background-color: var(--color-bg-hover);
  color: var(--color-text-primary);
  transform: translateX(4px);
}

.menu-item.active {
  background-color: rgba(24, 144, 255, 0.1); /* Light brand color */
  color: var(--color-brand);
  font-weight: 600;
}

.menu-item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  height: 20px;
  width: 4px;
  background-color: var(--color-brand);
  border-radius: 0 4px 4px 0;
}

.menu-icon {
  opacity: 0.7;
  transition: opacity var(--transition-fast);
}

.menu-item.active .menu-icon {
  opacity: 1;
}

.menu-item:hover .menu-icon {
  opacity: 0.9;
}

/* --- Version Info Area --- */
.version-area {
  padding: 16px 24px;
  border-top: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
}

.version-left {
  position: relative;
  display: flex;
  align-items: center;
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}

.version-left.clickable {
  cursor: pointer;
}

.version-left.clickable:hover .version-text {
  color: var(--color-text-primary);
}

.version-text {
  position: relative;
  display: inline-block;
}

.version-badge {
  position: absolute;
  bottom: -4px;
  left: -4px;
  width: 6px;
  height: 6px;
  background-color: #ff4d4f; /* Red dot */
  border-radius: 50%;
  box-shadow: 0 0 0 2px var(--color-bg-sidebar);
  z-index: 10;
}

.github-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-text-secondary);
  padding: 8px;
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
  display: flex;
  align-items: center;
  justify-content: center;
}

.github-btn:hover {
  background-color: var(--color-bg-hover);
  color: var(--color-text-primary);
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
  /* backdrop-filter: blur(4px); Removed blur */
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
