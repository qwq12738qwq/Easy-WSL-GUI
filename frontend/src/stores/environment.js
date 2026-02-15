import { defineStore } from 'pinia'

// Mock Data
const MOCK_ENVIRONMENTS = [
    {
        id: "docker",
        name: "Docker Engine",
        description: "完整的 Docker 容器运行环境，包含 Docker CE 和 Docker Compose，支持 WSL2 后端集成。",
        icon: "docker",
        tags: ["Container", "DevOps"],
        version: "24.0.5",
        author: "Docker Inc."
    },
    {
        id: "openwrt",
        name: "OpenWrt Build",
        description: "预配置的 OpenWrt 固件编译环境，包含 build-essential, ncurses, zlib, awk 等所有必要依赖。",
        icon: "router",
        tags: ["Embedded", "Build", "C/C++"],
        version: "23.05",
        author: "OpenWrt Community"
    },
    {
        id: "nodejs_dev",
        name: "Node.js Full Stack",
        description: "Node.js 全栈开发环境，包含 nvm (Node Version Manager), node, npm, yarn, pnpm。",
        icon: "nodejs",
        tags: ["Web", "Backend", "JavaScript"],
        version: "LTS",
        author: "Node.js Foundation"
    },
    {
        id: "python_data",
        name: "Python Data Science",
        description: "Python 数据科学环境，集成 Anaconda, Jupyter, Pandas, NumPy, Matplotlib。",
        icon: "python",
        tags: ["Data Science", "AI", "Python"],
        version: "3.11",
        author: "Anaconda"
    },
    {
        id: "go_dev",
        name: "Go Development",
        description: "Go 语言开发环境，包含 Go SDK, Delve Debugger, gopls 语言服务器。",
        icon: "go",
        tags: ["Backend", "Systems", "Go"],
        version: "1.21",
        author: "Google"
    }
]

export const useEnvironmentStore = defineStore('environment', {
  state: () => ({
    environments: [], // Network environments
    localEnvironments: [], // Local custom environments
    loading: false
  }),
  actions: {
    async fetchEnvironments() {
      this.loading = true
      // Simulate network delay
      await new Promise(resolve => setTimeout(resolve, 600))
      
      // In real app: const res = await fetch('/api/environments')
      // this.environments = await res.json()
      this.environments = MOCK_ENVIRONMENTS
      this.loading = false
      
      // Load local environments from localStorage if available
      const savedLocal = localStorage.getItem('local_environments')
      if (savedLocal) {
          try {
              this.localEnvironments = JSON.parse(savedLocal)
          } catch (e) {
              console.error("Failed to parse local environments", e)
          }
      }
    },
    
    // Simulate getting environment details
    getEnvironmentById(id) {
        return this.environments.find(e => e.id === id) || this.localEnvironments.find(e => e.id === id)
    },

    addLocalEnvironment(env) {
        this.localEnvironments.push(env)
        this.saveLocalEnvironments()
    },

    removeLocalEnvironment(id) {
        this.localEnvironments = this.localEnvironments.filter(e => e.id !== id)
        this.saveLocalEnvironments()
    },

    saveLocalEnvironments() {
        localStorage.setItem('local_environments', JSON.stringify(this.localEnvironments))
    }
  }
})
