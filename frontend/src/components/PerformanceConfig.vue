<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { usePerformanceStore } from '../stores/performance'
// Import backend functions (mocked if running in browser without wails)
import { SelectDirectory, GetPerformanceConfig, SavePerformanceConfig } from '../../wailsjs/go/main/App'

const store = usePerformanceStore()
const form = reactive({ ...store.$state })
const errors = reactive({
  memoryLimit: '',
  swap: '',
  processorCount: '',
  networkMode: '',
})
const showToast = ref(false)
const hasChanges = ref(false)
const showChangeModal = ref(false)
const showRestartWarning = ref(false) // Restart warning modal state
const showResetWarning = ref(false) // Reset warning modal state
const isSaving = ref(false) // Loading state for save operation
const toastMessage = ref('配置已保存') // Dynamic toast message

// Watch for changes to show the modal
watch(form, (newVal) => {
  // Simple check if form differs from store state (which represents last saved/loaded state)
  // Note: deep comparison might be needed for robust check, but for now simple diff
  const keys = Object.keys(store.$state)
  let changed = false
  for (const key of keys) {
    if (form[key] !== store.$state[key]) {
      changed = true
      break
    }
  }
  
  if (changed && !hasChanges.value) {
    hasChanges.value = true
    showChangeModal.value = true
    // No timeout for hiding, user requested "asynchronous small window" which usually persists until action
  } else if (!changed) {
    hasChanges.value = false
    showChangeModal.value = false
  }
}, { deep: true })

const handleReloadConfig = async () => {
    try {
        // Call backend to get config
        const config = await GetPerformanceConfig()
        
        // Normalize select values (case-insensitive match)
        // Network Mode
        const networkModes = ['mirrored', 'nat', 'bridged', 'virtioproxy', 'none']
        const matchedNetworkMode = networkModes.find(m => m.toLowerCase() === (config.networkMode || '').toLowerCase())
        if (matchedNetworkMode) config.networkMode = matchedNetworkMode
        
        // Auto Memory Reclaim
        const reclaimModes = ['dropCache', 'gradual', 'disabled']
        const matchedReclaimMode = reclaimModes.find(m => m.toLowerCase() === (config.autoMemoryReclaim || '').toLowerCase())
        if (matchedReclaimMode) config.autoMemoryReclaim = matchedReclaimMode

        store.setPerformanceConfig(config)
        Object.assign(form, config)
        hasChanges.value = false
        showChangeModal.value = false
        toastMessage.value = '配置已恢复'
        showToast.value = true
        setTimeout(() => showToast.value = false, 2000)
    } catch (e) {
        console.error("Failed to load config:", e)
        alert("加载配置失败: " + e)
    }
}

// Mock Backend Limits (TODO: Replace with actual backend call)
const systemLimits = reactive({
  maxMemory: 32, // Default fallback
  maxProcessors: 12 // Default fallback
})

onMounted(async () => {
  // Sync form with store on mount
  Object.assign(form, store.$state)
  
  // TODO: Call backend to get system specs
  // Example: const specs = await GetSystemSpecs()
  // systemLimits.maxMemory = specs.totalMemoryGB
  // systemLimits.maxProcessors = specs.logicalCores
  
  console.log('Fetching system limits from backend...')
  // Mock async delay
  setTimeout(() => {
    systemLimits.maxMemory = 64 // Mock 64GB RAM
    systemLimits.maxProcessors = 16 // Mock 16 Cores
    console.log('System limits updated:', systemLimits)
  }, 500)

  // TODO: Call backend to get current performance config
  // Example: const config = await GetPerformanceConfig()
  console.log('Fetching configuration from backend...')
  setTimeout(async () => {
    try {
        await handleReloadConfig()
        // Override toast behavior for initial load
        showToast.value = false
    } catch (e) {
        console.error("Initial config load failed:", e)
    }
  }, 500)
})

const validateField = (field) => {
  errors[field] = ''
  if (field === 'memoryLimit') {
    if (!Number.isInteger(Number(form.memoryLimit)) || form.memoryLimit < 1) {
      errors.memoryLimit = '请输入正整数'
      return false
    }
    if (form.memoryLimit > systemLimits.maxMemory) {
        errors.memoryLimit = `不能超过系统最大内存 (${systemLimits.maxMemory} GB)`
        return false
    }
  }
  if (field === 'swap') {
    if (form.swap < 0) {
      errors.swap = 'Swap 不能为负数'
      return false
    }
  }
  if (field === 'processorCount') {
    if (!Number.isInteger(Number(form.processorCount)) || form.processorCount < 1) {
      errors.processorCount = '处理器数量必须为正整数'
      return false
    }
    if (form.processorCount > systemLimits.maxProcessors) {
        errors.processorCount = `不能超过系统核心数 (${systemLimits.maxProcessors})`
        return false
    }
  }
  if (field === 'vmIdleTimeout') {
      if (form.vmIdleTimeout < 0) {
          return false // Simple check, error handling could be more verbose
      }
  }
  return true
}

const handleSelectSwapFile = async () => {
  try {
    // Call backend to select directory
    // Note: User requested "SelectDirectory", but logical behavior might be selecting a file path.
    // We will follow instruction to use SelectDirectory and append filename, or assume user meant SelectFile.
    // Given the instruction "点击框内调用SelectDirectory()后端函数选择路径", we use SelectDirectory.
    let path = await SelectDirectory()
    if (path) {
        // Ensure path ends with separator before appending default name if needed
        // Or if user just wants the directory where the swap file lives. 
        // Typically .wslconfig expects a full path to the file.
        // Let's assume we append '\wsl.swap' if a directory is chosen, or user manually edits.
        // For now, let's just set the path. If it's a directory, maybe we should add the filename.
        if (!path.endsWith('.swap')) {
            path = path.replace(/\\$/, '') + '\\wsl.swap'
        }
        form.swapFile = path
    }
  } catch (e) {
    console.error("Failed to select directory:", e)
    // Mock for browser dev without backend
    form.swapFile = 'D:\\MockPath\\wsl.swap'
  }
}

const handleSaveClick = () => {
    const isValidMemory = validateField('memoryLimit')
    const isValidSwap = validateField('swap')
    const isValidProcessor = validateField('processorCount')
    
    if (isValidMemory && isValidSwap && isValidProcessor) {
        showRestartWarning.value = true
    }
}

const executeSave = async () => {
  showRestartWarning.value = false
  isSaving.value = true
  try {
      // Call backend to save
      // We can pass the whole form or formatted config. 
      // For simplicity, let's assume backend accepts the struct matching form.
      // Or we use exportWslConfig() locally and send string? 
      // Instructions say "save function ... bind same function".
      // Let's assume SavePerformanceConfig accepts the object.
      await SavePerformanceConfig(form)
      
      // Update local store
      store.setPerformanceConfig({ ...form })
      hasChanges.value = false
      showChangeModal.value = false
      toastMessage.value = '配置已保存'
      showToast.value = true
      setTimeout(() => {
        showToast.value = false
      }, 2000)
  } catch (e) {
      console.error("Save failed:", e)
      alert("保存失败: " + e)
  } finally {
      isSaving.value = false
  }
}

const handleReset = () => {
  showResetWarning.value = true
}

const executeReset = async () => {
  showResetWarning.value = false
  await handleReloadConfig()
}
</script>

<template>
  <div class="performance-view-container">
    <div class="view-header">
      <h2>WSL2 性能配置</h2>
      <p class="subtitle">管理 .wslconfig 全局配置，优化子系统运行效率。</p>
    </div>

    <div class="config-card">
      <!-- 核心资源限制 -->
      <section class="config-section">
        <h4 class="section-title">核心资源限制</h4>
        <div class="form-grid">
          <div class="form-group">
            <label>内存限制 (Memory)</label>
            <div class="input-suffix-wrapper">
              <input 
                v-model.number="form.memoryLimit" 
                type="number" 
                class="input" 
                :class="{ 'input-error': errors.memoryLimit }"
                @blur="validateField('memoryLimit')"
              >
              <span class="suffix">GB</span>
            </div>
            <span class="annotation">设置 WSL2 虚拟机可使用的最大内存。建议不超过物理内存的 80% (当前上限: {{ systemLimits.maxMemory }} GB)。</span>
            <span class="error-text" v-if="errors.memoryLimit">{{ errors.memoryLimit }}</span>
          </div>

          <div class="form-group">
            <label>处理器数量 (Processors)</label>
            <input 
              v-model.number="form.processorCount" 
              type="number" 
              class="input"
              :class="{ 'input-error': errors.processorCount }"
              @blur="validateField('processorCount')"
            >
            <span class="annotation">分配给 WSL2 的虚拟处理器核心数 (当前系统核心数: {{ systemLimits.maxProcessors }})。</span>
            <span class="error-text" v-if="errors.processorCount">{{ errors.processorCount }}</span>
          </div>

          <div class="form-group">
            <label>交换空间 (Swap)</label>
            <div class="input-suffix-wrapper">
              <input 
                v-model.number="form.swap" 
                type="number" 
                class="input" 
                :class="{ 'input-error': errors.swap }"
                @blur="validateField('swap')"
              >
              <span class="suffix">GB</span>
            </div>
            <span class="annotation">设置交换空间大小。0 表示禁用。</span>
            <span class="error-text" v-if="errors.swap">{{ errors.swap }}</span>
          </div>

          <div class="form-group">
            <label>交换文件路径 (Swap File)</label>
            <div class="input-action-wrapper">
                <input 
                v-model="form.swapFile" 
                type="text" 
                class="input" 
                readonly
                @click="handleSelectSwapFile"
                placeholder="点击选择路径"
                >
            </div>
            <span class="annotation">指定交换文件的存储位置 (默认: %Temp%\swap.vhdx)。</span>
          </div>


          <div class="form-group">
            <label>默认 VHD 大小 (Default VHD Size)</label>
            <div class="input-suffix-wrapper">
              <input 
                v-model.number="form.defaultVhdSize" 
                type="number" 
                class="input"
              >
              <span class="suffix">GB</span>
            </div>
            <span class="annotation">限制分发文件系统允许占用的最大大小 (默认: 1024 GB / 1 TB)。</span>
          </div>
        </div>
      </section>

      <div class="divider"></div>

      <!-- 高级内核设置 (Advanced Kernel) -->
      <section class="config-section">
        <h4 class="section-title">高级内核设置 (Advanced Kernel)</h4>
        <div class="form-grid">
            <div class="form-group">
                <label>自定义内核路径 (Kernel)</label>
                <input v-model="form.kernel" type="text" class="input" placeholder="留空使用 Microsoft 内置内核">
                <span class="annotation">自定义 Linux 内核的绝对 Windows 路径。</span>
            </div>
            <div class="form-group">
                <label>内核命令行 (Kernel Command Line)</label>
                <input v-model="form.kernelCommandLine" type="text" class="input" placeholder="例如: debug">
                <span class="annotation">其他内核命令行参数。</span>
            </div>
            <div class="switch-item-inline">
                <div class="switch-info">
                    <span class="switch-label">安全模式 (Safe Mode)</span>
                    <span class="switch-annotation">禁用许多功能，用于恢复处于错误状态的发行版。</span>
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="form.safeMode">
                  <span class="slider round"></span>
                </label>
            </div>
            <div class="switch-item-inline">
                <div class="switch-info">
                    <span class="switch-label">嵌套虚拟化 (Nested Virtualization)</span>
                    <span class="switch-annotation">允许在 WSL 2 中运行其他嵌套 VM (如 Docker)。</span>
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="form.nestedVirtualization">
                  <span class="slider round"></span>
                </label>
            </div>
            <div class="switch-item-inline">
                <div class="switch-info">
                    <span class="switch-label">页面报告 (Page Reporting)</span>
                    <span class="switch-annotation">允许 Windows 回收未使用的内存页面。</span>
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="form.pageReporting">
                  <span class="slider round"></span>
                </label>
            </div>
        </div>
      </section>

      <div class="divider"></div>

      <!-- 网络配置 -->
      <section class="config-section">
        <h4 class="section-title">网络配置 (Networking)</h4>
        <div class="form-grid">
            <div class="form-group">
                <label>网络模式 (Networking Mode)</label>
                <select v-model="form.networkMode" class="input">
                  <option value="mirrored">mirrored (镜像模式 - 推荐)</option>
                  <option value="nat">nat (NAT 模式 - 默认)</option>
                  <option value="bridged">bridged (桥接模式 - 已弃用)</option>
                  <option value="virtioproxy">virtioproxy</option>
                  <option value="none">none (无网络)</option>
                </select>
                <span class="annotation">镜像模式可实现主机与 WSL 共享 IP；NAT 模式为传统虚拟网络。</span>
            </div>

            <div class="form-group" v-if="form.networkMode === 'mirrored'">
                <label>忽略端口 (Ignored Ports)</label>
                <input v-model="form.ignoredPorts" type="text" class="input" placeholder="例如: 3000,9000">
                <span class="annotation">指定 Linux 应用程序可以绑定到哪些端口（即使该端口已在 Windows 中使用）。</span>
            </div>

            <div class="switch-item-inline">
                <div class="switch-info">
                    <span class="switch-label">本地回环转发 (Localhost Forwarding)</span>
                    <span class="switch-annotation">允许从 Windows 访问 WSL 中监听 localhost 的服务。</span>
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="form.localhostForwarding">
                  <span class="slider round"></span>
                </label>
            </div>

            <div class="switch-item-inline">
                <div class="switch-info">
                    <span class="switch-label">DNS 隧道 (DNS Tunneling)</span>
                    <span class="switch-annotation">改善网络环境复杂时的域名解析稳定性。</span>
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="form.dnsTunneling">
                  <span class="slider round"></span>
                </label>
            </div>
            
            <div class="form-group" v-if="form.dnsTunneling">
                <label>DNS 隧道 IP (DNS Tunneling IP)</label>
                <input v-model="form.dnsTunnelingIpAddress" type="text" class="input">
                <span class="annotation">指定在启用 DNS 隧道时将在 Linux resolv.conf 文件中配置的名称服务器。</span>
            </div>

            <div class="switch-item-inline" v-if="form.dnsTunneling">
                <div class="switch-info">
                    <span class="switch-label">尽力而为 DNS 解析 (Best Effort DNS Parsing)</span>
                    <span class="switch-annotation">Windows 将尝试解析 DNS 请求，忽略未知记录。</span>
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="form.bestEffortDnsParsing">
                  <span class="slider round"></span>
                </label>
            </div>

            <div class="switch-item-inline">
                <div class="switch-info">
                    <span class="switch-label">DNS 代理 (DNS Proxy)</span>
                    <span class="switch-annotation">将 Linux 中的 DNS 服务器配置为主机上的 NAT (仅适用于 NAT 模式)。</span>
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="form.dnsProxy">
                  <span class="slider round"></span>
                </label>
            </div>

            <div class="switch-item-inline">
                <div class="switch-info">
                    <span class="switch-label">防火墙同步 (Firewall)</span>
                    <span class="switch-annotation">将 Windows 防火墙规则自动应用到 WSL 实例中。</span>
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="form.firewall">
                  <span class="slider round"></span>
                </label>
            </div>

            <div class="switch-item-inline">
                <div class="switch-info">
                    <span class="switch-label">自动代理 (Auto Proxy)</span>
                    <span class="switch-annotation">强制 WSL 使用 Windows 的 HTTP/HTTPS 代理设置。</span>
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="form.autoProxy">
                  <span class="slider round"></span>
                </label>
            </div>

            <div class="switch-item-inline" v-if="form.networkMode === 'mirrored'">
                <div class="switch-info">
                    <span class="switch-label">回环地址访问 (Host Address Loopback)</span>
                    <span class="switch-annotation">允许容器通过分配给主机的 IP 地址连接到主机。</span>
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="form.hostAddressLoopback">
                  <span class="slider round"></span>
                </label>
            </div>
        </div>
      </section>

      <div class="divider"></div>

      <!-- WSLg 配置 (新增) -->
      <section class="config-section">
          <h4 class="section-title">WSLg (GUI 应用程序)</h4>
          <div class="switch-list">
              <div class="switch-item">
                <div class="switch-info">
                    <span class="switch-label">启用 GUI 应用程序 (GUI Applications)</span>
                    <span class="switch-annotation">允许在 WSL 中运行 Linux GUI 应用程序。</span>
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="form.guiApplications">
                  <span class="slider round"></span>
                </label>
              </div>

              <div class="switch-item">
                <div class="switch-info">
                    <span class="switch-label">调试控制台 (Debug Console)</span>
                    <span class="switch-annotation">启用 WSLg 系统的调试控制台 (仅供开发调试使用)。</span>
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="form.debugConsole">
                  <span class="slider round"></span>
                </label>
              </div>
          </div>
      </section>

      <div class="divider"></div>

      <!-- 实验性功能 -->
      <section class="config-section">
        <h4 class="section-title">实验性功能 (Experimental)</h4>
        <div class="switch-list">
          <div class="form-group" style="margin-bottom: 16px;">
            <label>内存自动回收 (Auto Memory Reclaim)</label>
            <select v-model="form.autoMemoryReclaim" class="input">
              <option value="dropCache">dropCache (立即回收 - 默认)</option>
              <option value="gradual">gradual (缓慢回收)</option>
              <option value="disabled">disabled (禁用)</option>
            </select>
            <span class="annotation">控制空闲时如何释放缓存内存回宿主机。</span>
          </div>
          
          <div class="switch-item">
            <div class="switch-info">
                <span class="switch-label">稀疏磁盘 (Sparse VHD)</span>
                <span class="switch-annotation">启用后，新创建的虚拟磁盘文件将自动设置为稀疏。</span>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="form.sparseVhd">
              <span class="slider round"></span>
            </label>
          </div>

          <div class="form-group">
              <label>VM 空闲超时 (VM Idle Timeout)</label>
              <div class="input-suffix-wrapper">
                  <input v-model.number="form.vmIdleTimeout" type="number" class="input">
                  <span class="suffix">ms</span>
              </div>
              <span class="annotation">VM 在关闭之前处于空闲状态的毫秒数 (默认: 60000)。</span>
          </div>
        </div>
      </section>

      <div class="action-bar">
        <button class="btn btn-secondary" @click="handleReset">恢复默认</button>
        <button class="btn btn-primary" @click="handleSaveClick">保存配置</button>
      </div>
    </div>

    <Transition name="toast">
      <div v-if="showToast" class="toast-message">
        {{ toastMessage }}
      </div>
    </Transition>

    <!-- 异步修改提醒小窗 -->
    <Transition name="slide-up">
        <div v-if="showChangeModal" class="change-notification">
            <div class="notification-content">
                <span class="icon">📝</span>
                <div class="text">
                    <span class="title">配置已修改</span>
                    <span class="desc">检测到未保存的更改。</span>
                </div>
            </div>
            <div class="notification-actions">
                 <button class="btn-xs btn-primary" @click="handleSaveClick">保存</button>
                 <button class="btn-xs btn-secondary" @click="handleReloadConfig">重置为后端配置</button>
            </div>
        </div>
    </Transition>

    <!-- Reset Warning Modal -->
    <Transition name="fade">
      <div v-if="showResetWarning" class="modal-backdrop">
        <div class="modal-content">
          <div class="modal-header">
            <h3>⚠️ 恢复默认设置</h3>
          </div>
          <div class="modal-body">
            <p>确定要恢复默认设置吗？</p>
            <p>此操作将重新读取后端配置，所有未保存的更改将丢失。</p>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showResetWarning = false">取消</button>
            <button class="btn btn-primary" @click="executeReset">确认恢复</button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Restart Warning Modal -->
    <Transition name="fade">
      <div v-if="showRestartWarning" class="modal-backdrop">
        <div class="modal-content">
          <div class="modal-header">
            <h3>⚠️ 需要重启 WSL</h3>
          </div>
          <div class="modal-body">
            <p>保存配置后，所有正在运行的 WSL 发行版将被强制关闭以应用更改。</p>
            <p>请确保您已保存所有未保存的工作。</p>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showRestartWarning = false">取消</button>
            <button class="btn btn-primary" @click="executeSave">确认保存并重启</button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Loading Overlay -->
    <Transition name="fade">
      <div v-if="isSaving" class="loading-overlay">
        <div class="spinner"></div>
        <p>正在保存配置...</p>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.performance-view-container {
  padding: 32px;
  max-width: 1000px;
  margin: 0 auto;
  color: var(--color-text-primary);
}

.view-header {
  margin-bottom: 32px;
}

.view-header h2 {
  font-size: 24px;
  margin-bottom: 8px;
  font-weight: 600;
}

.subtitle {
  color: var(--color-text-secondary);
  font-size: 14px;
}

.config-card {
  background: var(--color-bg-card);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 32px;
  box-shadow: var(--shadow-sm);
}

.config-section {
  margin-bottom: 24px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text-primary);
  margin-bottom: 20px;
  padding-left: 12px;
  border-left: 4px solid var(--color-brand);
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 24px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text-primary);
}

.annotation {
  font-size: 12px;
  color: var(--color-text-secondary);
  line-height: 1.5;
}

.input-suffix-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.input-suffix-wrapper .input {
  padding-right: 40px;
}

.input-action-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  cursor: pointer;
}

.input-action-wrapper .input {
  cursor: pointer;
}

.suffix {
  position: absolute;
  right: 12px;
  color: var(--color-text-secondary);
  font-size: 13px;
  pointer-events: none;
}

.input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg-input, var(--color-bg-card));
  color: var(--color-text-primary);
  font-size: 14px;
  transition: all var(--transition-fast);
}

/* Fix for number input spin buttons in dark mode */
[data-theme='dark'] .input[type=number]::-webkit-inner-spin-button,
[data-theme='dark'] .input[type=number]::-webkit-outer-spin-button {
  filter: invert(1);
  opacity: 0.6;
}

.input:focus {
  outline: none;
  border-color: var(--color-brand);
  box-shadow: 0 0 0 2px var(--color-brand-alpha);
}

.input-error {
  border-color: var(--color-error);
}

.error-text {
  color: var(--color-error);
  font-size: 12px;
  margin-top: 4px;
}

.divider {
  height: 1px;
  background: var(--color-border);
  margin: 32px 0;
}

.switch-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.switch-item {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 16px;
  background: var(--color-bg-hover);
  border-radius: var(--radius-md);
}

.switch-item-inline {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background: var(--color-bg-hover);
  border-radius: var(--radius-md);
  height: 100%; /* Match height of other grid items if needed */
}

.switch-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
  padding-right: 16px;
}

.switch-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text-primary);
}

.switch-annotation {
    font-size: 12px;
    color: var(--color-text-secondary);
}

/* Material Switch */
.switch {
  position: relative;
  display: inline-block;
  width: 44px;
  height: 24px;
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
  background-color: var(--color-border);
  transition: .4s;
}

.slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: .4s;
  box-shadow: 0 2px 4px rgba(0,0,0,0.2);
}

input:checked + .slider {
  background-color: var(--color-brand);
}

input:focus + .slider {
  box-shadow: 0 0 1px var(--color-brand);
}

input:checked + .slider:before {
  transform: translateX(20px);
}

.slider.round {
  border-radius: 24px;
}

.slider.round:before {
  border-radius: 50%;
}

.action-bar {
    margin-top: 40px;
    display: flex;
    justify-content: flex-end;
    gap: 16px;
}

.btn {
    padding: 8px 24px;
    border-radius: 6px;
    font-size: 14px;
    cursor: pointer;
    border: none;
    transition: all 0.2s;
}

.btn-secondary {
    background: var(--color-bg-hover);
    color: var(--color-text-primary);
    border: 1px solid var(--color-border);
}
.btn-secondary:hover {
    border-color: var(--color-text-secondary);
}

.btn-primary {
    background: var(--color-brand);
    color: white;
}
.btn-primary:hover {
    opacity: 0.9;
    transform: translateY(-1px);
}

.toast-message {
  position: fixed;
  bottom: 40px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(0, 0, 0, 0.8);
  color: white;
  padding: 10px 24px;
  border-radius: 8px;
  font-size: 14px;
  z-index: 2000;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
}

.toast-enter-active,
.toast-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translate(-50%, 20px);
}
/* 异步修改提醒小窗 */
.change-notification {
  position: fixed;
  bottom: 24px;
  right: 24px;
  background: var(--color-bg-card);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 300px;
  z-index: 1000;
}

.notification-content {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.notification-content .icon {
  font-size: 20px;
}

.notification-content .text {
  display: flex;
  flex-direction: column;
}

.notification-content .title {
  font-weight: 600;
  color: var(--color-text-primary);
  font-size: 14px;
}

.notification-content .desc {
  font-size: 12px;
  color: var(--color-text-secondary);
  margin-top: 2px;
}

.notification-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.btn-xs {
  padding: 4px 12px;
  font-size: 12px;
  border-radius: 4px;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all var(--transition-fast);
}

.btn-xs.btn-primary {
  background: var(--color-brand);
  color: white;
}
.btn-xs.btn-primary:hover {
  background: var(--color-brand-hover);
}

.btn-xs.btn-secondary {
  background: transparent;
  border-color: var(--color-border);
  color: var(--color-text-primary);
}
.btn-xs.btn-secondary:hover {
  background: var(--color-bg-hover);
}

/* Slide Up Animation */
.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.slide-up-enter-from,
.slide-up-leave-to {
  opacity: 0;
  transform: translateY(150%) scale(0.95);
}

/* Modal Styles */
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 2500;
  backdrop-filter: blur(4px);
}

.modal-content {
  background: var(--color-bg-card);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  width: 400px;
  max-width: 90%;
  box-shadow: var(--shadow-lg);
  padding: 24px;
  animation: modal-pop 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.modal-header h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--color-text-primary);
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.modal-body {
  margin-bottom: 24px;
  color: var(--color-text-secondary);
  font-size: 14px;
  line-height: 1.6;
}

.modal-body p {
  margin-bottom: 8px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.modal-footer .btn-secondary {
  background: var(--color-bg-hover);
  color: var(--color-text-primary);
  border: 1px solid var(--color-border);
}

.modal-footer .btn-secondary:hover {
  border-color: var(--color-text-secondary);
}

.modal-footer .btn-primary {
  background: var(--color-brand);
  color: white;
}

.modal-footer .btn-primary:hover {
  background: var(--color-brand-hover);
}

@keyframes modal-pop {
  from { opacity: 0; transform: scale(0.95); }
  to { opacity: 1; transform: scale(1); }
}

/* Loading Overlay */
.loading-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  z-index: 3000;
  backdrop-filter: blur(4px);
  color: white;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  border-top-color: white;
  animation: spin 1s ease-in-out infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
