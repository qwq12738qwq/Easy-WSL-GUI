import { defineStore } from 'pinia'

export const usePerformanceStore = defineStore('performance', {
  state: () => ({
    memoryLimit: 8,
    swap: 0,
    swapFile: 'C:\\wsl.swap',
    processorCount: 4,
    networkMode: 'mirrored',
    localhostForwarding: true,
    autoMemoryReclaim: 'dropCache',
    sparseVhd: true,
    dnsTunneling: true,
    firewall: true,
    autoProxy: true,
    hostAddressLoopback: true,
    guiApplications: true,
    debugConsole: false,
    // New fields
    kernel: '',
    kernelModules: '',
    kernelCommandLine: '',
    safeMode: false,
    maxCrashDumpCount: 10,
    nestedVirtualization: true,
    vmIdleTimeout: 60000,
    dnsProxy: true,
    defaultVhdSize: 1024, // GB
    pageReporting: true,
    bestEffortDnsParsing: false,
    dnsTunnelingIpAddress: '10.255.255.254',
    initialAutoProxyTimeout: 1000,
    ignoredPorts: ''
  }),
  actions: {
    setPerformanceConfig(config) {
      this.$state = { ...this.$state, ...config }
    },
    resetToDefault() {
      this.$state = {
        memoryLimit: 8,
        swap: 0,
        swapFile: 'C:\\wsl.swap',
        processorCount: 4,
        networkMode: 'mirrored',
        localhostForwarding: true,
        autoMemoryReclaim: 'dropCache',
        sparseVhd: true,
        dnsTunneling: true,
        firewall: true,
        autoProxy: true,
        hostAddressLoopback: true,
        guiApplications: true,
        debugConsole: false,
        // New fields defaults
        kernel: '',
        kernelModules: '',
        kernelCommandLine: '',
        safeMode: false,
        maxCrashDumpCount: 10,
        nestedVirtualization: true,
        vmIdleTimeout: 60000,
        dnsProxy: true,
        defaultVhdSize: 1024,
        pageReporting: true,
        bestEffortDnsParsing: false,
        dnsTunnelingIpAddress: '10.255.255.254',
        initialAutoProxyTimeout: 1000,
        ignoredPorts: ''
      }
    }
  }
})
