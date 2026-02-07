<script setup>
import { EventsOn, EventsOff } from 'wailsjs/runtime/runtime'
import { ref, onMounted, onUnmounted, reactive, computed } from 'vue'
import InfoCard from './LinuxCard.vue'
import { Install_Bottom, SelectDirectory } from 'wailsjs/go/main/App' 

const instances = ref([
  { id: 1, name: 'Ubuntu', desc: '常用开发环境', state: 'online', img_name:'UbuntuCoF', 
    versions: [
      { label: 'Ubuntu-26.04', value: 'Ubuntu-26.04' },
      { label: 'Ubuntu-25.10', value: 'Ubuntu-25.10' },
      { label: 'Ubuntu-25.04', value: 'Ubuntu-25.04' },
      { label: 'Ubuntu-24.04', value: 'Ubuntu-24.04' },
    ]},
  { id: 2, name: 'Debian', desc: '测试服务器', state: 'offline', img_name:'Debian', versions: [{ label: 'Latest', value: 'Debian' }] },
  { id: 3, name: 'Kali-Linux', desc: '网络安全工具库', state: 'online', img_name:'Kali-drago', versions: [{ label: 'Latest', value: 'Kali' }] },
  { id: 4, name: 'Arch', desc: '自定义配置', state: 'online', img_name:'Arch', versions: [{ label: 'Latest', value: 'Arch' }] },
  { id: 5, name: 'Fedora', desc: '实验性特性', state: 'offline', img_name:'Fedora', versions: [{ label: 'Latest', value: 'Fedora' }] },
  { id: 6, name: 'AlmaLinux', desc: '实验性特性', state: 'offline', img_name:'AlmaLinux', 
    versions: [
    { label: 'AlmaLinux-10', value: 'AlmaLinux-10' },
    { label: 'AlmaLinux-Kitten-10', value: 'AlmaLinux-Kitten-10' },
    { label: 'AlmaLinux-9', value: 'AlmaLinux-9' },
    { label: 'AlmaLinux-8', value: 'AlmaLinux-8' },
  ]},
    { id: 7, name: 'openSUSE', desc: '实验性特性', state: 'offline', img_name:'openSUSE', versions: [
      { label: 'openSUSE-Leap-16.0', value: 'openSUSE-Leap-16.0' },
      { label: 'openSUSE-Tumbleweed', value: 'openSUSE-Tumbleweed' }
    ] },
  { id: 8, name: 'SUSE', desc: '实验性特性', state: 'offline', img_name:'SUSE', versions: [
    { label: 'SUSE-Linux-Enterprise-16.0', value: 'SUSE-Linux-Enterprise-16.0' },
    { label: 'SUSE-Linux-Enterprise-15-SP7', value: 'SUSE-Linux-Enterprise-15-SP7' },
  ] },

])

const showModal = ref(false)
// 当前操作的步骤 'config' (配置) | 'install' (安装进度)
const currentStepView = ref('config') 
const currentInstance = ref(null)
// 错误状态控制变量
const isError = ref(false)
const errorDetail = ref('')

// --- 新增：自定义安装步骤配置 ---
// keyword: 匹配后端日志的关键词，匹配到时会跳转到该步骤
const installSteps = ref([
    { title: '下载资源', status: 'pending', keyword: ['download','下载完成'] }, // 示例关键词
    { title: '解压安装', status: 'pending', keyword: ['installing','解压'] }, 
    { title: '配置用户', status: 'pending', keyword: ['password', '配置用户'] },
    { title: '完成', status: 'pending', keyword: ['success','成功'] }
])

const currentStepIndex = ref(0)
const progressPercent = ref(0)
const currentLogText = ref('等待开始...') // 显示当前正在进行的具体操作文本

// 表单数据
const installForm = reactive({
    username: '',
    password: '',
    version: '',
    installPath: '',
    threadCount: 4 // 默认推荐 4 线程
})

const handleSelectPath = async () => {
    try {
        const path = await SelectDirectory()
        if (path) installForm.installPath = path
    } catch (e) {
        console.error("选择路径失败", e)
    }
}

const errors = reactive({ username: '', password: '' })

const validateForm = () => {
    let isValid = true
    errors.username = ''
    errors.password = ''

    const userRegex = /^[a-z_][a-z0-9_-]*$/
    
    if (!installForm.username) {
        errors.username = '用户名不能为空'
        isValid = false
    } else if (installForm.username.length > 32) {
        errors.username = '用户名不能超过 32 个字符'
        isValid = false
    } else if (!userRegex.test(installForm.username)) {
        errors.username = '仅支持小写字母、数字、下划线(_)、短横线(-)，且需以字母开头'
        isValid = false
    }

    if (!installForm.password) {
        errors.password = '密码不能为空'
        isValid = false
    } else if (installForm.password.includes(' ')) {
        errors.password = '密码不能包含空格' 
        isValid = false
    }

    return isValid
}

let isMouseDownOnOverlay = false

const handleOverlayMouseDown = (e) => {
    if (e.target.classList.contains('modal-overlay') && e.button === 0) {
        isMouseDownOnOverlay = true
    }
}

const handleOverlayMouseUp = (e) => {
    if (isMouseDownOnOverlay && e.target.classList.contains('modal-overlay')) {
        // 安装中禁止点击背景关闭
        if (currentStepView.value !== 'install') {
            showModal.value = false
        }
    }
    isMouseDownOnOverlay = false 
}

const handleAction = (item) => {
  currentInstance.value = item
  installForm.username = ''
  installForm.password = ''
  installForm.installPath = '' 
  if (item.versions && item.versions.length > 0) {
     installForm.version = item.versions[0].value
  } else {
     installForm.version = 'latest'
  }
  errors.username = ''
  errors.password = ''
  
  // 重置步骤状态
  currentStepView.value = 'config'
  resetProgress()
  showModal.value = true
}

const resetProgress = () => {
    isError.value = false
    errorDetail.value = ''
    currentStepIndex.value = 0
    progressPercent.value = 5
    currentLogText.value = '准备就绪'
    installSteps.value.forEach(s => s.status = 'pending')
    installSteps.value[0].status = 'processing'
}

const startInstall = async () => {
  if (!validateForm()) return
  if (!currentInstance.value) return

  // 重置错误状态和步骤状态，确保重新开始时状态干净
  isError.value = false
  errorDetail.value = ''
  installSteps.value.forEach(s => s.status = 'pending')
  installSteps.value[0].status = 'processing'
  currentStepIndex.value = 0
  progressPercent.value = 2 // Start with a small visual progress
  currentLogText.value = '准备就绪...'

  currentStepView.value = 'install'
  
  try {
    await Install_Bottom(currentInstance.value.name, installForm.username, installForm.password, installForm.version, installForm.installPath)
  } catch (e) {
    // If immediate call fails
    currentLogText.value = "启动安装失败: " + e
    isError.value = true
    errorDetail.value = e.toString()
    installSteps.value[currentStepIndex.value].status = 'error'
  }
}



// 核心逻辑：处理进度条和步骤跳转
const processLogAndProgress = (line) => {
    if (!line) return;
    // 关键词匹配跳转步骤 (根据你实际后端的日志内容调整 keyword)
    const lowerLine = line.toLowerCase()
    // 进度条增长黑名单
    currentLogText.value = line;


    const skipIncrementKeywords = ['正在下载', '%', 'progress'];
    const shouldSkip = skipIncrementKeywords.some(key => lowerLine.includes(key));

    currentLogText.value = line
    
    // 尝试解析进度百分比 (假设日志格式如 "Progress: 25.5%" 或 "25%")
    // 浮点数支持：(\d+(\.\d+)?)
    const percentMatch = lowerLine.match(/(\d+(\.\d+)?)%/)
    if (percentMatch) {
        const p = parseFloat(percentMatch[1])
        if (!isNaN(p)) {
             // 只有当解析出的进度大于当前进度时才更新
             // 限制最大增长，确保不超过当前步骤的最大范围
             const stepCount = installSteps.value.length
             const stepWidth = 100 / stepCount
             const currentStepMax = (currentStepIndex.value + 1) * stepWidth
             
             // 将后端 0-100% 映射到当前步骤的范围 (例如步骤1是 0-25%)
             // 公式: 当前步骤起始 + (后端进度% * 步骤宽度)
             const mappedPercent = (currentStepIndex.value * stepWidth) + (p / 100 * stepWidth)

             if (mappedPercent > progressPercent.value && mappedPercent <= currentStepMax) {
                 progressPercent.value = mappedPercent
             }
        }
    } else {
        // 如果没有明确百分比，尝试通过关键词推断
        // 由于有高频浮点数更新，这里主要处理整点或特殊标记
        if (lowerLine.includes('25%')) progressPercent.value = Math.max(progressPercent.value, 25)
        else if (lowerLine.includes('50%')) progressPercent.value = Math.max(progressPercent.value, 50)
        else if (lowerLine.includes('75%')) progressPercent.value = Math.max(progressPercent.value, 75)
        else if (lowerLine.includes('100%')) progressPercent.value = 100
    }
    
    // 重试逻辑：如果检测到 Retry 关键词，回撤进度条
    const retryKeywords = ['retry', 'retrying', '重试', 'connection reset', 'time out', 'timeout']
    if (retryKeywords.some(k => lowerLine.includes(k))) {
        // 简单策略：回退到当前步骤的起始进度，或者减去一定数值
        // 假设当前步骤索引对应的基础进度
        const currentStepBase = (currentStepIndex.value / installSteps.value.length) * 100
        // 回退到该步骤的起点，给用户“重新开始这段”的感觉
        progressPercent.value = Math.max(currentStepBase, 5) 
        currentLogText.value = "检测到网络波动，正在重试..."
        return // 跳过后续增长逻辑
    }

    // 1. 只有不在黑名单中时，才进行模拟增长
    // 修改策略：如果已经有高频百分比更新（percentMatch），则不进行模拟增长，避免冲突
    // 只有在没有解析出百分比时，才启用模拟增长
    // 并且限制模拟增长不能超过当前步骤的 90% (预留给真实完成信号)
    if (!percentMatch && !shouldSkip) {
        const stepCount = installSteps.value.length
        const stepWidth = 100 / stepCount
        const currentStepMax = (currentStepIndex.value + 1) * stepWidth
        // 限制模拟增长的上限为当前步骤结束前的 5% 缓冲
        const simulationLimit = currentStepMax - (stepWidth * 0.1)
        
        if (progressPercent.value < simulationLimit) {
            progressPercent.value += 0.5 // 减缓模拟增长速度
        }
    }
    
    // 遍历所有步骤，看是否命中关键词
    installSteps.value.forEach((step, index) => {
        // 如果 keyword 是数组，用 some；如果是字符串，也要能兼容
        const keywords = Array.isArray(step.keyword) ? step.keyword : [step.keyword];
        // 只要日志包含数组中任意一个词，即视为匹配成功
        const isMatch = keywords.some(key => key && lowerLine.includes(key.toLowerCase()));

        if (isMatch) {
            // 如果命中更后的步骤，更新状态
            if (index > currentStepIndex.value) {
                for(let i = 0; i < index; i++) {
                    installSteps.value[i].status = 'finished'
                }
                currentStepIndex.value = index
                installSteps.value[index].status = 'processing'
                
                const basePercent = (index / installSteps.value.length) * 100
                if (progressPercent.value < basePercent) {
                    progressPercent.value = basePercent
                }
            }
        }
    })

    // 特殊处理：如果检测到“完成”或“Success”
    if (lowerLine.includes('success') || lowerLine.includes('completed')) {
        progressPercent.value = 100
        installSteps.value.forEach(s => s.status = 'finished')
        currentLogText.value = "安装已完成！"
    }
}

onMounted(() => {
  // 监听正常日志（保持不变）
  EventsOn("wsl-output", (line) => {
    if(!showModal.value || isError.value) return 
    processLogAndProgress(line)
  })

  // === 新增：监听错误事件 ===
  EventsOn("wsl-error", (errMsg) => {
    // 1. 标记为错误状态
    isError.value = true
    errorDetail.value = errMsg
    
    // 2. 将当前正在进行的步骤标红
    if (installSteps.value[currentStepIndex.value]) {
        installSteps.value[currentStepIndex.value].status = 'error'
    }
    
    // 3. 停止进度条增长（可选，视觉上停止）
    currentLogText.value = "任务异常终止"
  })
})

onUnmounted(() => {
  try {
    if (typeof EventsOff === 'function') {
      EventsOff("wsl-output")
    }
  } catch (e) {
    console.error("清理事件失败", e)
  }
})

const getIconUrl = (name) => {
    // 安全检查：防止 name 为空导致报错
    if (!name) return '' 
    try {
        // 这里的路径 ../assets/icons/ 必须和你实际存放图片的文件夹层级一致
        return new URL(`../assets/icons/${name}.png`, import.meta.url).href
    } catch (e) {
        console.error("图片加载失败:", name, e)
        return '' // 如果出错返回空，防止白屏
    }
}
</script>

<template>
  <div class="install-view-container">
    <div class="card-grid">
      <TransitionGroup name="list">
        <InfoCard 
          v-for="item in instances" 
          :key="item.id"
          :title="item.name"
          :description="item.desc"
          :status="item.state"
          :iconName="item.img_name"
          @action="handleAction(item)"
        />
      </TransitionGroup>
    </div>

    <Transition name="modal">
      <div v-if="showModal" class="modal-overlay" @mousedown="handleOverlayMouseDown" @mouseup="handleOverlayMouseUp">
          <div class="modal-window" @mousedown.stop>
          
          <div class="modal-header">
              <span>
              {{ currentStepView === 'config' ? '安装配置向导' : '系统部署中' }} 
              - {{ currentInstance && currentInstance.name }}
              </span>
              <button v-if="currentStepView === 'config'" class="btn-ghost close-btn" @click="showModal = false">✕</button>
          </div>
          
          <Transition name="fade-slide" mode="out-in">
              <div v-if="currentStepView === 'config'" class="modal-content-body config-body">
                  <div class="form-group">
                      <label>用户名 (UNIX Username)</label>
                      <input 
                      v-model="installForm.username" 
                      type="text" 
                      class="input"
                      placeholder="请输入用户名"
                      :class="{ 'input-error': errors.username }"
                      @input="errors.username = ''"
                      >
                      <span v-if="errors.username" class="error-msg">{{ errors.username }}</span>
                  </div>
                  <div class="form-group">
                      <label>密码 (Root Password)</label>
                      <input 
                      v-model="installForm.password" 
                      type="password" 
                      class="input"
                      placeholder="请输入密码"
                      :class="{ 'input-error': errors.password }"
                      @input="errors.password = ''"
                      >
                      <span v-if="errors.password" class="error-msg">{{ errors.password }}</span>
                  </div>

                  <div class="form-group">
                  <label>安装路径 (可选)</label>
                  <div class="path-input-group">
                      <input 
                          type="text" 
                          class="input"
                          :value="installForm.installPath" 
                          placeholder="默认路径 (C:\Users\<用户名>\AppData\Local\Packages\)" 
                          readonly
                      >
                      <button class="btn btn-secondary browse-btn" @click="handleSelectPath">浏览...</button>
                  </div>
                  </div>

                  <div class="form-group">
                      <label>安装版本</label>
                      <select v-if="currentInstance.versions && currentInstance.versions.length > 0" 
                              v-model="installForm.version" class="input">
                          <option v-for="ver in currentInstance.versions" :key="ver.value" :value="ver.value">
                              {{ ver.label }}
                          </option>
                      </select>
                      
                      <input v-else type="text" value="Default (Latest)" disabled class="input disabled-input">
                  </div>

                  <!-- 线程数选择 (优化 2) -->
                  <div class="form-group">
                      <label>下载线程数 <span class="recommend-badge">推荐: 4</span></label>
                      <div class="thread-selector">
                          <button 
                              v-for="n in 8" 
                              :key="n"
                              class="thread-btn"
                              :class="{ 'active': installForm.threadCount === n, 'recommended': n === 4 }"
                              @click="installForm.threadCount = n"
                              :title="n === 4 ? '推荐配置' : n + ' 线程'"
                          >
                              {{ n }}
                          </button>
                      </div>
                  </div>

                  <div class="action-bar">
                      <button class="btn btn-secondary" @click="showModal = false">取消</button>
                      <button class="btn btn-primary" @click="startInstall">开始安装</button>
                  </div>
              </div>

          <div v-else class="modal-content-body progress-body">
              
              <div v-if="!isError" class="progress-content">
                  <div class="install-hero">
                      <img :src="getIconUrl(currentInstance?.img_name)" class="hero-icon animate-pulse" />
                      <div class="hero-info">
                          <h3>正在安装 {{ currentInstance?.name }}</h3>
                          <p class="log-detail">{{ currentLogText }}</p>
                      </div>
                  </div>

                  <div class="progress-bar-container">
                      <div class="progress-track">
                          <div class="progress-fill" :style="{ width: progressPercent + '%' }">
                              <div class="progress-glow"></div>
                          </div>
                      </div>
                      <span class="progress-text">{{ Math.floor(progressPercent) }}%</span>
                  </div>

                  <div class="steps-container">
                      <div v-for="(step, index) in installSteps" :key="index" 
                          class="step-item" 
                          :class="step.status">
                          <div class="step-icon">
                              <span v-if="step.status === 'finished'">✓</span>
                              <span v-else-if="step.status === 'processing'" class="spinner"></span>
                              <span v-else-if="step.status === 'error'">!</span> <span v-else>{{ index + 1 }}</span>
                          </div>
                          <span class="step-title">{{ step.title }}</span>
                          <div v-if="index < installSteps.length - 1" class="step-line" :class="{ 'line-active': step.status === 'finished' }"></div>
                      </div>
                  </div>
                  
                  <div class="action-bar" v-if="progressPercent >= 100">
                      <button class="btn btn-primary" @click="showModal = false">完成</button>
                  </div>
              </div>

              <div v-else class="error-container">
                  <div class="error-icon-area">
                      <span class="error-symbol">⚠️</span>
                  </div>
                  <h3>安装失败</h3>
                  <p class="error-desc">在执行步骤 <b>{{ installSteps[currentStepIndex]?.title }}</b> 时发生错误。</p>
                  
                  <div class="error-box">
                      <code>{{ errorDetail }}</code>
                  </div>

                  <div class="action-bar">
                      <button class="btn btn-danger" @click="() => { currentStepView = 'config'; isError = false; }">返回设置</button>
                      <button class="btn btn-secondary" @click="showModal = false">关闭</button>
                  </div>
              </div>

          </div>
          </Transition>

          </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.install-view-container {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: var(--spacing-lg);
  padding: var(--spacing-xs);
}

/* Modal Styles */
.modal-overlay {
  position: fixed; top: 0; left: 0;
  width: 100vw; height: 100vh;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  display: flex; justify-content: center; align-items: center;
  z-index: 1000;
  transition: opacity 0.3s ease;
}

.modal-window {
  width: 520px;
  background: var(--color-bg-modal);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  display: flex; flex-direction: column;
  border: 1px solid var(--color-border);
  overflow: hidden;
  animation: modal-pop 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@keyframes modal-pop {
  from { opacity: 0; transform: scale(0.95) translateY(10px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}

.modal-header {
  padding: var(--spacing-md) var(--spacing-lg);
  background: var(--color-bg-hover);
  border-bottom: 1px solid var(--color-border);
  display: flex; justify-content: space-between; align-items: center;
  font-weight: 600;
  color: var(--color-text-primary);
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

.modal-content-body {
  padding: var(--spacing-xl);
  color: var(--color-text-primary);
}

/* Form Styles */
.form-group {
  margin-bottom: var(--spacing-md);
}

.form-group label {
  display: block;
  margin-bottom: var(--spacing-xs);
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
}

.path-input-group {
  display: flex;
  gap: var(--spacing-sm);
}

.input-error {
  border-color: var(--color-error) !important;
  background: rgba(255, 77, 79, 0.05);
}

.error-msg {
  display: block;
  margin-top: 4px;
  color: var(--color-error);
  font-size: var(--font-size-xs);
  animation: shake 0.3s;
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-5px); }
  75% { transform: translateX(5px); }
}

.action-bar {
  margin-top: var(--spacing-xl);
  display: flex;
  justify-content: flex-end;
  gap: var(--spacing-md);
}

/* Install Progress Styles */
.install-hero {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
  margin-bottom: var(--spacing-lg);
}

.hero-icon {
  width: 64px;
  height: 64px;
  object-fit: contain;
}

.hero-info h3 {
  margin: 0 0 4px 0;
  color: var(--color-text-primary);
  font-size: var(--font-size-lg);
}

.log-detail {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
  font-family: var(--font-family-mono);
  max-width: 360px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.progress-bar-container {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-xl);
}

.progress-track {
  flex: 1;
  height: 8px;
  background: var(--color-bg-hover);
  border-radius: var(--radius-full);
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--color-brand), #36cfc9);
  border-radius: var(--radius-full);
  transition: width 0.4s ease;
  position: relative;
}

.progress-glow {
  position: absolute; top: 0; left: 0; width: 100%; height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,0.4), transparent);
  animation: scan 1.5s infinite;
}

@keyframes scan { from { transform: translateX(-100%); } to { transform: translateX(100%); } }

.progress-text {
  color: var(--color-text-primary);
  font-weight: 600;
  font-size: var(--font-size-sm);
  width: 40px;
  text-align: right;
}

/* Steps */
.steps-container {
  display: flex;
  justify-content: space-between;
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
  background: var(--color-bg-card);
  border: 2px solid var(--color-border);
  color: var(--color-text-secondary);
  display: flex; align-items: center; justify-content: center;
  font-size: var(--font-size-xs);
  font-weight: bold;
  margin-bottom: 8px;
  transition: all var(--transition-normal);
}

.step-title {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
  transition: color var(--transition-normal);
}

.step-item.processing .step-icon {
  border-color: var(--color-brand);
  color: var(--color-brand);
  background: var(--color-bg-card);
  box-shadow: 0 0 0 3px rgba(24, 144, 255, 0.15);
}
.step-item.processing .step-title { color: var(--color-text-primary); }

.step-item.finished .step-icon {
  background: var(--color-brand);
  border-color: var(--color-brand);
  color: #fff;
}

.step-item.error .step-icon {
  border-color: var(--color-error);
  color: var(--color-error);
}

.step-line {
  position: absolute; top: 14px; left: 50%; width: 100%; height: 2px;
  background: var(--color-border); z-index: -1;
}
.step-item:last-child .step-line { display: none; }
.step-line.line-active { background: var(--color-brand); }

/* Spinner */
.spinner {
  width: 14px; height: 14px;
  border: 2px solid transparent;
  border-top-color: currentColor;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

/* Error State */
.error-container {
  text-align: center;
  padding: 10px 0;
  animation: fadeIn 0.3s ease;
}

.error-symbol {
  font-size: 48px;
  margin-bottom: var(--spacing-md);
  display: block;
}

.error-container h3 {
  color: var(--color-error);
  margin: 0 0 var(--spacing-sm) 0;
}

.error-desc {
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  margin-bottom: var(--spacing-lg);
}

.error-box {
  background: rgba(255, 77, 79, 0.05);
  border: 1px solid rgba(255, 77, 79, 0.2);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  text-align: left;
  margin-bottom: var(--spacing-lg);
  max-height: 200px;
  overflow-y: auto;
}

.error-box code {
  color: var(--color-error);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
  word-break: break-all;
}

/* Transitions */
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all 0.3s ease;
}
.fade-slide-enter-from {
  opacity: 0;
  transform: translateX(20px);
}
.fade-slide-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}

/* --- Input Redesign (优化 2) --- */
.input {
  width: 100%;
  padding: 10px 12px; /* 增加内边距 */
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md); /* 圆角 */
  background-color: var(--color-bg-input, var(--color-bg-card)); 
  color: var(--color-text-primary);
  font-size: 14px;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05); /* 增加层次感 */
}

.input:hover {
  border-color: var(--color-text-secondary);
}

.input:focus {
  outline: none;
  border-color: var(--color-brand);
  box-shadow: 0 0 0 3px rgba(24, 144, 255, 0.15); /* 聚焦高亮 */
}

.input:disabled, .disabled-input {
  background-color: var(--color-bg-hover);
  cursor: not-allowed;
  opacity: 0.7;
  box-shadow: none;
}

/* --- Thread Selector Styles --- */
.recommend-badge {
    font-size: 11px;
    background: rgba(24, 144, 255, 0.1);
    color: var(--color-brand);
    padding: 2px 6px;
    border-radius: 4px;
    margin-left: 8px;
    font-weight: 600;
}

.thread-selector {
    display: flex;
    gap: 4px;
    background: var(--color-bg-hover); /* 背景底色 */
    padding: 4px;
    border-radius: 8px;
    border: 1px solid var(--color-border);
}

.thread-btn {
    flex: 1;
    height: 32px;
    border: 1px solid transparent;
    background: transparent;
    border-radius: 6px;
    cursor: pointer;
    font-size: 13px;
    color: var(--color-text-secondary);
    transition: all 0.2s ease;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 500;
    position: relative;
}

.thread-btn:hover:not(.active) {
    background: rgba(125, 125, 125, 0.1);
    color: var(--color-text-primary);
}

.thread-btn.active {
    background: var(--color-bg-card); /* 激活时凸起 */
    color: var(--color-brand);
    box-shadow: 0 2px 5px rgba(0,0,0,0.08);
    font-weight: 600;
    border-color: rgba(0,0,0,0.02);
}

/* 推荐的小圆点标记 */
.thread-btn.recommended:not(.active)::after {
    content: "";
    position: absolute;
    bottom: 4px;
    width: 4px;
    height: 4px;
    background: var(--color-text-secondary);
    border-radius: 50%;
    opacity: 0.4;
}
</style>
