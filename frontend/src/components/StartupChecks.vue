<script setup>
import { ref, onMounted, computed } from 'vue'
import { startupText, currentLocale } from '../utils/startupConfig'

// Define emits
const emit = defineEmits(['complete'])

// State
const step = ref('idle') // idle, permission, wsl, error, success
const status = ref('loading') // loading, success, fail
const errorMessage = ref('')
const currentCheck = ref('')

// Computed text based on locale
const t = computed(() => startupText[currentLocale] || startupText['en'])

// --- Interface Functions (Reserved for Backend Integration) ---

/**
 * Checks for Windows Administrator Permissions
 * @returns {Promise<boolean>}
 */
const checkPermissions = async () => {
  // TODO: Replace with actual backend call, e.g., await CheckAdmin()
  // Mock implementation:
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve(true) // Change to false to test failure
    }, 1500)
  })
}

/**
 * Checks if WSL is enabled
 * @returns {Promise<boolean>}
 */
const checkWSLStatus = async () => {
  // TODO: Replace with actual backend call, e.g., await CheckWSL()
  // Mock implementation:
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve(true) // Change to false to test failure
    }, 1500)
  })
}

// --- Logic ---

const runChecks = async () => {
  status.value = 'loading'
  errorMessage.value = ''
  
  // 1. Check Permissions
  step.value = 'permission'
  currentCheck.value = t.value.permissionChecking
  
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
      <!-- Loading State -->
      <div v-if="status === 'loading'" class="state-content">
        <div class="spinner"></div>
        <p class="status-text">{{ currentCheck }}</p>
      </div>

      <!-- Failure State -->
      <div v-else-if="status === 'fail'" class="state-content">
        <div class="icon error-icon">❌</div>
        <p class="status-text error-text">{{ errorMessage }}</p>
        <button class="btn btn-primary retry-btn" @click="retry">
          {{ t.retry }}
        </button>
      </div>

      <!-- Success State (Transient) -->
      <div v-else-if="status === 'success'" class="state-content">
        <div class="icon success-icon">✅</div>
        <p class="status-text">{{ t.wslSuccess }}</p>
      </div>
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
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.icon {
  font-size: 40px;
}

.status-text {
  font-size: var(--font-size-md);
  color: var(--color-text-primary);
  margin: 0;
}

.error-text {
  color: var(--color-error);
}

.retry-btn {
  margin-top: 10px;
  width: 100%;
}
</style>
