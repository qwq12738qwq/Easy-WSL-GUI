import { usePerformanceStore } from '../stores/performance'

// TODO: 待后端对接
export function getPerformanceConfig() {
  console.log('getPerformanceConfig')
  const store = usePerformanceStore()
  return store.$state
}

// TODO: 待后端对接
export function setPerformanceConfig(payload) {
  console.log('setPerformanceConfig', payload)
  const store = usePerformanceStore()
  store.setPerformanceConfig(payload)
}

// TODO: 待后端对接
export function resetToDefault() {
  console.log('resetToDefault')
  const store = usePerformanceStore()
  store.resetToDefault()
}

// TODO: 待后端对接
export function exportWslConfig() {
  const store = usePerformanceStore()
  const { 
    memoryLimit, swap, swapFile, processorCount, 
    networkMode, localhostForwarding, 
    autoMemoryReclaim, sparseVhd, dnsTunneling,
    firewall, autoProxy, hostAddressLoopback,
    guiApplications, debugConsole,
    // New fields
    kernel, kernelModules, kernelCommandLine, safeMode, maxCrashDumpCount,
    nestedVirtualization, vmIdleTimeout, dnsProxy, defaultVhdSize,
    pageReporting, bestEffortDnsParsing, dnsTunnelingIpAddress,
    initialAutoProxyTimeout, ignoredPorts
  } = store.$state
  
  console.log('exportWslConfig')

  let config = `[wsl2]
memory=${memoryLimit}GB
swap=${swap}GB
swapFile=${swapFile}
processors=${processorCount}
networkingMode=${networkMode}
localhostForwarding=${localhostForwarding}
guiApplications=${guiApplications}
debugConsole=${debugConsole}
kernel=${kernel}
kernelModules=${kernelModules}
kernelCommandLine=${kernelCommandLine}
safeMode=${safeMode}
maxCrashDumpCount=${maxCrashDumpCount}
nestedVirtualization=${nestedVirtualization}
vmIdleTimeout=${vmIdleTimeout}
dnsProxy=${dnsProxy}
defaultVhdSize=${defaultVhdSize}GB
pageReporting=${pageReporting}
firewall=${firewall}
dnsTunneling=${dnsTunneling}
autoProxy=${autoProxy}

[experimental]
autoMemoryReclaim=${autoMemoryReclaim}
sparseVhd=${sparseVhd}
bestEffortDnsParsing=${bestEffortDnsParsing}
dnsTunnelingIpAddress=${dnsTunnelingIpAddress}
initialAutoProxyTimeout=${initialAutoProxyTimeout}
hostAddressLoopback=${hostAddressLoopback}
`
  if (ignoredPorts) {
      config += `ignoredPorts=${ignoredPorts}\n`
  }
  
  return config
}
