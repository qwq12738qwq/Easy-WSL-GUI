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
    { title: '迁移系统', status: 'pending', keyword: ['moving','迁移'] },
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
    installPath: '' 
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

  currentStepView.value = 'install'
  
  try {
    await Install_Bottom(currentInstance.value.name, installForm.username, installForm.password, installForm.version, installForm.installPath)
  } catch (e) {
    currentLogText.value = "调用失败: " + e
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
    
    // 1. 只有不在黑名单中时，才进行模拟增长
    if (!shouldSkip && progressPercent.value < 95) {
        progressPercent.value += 1.5 
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
  <div class="card-grid">
    <InfoCard 
      v-for="item in instances" 
      :key="item.id"
      :title="item.name"
      :description="item.desc"
      :status="item.state"
      :iconName="item.img_name"
      @action="handleAction(item)"
    />
  </div>

  <Transition name="modal">
    <div v-if="showModal" class="modal-overlay" @mousedown="handleOverlayMouseDown" @mouseup="handleOverlayMouseUp">
        <div class="modal-window" @mousedown.stop>
        
        <div class="modal-header">
            <span>
            {{ currentStepView === 'config' ? '安装配置向导' : '系统部署中' }} 
            - {{ currentInstance && currentInstance.name }}
            </span>
            <button v-if="currentStepView === 'config'" class="close-btn" @click="showModal = false">✕</button>
        </div>
        <Transition name="fade" mode="out-in">
            <div v-if="currentStepView === 'config'" class="config-body">
                <div class="form-group">
                    <label>用户名 (UNIX Username)</label>
                    <input 
                    v-model="installForm.username" 
                    type="text" 
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
                        :value="installForm.installPath" 
                        placeholder="默认路径 (C:\Users\<用户名>\AppData\Local\Packages\)" 
                        readonly
                    >
                    <button class="browse-btn" @click="handleSelectPath">浏览...</button>
                </div>
                </div>

                <div class="form-group">
                    <label>安装版本</label>
                    <select v-if="currentInstance.versions && currentInstance.versions.length > 0" 
                            v-model="installForm.version">
                        <option v-for="ver in currentInstance.versions" :key="ver.value" :value="ver.value">
                            {{ ver.label }}
                        </option>
                    </select>
                    
                    <input v-else type="text" value="Default (Latest)" disabled class="disabled-input">
                </div>

                <div class="action-bar">
                    <button class="cancel-btn" @click="showModal = false">取消</button>
                    <button class="confirm-btn" @click="startInstall">开始安装</button>
                </div>
            </div>

        <div v-else class="progress-body">
            
            <div v-if="!isError" class="progress-content">
                <div class="install-hero">
                    <img :src="getIconUrl(currentInstance?.img_name)" class="hero-icon" />
                    <div class="hero-info">
                        <h3>正在安装 {{ currentInstance?.name }}</h3>
                        <p class="log-detail">{{ currentLogText }}</p>
                    </div>
                </div>

                <div class="progress-bar-container">
                    <div class="progress-track">
                        <div class="progress-fill" :style="{ width: progressPercent + '%' }"></div>
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
                    <button class="confirm-btn" @click="showModal = false">完成</button>
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
                    <button class="retry-btn" @click="currentStepView = 'config'">返回设置</button>
                    <button class="cancel-btn" @click="showModal = false">关闭</button>
                </div>
            </div>

        </div>
        </Transition>

        </div>
    </div>
  </Transition>
</template>

<style scoped>

/* 亮色模式提升对比度 */
:root[data-theme='light'] {
    --card-bg: #fdfdfd;
    --card-border: #d0d0d0;
    --text-primary: #111111;   /* 接近纯黑 */
    --text-secondary: #555555; /* 深灰 */
    --btn-bg: #eeeeee;
}

/* 深色模式提升对比度 */
:root[data-theme='dark'] {
    --card-bg: #252526;
    --card-border: #444444;
    --text-primary: #ffffff;   /* 纯白 */
    --text-secondary: #bbbbbb; /* 浅灰 */
    --btn-bg: #333333;
}

  /* 路径选择组合框样式 */
.path-input-group {
    display: flex;
    gap: 10px;
}

.path-input-group input {
    flex: 1; /* 输入框占据剩余空间 */
    /* 继承原有的 input 样式，但 cursor 变一下提示只读 */
    cursor: default;
}

.browse-btn {
    padding: 0 15px;
    background: #3a3a3a;
    border: 1px solid #555;
    color: #eee;
    border-radius: 6px;
    cursor: pointer;
    white-space: nowrap; /* 防止文字换行 */
    transition: all 0.2s;
}

.browse-btn:hover {
    background: #4a4a4a;
    border-color: #777;
}

.disabled-input {
    background-color: #3a3a3a !important;
    color: #888 !important;
    cursor: not-allowed;
    border-color: #444 !important;
}

.home-container { padding: 20px; height: 100%; box-sizing: border-box; }
/* 1. 优化网格布局：增加间距，限制最大宽度 */
/* 1. 卡片网格容器优化：增加呼吸感 */
.card-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 20px;
    padding: 10px;
    max-width: 1200px;
    margin: 0 auto;
}

/* 2. 彻底移除所有 3D 和位移效果，改为“质感反馈” */
:deep(.card) { 
    background: var(--card-bg); /* 极淡的透明度 */
    border: 1px solid rgba(255, 255, 255, 0.08); /* 柔和的边框 */
    border-radius: 12px;
    transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
    position: relative;
    overflow: hidden;
}

/* 3. 悬浮效果：改为“光影渐变”而非位移 */
:deep(.card:hover) {
    transform: none !important; /* 强制禁用任何位移 */
    background: rgba(255, 255, 255, 0.06); /* 稍微加亮背景 */
    border-color: var(--accent-color); /* 使用主题色边框 */
    /* 细腻的蓝色发光阴影 */
    box-shadow: 0 0 20px rgba(24, 144, 255, 0.15); 
}

/* 4. 这里的 :deep 选择器确保能穿透到 LinuxCard 组件 */
:deep(.card-btn) {
    background: var(--card-btn-bg) !important;
    color: var(--card-btn-text) !important;
    font-weight: 600;
    border: 1px solid var(--card-border);
    border-radius: 6px;
    transition: all 0.2s;
}

:deep(.card:hover .card-btn) {
    background: #1890ff !important;
    color: #ffffff !important;
    border-color: transparent;
}

/* 标题：加粗并设为高对比度颜色 */
:deep(.card-body h4) {
    color: var(--text-primary) !important;
    font-weight: 700;
    font-size: 16px;
    margin-bottom: 6px;
}

/* 描述文字：提高灰度亮度 */
:deep(.card-body p) {
    color: var(--text-secondary) !important;
    font-size: 13px;
    line-height: 1.5;
}

/* 弹窗样式调整 */
.modal-overlay {
  position: fixed; top: 0; left: 0;
  width: 100vw; height: 100vh;
  background: rgba(0, 0, 0, 0.7); /* 加深一点背景 */
  display: flex; justify-content: center; align-items: center;
  z-index: 999;
  backdrop-filter: blur(2px); /* 增加一点模糊感 */
  cursor: default; /* 避免鼠标变为文本选择状 */
  user-select: none; /* 防止遮罩层本身被选中 */
}

.modal-window {
  width: 520px;
  background: rgba(30, 30, 30, 0.95);
  backdrop-filter: blur(10px); /* 磨砂玻璃效果 */
  overflow: hidden;
  box-shadow: 0 15px 40px rgba(0,0,0,0.6);
  display: flex; flex-direction: column;
  border: 1px solid rgba(255, 255, 255, 0.1);
  user-select: text; 
  cursor: auto;
}

.modal-header {
  padding: 15px 20px;
  background: #252526;
  color: #fff;
  display: flex; justify-content: space-between; align-items: center;
  font-size: 15px;
  font-weight: 600;
  border-bottom: 1px solid #333;
}

/* === 新增：表单样式 === */
.config-body {
    padding: 25px;
    color: #ccc;
}

.form-group {
    margin-bottom: 15px;
}

.form-group label {
    display: block;
    margin-bottom: 8px;
    font-size: 13px;
    color: #aaa;
}

.form-group input, 
.form-group select {
    width: 100%;
    padding: 12px;
    background: #1a1a1a;
    border: 1px solid #333;
    border-radius: 6px;
    color: #fff;
    font-size: 14px;
    outline: none;
    box-sizing: border-box;
    transition: all 0.2s ease; /* 增加过渡动画 */
}
/* 校验失败时的样式 */
.form-group input.input-error {
    border-color: #ff4d4f; /* 红色边框 */
    background: #2a1215;   /* 极淡的红色背景 */
}

/* 错误文字提示 */
.error-msg {
    display: block;
    margin-top: 5px;
    color: #ff4d4f;
    font-size: 12px;
    animation: shake 0.3s; /* 增加一个小抖动动画提示用户 */
}

/* 简单的抖动动画 */
@keyframes shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-5px); }
  75% { transform: translateX(5px); }
}

/* 焦点样式优化 */
.form-group input:focus {
    background: #222;
    box-shadow: 0 0 0 2px rgba(24, 144, 255, 0.2);
}
/* 如果聚焦时依然是错误状态，保持红色，或者你可以选择移除这行让它变蓝 */
.form-group input.input-error:focus {
    border-color: #ff7875; 
}

.form-group input:focus,
.form-group select:focus {
    border-color: #1890ff;
    background: #333;
}

.action-bar {
    margin-top: 30px;
    display: flex;
    justify-content: flex-end;
    gap: 12px;
}

.cancel-btn {
    padding: 8px 20px;
    background: transparent;
    border: 1px solid #555;
    color: #ccc;
    border-radius: 4px;
    cursor: pointer;
}
.cancel-btn:hover { background: #333; }

.confirm-btn {
    padding: 8px 20px;
    background: #1890ff;
    border: none;
    color: white;
    border-radius: 4px;
    cursor: pointer;
    font-weight: 500;
}
.confirm-btn:hover { background: #40a9ff; }

/* === 终端样式优化 === */
.terminal-body {
  height: 320px; /* 稍微调低一点 */
  padding: 15px;
  overflow-y: auto;
  background: #000;
  color: #4af626;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  line-height: 1.6;
}

.log-line { margin-bottom: 4px; word-break: break-all; }
.prefix { color: #888; margin-right: 8px; }

.cursor-blink {
    animation: blink 1s infinite;
    font-weight: bold;
}
@keyframes blink { 0%, 100% { opacity: 1; } 50% { opacity: 0; } }

.close-btn {
  background: transparent; border: none; color: #888;
  font-size: 16px; cursor: pointer; padding: 0 5px;
}
.close-btn:hover { color: #fff; }

/* 进度视图样式 */
.progress-body {
    padding: 30px;
    background: #1e1e1e;
    display: flex;
    flex-direction: column;
    gap: 25px;
}

.install-hero {
    display: flex;
    align-items: center;
    gap: 20px;
    margin-bottom: 10px;
}

.hero-icon {
    width: 64px;
    height: 64px;
    object-fit: contain;
}

.hero-info h3 {
    margin: 0 0 5px 0;
    color: #fff;
    font-size: 18px;
}

.log-detail {
    margin: 0;
    color: #888;
    font-size: 12px;
    font-family: 'Consolas', monospace;
    max-width: 380px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

/* 进度条 */
.progress-bar-container {
    display: flex;
    align-items: center;
    gap: 15px;
}

.progress-track {
    flex: 1;
    height: 6px;
    background: #000;
    border-radius: 4px;
    overflow: hidden;
}

.progress-fill {
    height: 100%;
    background: linear-gradient(90deg, #1890ff, #36cfc9);
    border-radius: 4px;
    transition: width 0.4s ease;
    box-shadow: 0 0 10px rgba(24, 144, 255, 0.3);
}

.progress-text {
    color: #fff;
    font-weight: bold;
    font-size: 14px;
    width: 40px;
    text-align: right;
}

/* 步骤指示器 */
.steps-container {
    display: flex;
    justify-content: space-between;
    margin-top: 10px;
    position: relative;
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

.step-title {
    font-size: 12px;
    color: #666;
    transition: color 0.3s;
}

/* 步骤状态变化 */
.step-item.processing .step-icon {
    border-color: #1890ff;
    color: #1890ff;
    background: #1e1e1e;
}

.step-item.processing .step-title {
    color: #fff;
}

.step-item.finished .step-icon {
    background: #1890ff;
    border-color: #1890ff;
    color: #fff;
}

.step-item.finished .step-title {
    color: #ccc;
}

.step-item.error .step-icon {
    border-color: #ff4d4f;
    color: #ff4d4f;
}

/* 连接线 */
.step-line {
    position: absolute;
    top: 14px; /* 这里的 top 应该是 step-icon 高度的一半 */
    left: 50%;
    width: 100%; /* 连接到下一个 */
    height: 2px;
    background: #333;
    z-index: -1;
}

/* 最后一个元素不需要向右的线 */
.step-item:last-child .step-line {
    display: none;
}

.step-line.line-active {
    background: #1890ff;
}

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

/* 错误容器样式 */
.error-container {
    animation: fadeIn 0.3s ease;
    text-align: center;
    padding: 10px 0;
}

.error-header {
    margin-bottom: 20px;
}

.error-icon {
    font-size: 40px;
    display: block;
    margin-bottom: 10px;
}

.error-header h3 {
    color: #ff4d4f;
    margin: 0;
}

.error-box {
    background: #2a1215;
    border: 1px solid #5c2526;
    border-radius: 6px;
    padding: 15px;
    text-align: left;
    margin-bottom: 20px;
    max-height: 200px;
    overflow-y: auto;
}

.error-box code {
    color: #ffccc7;
    font-family: 'Consolas', monospace;
    font-size: 13px;
    word-break: break-all;
}

.retry-btn {
    padding: 8px 20px;
    background: #ff4d4f;
    border: none;
    color: white;
    border-radius: 4px;
    cursor: pointer;
    font-weight: 500;
}

.retry-btn:hover {
    background: #ff7875;
}

@keyframes fadeIn {
    from { opacity: 0; transform: translateY(10px); }
    to { opacity: 1; transform: translateY(0); }
}

/* 错误界面样式 */
.error-container {
    text-align: center;
    padding: 10px;
    animation: fadeIn 0.3s ease;
}

.error-symbol {
    font-size: 48px;
    margin-bottom: 10px;
    display: block;
}

.error-container h3 {
    color: #ff4d4f;
    margin: 0 0 10px 0;
}

.error-desc {
    color: #aaa;
    font-size: 13px;
    margin-bottom: 20px;
}

.error-box {
    background: #2a1215; /* 深红色背景 */
    border: 1px solid #5c2526;
    border-radius: 6px;
    padding: 15px;
    text-align: left;
    margin-bottom: 25px;
    max-height: 150px;
    overflow-y: auto;
}

.error-box code {
    color: #ffccc7;
    font-family: 'Consolas', monospace;
    font-size: 12px;
    white-space: pre-wrap; /* 允许换行 */
    word-break: break-all;
}

.retry-btn {
    padding: 8px 25px;
    background: #ff4d4f;
    border: none;
    color: white;
    border-radius: 4px;
    cursor: pointer;
    font-weight: bold;
}
.retry-btn:hover { background: #ff7875; }

/* 定义淡入动画 */
@keyframes fadeIn {
    from { opacity: 0; transform: scale(0.95); }
    to { opacity: 1; transform: scale(1); }
}

/* 弹窗整体淡入缩放 */
.modal-enter-active, .modal-leave-active { transition: all 0.3s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; transform: scale(1.05); }

/* 视图切换平滑过渡 */
.fade-enter-active, .fade-slide-leave-active { transition: all 0.2s ease; }
.fade-enter-from { opacity: 0; transform: translateX(10px); }
.fade-leave-to { opacity: 0; transform: translateX(-10px); }

/* 进度条流光动画 */
.progress-fill { position: relative; overflow: hidden; }
.progress-glow {
    position: absolute; top: 0; left: 0; width: 100%; height: 100%;
    background: linear-gradient(90deg, transparent, rgba(255,255,255,0.3), transparent);
    animation: scan 1.5s infinite;
}
@keyframes scan { from { transform: translateX(-100%); } to { transform: translateX(100%); } }



/* 按钮点击反馈 */
button:active { transform: scale(0.95); }

/* 2. 入场交错动画 (Stagger) */
.stagger-enter-active {
    transition: all 0.5s ease;
}
.stagger-enter-from {
    opacity: 0;
    transform: translateY(20px);
}
/* 利用 style 中定义的 --i 实现延迟 */
.stagger-enter-active {
    transition-delay: calc(var(--i) * 0.1s);
}
</style>