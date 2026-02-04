<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { GetDistroStats, GetPath, GetMetrics, UninstallDistro } from '../../wailsjs/go/main/App'

const distros = ref([])
const isInitialLoading = ref(true)
const isSyncing = ref(false) // 防止并发同步

// --- 卸载模态框相关状态 ---
const showUninstallModal = ref(false)
const uninstallTarget = ref('')
const uninstallStepIndex = ref(0)
const isUninstalling = ref(false)

// 定义卸载流程步骤
const uninstallSteps = ref([
  { title: '确认操作', status: 'pending' },
  { title: '停止实例', status: 'pending' },
  { title: '注销分发', status: 'pending' },
  { title: '清理磁盘', status: 'pending' }
])

// 保持原有的数据同步逻辑
const syncData = async () => {
  if (isSyncing.value) return
  isSyncing.value = true
  
  try {
    const backendList = await GetDistroStats().catch(() => [])
    if (!backendList) { 
        // 如果后端返回空或错误，保持现有列表或清空视需求而定
        // 这里选择不做破坏性清空，除非明确返回空数组
        return 
    }

    // 前端防重保护：使用 Map 去重
    const uniqueBackendMap = new Map();
    backendList.forEach(item => {
        if(item.name) uniqueBackendMap.set(item.name, item);
    });
    const uniqueList = Array.from(uniqueBackendMap.values());

    // 移除本地存在但后端不存在的项目
    const backendNames = uniqueList.map(i => i.name)
    distros.value = distros.value.filter(d => backendNames.includes(d.name))

    // 更新或添加项目
    await Promise.all(uniqueList.map(async (item) => {
      let localItem = distros.value.find(d => d.name === item.name)
      if (!localItem) {
        let path = 'Loading...'
        try {
            path = await GetPath(item.name)
            path = (path && path.trim() !== "") ? path : 'N/A'
        } catch { path = 'N/A' }
        
        localItem = { 
            ...item, 
            path, 
            stats: { cpu: '0%', memUsed: '0', memTotal: '0', disk: '0%' } 
        }
        distros.value.push(localItem)
      } else {
        localItem.status = item.status
        localItem.version = item.version
      }

      // 获取指标逻辑
      if (localItem.status === 'Running') {
        try {
          const m = await GetMetrics(localItem.name)
          if (m) {
              localItem.stats.cpu = m.cpu || '0%'
              localItem.stats.memUsed = m.memUsed || '0'
              localItem.stats.memTotal = m.memTotal || '0'
              localItem.stats.disk = m.disk || '0%'
          }
        } catch (e) { 
            // 静默失败，保持旧值或归零
            // localItem.stats.cpu = '0%' 
        }
      } else { 
        localItem.stats.cpu = '0%' 
        localItem.stats.memUsed = '0'
      }
    }))
  } finally { 
    isInitialLoading.value = false 
    isSyncing.value = false
  }
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
  uninstallSteps.value.forEach(s => s.status = 'pending')
  uninstallSteps.value[0].status = 'processing'
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
    uninstallSteps.value[0].status = 'finished'
    uninstallStepIndex.value = 1
    uninstallSteps.value[1].status = 'processing'
    await new Promise(r => setTimeout(r, 800)) // UI 模拟耗时

    // 步骤 2: 停止完成，调用后端
    uninstallSteps.value[1].status = 'finished'
    uninstallStepIndex.value = 2
    uninstallSteps.value[2].status = 'processing'
    
    await UninstallDistro(uninstallTarget.value)
    
    // 步骤 3: 注销完成，清理UI
    uninstallSteps.value[2].status = 'finished'
    uninstallStepIndex.value = 3
    uninstallSteps.value[3].status = 'processing'
    await new Promise(r => setTimeout(r, 800)) // UI 模拟耗时

    // 全部完成
    uninstallSteps.value[3].status = 'finished'
    
    // 关闭并刷新
    setTimeout(() => {
        showUninstallModal.value = false
        syncData()
    }, 500)
    
  } catch (err) {
    // alert("卸载失败: " + err) // 移除 alert，改用 UI 显示
    uninstallSteps.value[uninstallStepIndex.value].status = 'error'
    console.error(err)
  } finally {
    isUninstalling.value = false
  }
}

const getDistroIcon = (name) => {
  const n = name.toLowerCase()
  let iconName = 'UbuntuCoF.png' // 默认值

  if (n.includes('ubuntu')) iconName = 'UbuntuCoF.png'
  else if (n.includes('debian')) iconName = 'Debian.png'
  else if (n.includes('kali'))   iconName = 'Kali-drago.png'
  else if (n.includes('arch'))   iconName = 'Arch.png'

  // 关键：利用 Vite 的动态资源解析
  // 假设你的图片放在：frontend/src/assets/icons/ 目录下
  return new URL(`../assets/icons/${iconName}`, import.meta.url).href
}

// 辅助函数：计算内存百分比
const getMemPercent = (used, total) => {
    const u = parseFloat(used) || 0
    const t = parseFloat(total) || 1
    if (t === 0) return 0
    return Math.min((u / t) * 100, 100)
}
</script>

<template>
  <div class="home-view-container">
    <header class="view-header">
      <div class="header-left">
          <h2>我的发行版</h2>
          <span class="distro-count" v-if="!isInitialLoading">{{ distros.length }} 个实例</span>
      </div>
      <div class="status-tag">
        <span class="status-dot-pulse"></span> 
        <span class="status-text">系统监控运行中</span>
      </div>
    </header>

    <div v-if="isInitialLoading" class="loading-grid">
      <div v-for="i in 3" :key="i" class="skeleton-card"></div>
    </div>

    <div v-else-if="distros.length === 0" class="empty-state">
        <div class="empty-icon">📦</div>
        <p>暂无已安装的 WSL 发行版</p>
        <span class="sub-text">请前往“安装”页面获取新的系统</span>
    </div>

    <div v-else class="distro-grid">
      <TransitionGroup name="list">
      <div v-for="item in distros" :key="item.name" class="distro-card" :class="{ 'running': item.status === 'Running' }">
        <button class="action-uninstall" @click="handleUninstallClick(item.name)" title="卸载实例">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18"></path><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"></path><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"></path></svg>
        </button>
        
        <div class="card-header">
          <div class="icon-wrapper">
             <img :src="getDistroIcon(item.name)" class="distro-icon" />
          </div>
          <div class="info-content">
            <div class="name-row">
              <span class="name" :title="item.name">{{ item.name }}</span>
              <span class="status-badge" :class="item.status.toLowerCase()">{{ item.status }}</span>
            </div>
            <div class="version-text">v{{ item.version }}</div>
            <div class="path-text" :title="item.path">{{ item.path }}</div>
          </div>
        </div>

        <div class="divider"></div>

        <div class="metrics-box" v-if="item.status === 'Running'">
          <div class="metric-row">
            <div class="label-group">
                <span class="label-icon">⚡</span>
                <span class="label">CPU</span>
            </div>
            <div class="progress-wrapper">
                <div class="progress"><div class="bar cpu-bar" :style="{ width: item.stats.cpu }"></div></div>
                <span class="value-text">{{ item.stats.cpu }}</span>
            </div>
          </div>
          <div class="metric-row">
            <div class="label-group">
                <span class="label-icon">🧠</span>
                <span class="label">内存</span>
            </div>
            <div class="progress-wrapper">
                <div class="progress">
                  <div class="bar mem-bar" :style="{ width: getMemPercent(item.stats.memUsed, item.stats.memTotal) + '%' }"></div>
                </div>
                <span class="value-text">{{ item.stats.memUsed }} / {{ item.stats.memTotal }}</span>
            </div>
          </div>
          <div class="disk-info">
              <span class="disk-icon">💾</span> 磁盘占用: {{ item.stats.disk }}
          </div>
        </div>
        
        <div class="offline-placeholder" v-else>
          <div class="offline-icon">💤</div>
          <span>实例已休眠</span>
        </div>
      </div>
      </TransitionGroup>
    </div>

    <!-- 卸载模态框 -->
    <Transition name="modal">
    <div v-if="showUninstallModal" class="modal-overlay">
      <div class="modal-window">
        <div class="modal-header">
          <span>卸载向导</span>
          <button v-if="!isUninstalling" class="close-btn" @click="closeUninstallModal">✕</button>
        </div>
        
        <div class="modal-body">
            <div class="warning-section">
                <div class="warning-icon">⚠️</div>
                <div class="warning-content">
                    <h4>危险操作警告</h4>
                    <p>您即将卸载 <strong>{{ uninstallTarget }}</strong>。此操作不可逆，将永久删除该发行版及其所有数据。</p>
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
                    {{ isUninstalling ? '正在处理...' : '确认卸载' }}
                </button>
            </div>
        </div>
      </div>
    </div>
    </Transition>

  </div>
</template>

<style scoped>
/* --- 布局容器 --- */
.home-view-container { 
  display: flex; 
  flex-direction: column; 
  gap: 24px; 
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
}

.view-header { 
  display: flex; 
  justify-content: space-between; 
  align-items: center;
  padding: 0 4px;
}

.header-left h2 {
    margin: 0;
    font-size: 24px;
    font-weight: 600;
    color: var(--color-text-primary);
}
.distro-count {
    font-size: 13px;
    color: var(--color-text-secondary);
    margin-left: 12px;
    background: rgba(0,0,0,0.05);
    padding: 2px 8px;
    border-radius: 12px;
}
:root[data-theme='dark'] .distro-count { background: rgba(255,255,255,0.1); }

/* --- 状态标签 --- */
.status-tag {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 16px;
  background: rgba(82, 196, 26, 0.1);
  border: 1px solid rgba(82, 196, 26, 0.2);
  border-radius: 20px;
  backdrop-filter: blur(4px);
}

.status-dot-pulse {
  width: 8px;
  height: 8px;
  background: var(--color-success);
  border-radius: 50%;
  box-shadow: 0 0 0 0 rgba(82, 196, 26, 0.7);
  animation: pulse-green 2s infinite;
}

@keyframes pulse-green {
  0% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(82, 196, 26, 0.7); }
  70% { transform: scale(1); box-shadow: 0 0 0 6px rgba(82, 196, 26, 0); }
  100% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(82, 196, 26, 0); }
}

.status-text {
  font-size: 12px;
  color: var(--color-success);
  font-weight: 600;
}

/* --- 卡片网格 --- */
.distro-grid { 
  display: grid; 
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); 
  gap: 24px; 
}

/* --- 卡片样式 --- */
.distro-card { 
  background: var(--color-bg-card); 
  border: 1px solid var(--color-border);
  border-radius: 16px; 
  padding: 24px; 
  position: relative; 
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  box-shadow: var(--shadow-sm);
  overflow: hidden;
}

.distro-card:hover { 
  transform: translateY(-5px);
  box-shadow: var(--shadow-md);
  border-color: var(--color-brand);
}

.distro-card.running::before {
  content: "";
  position: absolute;
  left: 0; top: 0; bottom: 0;
  width: 4px;
  background: var(--color-success);
}

.action-uninstall { 
    position: absolute; top: 12px; right: 12px; 
    border: none; background: transparent; 
    color: var(--color-text-secondary); 
    width: 28px; height: 28px;
    border-radius: 50%;
    display: flex; align-items: center; justify-content: center;
    cursor: pointer; transition: all 0.2s;
    opacity: 0;
}
.distro-card:hover .action-uninstall { opacity: 1; }
.action-uninstall:hover { background: rgba(255, 77, 79, 0.1); color: var(--color-error); }

/* 卡片内容 */
.card-header { display: flex; gap: 16px; margin-bottom: 20px; align-items: flex-start; }
.icon-wrapper { 
    width: 56px; height: 56px; 
    background: var(--color-bg-hover); 
    border-radius: 12px;
    display: flex; align-items: center; justify-content: center;
    padding: 8px;
}

.distro-icon { width: 100%; height: 100%; object-fit: contain; }

.info-content { flex: 1; min-width: 0; }
.name-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px; }
.name { font-weight: 700; color: var(--color-text-primary); font-size: 18px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.status-badge {
    font-size: 10px; padding: 2px 8px; border-radius: 10px; font-weight: 600; text-transform: uppercase;
    background: var(--color-bg-hover); color: var(--color-text-secondary);
}
.status-badge.running { background: rgba(82, 196, 26, 0.15); color: var(--color-success); }
.status-badge.stopped { background: var(--color-bg-hover); color: var(--color-text-secondary); }

.version-text { font-size: 12px; color: var(--color-text-secondary); margin-bottom: 6px; }
.path-text { 
    font-size: 11px; color: var(--color-text-secondary); opacity: 0.7; 
    font-family: 'Consolas', monospace; 
    white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}

.divider { height: 1px; background: var(--color-border); margin-bottom: 16px; }

/* 指标区域 */
.metrics-box { display: flex; flex-direction: column; gap: 12px; }
.metric-row { display: flex; flex-direction: column; gap: 6px; }
.label-group { display: flex; align-items: center; gap: 6px; font-size: 12px; color: var(--color-text-secondary); }
.label-icon { font-size: 14px; }

.progress-wrapper { display: flex; align-items: center; gap: 10px; }
.progress { flex: 1; height: 6px; background: var(--color-bg-hover); border-radius: 3px; overflow: hidden; }

.bar { height: 100%; border-radius: 3px; transition: width 0.5s ease; }
.cpu-bar { background: linear-gradient(90deg, #1890ff, #36cfc9); }
.mem-bar { background: linear-gradient(90deg, #722ed1, #b37feb); }

.value-text { font-size: 11px; font-family: 'Consolas', monospace; color: var(--color-text-primary); width: 60px; text-align: right; }

.disk-info { font-size: 11px; color: var(--color-text-secondary); display: flex; align-items: center; justify-content: flex-end; gap: 6px; margin-top: 4px; }

/* 离线状态 */
.offline-placeholder { 
    text-align: center; padding: 10px; 
    color: var(--color-text-secondary); opacity: 0.6;
    display: flex; flex-direction: column; align-items: center; gap: 8px;
}
.offline-icon { font-size: 24px; }

/* 骨架屏 */
.loading-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 24px; }
.skeleton-card { height: 200px; background: var(--color-bg-hover); border-radius: 16px; animation: pulse 1.5s infinite; }
@keyframes pulse { 0% { opacity: 0.6; } 50% { opacity: 0.3; } 100% { opacity: 0.6; } }

/* 空状态 */
.empty-state {
    text-align: center; padding: 60px 20px;
    color: var(--color-text-secondary);
}
.empty-icon { font-size: 48px; margin-bottom: 16px; opacity: 0.5; }
.sub-text { font-size: 13px; opacity: 0.7; margin-top: 8px; display: block; }

/* === 模态框优化 === */
.modal-overlay {
  position: fixed; top: 0; left: 0;
  width: 100vw; height: 100vh;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  display: flex; justify-content: center; align-items: center;
  z-index: 1000;
}

.modal-window {
  width: 500px;
  background: var(--color-bg-modal);
  border-radius: 16px;
  overflow: hidden;
  box-shadow: var(--shadow-lg);
  border: 1px solid var(--color-border);
  display: flex; flex-direction: column;
}

.modal-header {
  padding: 16px 24px;
  background: var(--color-bg-hover);
  border-bottom: 1px solid var(--color-border);
  display: flex; justify-content: space-between; align-items: center;
  font-weight: 600; color: var(--color-text-primary);
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--color-text-secondary);
  font-size: 20px;
  cursor: pointer;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.close-btn:hover {
  background: rgba(0, 0, 0, 0.05);
  color: var(--color-text-primary);
}

:root[data-theme='dark'] .close-btn:hover {
  background: rgba(255, 255, 255, 0.1);
}

.modal-body { padding: 24px; display: flex; flex-direction: column; gap: 24px; }

.warning-section {
    display: flex; gap: 16px;
    background: rgba(255, 77, 79, 0.1);
    border: 1px solid rgba(255, 77, 79, 0.2);
    padding: 16px; border-radius: 8px;
}
.warning-icon { font-size: 24px; }
.warning-content h4 { margin: 0 0 4px 0; color: var(--color-error); font-size: 15px; }
.warning-content p { margin: 0; font-size: 13px; color: var(--color-text-secondary); line-height: 1.5; }

/* 步骤条 */
.steps-container { display: flex; justify-content: space-between; position: relative; padding: 0 10px; margin-top: 10px; }
.step-item { display: flex; flex-direction: column; align-items: center; position: relative; flex: 1; z-index: 2; }
.step-icon {
    width: 24px; height: 24px; border-radius: 50%;
    background: var(--color-bg-card); border: 2px solid var(--color-text-secondary);
    color: var(--color-text-secondary);
    display: flex; align-items: center; justify-content: center;
    font-size: 11px; font-weight: bold; margin-bottom: 8px;
    transition: all 0.3s;
}
.step-title { font-size: 11px; color: var(--color-text-secondary); transition: color 0.3s; }

.step-item.processing .step-icon { border-color: var(--color-brand); color: var(--color-brand); }
.step-item.processing .step-title { color: var(--color-text-primary); }
.step-item.finished .step-icon { background: var(--color-brand); border-color: var(--color-brand); color: #fff; }

.step-line {
    position: absolute; top: 11px; left: 50%; width: 100%; height: 2px;
    background: var(--color-border); z-index: -1;
}
.step-line.line-active { background: var(--color-brand); }

/* 按钮 */
.action-bar { display: flex; justify-content: flex-end; gap: 12px; }
.cancel-btn {
    padding: 8px 20px; border-radius: 6px; cursor: pointer;
    background: transparent; border: 1px solid var(--color-border); color: var(--color-text-secondary);
}
.cancel-btn:hover { border-color: var(--color-text-primary); color: var(--color-text-primary); background: var(--color-bg-hover); }
.danger-btn {
    padding: 8px 24px; border-radius: 6px; cursor: pointer;
    background: var(--color-error); border: none; color: white; font-weight: 500;
    box-shadow: 0 4px 10px rgba(255, 77, 79, 0.3);
}
.danger-btn:hover { background: #ff7875; }
.danger-btn:disabled { opacity: 0.6; cursor: not-allowed; }

/* 动画 */
.modal-enter-active, .modal-leave-active { transition: opacity 0.3s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
.list-move, .list-enter-active, .list-leave-active { transition: all 0.5s ease; }
.list-enter-from, .list-leave-to { opacity: 0; transform: translateY(30px); }
.list-leave-active { position: absolute; }
</style>