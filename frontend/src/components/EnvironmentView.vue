<script setup>
import { ref, onMounted, reactive } from 'vue'
import { useEnvironmentStore } from '../stores/environment'
import { useAppStore } from '../stores/app'
import { Download, Info, Box, Server, Database, Code, Cpu, Plus, X, Trash2, Globe, Laptop } from 'lucide-vue-next'

const envStore = useEnvironmentStore()
const appStore = useAppStore()

// Tabs
const activeTab = ref('network') // 'local' | 'network'

// Modal State
const showCreateModal = ref(false)
const createForm = reactive({
    name: '',
    description: '',
    icon: 'box',
    version: '1.0.0',
    tags: '',
    author: 'Local User',
    packages: '' // multiline string
})

const availableIcons = [
    { value: 'box', label: '通用', icon: Box },
    { value: 'server', label: '服务', icon: Server },
    { value: 'database', label: '数据', icon: Database },
    { value: 'code', label: '开发', icon: Code },
    { value: 'cpu', label: '系统', icon: Cpu },
]

onMounted(() => {
    envStore.fetchEnvironments()
})

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

const handleInstallClick = (env) => {
    appStore.setCurrentTab('config')
}

const openCreateModal = () => {
    // Reset form
    createForm.name = ''
    createForm.description = ''
    createForm.icon = 'box'
    createForm.version = '1.0.0'
    createForm.tags = ''
    createForm.author = 'Local User'
    createForm.packages = ''
    
    showCreateModal.value = true
}

const closeCreateModal = () => {
    showCreateModal.value = false
}

const handleCreate = () => {
    if (!createForm.name) {
        alert("请输入环境名称")
        return
    }

    const newEnv = {
        id: 'local_' + Date.now(),
        name: createForm.name,
        description: createForm.description,
        icon: createForm.icon,
        version: createForm.version,
        tags: createForm.tags.split(',').map(t => t.trim()).filter(t => t),
        author: createForm.author,
        packages: createForm.packages.split('\n').map(p => p.trim()).filter(p => p),
        isLocal: true
    }

    envStore.addLocalEnvironment(newEnv)
    closeCreateModal()
    activeTab.value = 'local'
}

const handleDelete = (id) => {
    if (confirm("确定要删除这个本地环境预设吗？")) {
        envStore.removeLocalEnvironment(id)
    }
}
</script>

<template>
    <div class="env-view-container">
        <!-- 暂未开放提示 (临时添加) -->
        <div class="coming-soon-overlay">
            <div class="coming-soon-content">
                <div class="icon">🚧</div>
                <h2>环境市场暂未开放</h2>
                <p>该功能正在紧锣密鼓地开发中，敬请期待！</p>
            </div>
        </div>

        <!-- 原有内容 (暂时注释隐藏) -->
        <!-- 
        <div class="view-header">
            <h2>环境市场</h2>
            <p class="subtitle">管理本地自定义环境预设，或从网络市场一键部署。</p>
        </div>

        <div class="tabs">
            <div 
                class="tab-item" 
                :class="{ active: activeTab === 'local' }"
                @click="activeTab = 'local'"
            >
                <Laptop :size="18" />
                本地环境卡片
            </div>
            <div 
                class="tab-item" 
                :class="{ active: activeTab === 'network' }"
                @click="activeTab = 'network'"
            >
                <Globe :size="18" />
                网络环境市场
            </div>
        </div>

        <div v-if="activeTab === 'local'" class="tab-content fade-in">
            <div class="local-actions">
                <button class="create-btn" @click="openCreateModal">
                    <Plus :size="18" /> 创建新环境
                </button>
            </div>

            <div v-if="envStore.localEnvironments.length === 0" class="empty-local">
                <div class="empty-icon">📂</div>
                <p>暂无本地环境预设</p>
                <span class="sub-text">点击上方“创建新环境”开始制作</span>
            </div>

            <div v-else class="env-grid">
                <div v-for="env in envStore.localEnvironments" :key="env.id" class="env-card local-card">
                    <div class="card-header">
                        <div class="icon-wrapper local">
                            <component :is="getIcon(env.icon)" :size="32" />
                        </div>
                        <div class="header-text">
                            <h3>{{ env.name }}</h3>
                            <span class="version">v{{ env.version }}</span>
                        </div>
                        <button class="delete-btn" @click.stop="handleDelete(env.id)" title="删除预设">
                            <Trash2 :size="16" />
                        </button>
                    </div>
                    
                    <p class="description">{{ env.description }}</p>
                    
                    <div class="packages-preview" v-if="env.packages && env.packages.length">
                        <strong>包含软件包:</strong>
                        <div class="pkg-list">
                            <span v-for="pkg in env.packages.slice(0, 3)" :key="pkg" class="pkg-tag">{{ pkg }}</span>
                            <span v-if="env.packages.length > 3" class="pkg-more">+{{ env.packages.length - 3 }}</span>
                        </div>
                    </div>

                    <div class="card-footer">
                        <span class="author">By {{ env.author }}</span>
                        <button class="install-btn" @click="handleInstallClick(env)">
                            <Download :size="16" /> 部署
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <div v-else class="tab-content fade-in">
            <div v-if="envStore.loading" class="loading-state">
                <div class="spinner"></div>
                <p>正在加载网络市场...</p>
            </div>

            <div v-else class="env-grid">
                <div v-for="env in envStore.environments" :key="env.id" class="env-card">
                    <div class="card-header">
                        <div class="icon-wrapper" :class="env.id">
                            <component :is="getIcon(env.icon)" :size="32" />
                        </div>
                        <div class="header-text">
                            <h3>{{ env.name }}</h3>
                            <span class="version">v{{ env.version }}</span>
                        </div>
                    </div>
                    
                    <p class="description">{{ env.description }}</p>
                    
                    <div class="tags">
                        <span v-for="tag in env.tags" :key="tag" class="tag">{{ tag }}</span>
                    </div>

                    <div class="card-footer">
                        <span class="author">By {{ env.author }}</span>
                        <button class="install-btn" @click="handleInstallClick(env)">
                            <Download :size="16" /> 部署
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <Transition name="modal">
            <div v-if="showCreateModal" class="modal-overlay" @click.self="closeCreateModal">
                <div class="modal-window create-modal">
                    <div class="modal-header">
                        <h3>创建本地环境预设</h3>
                        <button class="close-btn" @click="closeCreateModal">
                            <X :size="20" />
                        </button>
                    </div>
                    <div class="modal-body">
                        <div class="form-row">
                            <div class="form-group">
                                <label>环境名称</label>
                                <input v-model="createForm.name" type="text" placeholder="例如: My Python Dev" class="input">
                            </div>
                            <div class="form-group sm">
                                <label>版本</label>
                                <input v-model="createForm.version" type="text" placeholder="1.0.0" class="input">
                            </div>
                        </div>

                        <div class="form-group">
                            <label>图标选择</label>
                            <div class="icon-selector">
                                <div 
                                    v-for="item in availableIcons" 
                                    :key="item.value"
                                    class="icon-option"
                                    :class="{ selected: createForm.icon === item.value }"
                                    @click="createForm.icon = item.value"
                                >
                                    <component :is="item.icon" :size="20" />
                                    <span>{{ item.label }}</span>
                                </div>
                            </div>
                        </div>

                        <div class="form-group">
                            <label>描述</label>
                            <input v-model="createForm.description" type="text" placeholder="简短描述该环境的用途..." class="input">
                        </div>

                        <div class="form-group">
                            <label>标签 (逗号分隔)</label>
                            <input v-model="createForm.tags" type="text" placeholder="Dev, Web, Tools..." class="input">
                        </div>

                        <div class="form-group">
                            <label>软件包列表 (每行一个)</label>
                            <textarea v-model="createForm.packages" rows="5" placeholder="git&#10;vim&#10;curl" class="textarea"></textarea>
                            <p class="hint">这些软件包将在部署时尝试自动安装</p>
                        </div>
                    </div>
                    <div class="modal-footer">
                        <button class="btn btn-secondary" @click="closeCreateModal">取消</button>
                        <button class="btn btn-primary" @click="handleCreate">创建环境</button>
                    </div>
                </div>
            </div>
        </Transition>
        -->
    </div>
</template>

<style scoped>
.env-view-container {
    padding: 2rem;
    height: 100%;
    overflow-y: auto;
    box-sizing: border-box;
}

.view-header h2 {
    margin-bottom: 0.5rem;
    font-size: 1.8rem;
}

.subtitle {
    color: var(--text-secondary);
    margin-bottom: 1.5rem;
}

/* Tabs */
.tabs {
    display: flex;
    gap: 1rem;
    margin-bottom: 2rem;
    border-bottom: 1px solid var(--border-color);
}

.tab-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1.5rem;
    cursor: pointer;
    color: var(--text-secondary);
    border-bottom: 2px solid transparent;
    transition: all 0.2s;
    font-weight: 500;
}

.tab-item:hover {
    color: var(--text-color);
    background: var(--hover-bg);
}

.tab-item.active {
    color: var(--primary-color);
    border-bottom-color: var(--primary-color);
}

/* Local Actions */
.local-actions {
    margin-bottom: 1.5rem;
}

.create-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1.5rem;
    background: var(--primary-color);
    color: white;
    border: none;
    border-radius: 8px;
    cursor: pointer;
    font-weight: 600;
    transition: opacity 0.2s;
}

.create-btn:hover {
    opacity: 0.9;
}

/* Grid & Cards */
.env-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 1.5rem;
}

.env-card {
    background: var(--card-bg);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    padding: 1.5rem;
    transition: all 0.2s;
    display: flex;
    flex-direction: column;
    position: relative;
}

.env-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 24px rgba(0,0,0,0.05);
    border-color: var(--primary-color);
}

.card-header {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-bottom: 1rem;
}

.icon-wrapper {
    width: 56px;
    height: 56px;
    border-radius: 12px;
    background: var(--bg-color);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--primary-color);
}

.icon-wrapper.local { color: var(--text-color); background: var(--bg-color); border: 1px solid var(--border-color); }
.icon-wrapper.docker { color: #0db7ed; background: rgba(13, 183, 237, 0.1); }
.icon-wrapper.openwrt { color: #cf3e52; background: rgba(207, 62, 82, 0.1); }
.icon-wrapper.nodejs_dev { color: #68a063; background: rgba(104, 160, 99, 0.1); }
.icon-wrapper.python_data { color: #3776ab; background: rgba(55, 118, 171, 0.1); }
.icon-wrapper.go_dev { color: #00add8; background: rgba(0, 173, 216, 0.1); }

.header-text h3 {
    font-size: 1.1rem;
    margin: 0 0 0.25rem 0;
}

.version {
    font-size: 0.8rem;
    color: var(--text-secondary);
    background: var(--bg-color);
    padding: 2px 6px;
    border-radius: 4px;
}

.delete-btn {
    margin-left: auto;
    background: none;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
    transition: all 0.2s;
}

.delete-btn:hover {
    background: rgba(255, 0, 0, 0.1);
    color: red;
}

.description {
    color: var(--text-secondary);
    font-size: 0.9rem;
    line-height: 1.5;
    margin-bottom: 1.5rem;
    flex: 1;
}

.tags {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-bottom: 1.5rem;
}

.tag {
    font-size: 0.75rem;
    padding: 2px 8px;
    border-radius: 100px;
    background: var(--bg-color);
    color: var(--text-secondary);
    border: 1px solid var(--border-color);
}

.packages-preview {
    background: var(--bg-color);
    padding: 0.75rem;
    border-radius: 8px;
    margin-bottom: 1.5rem;
    font-size: 0.85rem;
}

.packages-preview strong {
    display: block;
    margin-bottom: 0.5rem;
    color: var(--text-color);
}

.pkg-list {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
}

.pkg-tag {
    background: var(--card-bg);
    border: 1px solid var(--border-color);
    padding: 2px 6px;
    border-radius: 4px;
}

.pkg-more {
    color: var(--text-secondary);
    padding: 2px 4px;
}

.card-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-top: 1rem;
    border-top: 1px solid var(--border-color);
}

.author {
    font-size: 0.8rem;
    color: var(--text-secondary);
}

.install-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 1rem;
    background: var(--primary-color);
    color: white;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.9rem;
    font-weight: 500;
    transition: opacity 0.2s;
}

.install-btn:hover {
    opacity: 0.9;
}

/* Empty State */
.empty-local {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 4rem;
    color: var(--text-secondary);
    background: var(--card-bg);
    border-radius: 12px;
    border: 1px dashed var(--border-color);
}

.empty-icon {
    font-size: 3rem;
    margin-bottom: 1rem;
}

.sub-text {
    font-size: 0.9rem;
    opacity: 0.7;
    margin-top: 0.5rem;
}

/* Loading */
.loading-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 300px;
    color: var(--text-secondary);
}

.spinner {
    width: 40px;
    height: 40px;
    border: 3px solid rgba(0,0,0,0.1);
    border-radius: 50%;
    border-top-color: var(--primary-color);
    animation: spin 1s linear infinite;
    margin-bottom: 1rem;
}

/* Modal Styles */
.modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
}

.modal-window {
    background: var(--card-bg);
    border-radius: 12px;
    width: 500px;
    max-width: 90%;
    box-shadow: 0 20px 50px rgba(0,0,0,0.3);
    display: flex;
    flex-direction: column;
    max-height: 85vh;
    border: 1px solid var(--border-color);
}

.modal-header {
    padding: 1.5rem;
    border-bottom: 1px solid var(--border-color);
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.modal-header h3 {
    margin: 0;
    font-size: 1.25rem;
}

.close-btn {
    background: none;
    border: none;
    cursor: pointer;
    color: var(--text-secondary);
    padding: 4px;
    border-radius: 4px;
}

.close-btn:hover {
    background: var(--hover-bg);
    color: var(--text-color);
}

.modal-body {
    padding: 1.5rem;
    overflow-y: auto;
}

.modal-footer {
    padding: 1.5rem;
    border-top: 1px solid var(--border-color);
    display: flex;
    justify-content: flex-end;
    gap: 1rem;
}

.form-row {
    display: flex;
    gap: 1rem;
}

.form-group {
    margin-bottom: 1.25rem;
    flex: 1;
}

.form-group.sm {
    flex: 0 0 100px;
}

.form-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 500;
    font-size: 0.9rem;
}

.input, .textarea {
    width: 100%;
    padding: 0.75rem;
    border-radius: 8px;
    border: 1px solid var(--border-color);
    background: var(--bg-color);
    color: var(--text-color);
    font-family: inherit;
    box-sizing: border-box;
}

.input:focus, .textarea:focus {
    border-color: var(--primary-color);
    outline: none;
}

.hint {
    font-size: 0.8rem;
    color: var(--text-secondary);
    margin-top: 0.5rem;
}

.icon-selector {
    display: flex;
    gap: 0.75rem;
    flex-wrap: wrap;
}

.icon-option {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem;
    border: 1px solid var(--border-color);
    border-radius: 8px;
    cursor: pointer;
    width: 60px;
    transition: all 0.2s;
}

.icon-option:hover {
    background: var(--hover-bg);
}

.icon-option.selected {
    border-color: var(--primary-color);
    background: rgba(13, 183, 237, 0.1);
    color: var(--primary-color);
}

.icon-option span {
    font-size: 0.75rem;
}

.btn {
    padding: 0.75rem 1.5rem;
    border-radius: 8px;
    border: none;
    font-weight: 600;
    cursor: pointer;
    transition: opacity 0.2s;
}

.btn-primary {
    background: var(--primary-color);
    color: white;
}

.btn-secondary {
    background: var(--bg-color);
    color: var(--text-color);
    border: 1px solid var(--border-color);
}

.btn:hover {
    opacity: 0.9;
}

.fade-in {
    animation: fadeIn 0.3s ease-out;
}

@keyframes fadeIn {
    from { opacity: 0; transform: translateY(5px); }
    to { opacity: 1; transform: translateY(0); }
}

/* Coming Soon Overlay */
.coming-soon-overlay {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-color);
    z-index: 10;
}

.coming-soon-content {
    text-align: center;
    color: var(--text-secondary);
}

.coming-soon-content .icon {
    font-size: 4rem;
    margin-bottom: 1rem;
}

.coming-soon-content h2 {
    font-size: 1.5rem;
    margin-bottom: 0.5rem;
    color: var(--text-color);
}

.coming-soon-content p {
    font-size: 1rem;
    opacity: 0.8;
}
</style>
