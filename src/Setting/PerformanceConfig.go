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
	PageReporting           bool   `json:"pageReporting"`
	BestEffortDnsParsing    bool   `json:"bestEffortDnsParsing"`
	DnsTunnelingIpAddress   string `json:"dnsTunnelingIpAddress"`
	InitialAutoProxyTimeout int    `json:"initialAutoProxyTimeout"`
	IgnoredPorts            string `json:"ignoredPorts"`
	UseWindowsDnsCli        bool   `json:"useWindowsDnsCli"`
}

func Wriding_PerformanceConfig(config PerformanceConfig) error {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("无法打开用户文件夹: %v", err)
	}

	configFile := filepath.Join(userHome, ".wslconfig")

	var sb strings.Builder
	sb.WriteString("[wsl2]\n")

	if config.MemoryLimit > 0 {
		sb.WriteString(fmt.Sprintf("memory=%dGB\n", config.MemoryLimit))
	}

	// Swap: -1 or <0 means default/unmanaged. 0 means disabled.
	if config.Swap >= 0 {
		sb.WriteString(fmt.Sprintf("swap=%dGB\n", config.Swap))
	}

	if config.SwapFile != "" {
		sb.WriteString(fmt.Sprintf("swapFile=%s\n", config.SwapFile))
	}
	if config.ProcessorCount > 0 {
		sb.WriteString(fmt.Sprintf("processors=%d\n", config.ProcessorCount))
	}
	if config.NetworkMode != "" {
		sb.WriteString(fmt.Sprintf("networkingMode=%s\n", config.NetworkMode))
	}

	// Booleans: Only write if different from default
	if !config.LocalhostForwarding {
		sb.WriteString(fmt.Sprintf("localhostForwarding=%v\n", config.LocalhostForwarding))
	}
	if !config.GuiApplications {
		sb.WriteString(fmt.Sprintf("guiApplications=%v\n", config.GuiApplications))
	}
	if config.DebugConsole {
		sb.WriteString(fmt.Sprintf("debugConsole=%v\n", config.DebugConsole))
	}

	if config.Kernel != "" {
		sb.WriteString(fmt.Sprintf("kernel=%s\n", config.Kernel))
	}
	if config.KernelModules != "" {
		sb.WriteString(fmt.Sprintf("kernelModules=%s\n", config.KernelModules))
	}
	if config.KernelCommandLine != "" {
		sb.WriteString(fmt.Sprintf("kernelCommandLine=%s\n", config.KernelCommandLine))
	}

	if config.SafeMode {
		sb.WriteString(fmt.Sprintf("safeMode=%v\n", config.SafeMode))
	}

	if config.MaxCrashDumpCount > 0 {
		sb.WriteString(fmt.Sprintf("maxCrashDumpCount=%d\n", config.MaxCrashDumpCount))
	}

	if !config.NestedVirtualization {
		sb.WriteString(fmt.Sprintf("nestedVirtualization=%v\n", config.NestedVirtualization))
	}

	if config.VmIdleTimeout > 0 {
		sb.WriteString(fmt.Sprintf("vmIdleTimeout=%d\n", config.VmIdleTimeout))
	}

	if !config.DnsProxy {
		sb.WriteString(fmt.Sprintf("dnsProxy=%v\n", config.DnsProxy))
	}

	sb.WriteString(fmt.Sprintf("pageReporting=%v\n", config.PageReporting))

	if !config.Firewall {
		sb.WriteString(fmt.Sprintf("firewall=%v\n", config.Firewall))
	}
	if config.DnsTunneling {
		sb.WriteString(fmt.Sprintf("dnsTunneling=%v\n", config.DnsTunneling))
	}
	if !config.AutoProxy {
		sb.WriteString(fmt.Sprintf("autoProxy=%v\n", config.AutoProxy))
	}

	sb.WriteString("\n[experimental]\n")

	if config.AutoMemoryReclaim != "" {
		sb.WriteString(fmt.Sprintf("autoMemoryReclaim=%s\n", config.AutoMemoryReclaim))
	}

	if config.SparseVhd {
		sb.WriteString(fmt.Sprintf("sparseVhd=%v\n", config.SparseVhd))
	}
	if !config.BestEffortDnsParsing {
		sb.WriteString(fmt.Sprintf("bestEffortDnsParsing=%v\n", config.BestEffortDnsParsing))
	}

	if config.DnsTunnelingIpAddress != "" {
		sb.WriteString(fmt.Sprintf("dnsTunnelingIpAddress=%s\n", config.DnsTunnelingIpAddress))
	}

	if config.InitialAutoProxyTimeout > 0 {
		sb.WriteString(fmt.Sprintf("initialAutoProxyTimeout=%d\n", config.InitialAutoProxyTimeout))
	}

	if !config.HostAddressLoopback {
		sb.WriteString(fmt.Sprintf("hostAddressLoopback=%v\n", config.HostAddressLoopback))
	}

	if config.UseWindowsDnsCli {
		sb.WriteString(fmt.Sprintf("useWindowsDnsCli=%v\n", config.UseWindowsDnsCli))
	}

	content := sb.String()

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
		MemoryLimit:         8,
		Swap:                0,
		SwapFile:            `C:\\wsl.swap`,
		ProcessorCount:      4,
		NetworkMode:         "mirrored",
		LocalhostForwarding: true,
		AutoMemoryReclaim:   "dropCache",
		SparseVhd:           true,
		DnsTunneling:        true,
		Firewall:            true,
		AutoProxy:           true,
		HostAddressLoopback: true,
		GuiApplications:     true,
		DebugConsole:        false, NestedVirtualization: true,
		VmIdleTimeout:           60000,
		PageReporting:           true,
		DnsTunnelingIpAddress:   "10.255.255.254",
		InitialAutoProxyTimeout: 1000,
		UseWindowsDnsCli:        false,
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
		case "useWindowsDnsCli":
			config.UseWindowsDnsCli = parseBool(value)
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
