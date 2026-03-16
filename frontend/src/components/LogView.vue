<script setup>
import { ref, onMounted, nextTick, watch, onUnmounted } from 'vue'
import { GetLogs } from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { FileText, RotateCcw, Download, Search, Terminal, Trash2 } from 'lucide-vue-next'

const logs = ref([])
const filteredLogs = ref([])
const searchQuery = ref('')
const isLoading = ref(true)
const logContainer = ref(null)
const autoScroll = ref(true)

// 加载历史日志
const loadHistoryLogs = async () => {
    try {
        const data = await GetLogs()
        logs.value = data
        filterLogs()
    } catch (e) {
        console.error("加载历史日志失败:", e)
    } finally {
        isLoading.value = false
    }
}

// 监听实时日志事件
const setupRealtimeLogs = () => {
    // 监听 "log-event" 事件，后端通过 runtime.EventsEmit 推送
    EventsOn("log-event", (data) => {
        // data 是字符串，直接添加
        if (data) {
            logs.value.push(data)
            // 只有当自动滚动开启时才自动滚动
            if (autoScroll.value) {
                filterLogs() // 触发更新
            }
        }
    })
}

// 停止监听
const cleanupRealtimeLogs = () => {
    EventsOff("log-event")
}

// 过滤日志
const filterLogs = () => {
    if (!searchQuery.value) {
        filteredLogs.value = [...logs.value]
    } else {
        const query = searchQuery.value.toLowerCase()
        filteredLogs.value = logs.value.filter(log => 
            log.toLowerCase().includes(query)
        )
    }
    
    // 自动滚动到底部
    if (autoScroll.value && logContainer.value) {
        nextTick(() => {
            logContainer.value.scrollTop = logContainer.value.scrollHeight
        })
    }
}

// 清除日志 (仅清空本地显示)
const clearLogsLocal = () => {
    logs.value = []
    filteredLogs.value = []
}

// 导出日志
const downloadLogs = () => {
    const blob = new Blob([logs.value.join('\n')], { type: 'text/plain;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `easy-wsl-logs-${new Date().toISOString().slice(0, 10)}.txt`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
}

// 搜索变化
watch(searchQuery, () => {
    filterLogs()
})

// 手动滚动时关闭自动滚动
const handleScroll = () => {
    if (!logContainer.value) return
    const { scrollTop, scrollHeight, clientHeight } = logContainer.value
    if (scrollHeight - scrollTop - clientHeight < 50) {
        autoScroll.value = true
    } else {
        autoScroll.value = false
    }
}

onMounted(() => {
    isLoading.value = true
    loadHistoryLogs() // 1. 读取历史文件
    setupRealtimeLogs() // 2. 监听实时事件
})

onUnmounted(() => {
    cleanupRealtimeLogs() // 3. 离开页面时取消监听
})
</script>

<template>
    <div class="log-view-container">
        <header class="view-header">
            <div class="header-left">
                <Terminal class="header-icon" :size="24" />
                <h2>运行日志</h2>
            </div>
            <div class="header-actions">
                <button class="btn-icon" @click="loadLogs" title="刷新日志">
                    <RotateCcw :size="18" />
                </button>
                <!-- <button class="btn-icon" @click="clearLogsLocal" title="清除日志">
                    <Trash2 :size="18" />
                </button> -->
                <button class="btn-icon" @click="downloadLogs" title="下载日志">
                    <Download :size="18" />
                </button>
            </div>
        </header>
        
        <div class="search-bar">
            <Search class="search-icon" :size="16" />
            <input 
                type="text" 
                v-model="searchQuery" 
                placeholder="搜索日志..." 
                class="search-input"
            />
        </div>
        
        <div 
            class="log-container" 
            ref="logContainer"
            @scroll="handleScroll"
        >
            <div v-if="isLoading" class="loading-state">
                <div class="spinner"></div>
                <span>加载中...</span>
            </div>
            
            <div v-else-if="filteredLogs.length === 0" class="empty-state">
                <FileText :size="48" class="empty-icon" />
                <p>暂无日志记录</p>
            </div>
            
            <div v-else class="log-content">
                <div 
                    v-for="(log, index) in filteredLogs" 
                    :key="index" 
                    class="log-line"
                    :class="{ 
                        'log-error': log.toLowerCase().includes('error'),
                        'log-warn': log.toLowerCase().includes('warn'),
                        'log-info': !log.toLowerCase().includes('error') && !log.toLowerCase().includes('warn')
                    }"
                >
                    <span class="log-index">{{ index + 1 }}</span>
                    <span class="log-text">{{ log }}</span>
                </div>
            </div>
            
            <div v-if="autoScroll && filteredLogs.length > 0" class="scroll-hint">
                自动滚动已开启
            </div>
        </div>
    </div>
</template>

<style scoped>
.log-view-container {
    display: flex;
    flex-direction: column;
    height: 100%;
    color: var(--color-text-primary);
}

.view-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    flex-shrink: 0;
}

.header-left {
    display: flex;
    align-items: center;
    gap: 12px;
}

.header-icon {
    color: var(--color-brand);
}

h2 {
    margin: 0;
    font-size: 1.5rem;
    font-weight: 600;
}

.header-actions {
    display: flex;
    gap: 8px;
}

.btn-icon {
    background: var(--color-bg-hover);
    border: 1px solid var(--color-border);
    color: var(--color-text-secondary);
    padding: 8px;
    border-radius: var(--radius-md);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all var(--transition-fast);
}

.btn-icon:hover {
    background: var(--color-border-hover);
    color: var(--color-text-primary);
}

.btn-text {
    font-size: 0.9rem;
    margin-left: 4px;
}

.search-bar {
    position: relative;
    margin-bottom: 16px;
    flex-shrink: 0;
}

.search-icon {
    position: absolute;
    left: 12px;
    top: 50%;
    transform: translateY(-50%);
    color: var(--color-text-tertiary);
}

.search-input {
    width: 100%;
    padding: 10px 12px 10px 36px;
    background: var(--color-bg-input);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    color: var(--color-text-primary);
    font-size: 0.95rem;
    outline: none;
    transition: border-color var(--transition-fast);
    box-sizing: border-box;
}

.search-input:focus {
    border-color: var(--color-brand);
}

.log-container {
    flex: 1;
    background: var(--color-bg-card);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    padding: 16px;
    overflow-y: auto;
    font-family: 'Consolas', 'Monaco', monospace;
    font-size: 0.85rem;
    position: relative;
    min-height: 0;
}

.log-content {
    display: flex;
    flex-direction: column;
    gap: 4px;
}

.log-line {
    display: flex;
    gap: 12px;
    padding: 4px 8px;
    border-radius: 4px;
    line-height: 1.5;
    word-break: break-all;
}

.log-line:hover {
    background: var(--color-bg-hover);
}

.log-index {
    color: var(--color-text-tertiary);
    user-select: none;
    min-width: 30px;
    text-align: right;
}

.log-text {
    color: var(--color-text-secondary);
}

.log-error .log-text {
    color: var(--color-danger, #ff4d4f);
}

.log-warn .log-text {
    color: var(--color-warning, #faad14);
}

.log-info .log-text {
    color: var(--color-text-primary);
}

.loading-state, .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: var(--color-text-tertiary);
    gap: 12px;
}

.empty-icon {
    opacity: 0.5;
}

.spinner {
    width: 24px;
    height: 24px;
    border: 2px solid var(--color-border);
    border-top-color: var(--color-brand);
    border-radius: 50%;
    animation: spin 1s linear infinite;
}

@keyframes spin {
    to { transform: rotate(360deg); }
}

.scroll-hint {
    position: absolute;
    bottom: 8px;
    right: 16px;
    font-size: 0.75rem;
    color: var(--color-text-tertiary);
    background: var(--color-bg-card);
    padding: 2px 8px;
    border-radius: 4px;
    pointer-events: none;
}
</style>
