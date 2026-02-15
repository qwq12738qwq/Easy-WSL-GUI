import { defineStore } from 'pinia'

export const useDistroStore = defineStore('distro', {
  state: () => ({
    selectedDistro: ''
  }),
  actions: {
    selectDistro(name) {
      this.selectedDistro = name
    }
  }
})
