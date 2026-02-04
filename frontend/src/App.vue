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
    <Transition name="page" mode="out-in">
      <KeepAlive>
        <component :is="views[currentTab]" :key="currentTab" />
      </KeepAlive>
    </Transition>
  </section>
</main>
  </div>
</template>

<style>
/* 引入全局设计系统 */
@import './assets/styles/main.css';

/* 移除旧的变量定义，使用 variables.css 中的定义 */
/* 原有的 :root 和 :root[data-theme='light'] 已迁移至 variables.css */

/* 兼容性适配：确保旧组件也能用到新变量 */
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
}

/* --- 侧边栏：专业深色风格 --- */
.sidebar {
  width: 240px;
  background: var(--color-bg-sidebar);
  color: var(--color-text-primary);
  display: flex;
  flex-direction: column;
  z-index: 100;
  border-right: 1px solid var(--color-border);
  transition: all var(--transition-normal);
}

.brand {
  padding: 40px 24px 20px;
  font-size: 1.1rem;
  font-weight: 700;
  letter-spacing: 0.5px;
  color: var(--color-text-primary); /* Adapt to theme */
}

.menu { 
  padding: 10px; 
  flex: 1;
}

.menu-item {
  padding: 10px 16px;
  margin-bottom: 4px;
  border-radius: var(--radius-md);
  cursor: pointer;
  font-size: var(--font-size-md);
  color: var(--color-text-secondary);
  transition: all var(--transition-fast);
  display: flex;
  align-items: center;
}

.menu-item:hover {
  background-color: var(--color-bg-hover);
  color: var(--color-text-primary);
}

.menu-item.active {
  background-color: var(--color-brand);  /* 激活项保持主题蓝色 */
  color: #fff;
  font-weight: 500;
  box-shadow: 0 4px 12px rgba(24, 144, 255, 0.3);
}

/* --- 主内容区：干净、通透 --- */
.main-body {
  background-color: var(--color-bg-body);
  color: var(--color-text-primary);
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0; /* 防止内容撑破 flex 布局 */
}

.content-area {
  flex: 1;
  padding: var(--spacing-xl); /* 增加留白，显得更大气 */
  overflow-y: auto;
  position: relative;
}

/* 隐藏滚动条美化 */
.content-area::-webkit-scrollbar {
  width: 6px;
}
.content-area::-webkit-scrollbar-thumb {
  background: var(--color-border-hover);
  border-radius: 10px;
}

/* --- 卡片美化（应用于子组件） --- */
.view-card {
  background: var(--color-bg-card);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border);
  box-shadow: var(--shadow-sm);
  padding: var(--spacing-lg);
  max-width: 1200px;
  margin: 0 auto;
  transition: box-shadow var(--transition-normal);
}

</style>