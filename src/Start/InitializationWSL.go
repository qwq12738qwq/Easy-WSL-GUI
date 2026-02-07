//go:build windows
// +build windows

package start

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

const (
	MB_OK        = 0x00000000
	MB_ICONERROR = 0x00000010
)

// 调用 Windows user32.dll 弹窗
func ShowNativeMessageBox(title, text string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	messageBox := user32.NewProc("MessageBoxW")
	if err := user32.Load(); err != nil {
		fmt.Println(text)
		return
	}

	t, _ := syscall.UTF16PtrFromString(text)
	c, _ := syscall.UTF16PtrFromString(title)

	messageBox.Call(
		0,
		uintptr(unsafe.Pointer(t)),
		uintptr(unsafe.Pointer(c)),
		MB_OK|MB_ICONERROR,
	)
}

// 详细检测wsl
func DetectWSL() error {
	_, err := exec.LookPath("wsl.exe")
	if err != nil {
		return fmt.Errorf("wsl.exe not found")
	}
	if err := exec.Command("wsl.exe", "--status").Run(); err != nil {
		return fmt.Errorf("wsl exists but not enabled")
	}
	return nil
}

// 输出错误
func ShowFatalError(err error) {
	msg := "WSL 环境检测失败。\n\n"

	switch {
	case strings.Contains(err.Error(), "not found"):
		msg += "未检测到 wsl.exe。\n请在 Windows 中启用 WSL。"
	case strings.Contains(err.Error(), "not enabled"):
		msg += "检测到 wsl.exe，但 WSL 功能未启用。\n请以管理员身份安装。"
	default:
		msg += err.Error()
	}

	ShowNativeMessageBox("环境错误", msg)
}

// EnsureWslConfigExists 检查并创建默认的 .wslconfig
func EnsureWslConfigExists() error {
	// 1. 获取用户主目录 (C:\Users\YourName)
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取主目录失败: %v", err)
	}

	configPath := filepath.Join(home, ".wslconfig")

	// 2. 检查文件是否已存在，如果存在则不覆盖（保护用户原有配置）
	if _, err := os.Stat(configPath); err == nil {
		return nil // 文件已存在，跳过
	}

	// 3. 定义默认配置内容
	defaultConfig := `[wsl2]
memory=8GB
processors=4
localhostForwarding=true
pageReporting=true
nestedVirtualization=true
defaultVhdSize=1099511627776
autoProxy=true
networkingMode=nat
dnsProxy=true
vmIdleTimeout=60000
maxCrashDumpCount=10
debugConsole=false
guiApplications=true
safeMode=false


[experimental]
autoMemoryReclaim=dropCache
sparseVhd=false
dnsTunneling=true
firewall=false
autoProxy=false
hostAddressLoopback=false
dnsTunnelingIpAddress=10.255.255.254
bestEffortDnsParsing=false
ignoredPorts=Null
		`

	// 4. 写入文件
	err = os.WriteFile(configPath, []byte(defaultConfig), 0644)
	if err != nil {
		return fmt.Errorf("创建 .wslconfig 失败: %v", err)
	}

	fmt.Println(".wslconfig 已创建在:", configPath)
	return nil
}
