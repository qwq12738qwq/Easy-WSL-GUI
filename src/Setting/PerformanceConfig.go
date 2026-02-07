//go:build windows
// +build windows

package setting

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

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

func Wriding_PerformanceConfig(config PerformanceConfig) error {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("无法打开用户文件夹: %v", err)
	}

	configFile := filepath.Join(userHome, ".wslconfig")

	content := fmt.Sprintf(`[wsl2]
memory=%dGB
swap=%dGB
swapFile=%s
processors=%d
networkingMode=%s
localhostForwarding=%v
guiApplications=%v
debugConsole=%v
kernel=%s
kernelModules=%s
kernelCommandLine=%s
safeMode=%v
maxCrashDumpCount=%d
nestedVirtualization=%v
vmIdleTimeout=%d
dnsProxy=%v
defaultVhdSize=%dGB
pageReporting=%v
firewall=%v
dnsTunneling=%v
autoProxy=%v

[experimental]
autoMemoryReclaim=%s
sparseVhd=%v
bestEffortDnsParsing=%v
dnsTunnelingIpAddress=%s
initialAutoProxyTimeout=%d
hostAddressLoopback=%v
`,
		config.MemoryLimit,
		config.Swap,
		config.SwapFile,
		config.ProcessorCount,
		config.NetworkMode,
		config.LocalhostForwarding,
		config.GuiApplications,
		config.DebugConsole,
		config.Kernel,
		config.KernelModules,
		config.KernelCommandLine,
		config.SafeMode,
		config.MaxCrashDumpCount,
		config.NestedVirtualization,
		config.VmIdleTimeout,
		config.DnsProxy,
		config.DefaultVhdSize,
		config.PageReporting,
		config.Firewall,
		config.DnsTunneling,
		config.AutoProxy,
		config.AutoMemoryReclaim,
		config.SparseVhd,
		config.BestEffortDnsParsing,
		config.DnsTunnelingIpAddress,
		config.InitialAutoProxyTimeout,
		config.HostAddressLoopback,
	)

	if config.IgnoredPorts != "" {
		content += fmt.Sprintf("ignoredPorts=%s\n", config.IgnoredPorts)
	}

	err = os.WriteFile(configFile, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("无法写入.wslconfig,错误码: %v", err)
	}
	return nil
}

func Rading_PerformanceConfig() PerformanceConfig {
	// 1. 初始化默认值 (与前端 stores/performance.js 保持一致)
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

	// 2. 获取用户主目录路径
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home dir: %v\n", err)
		return config // 返回默认值
	}

	configPath := filepath.Join(homeDir, ".wslconfig")

	// 3. 打开文件
	file, err := os.Open(configPath)
	if os.IsNotExist(err) {
		return config // 文件不存在，返回默认值
	} else if err != nil {
		fmt.Printf("Error opening config file: %v\n", err)
		return config
	}
	defer file.Close()

	// 4. 逐行解析
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过注释和空行
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "[") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// 简单的键值映射
		switch key {
		case "memory":
			config.MemoryLimit = parseSizeToGB(value)
		case "swap":
			config.Swap = parseSizeToGB(value)
		case "swapFile":
			config.SwapFile = value
		case "processors":
			if v, err := strconv.Atoi(value); err == nil {
				config.ProcessorCount = v
			}
		case "networkingMode":
			config.NetworkMode = value
		case "localhostForwarding":
			config.LocalhostForwarding = parseBool(value)
		case "guiApplications":
			config.GuiApplications = parseBool(value)
		case "debugConsole":
			config.DebugConsole = parseBool(value)
		case "kernel":
			config.Kernel = value
		case "kernelModules":
			config.KernelModules = value
		case "kernelCommandLine":
			config.KernelCommandLine = value
		case "safeMode":
			config.SafeMode = parseBool(value)
		case "maxCrashDumpCount":
			if v, err := strconv.Atoi(value); err == nil {
				config.MaxCrashDumpCount = v
			}
		case "nestedVirtualization":
			config.NestedVirtualization = parseBool(value)
		case "vmIdleTimeout":
			if v, err := strconv.Atoi(value); err == nil {
				config.VmIdleTimeout = v
			}
		case "dnsProxy":
			config.DnsProxy = parseBool(value)
		case "defaultVhdSize":
			config.DefaultVhdSize = parseSizeToGB(value)
		case "pageReporting":
			config.PageReporting = parseBool(value)
		case "firewall":
			config.Firewall = parseBool(value)
		case "dnsTunneling":
			config.DnsTunneling = parseBool(value)
		case "autoProxy":
			config.AutoProxy = parseBool(value)
		case "autoMemoryReclaim":
			config.AutoMemoryReclaim = value
		case "sparseVhd":
			config.SparseVhd = parseBool(value)
		case "bestEffortDnsParsing":
			config.BestEffortDnsParsing = parseBool(value)
		case "dnsTunnelingIpAddress":
			config.DnsTunnelingIpAddress = value
		case "initialAutoProxyTimeout":
			if v, err := strconv.Atoi(value); err == nil {
				config.InitialAutoProxyTimeout = v
			}
		case "hostAddressLoopback":
			config.HostAddressLoopback = parseBool(value)
		case "ignoredPorts":
			config.IgnoredPorts = value
		}
	}

	return config
}

// 辅助函数：解析布尔值
func parseBool(v string) bool {
	return strings.ToLower(v) == "true"
}

// 辅助函数：解析大小字符串为 GB (例如 "8GB" -> 8, "2048MB" -> 2)
func parseSizeToGB(v string) int {
	v = strings.ToUpper(v)
	if strings.HasSuffix(v, "GB") {
		numStr := strings.TrimSuffix(v, "GB")
		if val, err := strconv.Atoi(numStr); err == nil {
			return val
		}
	} else if strings.HasSuffix(v, "MB") {
		numStr := strings.TrimSuffix(v, "MB")
		if val, err := strconv.Atoi(numStr); err == nil {
			// 简单处理：MB 转 GB，不足1GB按0算，或者向上取整，这里简单除以1024
			return val / 1024
		}
	} else {
		// 只有数字的情况，默认视为 GB (或者根据实际情况调整)
		if val, err := strconv.Atoi(v); err == nil {
			return val
		}
	}
	return 0
}
