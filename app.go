package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"Golang-WSL-GUI/src/installWSL"
	"Golang-WSL-GUI/src/runtimeGUI"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SelectDirectory 弹出系统原生目录选择框
func (a *App) SelectDirectory() string {
	// 调用 Wails 运行时打开目录选择对话框
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Ciallo(∠・ω&lt; )⌒☆",
	})

	if err != nil {
		fmt.Printf("选择目录时出错: %v\n", err)
		return ""
	}

	return path
}

func (a *App) Install_Bottom(name string, user string, pass string, ver string, path string) string {
	if path == "" {
		path = fmt.Sprintf(`C:\Users\%s\AppData\Local\Packages`, os.Getenv("USERNAME"))
	}

	Info := installWSL.WSLinfo{
		Linux_Version: ver,
		Install_Path:  path,
		Password:      pass,
		User:          user,
	}
	if err := installWSL.WSL2_Downloader(a.ctx, Info); err != nil {
		if err.Error() == "发行版存在,但未配置默认用户" {
			installWSL.WSL2_Setting_User(a.ctx, Info)
		}
		return ""
	}
	time.Sleep(2 * time.Second)
	installWSL.WSL2_Installer(a.ctx, Info)
	runtime.EventsEmit(a.ctx, "wsl-output", fmt.Sprintf("发行版 %s 安装成功", ver))
	return "success"
}

// 获取WSL基本信息
func (a *App) GetDistroStats() ([]*runtimeGUI.List, error) {
	Info, err := runtimeGUI.GetWSLallStatus()
	if err != nil {
		return nil, err
	}
	return Info, nil
}

// GetInstallPath 获取发行版安装路径 (只在前端初次加载时调用)
func (a *App) GetPath(name string) (string, error) {
	// 逻辑示例：通过注册表查询或 wsl -l -v 解析
	// 快捷方式：cmd /c "wsl -d <name> pwd" 并不准确，
	// 建议查询注册表: HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Lxss
	return "C:\\Users\\Admin\\AppData\\Local\\Packages\\...", nil
}

// 获取WSL发行版运行信息
func (a *App) GetMetrics(name string) {}

// UninstallDistro 卸载发行版
func (a *App) UninstallDistro(name string) error {
	installWSL.UninstallWSL(name)
	// 执行 wsl --unregister <name>
	return nil
}
