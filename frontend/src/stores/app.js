import { defineStore } from 'pinia'

export const useAppStore = defineStore('app', {
  state: () => ({
    currentTab: 'home', // default tab
    selectedDistro: ''
  }),
  actions: {
    setCurrentTab(tab) {
      this.currentTab = tab
    },
    selectDistro(name) {
      this.selectedDistro = name
    },
    navigateToConfig(distroName) {
        this.selectedDistro = distroName
        this.currentTab = 'config'
    }
  }
})
