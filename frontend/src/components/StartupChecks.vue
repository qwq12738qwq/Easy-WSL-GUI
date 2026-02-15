<script setup>
import { ref, onMounted, computed } from 'vue'
import { startupText, currentLocale } from '../utils/startupConfig'
import { CheckAdmin, CheckWSL, EnableWSLFeature } from '../../wailsjs/go/main/App'

// Define emits
const emit = defineEmits(['complete'])

// State
const step = ref('idle') // idle, permission, wsl, error, success
const status = ref('loading') // loading, success, fail
const errorMessage = ref('')
const currentCheck = ref('')
const isEnablingWSL = ref(false)

// Computed text based on locale
const t = computed(() => startupText[currentLocale] || startupText['en'])

// --- Interface Functions ---

/**
 * Checks for Windows Administrator Permissions
 * @returns {Promise<boolean>}
 */
const checkPermissions = async () => {
  try {
    return await CheckAdmin()
  } catch (e) {
    console.error("CheckAdmin failed", e)
    return false
  }
}

/**
 * Checks if WSL is enabled
 * @returns {Promise<boolean>}
 */
const checkWSLStatus = async () => {
  try {
    return await CheckWSL()
  } catch (e) {
    console.error("CheckWSL failed", e)
    return false
  }
}

const handleEnableWSL = async () => {
  if (isEnablingWSL.value) return
  isEnablingWSL.value = true
  try {
    await EnableWSLFeature()
    // 提示用户重启
    errorMessage.value = "已发送启用指令，请重启电脑后重试。"
  } catch (e) {
    alert("启用失败: " + e)
    isEnablingWSL.value = false
  }
}

// --- Logic ---

const runChecks = async () => {
  status.value = 'loading'
  errorMessage.value = ''
  isEnablingWSL.value = false
  
  // 1. Check Permissions
  step.value = 'permission'
  currentCheck.value = t.value.permissionChecking
  
  // 模拟一点延迟以展示动画
  await new Promise(r => setTimeout(r, 800))

  try {
    const permResult = await withTimeout(checkPermissions(), 5000)
    if (!permResult) {
      handleFail(t.value.permissionFail)
      return
    }
  } catch (e) {
    handleFail(e.message || t.value.error)
    return
  }
  
  // 2. Check WSL
  step.value = 'wsl'
  currentCheck.value = t.value.wslChecking
  
  // 模拟一点延迟以展示动画
  await new Promise(r => setTimeout(r, 800))

  try {
    const wslResult = await withTimeout(checkWSLStatus(), 5000)
    if (!wslResult) {
      handleFail(t.value.wslFail)
      return
    }
  } catch (e) {
    handleFail(e.message || t.value.error)
    return
  }
  
  // Success
  step.value = 'success'
  status.value = 'success'
  currentCheck.value = t.value.wslSuccess
  
  setTimeout(() => {
    emit('complete')
  }, 1000)
}

const withTimeout = (promise, ms) => {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => {
      reject(new Error(t.value.timeout))
    }, ms)
    
    promise
      .then((res) => {
        clearTimeout(timer)
        resolve(res)
      })
      .catch((err) => {
        clearTimeout(timer)
        reject(err)
      })
  })
}

const handleFail = (msg) => {
  status.value = 'fail'
  errorMessage.value = msg
}

const retry = () => {
  runChecks()
}

onMounted(() => {
  runChecks()
})

</script>

<template>
  <div class="startup-overlay">
    <div class="check-card">
      <Transition name="fade" mode="out-in">
        <!-- Loading State -->
        <div v-if="status === 'loading'" class="state-content" :key="step">
          <div class="spinner"></div>
          <p class="status-text">{{ currentCheck }}</p>
        </div>

        <!-- Failure State -->
        <div v-else-if="status === 'fail'" class="state-content" key="fail">
          <div class="icon error-icon">❌</div>
          <p class="status-text error-text">{{ errorMessage }}</p>
          
          <div class="action-buttons">
            <button v-if="step === 'wsl'" class="btn btn-secondary enable-btn" @click="handleEnableWSL" :disabled="isEnablingWSL">
              <span v-if="isEnablingWSL">正在处理...</span>
              <span v-else>启用 WSL 组件</span>
            </button>
            <button class="btn btn-primary retry-btn" @click="retry">
              {{ t.retry }}
            </button>
          </div>
        </div>

        <!-- Success State (Transient) -->
        <div v-else-if="status === 'success'" class="state-content" key="success">
          <div class="icon success-icon">✅</div>
          <p class="status-text">{{ t.wslSuccess }}</p>
        </div>
      </Transition>
    </div>
  </div>
</template>

<style scoped>
.startup-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background-color: var(--color-bg-body);
  z-index: 9999;
  display: flex;
  justify-content: center;
  align-items: center;
}

.check-card {
  background: var(--color-bg-card);
  padding: 40px;
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  text-align: center;
  width: 360px;
  border: 1px solid var(--color-border);
  transition: all 0.3s ease;
}

.state-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
  min-height: 160px;
  justify-content: center;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid var(--color-bg-hover);
  border-top-color: var(--color-brand);
  border-radius: 50%;
  animation: spin 1s cubic-bezier(0.68, -0.55, 0.27, 1.55) infinite; /* 更丝滑的旋转 */
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.icon {
  font-size: 40px;
  animation: popIn 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

@keyframes popIn {
  0% { opacity: 0; transform: scale(0.5); }
  100% { opacity: 1; transform: scale(1); }
}

.status-text {
  font-size: var(--font-size-md);
  color: var(--color-text-primary);
  margin: 0;
  animation: slideUp 0.3s ease-out;
}

.error-text {
  color: var(--color-error);
}

.action-buttons {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: 10px;
}

.btn {
  width: 100%;
  padding: 10px;
  border-radius: 6px;
  cursor: pointer;
  border: none;
  font-weight: 500;
  transition: transform 0.1s, opacity 0.2s;
}

.btn:active {
  transform: scale(0.98);
}

.btn-primary {
  background: var(--color-brand);
  color: white;
}

.btn-secondary {
  background: var(--color-bg-hover);
  color: var(--color-text-primary);
  border: 1px solid var(--color-border);
}

/* Transition Animations */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.fade-enter-from {
  opacity: 0;
  transform: translateY(10px);
}

.fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
