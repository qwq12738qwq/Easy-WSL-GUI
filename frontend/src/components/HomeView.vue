<script setup>
import { ref, onMounted, onUnmounted, onActivated, onDeactivated, reactive, computed } from 'vue'
import { GetDistroStats, GetPath, GetMetrics, UninstallDistro, StartMigration, SelectDirectory, OpenDistroFolder, StartDistro, StopDistro } from '../../wailsjs/go/main/App'
import { formatBytes } from '../utils/format'
import { ArrowRightLeft, Play, FolderOpen, Square } from 'lucide-vue-next'
import { EventsOn, EventsOff, BrowserOpenURL, EventsEmit } from '../../wailsjs/runtime/runtime'
import { useAppStore } from '../stores/app'
import { getDistroIcon } from '../utils/icon'

const appStore = useAppStore()
const handleCardClick = (name) => {
    appStore.navigateToConfig(name)
}

// 启动实例
const startingDistros = ref(new Set())

const startDistro = async (name) => {
    if (startingDistros.value.has(name)) return
    
    try {
        console.log(`尝试启动 ${name}...`)
        startingDistros.value.add(name)
        await StartDistro(name)
        // 启动命令发送后，稍微延迟一下刷新状态，或者等待后端事件
        // 这里简单做延迟刷新
        setTimeout(() => {
            syncData()
            startingDistros.value.delete(name)
        }, 1500)
    } catch (e) {
        console.error(`启动 ${name} 失败:`, e)
        alert(`启动失败: ${e}`)
        startingDistros.value.delete(name)
    }
}

// 停止实例
const stopDistro = async (name) => {
    try {
        console.log(`尝试停止 ${name}...`)
        await StopDistro(name)
        setTimeout(() => {
            syncData()
        }, 1500)
    } catch (e) {
        console.error(`停止 ${name} 失败:`, e)
        alert(`停止失败: ${e}`)
    }
}


// 打开实例文件夹
const openDistroFolder = async (name) => {
    try {
        if (OpenDistroFolder) {
            await OpenDistroFolder(name)
        } else {
            console.warn("OpenDistroFolder method not available in frontend yet.")
        }
    } catch (e) {
        console.error("Failed to open folder:", e)
    }
}

const distros = ref([])
const isInitialLoading = ref(true)
const isSyncing = ref(false) // 防止并发同步

// --- 排序逻辑 (优化 1) ---
const sortedDistros = computed(() => {
    // 创建副本以免影响原始数据
    return [...distros.value].sort((a, b) => {
        const isARunning = a.status === 'Running'
        const isBRunning = b.status === 'Running'

        // 1. 运行状态优先：运行中 (Running) 排在前面
        if (isARunning && !isBRunning) return -1
        if (!isARunning && isBRunning) return 1

        // 2. 如果都运行中，按名称 A-Z 排序
        if (isARunning && isBRunning) {
            return a.name.localeCompare(b.name)
        }

        // 3. 非运行状态保持原样 (在 sort 中返回 0 即视为相等，对于稳定排序会保持相对位置)
        // 注意：Chrome 的 sort 是稳定的，但为了保险，如果不涉及其他排序需求，0 即可
        return 0
    })
})

// --- 迁移相关状态 ---
const showMigrationModal = ref(false)
const migrationStepView = ref('config') // 'config' | 'progress'
const isMigrating = ref(false)
const migrationError = ref('')
const migrationLog = ref('准备就绪...')
const migrationProgress = ref(0)

const migrationForm = reactive({
    distroName: '',
    sourcePath: '',
    targetPath: '',
    verifyChecksum: true
})

const migrationSteps = ref([
    { title: '准备环境', status: 'pending', keyword: ['prepare', 'checking', '准备'] },
    { title: '导出系统', status: 'pending', keyword: ['exporting', '导出'] },
    { title: '卸载系统', status: 'pending', keyword: ['uninstall', '卸载'] },
    { title: '迁移系统', status: 'pending', keyword: ['moving', 'transferring', '迁移'] },
    { title: '选择默认用户', status: 'pending', keyword: ['select-user', '用户选择'] },
    { title: '还原用户', status: 'pending', keyword: ['还原'] }
])

const distroUsers = ref([])
const selectedUser = ref('')

const handleUserSelect = () => {
    if (!selectedUser.value) return
    EventsEmit("migration:select-user", selectedUser.value)
    migrationStepView.value = 'progress'
}

// 处理迁移日志与进度 (仿照 InstallView)
const processMigrationLog = (line) => {
    if (!line) return
    const lowerLine = line.toLowerCase()
    migrationLog.value = line

    const skipIncrementKeywords = ['%', 'progress', '进度', 'downloading', 'exporting', 'importing']
    const shouldSkip = skipIncrementKeywords.some(key => lowerLine.includes(key))

    // 尝试解析进度百分比 (假设日志格式如 "Progress: 25.5%" 或 "25%")
    const percentMatch = lowerLine.match(/(\d+(\.\d+)?)%/)
    if (percentMatch) {
        const p = parseFloat(percentMatch[1])
        if (!isNaN(p)) {
             // 限制最大增长，确保不超过当前步骤的最大范围
             const stepCount = migrationSteps.value.length
             const stepWidth = 100 / stepCount
             // 找到当前正在进行的步骤索引
             const activeStepIndex = migrationSteps.value.findIndex(s => s.status === 'processing')
             const currentIndex = activeStepIndex !== -1 ? activeStepIndex : 0
             
             const currentStepMax = (currentIndex + 1) * stepWidth
             
             // 将后端 0-100% 映射到当前步骤的范围
             const mappedPercent = (currentIndex * stepWidth) + (p / 100 * stepWidth)

             if (mappedPercent > migrationProgress.value && mappedPercent <= currentStepMax) {
                 migrationProgress.value = mappedPercent
             }
        }
    } else {
        // 如果没有明确百分比，尝试通过关键词推断
        if (lowerLine.includes('25%')) migrationProgress.value = Math.max(migrationProgress.value, 25)
        else if (lowerLine.includes('50%')) migrationProgress.value = Math.max(migrationProgress.value, 50)
        else if (lowerLine.includes('75%')) migrationProgress.value = Math.max(migrationProgress.value, 75)
        else if (lowerLine.includes('100%')) migrationProgress.value = 100
    }

    // 模拟增长逻辑
    if (!percentMatch && !shouldSkip) {
        const stepCount = migrationSteps.value.length
        const stepWidth = 100 / stepCount
        const activeStepIndex = migrationSteps.value.findIndex(s => s.status === 'processing')
        const currentIndex = activeStepIndex !== -1 ? activeStepIndex : 0
        const currentStepMax = (currentIndex + 1) * stepWidth
        const simulationLimit = currentStepMax - (stepWidth * 0.1)
        
        if (migrationProgress.value < simulationLimit) {
            migrationProgress.value += 0.2 // 迁移通常较慢，增长慢一点
        }
    }

    // 步骤匹配
    migrationSteps.value.forEach((step, index) => {
        const keywords = Array.isArray(step.keyword) ? step.keyword : [step.keyword]
        const isMatch = keywords.some(key => key && lowerLine.includes(key.toLowerCase()))

        if (isMatch) {
            // 将当前步骤之前的都标记为完成
            for(let i = 0; i < index; i++) {
                migrationSteps.value[i].status = 'finished'
            }
            
            // 标记当前步骤
            if (migrationSteps.value[index].status !== 'finished') {
                migrationSteps.value[index].status = 'processing'
                
                // 调整进度条基准
                const basePercent = (index / migrationSteps.value.length) * 100
                if (migrationProgress.value < basePercent) {
                    migrationProgress.value = basePercent
                }
            }
        }
    })
    
    // 特殊处理：如果检测到“完成”或“Success”
    if (lowerLine.includes('success') || lowerLine.includes('completed') || lowerLine.includes('迁移成功')) {
        migrationProgress.value = 100
        migrationSteps.value.forEach(s => s.status = 'finished')
    }
}

// 打开迁移弹窗
const openMigrationModal = (distro) => {
    migrationForm.distroName = distro.name
    migrationForm.sourcePath = distro.path
    migrationForm.targetPath = ''
    migrationForm.verifyChecksum = true
    
    migrationStepView.value = 'config'
    isMigrating.value = false
    migrationError.value = ''
    migrationProgress.value = 0
    migrationLog.value = '准备就绪...'
    
    // 重置步骤
    migrationSteps.value.forEach(s => s.status = 'pending')
    
    showMigrationModal.value = true
}

// 选择目标路径
const handleSelectTarget = async () => {
    try {
        const path = await SelectDirectory()
        if (path) migrationForm.targetPath = path
    } catch (e) {
        console.error("选择路径失败", e)
    }
}

// 开始迁移
const startMigration = async () => {
    // 重置错误
    migrationError.value = ''

    if (!migrationForm.targetPath) {
        migrationError.value = "请选择迁移目标路径"
        return
    }
    
    if (migrationForm.sourcePath === migrationForm.targetPath) {
        migrationError.value = "目标路径不能与源路径相同"
        return
    }

    // 重置状态
    migrationProgress.value = 0
    migrationLog.value = '准备就绪...'
    migrationSteps.value.forEach(s => s.status = 'pending')

    isMigrating.value = true
    migrationStepView.value = 'progress'
    migrationSteps.value[0].status = 'processing'
    
    // 监听进度事件
    EventsOn("migration:progress", (data) => {
        // data 可能是对象 { message: "xxx" } 或者直接是字符串
        const msg = (typeof data === 'object' && data.message) ? data.message : data
        processMigrationLog(msg)
    })

    // 监听用户列表 (兼容 migration:users 和 migration:user-groups)
    const handleUserList = (users) => {
        distroUsers.value = users || []
        if (distroUsers.value.length > 0) {
            selectedUser.value = distroUsers.value[0] // 默认选中第一个
        }
        migrationStepView.value = 'select-user'
        // 标记 "选择默认用户" 步骤为进行中
        const stepIndex = migrationSteps.value.findIndex(s => s.keyword.includes('select-user'))
        if (stepIndex !== -1) {
            // 完成之前的
            for(let i=0; i<stepIndex; i++) migrationSteps.value[i].status = 'finished'
            migrationSteps.value[stepIndex].status = 'processing'
        }
    }

    EventsOn("migration:users", handleUserList)
    EventsOn("migration:user-groups", handleUserList)
    
    EventsOn("migration:done", async (data) => {
        EventsOff("migration:progress")
        EventsOff("migration:users")
        EventsOff("migration:user-groups")
        EventsOff("migration:done")
        
        if (data.status === 'failed') {
            isMigrating.value = false
            migrationError.value = data.error || "未知错误"
            // 标记当前步骤为错误
            const currentStep = migrationSteps.value.find(s => s.status === 'processing')
            if (currentStep) currentStep.status = 'error'
        } else {
            migrationProgress.value = 100
            migrationSteps.value.forEach(s => s.status = 'finished')
            migrationLog.value = "迁移成功！"
            
            // 迁移成功之后前端重新执行后端的GetPath()函数刷新安装路径
            try {
                const newPath = await GetPath(migrationForm.distroName)
                const targetDistro = distros.value.find(d => d.name === migrationForm.distroName)
                if (targetDistro) {
                    targetDistro.path = newPath
                }
            } catch (e) {
                console.error("刷新路径失败:", e)
            }

            // 延迟关闭
            setTimeout(() => {
                showMigrationModal.value = false
                syncData()
            }, 1500)
        }
    })

    try {
        const options = { 
            distroName: migrationForm.distroName,
            sourcePath: migrationForm.sourcePath, 
            targetPath: migrationForm.targetPath, 
            verifyChecksum: migrationForm.verifyChecksum 
        }
        await StartMigration(options)
        // 开始后，直接显示为"处理中"，不再显示具体百分比
        migrationProgress.value = 50 // 假进度
        migrationLog.value = "系统迁移中，请耐心等待..."
    } catch (e) {
        console.error("Migration start failed:", e)
        isMigrating.value = false
        migrationError.value = e.toString()
        EventsOff("migration:progress")
        EventsOff("migration:done")
        
        if (e.toString().includes("is not a function") || e.toString().includes("404")) {
            alert("迁移服务暂未开放")
            showMigrationModal.value = false
        }
    }
}

// --- 卸载模态框相关状态 ---
const showUninstallModal = ref(false)
const uninstallTarget = ref('')
const uninstallStepIndex = ref(0)
const isUninstalling = ref(false)
const uninstallLog = ref('') // 新增：卸载日志

// 定义卸载流程步骤 (带关键词)
const uninstallSteps = ref([
  { title: '确认操作', status: 'pending' },
  { title: '停止实例', status: 'pending', keyword: ['stopping', 'terminating', '停止'] },
  { title: '注销分发', status: 'pending', keyword: ['unregistering', 'destroying', '注销', '卸载'] },
  { title: '清理磁盘', status: 'pending', keyword: ['cleaning', 'removing', 'cleanup', '清理'] }
])

// 处理卸载日志
const processUninstallLog = (line) => {
    if (!line) return
    const lowerLine = line.toLowerCase()
    uninstallLog.value = line
    
    uninstallSteps.value.forEach((step, index) => {
        if (!step.keyword) return
        const keywords = Array.isArray(step.keyword) ? step.keyword : [step.keyword]
        if (keywords.some(k => lowerLine.includes(k.toLowerCase()))) {
            // 完成之前的步骤
            for(let i = 1; i < index; i++) {
                 if (uninstallSteps.value[i].status !== 'finished') {
                     uninstallSteps.value[i].status = 'finished'
                 }
            }
            // 标记当前步骤
            uninstallStepIndex.value = index
            uninstallSteps.value[index].status = 'processing'
        }
    })
}

const handleSystemMigrate = null

// 保持原有的数据同步逻辑
const syncData = async () => {
  if (isSyncing.value) return
  isSyncing.value = true
  
  try {
    const backendList = await GetDistroStats().catch(() => [])
    if (!backendList) { 
        // 如果后端返回空或错误，保持现有列表或清空视需求而定
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
            stats: { cpu: '0%', memUsed: '0', memTotal: '0', disk: '0%', diskText: '0 B / 0 B' } 
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
              
              // 3. 磁盘占用展示改造
              // 假设后端返回 usedBytes 和 totalBytes，如果只有 disk 百分比字符串，则无法准确显示
              // 这里做兼容处理：如果有 bytes 则使用 formatBytes，否则保留原样或显示 N/A
              if (m.usedBytes !== undefined && m.totalBytes !== undefined) {
                  const diskInfo = formatBytes(m.usedBytes, m.totalBytes)
                  localItem.stats.diskText = diskInfo.text
                  localItem.stats.disk = diskInfo.percent + '%' // 更新百分比供其他用途
              } else {
                  // Fallback: 如果后端还没更新，尝试保留原值或显示 N/A
                  localItem.stats.diskText = m.disk || 'N/A'
                  localItem.stats.disk = m.disk || '0%'
              }
          }
        } catch (e) { 
            // 静默失败，保持旧值或归零
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

const startPolling = () => {
    if (timer) return
    syncData()
    timer = setInterval(syncData, 3000)
}

const stopPolling = () => {
    if (timer) {
        clearInterval(timer)
        timer = null
    }
}

const refreshAllPaths = async () => {
    // 仅在列表不为空时执行刷新
    if (distros.value.length === 0) return
    
    for (const item of distros.value) {
        try {
            const p = await GetPath(item.name)
            if (p && p.trim() !== '') {
                item.path = p
            }
        } catch(e) {
            console.warn('Failed to refresh path for', item.name)
        }
    }
}

onMounted(() => {
  // Initial load is handled by onActivated if using KeepAlive, 
  // but keeping syncData here ensures immediate fetch on mount if needed before activation logic kicks in.
  // However, onActivated is called after onMounted on first load for KeepAlive components.
  // We can just rely on onActivated.
})

onActivated(async () => {
    // 激活时先尝试刷新所有路径，以同步 ConfigView 可能发生的迁移变更
    await refreshAllPaths()
    startPolling()
})

onDeactivated(() => {
    stopPolling()
})

onUnmounted(() => {
    stopPolling()
})

// --- 卸载逻辑控制 ---

const handleUninstallClick = (name) => {
  uninstallTarget.value = name
  uninstallStepIndex.value = 0
  isUninstalling.value = false
  uninstallLog.value = ''
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
  uninstallSteps.value[0].status = 'finished'
  uninstallLog.value = '正在初始化卸载...'
  
  // 监听卸载进度事件 (假设后端使用 uninstall:progress)
  EventsOn("uninstall:progress", (msg) => {
      processUninstallLog(msg)
  })

  // 也可以监听通用输出作为补充
  EventsOn("wsl-output", (msg) => processUninstallLog(msg))

  // 监听卸载失败事件
  EventsOn("uninstall:failed", (errMsg) => {
      uninstallSteps.value[uninstallStepIndex.value].status = 'error'
      uninstallLog.value = "错误: " + errMsg
      isUninstalling.value = false // 停止 loading 状态，但保持弹窗打开以显示错误
  })

  try {
    // 调用后端卸载
    await UninstallDistro(uninstallTarget.value)
    
    // 卸载完成
    uninstallSteps.value.forEach(s => s.status = 'finished')
    uninstallLog.value = '卸载成功'
    
    // 关闭并刷新
    setTimeout(() => {
        showUninstallModal.value = false
        syncData()
    }, 1000)
    
  } catch (err) {
    uninstallSteps.value[uninstallStepIndex.value].status = 'error'
    uninstallLog.value = "错误: " + err
    console.error(err)
  } finally {
    isUninstalling.value = false
    EventsOff("uninstall:progress")
    EventsOff("uninstall:failed")
    EventsOff("wsl-output")
  }
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
      <div v-for="item in sortedDistros" :key="item.name" class="distro-card" :class="{ 'running': item.status === 'Running' }" @click="handleCardClick(item.name)">
        <div class="card-actions">
            <button v-if="item.status === 'Running'" class="action-btn stop-action" @click.stop="stopDistro(item.name)" title="停止实例">
                <Square :size="16" fill="currentColor" />
            </button>
            <button class="action-btn folder-action" @click.stop="openDistroFolder(item.name)" title="打开安装目录">
                <FolderOpen :size="16" />
            </button>
            <button class="action-btn migrate-action" @click.stop="openMigrationModal(item)" title="系统迁移">
                <ArrowRightLeft :size="16" />
            </button>
            <button class="action-btn uninstall-action" @click.stop="handleUninstallClick(item.name)" title="卸载实例">
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18"></path><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"></path><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"></path></svg>
            </button>
        </div>
        
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
              <span class="disk-icon">💾</span> 磁盘占用: {{ item.stats.diskText || item.stats.disk }}
          </div>
        </div>
        
        <div class="offline-placeholder" v-else>
          <div class="offline-icon">💤</div>
          <span>实例已休眠</span>
          <button class="start-btn" @click.stop="startDistro(item.name)" :disabled="startingDistros.has(item.name)" :class="{ 'is-loading': startingDistros.has(item.name) }">
              <span v-if="startingDistros.has(item.name)" class="spinner-sm start-spinner"></span>
              <Play v-else :size="14" class="start-icon" /> 
              {{ startingDistros.has(item.name) ? '正在启动...' : '启动实例' }}
          </button>
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

            <div v-if="uninstallLog" class="uninstall-log">
               {{ uninstallLog }}
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

    <!-- 迁移模态框 -->
    <Transition name="modal">
    <div v-if="showMigrationModal" class="modal-overlay">
      <div class="modal-window">
        <div class="modal-header">
          <span>系统迁移 - {{ migrationForm.distroName }}</span>
          <button v-if="!isMigrating" class="close-btn" @click="showMigrationModal = false">✕</button>
        </div>
        
        <div class="modal-body" v-if="migrationStepView === 'config'">
             <div class="form-group">
                  <label>当前位置 (源)</label>
                  <input type="text" class="input" :value="migrationForm.sourcePath" readonly disabled>
             </div>
             
             <div class="form-group">
                  <label>迁移目标位置</label>
                  <div class="path-input-group">
                      <input type="text" class="input" :value="migrationForm.targetPath" placeholder="请选择目标文件夹..." readonly>
                      <button class="btn btn-secondary browse-btn" @click="handleSelectTarget">浏览...</button>
                  </div>
             </div>

             <div v-if="migrationError" class="config-error">
                <span class="error-icon-sm">⚠️</span> {{ migrationError }}
             </div>

             <div class="action-bar">
                <button class="btn btn-secondary" @click="showMigrationModal = false">取消</button>
                <button class="btn btn-primary" @click="startMigration">开始迁移</button>
            </div>
        </div>

        <div class="modal-body" v-else-if="migrationStepView === 'select-user'">
            <div class="user-select-container">
                <div class="icon-header">
                    <span class="header-icon">👤</span>
                    <h3>选择默认用户</h3>
                    <p>检测到系统默认用户配置丢失，请从下方列表中选择一个用户作为默认登录用户。</p>
                </div>

                <div class="user-list">
                     <div 
                        v-for="user in distroUsers" 
                        :key="user" 
                        class="user-option"
                        :class="{ selected: selectedUser === user }"
                        @click="selectedUser = user"
                     >
                        <div class="radio-indicator"></div>
                        <span class="username">{{ user }}</span>
                     </div>
                </div>

                <div class="action-bar centered">
                    <button class="btn btn-primary" @click="handleUserSelect" :disabled="!selectedUser">
                        确认选择
                    </button>
                </div>
            </div>
        </div>

        <div class="modal-body" v-else>
            <!-- 进度视图 -->
             <div class="progress-content" v-if="!migrationError">
                  <div class="install-hero">
                      <img :src="getDistroIcon(migrationForm.distroName)" class="hero-icon" />
                      <div class="hero-info">
                          <h3>正在迁移...</h3>
                          <p class="log-detail">{{ migrationLog }}</p>
                      </div>
                  </div>

                  <div class="progress-bar-container">
                      <div class="progress-track">
                          <div class="progress-fill" :style="{ width: migrationProgress + '%' }">
                              <div class="progress-glow"></div>
                          </div>
                      </div>
                      <span class="progress-text">{{ Math.floor(migrationProgress) }}%</span>
                  </div>

                  <div class="steps-container">
                      <div v-for="(step, index) in migrationSteps" :key="index" class="step-item" :class="step.status">
                          <div class="step-icon">
                                <span v-if="step.status === 'finished'">✓</span>
                                <span v-else-if="step.status === 'processing'" class="spinner"></span>
                                <span v-else-if="step.status === 'error'">!</span>
                                <span v-else>{{ index + 1 }}</span>
                          </div>
                          <span class="step-title">{{ step.title }}</span>
                          <div v-if="index < migrationSteps.length - 1" class="step-line" :class="{ 'line-active': step.status === 'finished' }"></div>
                      </div>
                  </div>
             </div>

             <div class="error-container" v-else>
                  <div class="error-icon-area"><span class="error-symbol">⚠️</span></div>
                  <h3>迁移失败</h3>
                  <p class="error-desc">{{ migrationError }}</p>
                  <div class="action-bar">
                      <button class="btn btn-danger" @click="migrationStepView = 'config'">返回设置</button>
                      <button class="btn btn-secondary" @click="showMigrationModal = false">关闭</button>
                  </div>
             </div>
        </div>
      </div>
    </div>
    </Transition>

  </div>
</template>

<style scoped>
/* --- 用户选择样式 --- */
.user-select-container {
    text-align: center;
    padding: 10px 0;
}

.icon-header {
    margin-bottom: 24px;
}

.header-icon {
    font-size: 48px;
    display: block;
    margin-bottom: 16px;
}

.icon-header h3 {
    margin: 0 0 8px 0;
    font-size: 20px;
    color: var(--color-text-primary);
}

.icon-header p {
    margin: 0;
    color: var(--color-text-secondary);
    font-size: 14px;
}

.user-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
    max-height: 200px;
    overflow-y: auto;
    margin-bottom: 24px;
    text-align: left;
    border: 1px solid var(--color-border);
    border-radius: 8px;
    padding: 8px;
}

.user-option {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.2s;
    background: var(--color-bg-secondary);
}

.user-option:hover {
    background: var(--color-bg-hover);
}

.user-option.selected {
    background: rgba(var(--color-brand-rgb), 0.1);
    border: 1px solid var(--color-brand);
}

.radio-indicator {
    width: 18px;
    height: 18px;
    border: 2px solid var(--color-text-secondary);
    border-radius: 50%;
    position: relative;
}

.user-option.selected .radio-indicator {
    border-color: var(--color-brand);
}

.user-option.selected .radio-indicator::after {
    content: "";
    position: absolute;
    top: 3px; left: 3px;
    width: 8px; height: 8px;
    background: var(--color-brand);
    border-radius: 50%;
}

.username {
    font-weight: 500;
    color: var(--color-text-primary);
}

.action-bar.centered {
    justify-content: center;
}

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

.header-left {
  display: flex;
  align-items: center;
}

.header-left h2 {
    margin: 0;
    font-size: 24px;
    font-weight: 600;
    color: var(--color-text-primary);
}

.migrate-btn {
    margin-left: 16px;
    padding: 6px 12px;
    border-radius: 6px;
    background: var(--color-bg-hover);
    border: 1px solid var(--color-border);
    cursor: pointer;
    display: flex; align-items: center; gap: 6px;
    font-size: 13px; color: var(--color-text-primary);
    transition: all 0.2s;
}
.migrate-btn:hover { background: var(--color-bg-active); border-color: var(--color-brand); }
.migrate-btn:disabled { opacity: 0.6; cursor: not-allowed; }

.spinner-sm {
    width: 14px; height: 14px;
    border: 2px solid var(--color-text-secondary);
    border-top-color: transparent;
    border-radius: 50%;
    animation: spin 1s linear infinite;
}

@media (max-width: 768px) {
    .migrate-btn { display: none; }
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

.config-error {
    background: rgba(255, 77, 79, 0.1);
    border: 1px solid rgba(255, 77, 79, 0.2);
    color: var(--color-error);
    padding: 10px 12px;
    border-radius: 6px;
    font-size: 13px;
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: -8px; /* Slightly closer to inputs */
    margin-bottom: 8px;
    animation: shake 0.4s cubic-bezier(0.36, 0.07, 0.19, 0.97) both;
}

.error-icon-sm {
    font-size: 14px;
}

.uninstall-log {
    margin-top: -12px;
    margin-bottom: 12px;
    font-size: 12px;
    color: var(--color-text-secondary);
    background: var(--color-bg-tertiary);
    padding: 8px 12px;
    border-radius: 6px;
    font-family: monospace;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    text-align: center;
}

@keyframes shake {
  10%, 90% { transform: translate3d(-1px, 0, 0); }
  20%, 80% { transform: translate3d(2px, 0, 0); }
  30%, 50%, 70% { transform: translate3d(-4px, 0, 0); }
  40%, 60% { transform: translate3d(4px, 0, 0); }
}

.distro-card.running::before {
  content: "";
  position: absolute;
  left: 0; top: 0; bottom: 0;
  width: 4px;
  background: var(--color-success);
  animation: height-grow 0.4s ease-out;
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
.metrics-box { 
    display: flex; flex-direction: column; gap: 12px; 
    animation: fade-in-up 0.5s ease-out;
}

@keyframes fade-in-up {
    from { opacity: 0; transform: translateY(10px); }
    to { opacity: 1; transform: translateY(0); }
}

.offline-placeholder { 
    text-align: center; padding: 10px; 
    color: var(--color-text-secondary); opacity: 0.8;
    display: flex; flex-direction: column; align-items: center; gap: 8px;
    animation: fade-in 0.3s ease-in;
}

@keyframes fade-in {
    from { opacity: 0; }
    to { opacity: 0.8; }
}

.distro-card.running::before {
  content: "";
  position: absolute;
  left: 0; top: 0; bottom: 0;
  width: 4px;
  background: var(--color-success);
  animation: height-grow 0.4s ease-out;
}

@keyframes height-grow {
    from { height: 0; top: 50%; bottom: 50%; }
    to { height: 100%; top: 0; bottom: 0; }
}

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
.offline-icon { font-size: 24px; opacity: 0.6; }

.start-btn {
    margin-top: 4px;
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 16px;
    border-radius: 20px;
    background: var(--color-bg-hover);
    border: 1px solid var(--color-border);
    color: var(--color-brand);
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
}

.start-btn:hover {
    background: var(--color-brand);
    color: white;
    border-color: var(--color-brand);
    transform: translateY(-1px);
    box-shadow: 0 2px 8px rgba(24, 144, 255, 0.25);
}

.start-btn:active {
    transform: translateY(0);
}

.start-btn.is-loading {
    cursor: wait;
    opacity: 0.8;
    background: var(--color-bg-active);
    border-color: var(--color-brand);
    color: var(--color-brand);
}

.start-spinner {
    border-color: var(--color-brand);
    border-top-color: transparent;
    margin-right: 4px;
    width: 12px; height: 12px;
}

.start-icon {
    fill: currentColor;
}


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
  width: 28px; height: 28px;
  border-radius: 50%;
  border: 1px solid transparent;
  background: transparent;
  color: var(--color-text-secondary);
  display: flex; align-items: center; justify-content: center;
  font-size: 16px;
  cursor: pointer;
  transition: all 0.2s;
}
.close-btn:hover {
  background: var(--color-bg-hover);
  color: var(--color-text-primary);
  border-color: var(--color-border);
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
.list-move, .list-enter-active, .list-leave-active { transition: all 0.5s ease; }
.list-enter-from, .list-leave-to { opacity: 0; transform: translateY(30px); }
.list-leave-active { position: absolute; }
/* --- Card Actions --- */
.card-actions {
    position: absolute; top: 12px; right: 12px;
    display: flex; gap: 4px;
    opacity: 0; transition: opacity 0.2s;
}
.distro-card:hover .card-actions { opacity: 1; }

.action-btn {
    width: 28px; height: 28px;
    border-radius: 50%;
    border: none; background: transparent;
    color: var(--color-text-secondary);
    display: flex; align-items: center; justify-content: center;
    cursor: pointer; transition: all 0.2s;
}
.action-btn:hover { background: var(--color-bg-hover); color: var(--color-text-primary); }
.stop-action:hover { background: rgba(255, 77, 79, 0.1); color: var(--color-error); }
.folder-action:hover { background: rgba(24, 144, 255, 0.1); color: var(--color-brand); }
.uninstall-action:hover { background: rgba(255, 77, 79, 0.1); color: var(--color-error); }
.migrate-action:hover { background: var(--color-bg-active); color: var(--color-brand); }

/* --- Form Styles (from InstallView) --- */
.form-group { margin-bottom: 16px; }
.form-group label { display: block; margin-bottom: 6px; font-size: 13px; color: var(--color-text-secondary); }
.input { 
    width: 100%; padding: 8px 12px; 
    border-radius: 6px; border: 1px solid var(--color-border); 
    background: var(--color-bg-input, var(--color-bg-card)); 
    color: var(--color-text-primary);
    font-size: 13px;
}
.input:disabled { opacity: 0.7; cursor: not-allowed; }
.path-input-group { display: flex; gap: 8px; }

.checkbox-group { margin-top: 8px; }
.checkbox-label { display: flex; align-items: center; gap: 8px; cursor: pointer; font-size: 13px; color: var(--color-text-primary); }

/* --- Buttons --- */
.btn { padding: 6px 16px; border-radius: 6px; border: none; cursor: pointer; font-size: 13px; transition: all 0.2s; font-weight: 500; }
.btn-primary { background: var(--color-brand); color: #fff; }
.btn-primary:hover { opacity: 0.9; }
.btn-secondary { background: var(--color-bg-hover); color: var(--color-text-primary); border: 1px solid var(--color-border); }
.btn-secondary:hover { border-color: var(--color-text-secondary); }
.btn-danger { background: var(--color-error); color: #fff; }

/* --- Progress & Hero --- */
.install-hero { margin-bottom: 24px; text-align: center; }
.hero-icon { width: 64px; height: 64px; object-fit: contain; margin-bottom: 16px; }
.hero-info h3 { margin: 0 0 4px 0; font-size: 18px; color: var(--color-text-primary); }
.log-detail { font-size: 12px; color: var(--color-text-secondary); margin: 0; font-family: monospace; }

.progress-bar-container { display: flex; align-items: center; gap: 12px; margin-bottom: 24px; }
.progress-track { flex: 1; height: 8px; background: var(--color-bg-hover); border-radius: 4px; overflow: hidden; }
.progress-fill { height: 100%; background: var(--color-brand); border-radius: 4px; position: relative; transition: width 0.3s; }
.progress-glow { position: absolute; top: 0; left: 0; width: 100%; height: 100%; background: linear-gradient(90deg, transparent, rgba(255,255,255,0.3), transparent); animation: scan 2s infinite; }
.progress-text { font-size: 13px; font-weight: 600; color: var(--color-text-primary); width: 36px; text-align: right; }

@keyframes scan { from { transform: translateX(-100%); } to { transform: translateX(100%); } }

/* Error State */
.error-container { text-align: center; padding: 20px; }
.error-symbol { font-size: 48px; display: block; margin-bottom: 16px; }
.error-desc { color: var(--color-error); margin-bottom: 24px; font-size: 16px; font-weight: 500; }

</style>