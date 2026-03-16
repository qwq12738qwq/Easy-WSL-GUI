<script setup>
import { ref, onMounted, onActivated, computed, watch, reactive, onUnmounted } from 'vue'
import { GetDistroStats, GetAPTSource, ChangeAPTSource, GetMetrics, StopDistro, OpenDistroFolder, StartMigration, UninstallDistro, SelectDirectory, GetPath, GetInstalledPackages, UninstallPackage } from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff, EventsEmit } from '../../wailsjs/runtime/runtime'
import { useAppStore } from '../stores/app'
import { useEnvironmentStore } from '../stores/environment'
import { Save, Download, Box, Server, Database, Code, Cpu, ArrowLeft, Settings, Info, Check, AlertTriangle, Monitor, FolderOpen, ArrowRightLeft, Trash2, Square, Search, Package } from 'lucide-vue-next'
import { getDistroIcon } from '../utils/icon'
import { formatBytes } from '../utils/format'
import DockerIcon from '../assets/icons/ConfigView/docker.png'

const appStore = useAppStore()
const envStore = useEnvironmentStore()
const distros = ref([])
const loading = ref(true)
const activeTab = ref('general') // 'general' | 'environment' | 'docker'

// Apt Source Config
const aptSources = [
    { name: 'Official (官方源)', value: 'official', desc: '系统默认的软件源，速度取决于网络环境' },
    { name: 'Aliyun (阿里云)', value: 'aliyun', desc: '阿里云开源镜像站，国内访问速度快' },
    { name: 'Tsinghua (清华大学)', value: 'tsinghua', desc: '清华大学开源软件镜像站，教育网和公网速度优秀' },
    { name: 'USTC (中科大)', value: 'ustc', desc: '中国科学技术大学开源软件镜像站' }
]
const selectedSource = ref('official')
const isSavingSource = ref(false)

// Docker Config
const dockerInstalled = ref(false)
const checkingDocker = ref(false)
const dockerSources = [
    { name: 'Official (官方源)', value: 'official', desc: 'Docker Hub 官方镜像源' },
    { name: 'Aliyun (阿里云)', value: 'aliyun', desc: '阿里云容器镜像加速器' },
    { name: 'Tencent (腾讯云)', value: 'tencent', desc: '腾讯云容器镜像加速器' },
    { name: 'Netease (网易云)', value: 'netease', desc: '网易云容器镜像加速器' }
]
const selectedDockerSource = ref('official')

const packageManagerName = computed(() => {
    const distro = (appStore.selectedDistro || '').toLowerCase()
    
    if (distro.includes('ubuntu') || distro.includes('debian') || distro.includes('kali')) {
        return 'APT'
    }
    if (distro.includes('arch')) {
        return 'Pacman'
    }
    if (distro.includes('fedora') || distro.includes('almalinux')) {
        return 'DNF'
    }
    if (distro.includes('suse')) { // Covers openSUSE and SUSE
        return 'Zypper'
    }
    return '系统'
})

const loadDistros = async () => {
    loading.value = true
    try {
        const stats = await GetDistroStats()
        if (stats) {
             distros.value = stats
        }
    } catch (e) {
        console.error("Failed to load distros", e)
    } finally {
        loading.value = false
    }
}

// --- Actions Logic ---

// Stop Distro
const stopDistro = async () => {
    const name = appStore.selectedDistro
    if (!name) return
    try {
        await StopDistro(name)
        // Refresh status after a delay
        setTimeout(() => {
            loadDistros()
        }, 1500)
    } catch (e) {
        console.error(`Failed to stop ${name}:`, e)
        alert(`停止失败: ${e}`)
    }
}

// Open Folder
const openDistroFolder = async () => {
    const name = appStore.selectedDistro
    if (!name) return
    try {
        if (OpenDistroFolder) {
            await OpenDistroFolder(name)
        }
    } catch (e) {
        console.error("Failed to open folder:", e)
    }
}

// --- Migration Logic ---
const showMigrationModal = ref(false)
const migrationStepView = ref('config')
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
    { title: '配置默认用户', status: 'pending', keyword: ['select-user', '配置用户'] },
    { title: '还原用户', status: 'pending', keyword: ['还原'] }
])

const distroUsers = ref([])
const selectedUser = ref('')

const handleUserSelect = () => {
    if (!selectedUser.value) return
    EventsEmit("migration:select-user", selectedUser.value)
    migrationStepView.value = 'progress'
}

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

    migrationSteps.value.forEach((step, index) => {
        const keywords = Array.isArray(step.keyword) ? step.keyword : [step.keyword]
        const isMatch = keywords.some(key => key && lowerLine.includes(key.toLowerCase()))

        if (isMatch) {
            for(let i = 0; i < index; i++) {
                migrationSteps.value[i].status = 'finished'
            }
            if (migrationSteps.value[index].status !== 'finished') {
                migrationSteps.value[index].status = 'processing'
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
        migrationLog.value = "迁移成功！"
        
        // 刷新当前显示的路径信息
        loadCurrentPath()

        // 延迟关闭
        setTimeout(() => {
            showMigrationModal.value = false
            loadDistros()
        }, 1500)
    }
}

const openMigrationModal = async () => {
    const name = appStore.selectedDistro
    if (!name) return
    
    // Find distro info to get current path (if we have path in distro list)
    // Assuming backend returns path in GetDistroStats or we might need GetPath
    // ConfigView doesn't seem to fetch Path in loadDistros, let's try GetPath
    let currentPath = 'Loading...'
    try {
        const res = await GetPath(name)
        currentPath = (res && res.trim() !== "") ? res : 'N/A'
    } catch (e) {
        console.warn("Could not get path", e)
        currentPath = 'N/A'
    }

    migrationForm.distroName = name
    migrationForm.sourcePath = currentPath
    migrationForm.targetPath = ''
    migrationForm.verifyChecksum = true
    
    migrationStepView.value = 'config'
    isMigrating.value = false
    migrationError.value = ''
    migrationProgress.value = 0
    migrationLog.value = '准备就绪...'
    
    migrationSteps.value.forEach(s => s.status = 'pending')
    
    showMigrationModal.value = true
}

const handleSelectTarget = async () => {
    try {
        const path = await SelectDirectory()
        if (path) migrationForm.targetPath = path
    } catch (e) {
        console.error("选择路径失败", e)
    }
}

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
            const currentStep = migrationSteps.value.find(s => s.status === 'processing')
            if (currentStep) currentStep.status = 'error'
        } else {
            migrationProgress.value = 100
            migrationSteps.value.forEach(s => s.status = 'finished')
            migrationLog.value = "迁移成功！"
            
            // 刷新当前显示的路径信息
            loadCurrentPath()

            // 延迟关闭
            setTimeout(() => {
                showMigrationModal.value = false
                loadDistros()
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

// --- Uninstall Logic ---
const showUninstallModal = ref(false)
const uninstallTarget = ref('')
const uninstallStepIndex = ref(0)
const isUninstalling = ref(false)
const uninstallLog = ref('')

const uninstallSteps = ref([
  { title: '确认操作', status: 'pending' },
  { title: '停止实例', status: 'pending', keyword: ['stopping', 'terminating', '停止'] },
  { title: '注销分发', status: 'pending', keyword: ['unregistering', 'destroying', '注销', '卸载'] },
  { title: '清理磁盘', status: 'pending', keyword: ['cleaning', 'removing', 'cleanup', '清理'] }
])

const processUninstallLog = (line) => {
    if (!line) return
    const lowerLine = line.toLowerCase()
    uninstallLog.value = line
    
    uninstallSteps.value.forEach((step, index) => {
        if (!step.keyword) return
        const keywords = Array.isArray(step.keyword) ? step.keyword : [step.keyword]
        if (keywords.some(k => lowerLine.includes(k.toLowerCase()))) {
            for(let i = 1; i < index; i++) {
                 if (uninstallSteps.value[i].status !== 'finished') {
                     uninstallSteps.value[i].status = 'finished'
                 }
            }
            uninstallStepIndex.value = index
            uninstallSteps.value[index].status = 'processing'
        }
    })
}

const handleUninstallClick = () => {
  const name = appStore.selectedDistro
  if (!name) return
  uninstallTarget.value = name
  uninstallStepIndex.value = 0
  isUninstalling.value = false
  uninstallLog.value = ''
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
  
  EventsOn("uninstall:progress", (msg) => {
      processUninstallLog(msg)
  })

  EventsOn("wsl-output", (msg) => processUninstallLog(msg))

  EventsOn("uninstall:failed", (errMsg) => {
      uninstallSteps.value[uninstallStepIndex.value].status = 'error'
      uninstallLog.value = "错误: " + errMsg
      isUninstalling.value = false
  })

  try {
    await UninstallDistro(uninstallTarget.value)
    
    uninstallSteps.value.forEach(s => s.status = 'finished')
    uninstallLog.value = '卸载成功'
    
    setTimeout(() => {
        showUninstallModal.value = false
        // Return to list view
        appStore.selectDistro('')
        loadDistros()
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

// 获取当前源配置
const loadCurrentSource = async () => {
    if (!appStore.selectedDistro) return
    try {
        const source = await GetAPTSource(appStore.selectedDistro)
        if (source) {
            selectedSource.value = source
        }
    } catch (e) {
        console.error("Failed to load APT source", e)
    }
}

const currentDistroPath = ref('')

const loadCurrentPath = async () => {
    if (!appStore.selectedDistro) return
    currentDistroPath.value = 'Loading...'
    try {
        const path = await GetPath(appStore.selectedDistro)
        currentDistroPath.value = (path && path.trim() !== "") ? path : 'N/A'
    } catch (e) {
        console.error("Failed to load path", e)
        currentDistroPath.value = 'N/A'
    }
}

// 监听选中的发行版变化，重新加载源配置
watch(() => appStore.selectedDistro, (newVal) => {
    if (newVal) {
        loadCurrentSource()
        loadCurrentPath()
        checkDockerStatus()
    }
})

onMounted(async () => {
    await loadDistros()
    // Load environments if not loaded
    if (envStore.environments.length === 0) {
        envStore.fetchEnvironments()
    }
})

onActivated(async () => {
    await loadDistros()
    if (appStore.selectedDistro) {
        loadCurrentSource()
        loadCurrentPath()
        checkDockerStatus()
    }
})

const handleBack = () => {
    appStore.selectDistro('')
    activeTab.value = 'general'
}

const checkDockerStatus = async () => {
    if (!appStore.selectedDistro) return
    checkingDocker.value = true
    try {
        // Use GetMetrics to indirectly check if docker command is available or if service is running
        // Or we can add a specific backend check. For now, assume we check "docker --version"
        // Since we don't have a specific IsDockerInstalled API exposed, we might need to rely on 
        // user feedback or mock it.
        // But the requirement says "检测是否安装Docker".
        // Let's assume GetMetrics might return installed packages or we can infer it.
        // Actually, let's mock it for now as per instructions "backend留出接口"
        // But here we need to implement the frontend logic.
        
        // Mock check
        console.log(`Checking Docker status for ${appStore.selectedDistro}...`)
        setTimeout(() => {
            // Randomly true for demo, or based on distro name if it has "docker"
            dockerInstalled.value = true 
            checkingDocker.value = false
        }, 500)
    } catch (e) {
        console.error("Docker check failed", e)
        dockerInstalled.value = false
        checkingDocker.value = false
    }
}

const handleSaveDockerSource = () => {
    alert(`正在为 ${appStore.selectedDistro} 配置 Docker 源: ${selectedDockerSource.value}\n(前端演示模式)`)
}

const handleSave = async () => {
    if (!appStore.selectedDistro) return
    isSavingSource.value = true
    try {
        await ChangeAPTSource(appStore.selectedDistro, selectedSource.value)
        alert(`成功将 ${appStore.selectedDistro} 的源切换为 ${selectedSource.value}`)
    } catch (e) {
        console.error("Failed to change APT source", e)
        alert("切换源失败: " + e)
    } finally {
        isSavingSource.value = false
    }
}

// Environment Installation Logic
const getIcon = (iconName) => {
    const map = {
        'docker': Box,
        'router': Server,
        'nodejs': Code,
        'python': Database,
        'go': Cpu,
        'box': Box,
        'server': Server,
        'database': Database,
        'code': Code,
        'cpu': Cpu
    }
    return map[iconName] || Box
}

const handleInstallEnvironment = (env) => {
    const distro = appStore.selectedDistro
    if (!distro) {
        alert("请先选择一个发行版")
        return
    }
    
    if (confirm(`确定要在 ${distro} 上安装 ${env.name} 吗？\n\n这将自动配置所需的所有依赖和环境脚本。`)) {
        // Mock backend call
        alert(`正在向 ${distro} 发送安装指令：\n安装 ${env.name} (v${env.version})...\n\n(前端演示模式)`)
    }
}

const currentDistroStatus = computed(() => {
    const d = distros.value.find(item => item.name === appStore.selectedDistro)
    return d ? d.status : 'Unknown'
})

// --- Software Packages Logic ---
const softwarePackages = ref([])
const loadingPackages = ref(false)
const packageSearchQuery = ref('')
const isUninstallingPackage = ref(false)

const filteredPackages = computed(() => {
    if (!packageSearchQuery.value) return softwarePackages.value
        const query = packageSearchQuery.value.toLowerCase()
        return softwarePackages.value.filter(pkg => 
            pkg.name.toLowerCase().includes(query) || 
            pkg.version.toLowerCase().includes(query) ||
            pkg.source.toLowerCase().includes(query)
        )
})

const loadPackages = async () => {
    if (!appStore.selectedDistro) return
    loadingPackages.value = true
    try {
        const pkgs = await GetInstalledPackages(appStore.selectedDistro)
        // Sort alphabetically
        softwarePackages.value = pkgs.sort((a, b) => a.Name.localeCompare(b.Name))
    } catch (e) {
        console.error("Failed to load packages", e)
        // Fallback or empty
        softwarePackages.value = []
    } finally {
        loadingPackages.value = false
    }
}

const handleUninstallPackage = async (pkgName) => {
    if (!appStore.selectedDistro) return
    
    if (confirm(`确定要卸载软件包 "${pkgName}" 吗？\n此操作不可逆。`)) {
        isUninstallingPackage.value = true
        try {
            // Listen for progress
            EventsOn("package:progress", (msg) => {
                console.log("Uninstall progress:", msg)
            })
            
            await UninstallPackage(appStore.selectedDistro, pkgName)
            alert(`成功卸载 ${pkgName}`)
            
            // Refresh list
            await loadPackages()
        } catch (e) {
            console.error("Uninstall failed", e)
            alert(`卸载失败: ${e}`)
        } finally {
            isUninstallingPackage.value = false
            EventsOff("package:progress")
        }
    }
}

watch(activeTab, (newTab) => {
    if (newTab === 'software') {
        loadPackages()
    }
})
</script>

<template>
    <div class="config-view-container">
        <Transition name="page" mode="out-in">
        <!-- List View -->
        <div v-if="!appStore.selectedDistro" class="view-content" key="list">
            <div class="view-header">
                <h2>发行版配置</h2>
                <p class="subtitle">选择一个发行版以管理其系统源、用户及环境配置。</p>
            </div>

            <div v-if="loading" class="loading-state">
                <div class="spinner"></div>
            </div>

            <div v-else class="distro-grid">
                <div 
                    v-for="distro in distros" 
                    :key="distro.name"
                    class="distro-card"
                    @click="appStore.selectDistro(distro.name)"
                >
                    <div class="card-icon-wrapper">
                         <img :src="getDistroIcon(distro.name)" class="distro-img" />
                    </div>
                    <div class="card-info">
                        <div class="card-name">{{ distro.name }}</div>
                        <div class="card-meta">
                            <span class="status-badge" :class="distro.status.toLowerCase()">
                                <span class="status-dot"></span>
                                {{ distro.status }}
                            </span>
                            <span class="version-badge" v-if="distro.version !== '2'">v{{ distro.version }}</span>
                            <span class="version-badge v2-fix" v-else>v{{ distro.version }}</span>
                        </div>
                    </div>
                    <div class="card-arrow">
                        <Settings :size="20" />
                    </div>
                </div>
            </div>
        </div>

        <!-- Detail View -->
        <div v-else class="view-content" key="detail">
            <div class="detail-header-row">
                <button class="back-btn" @click="handleBack" title="返回列表">
                    <ArrowLeft :size="20" />
                    <span>返回列表</span>
                </button>
            </div>
            
            <div class="detail-layout">
                <!-- Left Sidebar Info -->
                <div class="distro-sidebar">
                    <div class="distro-profile">
                        <div class="profile-icon">
                            <img :src="getDistroIcon(appStore.selectedDistro)" />
                        </div>
                        <h3>{{ appStore.selectedDistro }}</h3>
                        <div class="profile-status" :class="currentDistroStatus.toLowerCase()">
                            {{ currentDistroStatus }}
                        </div>
                        <div class="profile-path" :title="currentDistroPath">{{ currentDistroPath }}</div>
                    </div>
                    
                    <div class="sidebar-actions">
                        <button v-if="currentDistroStatus === 'Running'" class="action-btn-sidebar stop" @click="stopDistro" title="停止实例">
                            <Square :size="18" fill="currentColor" />
                            <span>停止</span>
                        </button>
                        <button class="action-btn-sidebar" @click="openDistroFolder" title="打开文件夹">
                            <FolderOpen :size="18" />
                            <span>目录</span>
                        </button>
                        <button class="action-btn-sidebar" @click="openMigrationModal" title="系统迁移">
                            <ArrowRightLeft :size="18" />
                            <span>迁移</span>
                        </button>
                        <button class="action-btn-sidebar danger" @click="handleUninstallClick" title="卸载系统">
                            <Trash2 :size="18" />
                            <span>卸载</span>
                        </button>
                    </div>

                    <div class="sidebar-nav">
                        <div 
                            class="nav-item" 
                            :class="{ active: activeTab === 'general' }"
                            @click="activeTab = 'general'"
                        >
                            <Settings :size="18" />
                            <span>通用设置</span>
                        </div>
                        <div 
                            class="nav-item" 
                            :class="{ active: activeTab === 'environment' }"
                            @click="activeTab = 'environment'"
                        >
                            <Box :size="18" />
                            <span>环境配置</span>
                        </div>
                        <div 
                            class="nav-item" 
                            :class="{ active: activeTab === 'docker' }"
                            @click="activeTab = 'docker'"
                        >
                            <div class="nav-icon-wrapper">
                                <img :src="DockerIcon" class="docker-icon-img" />
                            </div>
                            <span>Docker 配置</span>
                        </div>
                        <div 
                            class="nav-item" 
                            :class="{ active: activeTab === 'software' }"
                            @click="activeTab = 'software'"
                        >
                            <Package :size="18" />
                            <span>软件包</span>
                        </div>
                    </div>
                </div>

                <!-- Right Content Area -->
                <div class="config-panel">
                    <!-- General Settings -->
                    <div v-if="activeTab === 'general'" class="tab-pane fade-in">
                        <div class="panel-header">
                            <h4>{{ packageManagerName }} 软件源设置 (Beta)</h4>
                            <p>配置系统的软件源镜像，加速软件下载和更新。</p>
                        </div>
                        
                        <div class="source-selector">
                            <div 
                                v-for="source in aptSources" 
                                :key="source.value"
                                class="source-option"
                                :class="{ selected: selectedSource === source.value }"
                                @click="selectedSource = source.value"
                            >
                                <div class="radio-circle">
                                    <div class="radio-inner"></div>
                                </div>
                                <div class="source-info">
                                    <span class="source-name">{{ source.name }}</span>
                                    <span class="source-desc">{{ source.desc }}</span>
                                </div>
                            </div>
                        </div>

                        <div class="actions-footer">
                            <button class="btn-primary" @click="handleSave">
                                <Save class="icon" /> 保存并应用更改
                            </button>
                        </div>
                    </div>

                    <!-- Environment Settings -->
                    <div v-else-if="activeTab === 'environment'" class="tab-pane fade-in">
                        <!-- 暂未开放提示 -->
                        <div class="coming-soon-placeholder">
                            <div class="icon">🚧</div>
                            <h3>环境配置暂未开放</h3>
                            <p>该功能正在紧锣密鼓地开发中，敬请期待！</p>
                        </div>

                        <!-- 原有内容 (暂时注释) -->
                        <!--
                        <div class="panel-header">
                            <h4>一键环境部署</h4>
                            <p>快速在当前发行版中安装常用开发环境。</p>
                        </div>
                        
                        <div v-if="currentDistroStatus !== 'Running'" class="warning-banner">
                            <AlertTriangle :size="20" />
                            <div class="warning-content">
                                <strong>系统未启动</strong>
                                <p>请先启动 {{ appStore.selectedDistro }} 以执行环境安装操作。</p>
                            </div>
                        </div>

                        <div class="env-section-title">本地预设</div>
                        <div class="env-list-modern">
                            <div v-if="envStore.localEnvironments.length === 0" class="empty-hint">暂无本地预设</div>
                            <div v-for="env in envStore.localEnvironments" :key="env.id" class="env-item-modern">
                                <div class="env-icon-box local">
                                    <component :is="getIcon(env.icon)" :size="24" />
                                </div>
                                <div class="env-details">
                                    <div class="env-title">{{ env.name }} <span class="env-ver">v{{ env.version }}</span></div>
                                    <div class="env-desc">{{ env.description }}</div>
                                </div>
                                <button 
                                    class="action-btn" 
                                    @click="handleInstallEnvironment(env)"
                                    :disabled="currentDistroStatus !== 'Running'"
                                >
                                    安装
                                </button>
                            </div>
                        </div>

                        <div class="env-section-title">官方推荐</div>
                        <div class="env-list-modern">
                            <div v-for="env in envStore.environments" :key="env.id" class="env-item-modern">
                                <div class="env-icon-box" :class="env.id">
                                    <component :is="getIcon(env.icon)" :size="24" />
                                </div>
                                <div class="env-details">
                                    <div class="env-title">{{ env.name }} <span class="env-ver">v{{ env.version }}</span></div>
                                    <div class="env-desc">{{ env.description }}</div>
                                </div>
                                <button 
                                    class="action-btn" 
                                    @click="handleInstallEnvironment(env)"
                                    :disabled="currentDistroStatus !== 'Running'"
                                >
                                    安装
                                </button>
                            </div>
                        </div>
                        -->
                    </div>

                    <!-- Docker Settings -->
                    <div v-else-if="activeTab === 'docker'" class="tab-pane fade-in">
                        <!-- 暂未开放提示 -->
                        <div class="coming-soon-placeholder">
                            <div class="icon">🐳</div>
                            <h3>Docker 配置暂未开放</h3>
                            <p>该功能正在紧锣密鼓地开发中，敬请期待！</p>
                        </div>
                    </div>

                    <!-- Software Packages -->
                    <div v-else-if="activeTab === 'software'" class="tab-pane fade-in">
                        <div class="panel-header">
                            <h4>已安装软件包</h4>
                            <p>查看和管理当前发行版中已安装的系统软件包。</p>
                        </div>

                        <div class="search-bar-container">
                            <Search class="search-icon" :size="18" />
                            <input 
                                type="text" 
                                v-model="packageSearchQuery" 
                                placeholder="搜索软件包名称或版本..." 
                                class="search-input"
                            />
                        </div>

                        <div v-if="loadingPackages" class="loading-state">
                            <div class="spinner"></div>
                            <p>正在加载软件包列表...</p>
                        </div>

                        <div v-else class="package-list-container">
                            <div v-if="filteredPackages.length === 0" class="empty-state-small">
                                <p>未找到匹配的软件包</p>
                            </div>
                            <div v-else class="package-grid">
                                <div v-for="pkg in filteredPackages" :key="pkg.name" class="package-item">
                                    <div class="package-info">
                                        <span class="pkg-name">{{ pkg.name }}</span>
                                        <span class="pkg-version">v{{ pkg.version }}</span>
                                        <span class="pkg-source" :title="pkg.source">来源: {{ pkg.source }}</span>
                                    </div>
                                    <button 
                                        class="btn-uninstall-pkg" 
                                        @click="handleUninstallPackage(pkg.name)"
                                        :disabled="isUninstallingPackage"
                                        title="卸载此软件包"
                                    >
                                        <Trash2 :size="14" />
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        </Transition>

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
.config-view-container {
    padding: 2rem;
    height: 100%;
    overflow-y: auto;
    box-sizing: border-box;
    background: var(--color-bg-body);
}

.view-header h2 {
    margin-bottom: 0.5rem;
    font-size: 2rem;
    font-weight: 700;
    letter-spacing: -0.5px;
    color: var(--color-text-primary);
}

.subtitle {
    color: var(--color-text-secondary);
    margin-bottom: 2.5rem;
    font-size: 1.05rem;
}

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

/* Distro Grid */
.distro-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 1.5rem;
}

.distro-card {
    background: var(--color-bg-card);
    border: 1px solid var(--color-border);
    border-radius: 16px;
    padding: 1.5rem;
    display: flex;
    align-items: center;
    gap: 1.25rem;
    cursor: pointer;
    transition: all 0.3s cubic-bezier(0.25, 0.8, 0.25, 1);
    position: relative;
    overflow: hidden;
}

.distro-card:hover {
    transform: translateY(-4px);
    box-shadow: 0 12px 32px rgba(0,0,0,0.08);
    border-color: var(--color-brand);
}

.distro-card::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    width: 4px;
    height: 100%;
    background: var(--color-brand);
    opacity: 0;
    transition: opacity 0.3s;
}

.distro-card:hover::before {
    opacity: 1;
}

.card-icon-wrapper {
    width: 64px;
    height: 64px;
    background: var(--color-bg-body);
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
}

.distro-img {
    width: 40px;
    height: 40px;
    object-fit: contain;
}

.card-info {
    flex: 1;
}

.card-name {
    font-weight: 700;
    font-size: 1.15rem;
    margin-bottom: 0.5rem;
    color: var(--color-text-primary);
}

.card-meta {
    display: flex;
    gap: 0.75rem;
    align-items: center;
    height: 24px; /* 统一高度，防止错位 */
}

.status-badge {
    font-size: 0.8rem;
    padding: 0 10px;
    height: 24px;
    border-radius: 12px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    font-weight: 500;
    background: var(--color-bg-body);
    color: var(--color-text-secondary);
    line-height: normal;
}

.status-badge.running {
    background: rgba(76, 175, 80, 0.1);
    color: #4caf50;
}

.status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: currentColor;
    flex-shrink: 0;
}

.version-badge {
    font-size: 0.75rem;
    color: var(--color-text-secondary);
    background: var(--color-bg-body);
    padding: 0 8px;
    height: 24px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 6px;
    line-height: normal;
    border: 1px solid transparent; /* 防止布局抖动 */
}

.version-badge.v2-fix {
  position: relative;
  top: 1px;
}

.card-arrow {
    color: var(--color-text-secondary);
    opacity: 0.3;
    transition: all 0.3s;
}

.distro-card:hover .card-arrow {
    opacity: 1;
    color: var(--color-brand);
    transform: translateX(4px);
}

/* Detail View Layout */
.detail-header-row {
    margin-bottom: 1.5rem;
}

.back-btn {
    background: none;
    border: none;
    cursor: pointer;
    color: var(--color-text-secondary);
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 1rem;
    padding: 0;
    transition: color 0.2s;
}

.back-btn:hover {
    color: var(--color-brand);
}

.detail-layout {
    display: flex;
    gap: 2rem;
    height: calc(100% - 60px);
}

/* Sidebar */
.distro-sidebar {
    width: 280px;
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
}

.distro-profile {
    background: var(--color-bg-card);
    border: 1px solid var(--color-border);
    border-radius: 16px;
    padding: 2rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
}

.profile-icon {
    width: 80px;
    height: 80px;
    background: var(--color-bg-body);
    border-radius: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 1rem;
}

.profile-icon img {
    width: 50px;
    height: 50px;
    object-fit: contain;
}

.distro-profile h3 {
    margin: 0 0 0.5rem 0;
    font-size: 1.25rem;
}

.profile-status {
    font-size: 0.85rem;
    padding: 4px 12px;
    border-radius: 100px;
    background: var(--color-bg-body);
    color: var(--color-text-secondary);
    font-weight: 500;
}

.profile-status.running {
    background: rgba(76, 175, 80, 0.1);
    color: #4caf50;
}

.profile-path {
    font-size: 0.75rem;
    color: var(--color-text-secondary);
    margin-top: 0.5rem;
    word-break: break-all;
    opacity: 0.7;
    font-family: 'Consolas', monospace;
    max-width: 100%;
}

.sidebar-nav {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.nav-item {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1rem 1.5rem;
    border-radius: 12px;
    cursor: pointer;
    color: var(--color-text-secondary);
    transition: all 0.2s;
    font-weight: 500;
}

.nav-item:hover {
    background: var(--color-bg-card);
    color: var(--color-text-primary);
}

.nav-item.active {
    background: var(--color-brand);
    color: white;
    box-shadow: 0 4px 12px rgba(24, 144, 255, 0.3);
}

.nav-icon-wrapper {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 18px;
    height: 18px;
}

.docker-icon-img {
    width: 18px;
    height: 18px;
    object-fit: contain;
}

/* Config Panel */
.config-panel {
    flex: 1;
    background: var(--color-bg-card);
    border-radius: 16px;
    border: 1px solid var(--color-border);
    overflow: hidden;
    display: flex;
    flex-direction: column;
}

.tab-pane {
    padding: 2.5rem;
    overflow-y: auto;
    height: 100%;
}

.panel-header {
    margin-bottom: 2.5rem;
    border-bottom: 1px solid var(--color-border);
    padding-bottom: 1.5rem;
}

.panel-header h4 {
    margin: 0 0 0.5rem 0;
    font-size: 1.5rem;
}

.panel-header p {
    margin: 0;
    color: var(--color-text-secondary);
}

/* Source Selector */
.source-selector {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    margin-bottom: 2rem;
}

.source-option {
    display: flex;
    align-items: flex-start;
    gap: 1rem;
    padding: 1.25rem;
    border: 1px solid var(--color-border);
    border-radius: 12px;
    cursor: pointer;
    transition: all 0.2s;
}

.source-option:hover {
    border-color: var(--color-brand);
    background: var(--color-bg-body);
}

.source-option.selected {
    border-color: var(--color-brand);
    background: rgba(24, 144, 255, 0.08);
    box-shadow: 0 0 0 2px rgba(24, 144, 255, 0.2);
}

.radio-circle {
    width: 20px;
    height: 20px;
    border-radius: 50%;
    border: 2px solid var(--color-border);
    display: flex;
    align-items: center;
    justify-content: center;
    margin-top: 2px;
    transition: all 0.2s;
    background: var(--color-bg-body);
}

.source-option.selected .radio-circle {
    border-color: var(--color-brand);
}

.radio-inner {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--color-brand);
    opacity: 0;
    transform: scale(0.5);
    transition: all 0.2s;
}

.source-option.selected .radio-inner {
    opacity: 1;
    transform: scale(1);
}

.source-info {
    flex: 1;
}

.source-name {
    display: block;
    font-weight: 600;
    margin-bottom: 0.25rem;
    color: var(--color-text-primary);
}

.source-desc {
    font-size: 0.9rem;
    color: var(--color-text-secondary);
}

.actions-footer {
    display: flex;
    justify-content: flex-end;
    padding-top: 2rem;
    border-top: 1px solid var(--color-border);
}

/* Environment List Modern */
.warning-banner {
    background: rgba(255, 193, 7, 0.1);
    border: 1px solid rgba(255, 193, 7, 0.3);
    border-radius: 12px;
    padding: 1rem 1.5rem;
    display: flex;
    align-items: flex-start;
    gap: 1rem;
    color: #b48505;
    margin-bottom: 2rem;
}

.warning-banner.error-banner {
    background: rgba(255, 77, 79, 0.1);
    border-color: rgba(255, 77, 79, 0.3);
    color: #ff4d4f;
}

.warning-content strong {
    display: block;
    margin-bottom: 0.25rem;
}

.warning-content p {
    margin: 0;
    font-size: 0.9rem;
    opacity: 0.9;
}

.env-section-title {
    font-size: 0.9rem;
    font-weight: 600;
    text-transform: uppercase;
    color: var(--color-text-secondary);
    letter-spacing: 0.5px;
    margin-bottom: 1rem;
    margin-top: 2rem;
}
.env-section-title:first-of-type {
    margin-top: 0;
}

.env-list-modern {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.env-item-modern {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 1.25rem;
    padding: 1.25rem;
    border: 1px solid var(--color-border);
    border-radius: 12px;
    background: var(--color-bg-body);
    transition: all 0.2s;
}

.env-item-modern:hover {
    border-color: var(--color-brand);
    transform: translateX(4px);
    background: var(--color-bg-card);
}

.env-icon-box {
    width: 48px;
    height: 48px;
    border-radius: 10px;
    background: var(--color-bg-card);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--color-brand);
    flex-shrink: 0;
}

.env-icon-box.local { color: var(--color-text-primary); border: 1px solid var(--color-border); }
.env-icon-box.docker { color: #0db7ed; background: rgba(13, 183, 237, 0.1); }
.env-icon-box.openwrt { color: #cf3e52; background: rgba(207, 62, 82, 0.1); }
.env-icon-box.nodejs_dev { color: #68a063; background: rgba(104, 160, 99, 0.1); }
.env-icon-box.python_data { color: #3776ab; background: rgba(55, 118, 171, 0.1); }
.env-icon-box.go_dev { color: #00add8; background: rgba(0, 173, 216, 0.1); }

.env-details {
    flex: 1;
    min-width: 200px;
}

.env-title {
    font-weight: 600;
    font-size: 1.05rem;
    margin-bottom: 0.25rem;
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.env-ver {
    font-size: 0.75rem;
    background: rgba(0,0,0,0.05);
    padding: 2px 6px;
    border-radius: 4px;
    color: var(--color-text-secondary);
    font-weight: normal;
}

.env-desc {
    font-size: 0.9rem;
    color: var(--color-text-secondary);
}

.action-btn {
    padding: 0.5rem 1.25rem;
    background: var(--color-bg-card);
    border: 1px solid var(--color-border);
    border-radius: 8px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    font-size: 0.9rem;
    color: var(--color-text-primary); /* 修复暗色模式下文字看不清的问题 */
}

.action-btn:hover:not(:disabled) {
    background: var(--color-brand);
    color: white;
    border-color: var(--color-brand);
}

.action-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.empty-hint {
    font-style: italic;
    color: var(--color-text-secondary);
    font-size: 0.9rem;
}

/* Search Bar */
.search-bar-container {
    position: relative;
    margin-bottom: 2rem;
}

.search-icon {
    position: absolute;
    left: 12px;
    top: 50%;
    transform: translateY(-50%);
    color: var(--color-text-secondary);
}

.search-input {
    width: 100%;
    padding: 12px 12px 12px 40px;
    font-size: 1rem;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    background: var(--color-bg-body);
    color: var(--color-text-primary);
    transition: all var(--transition-fast);
}

.search-input:focus {
    border-color: var(--color-brand);
    box-shadow: 0 0 0 2px rgba(24, 144, 255, 0.1);
    outline: none;
}

/* Package List */
.package-list-container {
    background: var(--color-bg-body);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    max-height: 500px;
    overflow-y: auto;
}

.empty-state-small {
    padding: 3rem;
    text-align: center;
    color: var(--color-text-secondary);
}

.package-grid {
    display: flex;
    flex-direction: column;
}

.package-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    border-bottom: 1px solid var(--color-border);
    transition: background var(--transition-fast);
}

.package-item:last-child {
    border-bottom: none;
}

.package-item:hover {
    background: var(--color-bg-hover);
}

.package-info {
    display: flex;
    flex-direction: column;
}

.pkg-name {
    font-weight: 600;
    color: var(--color-text-primary);
    font-size: 0.95rem;
}

.pkg-version {
    font-size: 0.8rem;
    color: var(--color-text-secondary);
    margin-top: 2px;
}

.pkg-source {
    font-size: 0.75rem;
    color: var(--color-text-tertiary);
    margin-top: 2px;
    display: block;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 150px;
}

.btn-uninstall-pkg {
    background: transparent;
    border: none;
    color: var(--color-text-secondary);
    cursor: pointer;
    padding: 6px;
    border-radius: var(--radius-sm);
    transition: all var(--transition-fast);
    display: flex;
    align-items: center;
    justify-content: center;
}

.btn-uninstall-pkg:hover:not(:disabled) {
    background: rgba(255, 77, 79, 0.1);
    color: var(--color-error);
}

.btn-uninstall-pkg:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

/* Button & Spinner Reusables */
.btn-primary {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1.5rem;
    background: var(--color-brand);
    color: white;
    border: none;
    border-radius: 8px;
    cursor: pointer;
    font-weight: 600;
    transition: opacity 0.2s;
}

.btn-primary:hover {
    opacity: 0.9;
}

/* Spinner (Unified in main.css) */
/* .spinner, .loading-state, .fade-in, @keyframes spin moved to main.css */

/* .progress-bar-container and related classes moved to main.css */

/* Sidebar Actions */
.sidebar-actions {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 0 0.5rem;
}

.action-btn-sidebar {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 0.75rem 1.5rem;
    border-radius: 12px;
    cursor: pointer;
    color: var(--color-text-secondary);
    transition: all 0.2s;
    font-weight: 500;
    border: 1px solid transparent;
    background: transparent;
    width: 100%;
    text-align: left;
}

.action-btn-sidebar:hover {
    background: var(--color-bg-card);
    color: var(--color-text-primary);
    border-color: var(--color-border);
}

.action-btn-sidebar.stop {
    color: #ff4d4f;
}
.action-btn-sidebar.stop:hover {
    background: rgba(255, 77, 79, 0.1);
    border-color: rgba(255, 77, 79, 0.2);
}

.action-btn-sidebar.danger {
    color: #ff4d4f;
    opacity: 0.8;
}
.action-btn-sidebar.danger:hover {
    background: rgba(255, 77, 79, 0.1);
    border-color: rgba(255, 77, 79, 0.2);
    opacity: 1;
}

/* === 模态框优化 (From HomeView) === */
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
  background: var(--color-bg-card); /* Adapted variable */
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 10px 40px rgba(0,0,0,0.2); /* Adapted shadow */
  border: 1px solid var(--color-border);
  display: flex; flex-direction: column;
}

.modal-header {
  padding: 16px 24px;
  background: var(--color-bg-body); /* Adapted variable */
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
  background: var(--color-bg-body);
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
.warning-content h4 { margin: 0 0 4px 0; color: #ff4d4f; font-size: 15px; }
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
.cancel-btn:hover { border-color: var(--color-text-primary); color: var(--color-text-primary); background: var(--color-bg-body); }
.danger-btn {
    padding: 8px 24px; border-radius: 6px; cursor: pointer;
    background: #ff4d4f; border: none; color: white; font-weight: 500;
    box-shadow: 0 4px 10px rgba(255, 77, 79, 0.3);
}
.danger-btn:hover { background: #ff7875; }
.danger-btn:disabled { opacity: 0.6; cursor: not-allowed; }

/* Form Styles */
.form-group { margin-bottom: 16px; }
.form-group label { display: block; margin-bottom: 6px; font-size: 13px; color: var(--color-text-secondary); }
.input { 
    width: 100%; padding: 8px 12px; 
    border-radius: 6px; border: 1px solid var(--color-border); 
    background: var(--color-bg-body); 
    color: var(--color-text-primary);
    font-size: 13px;
}
.input:disabled { opacity: 0.7; cursor: not-allowed; }
.path-input-group { display: flex; gap: 8px; }

/* Config Error */
.config-error {
    background: rgba(255, 77, 79, 0.1);
    border: 1px solid rgba(255, 77, 79, 0.2);
    color: #ff4d4f;
    padding: 10px 12px;
    border-radius: 6px;
    font-size: 13px;
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: -8px;
    margin-bottom: 8px;
}

.error-icon-sm { font-size: 14px; }

/* Uninstall Log */
.uninstall-log {
    margin-top: -12px;
    margin-bottom: 12px;
    font-size: 12px;
    color: var(--color-text-secondary);
    background: var(--color-bg-body);
    padding: 8px 12px;
    border-radius: 6px;
    font-family: monospace;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    text-align: center;
}

/* Progress & Hero */
.install-hero { margin-bottom: 24px; text-align: center; }
.hero-icon { width: 64px; height: 64px; object-fit: contain; margin-bottom: 16px; }
.hero-info h3 { margin: 0 0 4px 0; font-size: 18px; color: var(--color-text-primary); }
.log-detail { font-size: 12px; color: var(--color-text-secondary); margin: 0; font-family: monospace; }

/* .progress-bar-container and related classes moved to main.css */
/* .progress-track, .progress-fill, .progress-glow, .progress-text, @keyframes scan moved to main.css */

/* Error State */
.error-container { text-align: center; padding: 20px; }
.error-symbol { font-size: 48px; display: block; margin-bottom: 16px; }
.error-desc { color: #ff4d4f; margin-bottom: 24px; font-size: 16px; font-weight: 500; }

.btn { padding: 6px 16px; border-radius: 6px; border: none; cursor: pointer; font-size: 13px; transition: all 0.2s; font-weight: 500; }
.btn-secondary { background: var(--color-bg-body); color: var(--color-text-primary); border: 1px solid var(--color-border); }
.btn-secondary:hover { border-color: var(--color-text-secondary); }
.btn-danger { background: #ff4d4f; color: #fff; }

/* Modal Transition */
.modal-enter-active, .modal-leave-active { transition: opacity 0.3s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
</style>