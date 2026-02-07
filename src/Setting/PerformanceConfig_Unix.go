//go:build !windows
// +build !windows

package setting

type PerformanceConfig struct {
	MemoryLimit             int    `json:"memoryLimit"`
	Swap                    int    `json:"swap"`
	SwapFile                string `json:"swapFile"`
	ProcessorCount          int    `json:"processorCount"`
	NetworkMode             string `json:"networkMode"`
	LocalhostForwarding     bool   `json:"localhostForwarding"`
	AutoMemoryReclaim       string `json:"autoMemoryReclaim"`
	SparseVhd               bool   `json:"sparseVhd"`
	DnsTunneling            bool   `json:"dnsTunneling"`
	Firewall                bool   `json:"firewall"`
	AutoProxy               bool   `json:"autoProxy"`
	HostAddressLoopback     bool   `json:"hostAddressLoopback"`
	GuiApplications         bool   `json:"guiApplications"`
	DebugConsole            bool   `json:"debugConsole"`
	Kernel                  string `json:"kernel"`
	KernelModules           string `json:"kernelModules"`
	KernelCommandLine       string `json:"kernelCommandLine"`
	SafeMode                bool   `json:"safeMode"`
	MaxCrashDumpCount       int    `json:"maxCrashDumpCount"`
	NestedVirtualization    bool   `json:"nestedVirtualization"`
	VmIdleTimeout           int    `json:"vmIdleTimeout"`
	DnsProxy                bool   `json:"dnsProxy"`
	DefaultVhdSize          int    `json:"defaultVhdSize"`
	PageReporting           bool   `json:"pageReporting"`
	BestEffortDnsParsing    bool   `json:"bestEffortDnsParsing"`
	DnsTunnelingIpAddress   string `json:"dnsTunnelingIpAddress"`
	InitialAutoProxyTimeout int    `json:"initialAutoProxyTimeout"`
	IgnoredPorts            string `json:"ignoredPorts"`
}

// 写入.wslconfig函数
func Wriding_PerformanceConfig(config PerformanceConfig) error { return nil }

// 读取.wslconfig函数
func Rading_PerformanceConfig() PerformanceConfig {
	config := PerformanceConfig{
		MemoryLimit:             8,
		Swap:                    0,
		SwapFile:                `C:\\wsl.swap`,
		ProcessorCount:          4,
		NetworkMode:             "mirrored",
		LocalhostForwarding:     true,
		AutoMemoryReclaim:       "dropCache",
		SparseVhd:               true,
		DnsTunneling:            true,
		Firewall:                true,
		AutoProxy:               true,
		HostAddressLoopback:     true,
		GuiApplications:         true,
		DebugConsole:            false,
		VmIdleTimeout:           60000,
		DefaultVhdSize:          1024,
		DnsTunnelingIpAddress:   "10.255.255.254",
		InitialAutoProxyTimeout: 1000,
	}
	return config
}
