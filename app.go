package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	run "runtime"
	"strings"
	"time"

	setting "Golang-WSL-GUI/src/Setting"
	start "Golang-WSL-GUI/src/Start"
	"Golang-WSL-GUI/src/installWSL"
	runtimeGUI "Golang-WSL-GUI/src/runtimeGUI"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type MigrationOptions struct {
	SourcePath string `json:"sourcePath"`
	TargetPath string `json:"targetPath"`
	DistroName string `json:"distroName"`
}

type App struct {
	ctx context.Context
}

var WSL_Regedit_Info = map[string]runtimeGUI.Regedit_WSL{}

var WSLinfoMap = map[string]installWSL.WSLinfo{}

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

func (a *App) Install_Bottom(name string, user string, pass string, ver string, path string, threadCount int) string {
	if path == "" {
		path = fmt.Sprintf(`C:\Users\%s\AppData\Local\Packages`, os.Getenv("USERNAME"))
	}

	Info := installWSL.WSLinfo{
		Linux_Version:   ver,
		Install_Path:    &installWSL.WSLpath{Path: path},
		Auth:            &installWSL.WSLAuth{User: user, Password: pass},
		DownloadThreads: &installWSL.WSLDownload{DownloadThreads: threadCount},
	}
	if err := installWSL.WSL2_Downloader(a.ctx, Info); err != nil {
		if err.Error() == "发行版存在,但未配置默认用户" {
			if installWSL.WSL2_Setting_User(a.ctx, Info) != nil {
				return err.Error()
			}
			runtime.EventsEmit(a.ctx, "wsl-output", "success")
			return "success"
		}
		return ""
	}
	time.Sleep(2 * time.Second)
	if err := installWSL.WSL2_Installer(a.ctx, Info); err != nil {
		return err.Error()
	}
	if err := installWSL.WSL2_Setting_User(a.ctx, Info); err != nil {
		return err.Error()
	}
	runtime.EventsEmit(a.ctx, "wsl-output", "success")
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
	infoptr, err := runtimeGUI.Seach_WSL_Regedit_Info(name)
	if err != nil {
		return err.Error(), err
	}
	WSL_Regedit_Info[name] = *infoptr
	return WSL_Regedit_Info[name].BasePath, nil
}

// 获取WSL发行版运行信息
func (a *App) GetMetrics(name string) runtimeGUI.Metrics {
	Info := installWSL.WSLinfo{
		Linux_Version:   name,
		Install_Path:    nil,
		Auth:            nil,
		DownloadThreads: nil,
	}
	ptr, err := runtimeGUI.GetMetrics_Runtime(Info)
	if err != nil {
		return runtimeGUI.Metrics{}
	}
	return *ptr
}

// UninstallDistro 卸载发行版
func (a *App) UninstallDistro(name string) error {
	Info := installWSL.WSLinfo{
		Linux_Version:   name,
		Install_Path:    nil,
		Auth:            nil,
		DownloadThreads: nil,
	}
	installWSL.Start_cmd(Info, "Shutdown")
	runtime.EventsEmit(a.ctx, "uninstall:progress", fmt.Sprintf("正在停止 %s 发行版", Info.Linux_Version))
	// 确保完全关闭
	for {
		listptr, _ := runtimeGUI.GetWSLallStatus()

		isRunning := false
		for _, distro := range listptr {
			if distro.Name == Info.Linux_Version {
				if distro.Status == "Running" {
					isRunning = true
				}
				break
			}
		}

		if !isRunning {
			break
		}

		time.Sleep(2 * time.Second)
	}
	runtime.EventsEmit(a.ctx, "uninstall:progress", fmt.Sprintf("开始卸载 %s 发行版", Info.Linux_Version))
	if err := installWSL.UninstallWSL(a.ctx, Info); err != nil {
		return err
	}
	runtime.EventsEmit(a.ctx, "uninstall:progress", "success")
	// 执行 wsl --unregister <name>
	return nil
}

// 检查管理员权限
func (a *App) CheckAdmin() bool {
	if run.GOOS == "windows" {
		cmd := exec.Command("net", "session")
		err := cmd.Run()
		return err == nil
	}
	// 非 Windows 系统根据逻辑返回
	return false
}

// 检查WSL功能组件
func (a *App) CheckWSL() bool {
	if run.GOOS == "windows" {
		err := start.DetectWSL()
		return err == nil
	}
	return false
}

// 迁移WSL系统函数
func (a *App) StartMigration(option MigrationOptions) error {
	Info := installWSL.WSLinfo{
		Linux_Version:   option.DistroName,
		Install_Path:    &installWSL.WSLpath{Path: option.TargetPath},
		Auth:            nil,
		DownloadThreads: nil,
	}

	user, err := runtimeGUI.GetDefaultUser(Info)
	if err != nil {
		return errors.New(user)
	}
	// 刷新Info
	Info = installWSL.WSLinfo{
		Linux_Version:   option.DistroName,
		Install_Path:    &installWSL.WSLpath{Path: option.TargetPath},
		Auth:            &installWSL.WSLAuth{User: user, Password: "0"},
		DownloadThreads: nil,
	}

	installWSL.Start_cmd(Info, "Stop")
	// 确保完全关闭
	for {
		listptr, _ := runtimeGUI.GetWSLallStatus()

		isRunning := false
		for _, distro := range listptr {
			if distro.Name == option.DistroName {
				if distro.Status == "Running" {
					isRunning = true
				}
				break
			}
		}

		if !isRunning {
			break
		}

		time.Sleep(2 * time.Second)
	}
	runtime.EventsEmit(a.ctx, "migration:progress", "迁移准备工作完成")
	time.Sleep(2 * time.Second)
	// 异步处理,防止堵塞
	go installWSL.MovingPathWSL(a.ctx, Info)
	// 已接收
	return nil
}

// 打开发行版内部目录
func (a *App) OpenDistroFolder(distroName string) error {
	Info := installWSL.WSLinfo{
		Linux_Version:   distroName,
		Install_Path:    nil,
		Auth:            nil,
		DownloadThreads: nil,
	}
	defaultUser, err := runtimeGUI.GetDefaultUser(Info)
	if err != nil {
		cmd := exec.Command("explorer.exe", fmt.Sprintf(`\\wsl$\%s\home`, distroName))
		return cmd.Start()
	}
	// Windows下调用explorer
	cmd := exec.Command("explorer.exe", fmt.Sprintf(`\\wsl$\%s\home\%s`, distroName, defaultUser))
	return cmd.Start()
}

// 启动发行版按钮
func (a *App) StartDistro(name string) {
	Info := installWSL.WSLinfo{
		Linux_Version:   name,
		Install_Path:    nil,
		Auth:            nil,
		DownloadThreads: nil,
	}
	installWSL.Start_cmd(Info, "Start")
}

// .wslconfig全局性能写入配置
func (a *App) SavePerformanceConfig(config setting.PerformanceConfig) error {
	if err := setting.Wriding_PerformanceConfig(config); err != nil {
		return err
	}
	Info := installWSL.WSLinfo{
		Linux_Version:   "",
		Install_Path:    nil,
		Auth:            nil,
		DownloadThreads: nil,
	}
	installWSL.Start_cmd(Info, "ShutdownAll")
	return nil
}

// .wslconfig全局性能读取配置
func (a *App) GetPerformanceConfig() setting.PerformanceConfig {
	return setting.Rading_PerformanceConfig()
}

// 获取 WSL 版本
func (a *App) GetWSLVersion() string {
	Info := installWSL.WSLinfo{
		Linux_Version:   "",
		Install_Path:    nil,
		Auth:            nil,
		DownloadThreads: nil,
	}
	return setting.GetOnlyWslVersion(Info)
}

// 显示详细版本信息
func (a *App) ShowWSLInfo() string {
	Info := installWSL.WSLinfo{
		Linux_Version:   "",
		Install_Path:    nil,
		Auth:            nil,
		DownloadThreads: nil,
	}

	line, err := installWSL.Start_cmd(Info, "Version")
	if err != nil {
		return err.Error()
	}

	// 定义 UTF-16LE 解码器
	decoder := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()
	// 尝试解码
	decoded, err := io.ReadAll(transform.NewReader(bytes.NewReader(line), decoder))

	var result string
	if err != nil || len(decoded) < 2 {
		result = string(line)
	} else {
		result = string(decoded)
	}

	// 去除不可见的 BOM 头、回车符 \r 和多余空格
	result = strings.ReplaceAll(result, "\uFEFF", "") // 去除 UTF-16 BOM
	result = strings.ReplaceAll(result, "\r", "")     // 统一换行符
	result = strings.TrimSpace(result)                // 去除首尾空白

	return result
}

// CheckAndUpdateWSL 检查并更新 WSL
func (a *App) CheckAndUpdateWSL() {
	// 简单对接建议:
	// 在后台打开终端执行 `wsl --update`，让用户看到原生更新进度
	// 或者使用 exec.Command 执行并实时推送进度给前端（较复杂）

	// 推荐方式 (简单有效): 打开外部终端执行更新
	// Windows 命令: start cmd /k "wsl --update"
	exec.Command("cmd", "/c", "start", "cmd", "/k", "wsl --update").Start()
}
