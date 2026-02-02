<script setup>

import { ref } from 'vue'
import HomeView from './components/HomeView.vue'
import InstallView from './components/InstallView.vue'
import SettingView from './components/SettingView.vue'
// 当前选中的标签
const currentTab = ref('install')

// 建立 ID 到 组件的映射
const views = {
  home: HomeView,
  install: InstallView,
  setting: SettingView
}

</script>

<template>
  <div class="app-wrapper">
    <aside class="sidebar">
      <div class="brand">WSL-Manager</div>
      <nav class="menu">
        <div 
          :class="['menu-item', { active: currentTab === 'home' }]" 
          @click="currentTab = 'home'"
        >首页</div>
        <div 
          :class="['menu-item', { active: currentTab === 'install' }]" 
          @click="currentTab = 'install'"
        >安装</div>
        <div 
          :class="['menu-item', { active: currentTab === 'setting' }]" 
          @click="currentTab = 'setting'"
        >设置</div>
      </nav>
    </aside>

<main class="main-body">
  <section class="content-area">
    <KeepAlive>
      <component :is="views[currentTab]" :key="currentTab" />
    </KeepAlive>
  </section>
</main>
  </div>
</template>

<style>

/* 默认暗色模式变量 */
:root {
  --main-text: #eeeeee;
  --main-bg: #121212;
  --card-bg: #252526;
  --card-title: #ffffff;
  --card-desc: #bbbbbb;
  --brand-color: #ffffff;
  --accent-color: #1890ff;
  --sidebar-bg: #1e1e1e;       /* 侧边栏背景 */
  --sidebar-text: #ffffff;     /* 侧边栏文字 */
  --sidebar-hover: #2d2d2d;    /* 悬停背景 */
}

/* 亮色模式自动切换 */
:root[data-theme='light'] {
  --sidebar-bg: #f3f3f3;       /* 浅灰色背景 */
  --sidebar-text: #333333;     /* 深色文字 */
  --sidebar-hover: #e0e0e0;    /* 稍微深一点的悬停色 */
  --main-text: #222222;
  --main-bg: #f5f5f7;
  --card-bg: #ffffff;
  --card-title: #000000;
  --card-desc: #666666;
  --brand-color: #1a1a1a;
}
/* 将原来组件中的硬编码颜色改为变量 */
.modal-window {
  background: var(--sidebar-bg);
  color: var(--text-color);
}

body {
  font-family: "Segoe UI Variable Text", "Segoe UI", -apple-system, sans-serif;
  margin: 0;
  background-color: var(--body-bg);
}

.app-wrapper {
  display: flex;
  width: 100vw;
  height: 100vh;
}

/* --- 侧边栏：专业深色风格 --- */
.sidebar {
  width: 240px;
  background: var(--sidebar-bg);
  color: var(--sidebar-text);
  display: flex;
  flex-direction: column;
  z-index: 100;
  transition: all 0.3s ease;
}

.brand {
  padding: 40px 24px 20px;
  font-size: 1.1rem;
  font-weight: 700;
  letter-spacing: 0.5px;
  color: var(--brand-color);
}

.menu { 
  
  padding: 10px; 
  flex: 1;
}

.menu-item {
  padding: 10px 16px;
  margin-bottom: 4px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  color: #bbbbbb;
  transition: background 0.15s, color 0.15s;
  display: flex;
  align-items: center;
}

.menu-item:hover {
  background-color: var(--sidebar-hover);
  color: white;
}

.menu-item.active {
  background-color: var(--accent-color);  /* 激活项保持主题蓝色 */
  color: white;
  font-weight: 500;
  /* 左侧激活指示条 */
  box-shadow: inset 4px 0 0 var(--accent-color);
}

/* --- 主内容区：干净、通透 --- */
.main-body {
  background-color: var(--main-bg);
  color: var(--main-text);
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0; /* 防止内容撑破 flex 布局 */
}

.content-area {
  flex: 1;
  padding: 32px; /* 增加留白，显得更大气 */
  overflow-y: auto;
  position: relative;
}

/* 隐藏滚动条美化 */
.content-area::-webkit-scrollbar {
  width: 6px;
}
.content-area::-webkit-scrollbar-thumb {
  background: #ddd;
  border-radius: 10px;
}

/* --- 卡片美化（应用于子组件） --- */
.view-card {
  background: #ffffff;
  border-radius: 8px;
  border: 1px solid #e0e0e0;
  box-shadow: var(--card-shadow);
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
}
</style>