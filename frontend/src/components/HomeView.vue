<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { GetDistroStats, GetPath, GetMetrics, UninstallDistro } from '../../wailsjs/go/main/App'

const distros = ref([])
const isInitialLoading = ref(true)

// --- 卸载模态框相关状态 ---
const showUninstallModal = ref(false)
const uninstallTarget = ref('')
const uninstallStepIndex = ref(0)
const isUninstalling = ref(false)

// 定义卸载流程步骤
const uninstallSteps = [
  { title: '确认操作', status: 'pending' },
  { title: '停止实例', status: 'pending' },
  { title: '注销分发', status: 'pending' },
  { title: '清理磁盘', status: 'pending' }
]

// 保持原有的数据同步逻辑
const syncData = async () => {
  try {
    const backendList = await GetDistroStats()
    if (!backendList) { distros.value = []; return }
    // 前端防重保护：使用 Map 去重（即使后端发了重复的，前端也能过滤）
    const uniqueBackendMap = new Map();
    backendList.forEach(item => {
        if(item.name) uniqueBackendMap.set(item.name, item);
    });
    // 将 Map 转回数组进行后续处理
    const uniqueList = Array.from(uniqueBackendMap.values());

    const backendNames = uniqueList.map(i => i.name)
    distros.value = distros.value.filter(d => backendNames.includes(d.name))

    await Promise.all(uniqueList.map(async (item) => {
      let localItem = distros.value.find(d => d.name === item.name)
      if (!localItem) {
        let path = await GetPath(item.name).catch(() => 'N/A')
        path = (path && path.trim() !== "") ? path : 'N/A'
        localItem = { ...item, path, stats: { cpu: '0%', memUsed: '0', memTotal: '0', disk: '0%' } }
        distros.value.push(localItem)
      } else {
        localItem.status = item.status
        localItem.version = item.version
      }
      // 获取指标逻辑...
      if (localItem.status === 'Running') {
        try {
          const m = await GetMetrics(localItem.name)
          localItem.stats.cpu = m?.cpu || '0%'
          localItem.stats.memUsed = m?.memUsed || '0'
          localItem.stats.memTotal = m?.memTotal || '0'
          localItem.stats.disk = m?.disk || '0%'
        } catch (e) { localItem.stats.cpu = '0%' }
      } else { localItem.stats.cpu = '0%' }
    }))
  } finally { isInitialLoading.value = false }
}

let timer = null
onMounted(() => {
  syncData()
  timer = setInterval(syncData, 3000)
})
onUnmounted(() => clearInterval(timer))

// --- 卸载逻辑控制 ---

const handleUninstallClick = (name) => {
  uninstallTarget.value = name
  uninstallStepIndex.value = 0
  isUninstalling.value = false
  // 重置步骤状态
  uninstallSteps.forEach(s => s.status = 'pending')
  uninstallSteps[0].status = 'processing'
  showUninstallModal.value = true
}

const closeUninstallModal = () => {
  if (isUninstalling.value) return
  showUninstallModal.value = false
}

const confirmUninstall = async () => {
  isUninstalling.value = true
  
  try {
    // 步骤 1: 确认完成，开始停止
    uninstallSteps[0].status = 'finished'
    uninstallStepIndex.value = 1
    uninstallSteps[1].status = 'processing'
    await new Promise(r => setTimeout(r, 800)) // UI 模拟耗时

    // 步骤 2: 停止完成，调用后端
    uninstallSteps[1].status = 'finished'
    uninstallStepIndex.value = 2
    uninstallSteps[2].status = 'processing'
    
    await UninstallDistro(uninstallTarget.value)
    
    // 步骤 3: 注销完成，清理UI
    uninstallSteps[2].status = 'finished'
    uninstallStepIndex.value = 3
    uninstallSteps[3].status = 'processing'
    await new Promise(r => setTimeout(r, 800)) // UI 模拟耗时

    // 全部完成
    uninstallSteps[3].status = 'finished'
    
    // 关闭并刷新
    setTimeout(() => {
        showUninstallModal.value = false
        syncData()
    }, 500)
    
  } catch (err) {
    alert("卸载失败: " + err)
    uninstallSteps[uninstallStepIndex.value].status = 'error'
  } finally {
    isUninstalling.value = false
  }
}

const getDistroIcon = (name) => {
  const n = name.toLowerCase()
  if (n.includes('ubuntu')) return '../src/assets/icons/UbuntuCoF.png'
  if (n.includes('debian')) return '../src/assets/icons/Debian.png'
  if (n.includes('kali'))   return '../src/assets/icons/Kali-drago.png'
  if (n.includes('arch'))   return '../src/assets/icons/Arch.png'
  return '../src/assets/icons/UbuntuCoF.png' 
}
</script>

<template>
  <div class="home-view-container">
    <header class="view-header">
      <div class="status-tag">
        <span class="status-dot-static"></span> 
        <span class="status-text">系统监控已就绪</span>
      </div>
    </header>

    <div v-if="isInitialLoading" class="loading-grid">
      <div v-for="i in 3" :key="i" class="skeleton-card"></div>
    </div>

    <div v-else class="distro-grid">
      <div v-for="item in distros" :key="item.name" class="distro-card" :class="{ 'running': item.status === 'Running' }">
        <button class="action-uninstall" @click="handleUninstallClick(item.name)" title="卸载">×</button>
        
        <div class="card-main">
          <img :src="getDistroIcon(item.name)" class="distro-icon" />
          <div class="info-content">
            <div class="name-row">
              <span class="name">{{ item.name }}</span>
              <span class="version">v{{ item.version }}</span>
            </div>
            <div class="path-text">{{ item.path }}</div>
          </div>
        </div>

        <div class="metrics-box" v-if="item.status === 'Running'">
          <div class="metric-row">
            <div class="label">CPU <span>{{ item.stats.cpu }}</span></div>
            <div class="progress"><div class="bar cpu-bar" :style="{ width: item.stats.cpu }"></div></div>
          </div>
          <div class="metric-row">
            <div class="label">内存 <span>{{ item.stats.memUsed }}/{{ item.stats.memTotal }}</span></div>
            <div class="progress">
              <div class="bar mem-bar" :style="{ width: (parseFloat(item.stats.memUsed)/parseFloat(item.stats.memTotal)*100 || 0) + '%' }"></div>
            </div>
          </div>
          <div class="disk-info">磁盘占用: {{ item.stats.disk }}</div>
        </div>
        
        <div class="offline-placeholder" v-else>
          发行版已停止
        </div>
      </div>
    </div>

    <div v-if="showUninstallModal" class="modal-overlay">
      <div class="modal-window">
        <div class="modal-header">
          <span>卸载向导 - {{ uninstallTarget }}</span>
          <button v-if="!isUninstalling" class="close-btn" @click="closeUninstallModal">✕</button>
        </div>
        
        <div class="modal-body">
            <div class="warning-section">
                <div class="warning-icon">⚠️</div>
                <div class="warning-content">
                    <h4>危险操作警告</h4>
                    <p>您即将卸载 <strong>{{ uninstallTarget }}</strong>。此操作将永久删除该发行版及其所有数据（文件、配置、软件）。</p>
                </div>
            </div>

            <div class="steps-container">
                 <div v-for="(step, index) in uninstallSteps" :key="index" 
                     class="step-item" 
                     :class="step.status">
                    <div class="step-icon">
                        <span v-if="step.status === 'finished'">✓</span>
                        <span v-else-if="step.status === 'processing'" class="spinner"></span>
                        <span v-else-if="step.status === 'error'">!</span>
                        <span v-else>{{ index + 1 }}</span>
                    </div>
                    <span class="step-title">{{ step.title }}</span>
                    <div v-if="index < uninstallSteps.length - 1" class="step-line" :class="{ 'line-active': step.status === 'finished' }"></div>
                </div>
            </div>

            <div class="action-bar">
                <button class="cancel-btn" @click="closeUninstallModal" :disabled="isUninstalling">取消</button>
                <button class="danger-btn" @click="confirmUninstall" :disabled="isUninstalling">
                    {{ isUninstalling ? '正在卸载...' : '确认卸载' }}
                </button>
            </div>
        </div>
      </div>
    </div>

  </div>
</template>

<style scoped>
.status-indicator { font-size: 12px; color: #666; display: flex; align-items: center; gap: 6px; }
.distro-card:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.08); border-color: #1890ff; }
.distro-card.running { border-left: 4px solid #52c41a; }
.action-uninstall { position: absolute; top: 8px; right: 8px; border: none; background: none; color: #ccc; font-size: 18px; cursor: pointer; line-height: 1; }
.action-uninstall:hover { color: #ff4d4f; }
.card-main { display: flex; gap: 12px; margin-bottom: 15px; }
.distro-icon { width: 40px; height: 40px; object-fit: contain; }
.name { font-weight: bold; color: #333; font-size: 16px; }
.version { font-size: 10px; background: #f0f0f0; padding: 2px 6px; border-radius: 4px; margin-left: 8px; color: #888; }
.path-text { font-size: 11px; color: #999; margin-top: 4px; word-break: break-all; font-family: monospace; }
.metrics-box { display: flex; flex-direction: column; gap: 10px; }
.metric-row .label { display: flex; justify-content: space-between; font-size: 12px; margin-bottom: 4px; color: #666; }
.metric-row .label span { font-family: monospace; color: #333; }
.progress { height: 6px; background: #f0f0f0; border-radius: 3px; overflow: hidden; }
.bar { height: 100%; transition: width 0.8s cubic-bezier(0.4, 0, 0.2, 1); }
.cpu-bar { background: linear-gradient(90deg, #1890ff, #36cfc9); }
.mem-bar { background: linear-gradient(90deg, #722ed1, #b37feb); }
.disk-info { font-size: 11px; color: #aaa; margin-top: 5px; text-align: right; }
.offline-placeholder { text-align: center; padding: 20px; color: #bbb; font-style: italic; font-size: 13px; border: 1px dashed #eee; border-radius: 8px; }
/* --- 容器与头部 --- */
.home-view-container { 
  display: flex; 
  flex-direction: column; 
  gap: 24px; 
}

.view-header { 
  display: flex; 
  justify-content: flex-end; 
  margin-bottom: 4px; 
}

/* 替代呼吸灯：静态精致标签 */
.status-tag {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 12px;
  background: rgba(82, 196, 26, 0.1); /* 极淡的绿色背景 */
  border: 1px solid rgba(82, 196, 26, 0.2);
  border-radius: 20px;
}

.status-dot-static {
  width: 6px;
  height: 6px;
  background: #52c41a;
  border-radius: 50%;
}

.status-text {
  font-size: 12px;
  color: #52c41a;
  font-weight: 500;
}

/* --- 卡片美化 --- */
.distro-grid { 
  display: grid; 
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); 
  gap: 20px; 
}

.distro-card { 
  background: var(--card-bg); 
  border: 1px solid rgba(0, 0, 0, 0.05);
  border-radius: 12px; 
  padding: 20px; 
  position: relative; 
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

/* 深色模式下的微调 */
:root[data-theme='dark'] .distro-card {
  border: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.distro-card:hover { 
  transform: translateY(-2px);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.1);
  border-color: var(--accent-color); 
}

/* 运行状态指示：改为精致的左侧边条 */
.distro-card.running::before {
  content: "";
  position: absolute;
  left: 0;
  top: 20%;
  height: 60%;
  width: 4px;
  background: #52c41a;
  border-radius: 0 4px 4px 0;
}

/* 进度条美化：增加圆角和高度 */
.progress { 
  height: 8px; 
  background: rgba(0, 0, 0, 0.05); 
  border-radius: 4px; 
}

:root[data-theme='dark'] .progress {
  background: rgba(255, 255, 255, 0.1);
}
/* === 模态框样式 (移植自 InstallView，实现深色风格统一) === */
.modal-overlay {
  position: fixed; top: 0; left: 0;
  width: 100vw; height: 100vh;
  background: rgba(0, 0, 0, 0.7);
  display: flex; justify-content: center; align-items: center;
  z-index: 999;
  backdrop-filter: blur(2px);
}

.modal-window {
  width: 550px;
  background: #1e1e1e; /* 深色背景 */
  border-radius: 10px;
  overflow: hidden;
  box-shadow: 0 15px 40px rgba(0,0,0,0.6);
  display: flex; flex-direction: column;
  border: 1px solid #333;
}

.modal-header {
  padding: 15px 20px;
  background: #252526; /* 头部更深 */
  color: #fff;
  display: flex; justify-content: space-between; align-items: center;
  font-size: 15px;
  font-weight: 600;
  border-bottom: 1px solid #333;
}

.close-btn {
  background: transparent; border: none; color: #888;
  font-size: 16px; cursor: pointer; padding: 0 5px;
}
.close-btn:hover { color: #fff; }

.modal-body {
    padding: 30px;
    color: #ccc;
    display: flex;
    flex-direction: column;
    gap: 25px;
}

/* 警告区域 */
.warning-section {
    display: flex;
    gap: 15px;
    background: #2a1215; /* 深红背景 */
    border: 1px solid #5c2526;
    padding: 15px;
    border-radius: 6px;
}
.warning-icon { font-size: 24px; }
.warning-content h4 { margin: 0 0 5px 0; color: #ff4d4f; font-size: 14px; }
.warning-content p { margin: 0; font-size: 12px; color: #aaa; line-height: 1.5; }
.warning-content strong { color: #ffccc7; }

/* 步骤条样式 (完全移植) */
.steps-container {
    display: flex;
    justify-content: space-between;
    margin-top: 10px;
    position: relative;
    padding: 0 10px;
}

.step-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    position: relative;
    flex: 1;
    z-index: 2;
}

.step-icon {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    background: #333;
    border: 2px solid #444;
    color: #888;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 12px;
    font-weight: bold;
    margin-bottom: 8px;
    transition: all 0.3s;
}

.step-title { font-size: 12px; color: #666; transition: color 0.3s; }

/* 状态：进行中 */
.step-item.processing .step-icon {
    border-color: #1890ff;
    color: #1890ff;
    background: #1e1e1e;
    box-shadow: 0 0 8px rgba(24, 144, 255, 0.2);
}
.step-item.processing .step-title { color: #fff; }

/* 状态：完成 */
.step-item.finished .step-icon {
    background: #1890ff;
    border-color: #1890ff;
    color: #fff;
}
.step-item.finished .step-title { color: #ccc; }

/* 状态：错误 */
.step-item.error .step-icon {
    border-color: #ff4d4f;
    color: #ff4d4f;
}

/* 连接线 */
.step-line {
    position: absolute;
    top: 14px;
    left: 50%;
    width: 100%;
    height: 2px;
    background: #333;
    z-index: -1;
}
.step-item:last-child .step-line { display: none; }
.step-line.line-active { background: #1890ff; }

/* 按钮栏 */
.action-bar { display: flex; justify-content: flex-end; gap: 12px; margin-top: 10px; }

.cancel-btn {
    padding: 8px 20px;
    background: transparent;
    border: 1px solid #555;
    color: #ccc;
    border-radius: 4px;
    cursor: pointer;
}
.cancel-btn:hover:not(:disabled) { background: #333; color: #fff; }
.cancel-btn:disabled { opacity: 0.5; cursor: not-allowed; }

/* 危险按钮 (红色) */
.danger-btn {
    padding: 8px 20px;
    background: #ff4d4f;
    border: none;
    color: white;
    border-radius: 4px;
    cursor: pointer;
    font-weight: 500;
}
.danger-btn:hover:not(:disabled) { background: #ff7875; }
.danger-btn:disabled { opacity: 0.6; cursor: not-allowed; background: #a8071a; }

/* Spinner 动画 */
.spinner {
    width: 14px;
    height: 14px;
    border: 2px solid transparent;
    border-top-color: #1890ff;
    border-radius: 50%;
    animation: spin 1s linear infinite;
}
@keyframes spin { 100% { transform: rotate(360deg); } }
</style>