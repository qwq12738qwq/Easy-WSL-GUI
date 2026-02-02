//go:build windows
// +build windows

package start

import (
	"fmt"
	"os/exec"
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
