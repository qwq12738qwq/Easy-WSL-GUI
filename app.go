package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	run "runtime"
	"strings"
	"time"

	global "Golang-WSL-GUI/src/Global"
	"Golang-WSL-GUI/src/Logger"
	networkGUI "Golang-WSL-GUI/src/Network"
	setting "Golang-WSL-GUI/src/Setting"
	start "Golang-WSL-GUI/src/Start"
	"Golang-WSL-GUI/src/installWSL"
	runtimeGUI "Golang-WSL-GUI/src/runtimeGUI"

	"github.com/shirou/gopsutil/v3/mem"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	CurrentAppVersion = "v1.0.0-Beta"
	UpdateJsonUrl     = "https://gitee.com/MrLi114514/wsl_package/raw/master/Update.json"
)

type MigrationOptions struct {
	SourcePath string `json:"sourcePath"`
	TargetPath string `json:"targetPath"`
	DistroName string `json:"distroName"`
}

type SystemSpecs struct {
	TotalMemoryGB int `json:"totalMemoryGB"`
	LogicalCores  int `json:"logicalCores"`
}

type UpdateInfo struct {
	Version     string `json:"version"`
	UpdateLog   string `json:"updateLog"`
	ReleaseDate string `json:"releaseDate"`
	Url         string `json:"url"`
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

	//	待测试
	// 日志线程系统
	go func() {
		appLogger, err := Logger.NewLogger()
		if err != nil {
			// 日志函数运行失败时,防止主程序崩溃
			appLogger = nil
		} else {
			global.AppLogger = appLogger
			// 启动日志推送服务
			a.PushLogsToFrontend()
			global.AppLogger.Info("日志系统启动成功")
		}
	}()
	//	待测试

	// 延迟启动更新检测线程
	go func() {
		time.Sleep(10 * time.Second)
		a.TriggerUpdateAlert()
	}()

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

// 安装函数
func (a *App) Install_Bottom(
	name string,
	user string,
	pass string,
	ver string,
	path string,
	threadCount int,
	downloadUrl string,
	sha256 string,
) string {
	if global.AppLogger != nil {
		global.AppLogger.Info("开始安装 %s 发行版,用户: %s, 版本: %s", ver, user, ver)
	}
	if path == "" {
		path = fmt.Sprintf(`C:\Users\%s\AppData\Local\Packages`, os.Getenv("USERNAME"))
	}

	Info := installWSL.WSLinfo{
		Linux_Version:   ver,
		Install_Path:    &installWSL.WSLpath{Path: path},
		Auth:            &installWSL.WSLAuth{User: user, Password: pass},
		DownloadThreads: &installWSL.WSLDownload{DownloadThreads: threadCount},
		DownloadInfo:    &installWSL.Download_WSL{URL: downloadUrl, Sha256: sha256},
	}
	if err := installWSL.WSL2_Downloader(a.ctx, Info); err != nil {
		if global.AppLogger != nil {
			global.AppLogger.Error("Install_Bottom: 下载 %s 发行版失败: %s", ver, err.Error())
		}
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
		if global.AppLogger != nil {
			global.AppLogger.Error("Install_Bottom: 安装 %s 发行版失败: %s", ver, err.Error())
		}
		return err.Error()
	}
	if err := installWSL.WSL2_Setting_User(a.ctx, Info); err != nil {
		if global.AppLogger != nil {
			global.AppLogger.Error("Install_Bottom: 配置 %s 发行版用户失败: %s", ver, err.Error())
		}
		return err.Error()
	}

	runtime.EventsEmit(a.ctx, "wsl-output", fmt.Sprintf("正在重启 %s 发行版", Info.Linux_Version))
	if global.AppLogger != nil {
		global.AppLogger.Info("Install_Bottom: 正在重启 %s 发行版", Info.Linux_Version)
	}
	if a.StopDistro(name) != nil {
		runtime.EventsEmit(a.ctx, "wsl-output", fmt.Sprintf("重启 %s 发行版出错,请手动重启发行版完成安装", Info.Linux_Version))
		if global.AppLogger != nil {
			global.AppLogger.Warning("Install_Bottom: 重启 %s 发行版出错,请手动重启", Info.Linux_Version)
		}
		time.Sleep(5 * time.Second)
	}
	runtime.EventsEmit(a.ctx, "wsl-output", "success")
	if global.AppLogger != nil {
		global.AppLogger.Info("Install_Bottom: %s 发行版安装成功", Info.Linux_Version)
	}
	return "success"
}

// 获取WSL基本信息
func (a *App) GetDistroStats() ([]*runtimeGUI.List, error) {
	Info, err := runtimeGUI.GetWSLallStatus()
	if err != nil {
		if global.AppLogger != nil {
			global.AppLogger.Error("GetDistroStats: 获取WSL发行版列表失败: %s", err.Error())
		}
		return nil, err
	}
	if global.AppLogger != nil {
		global.AppLogger.Info("GetDistroStats: 成功获取 %d 个发行版", len(Info))
	}
	return Info, nil
}

// GetInstallPath 获取发行版安装路径
func (a *App) GetPath(name string) (string, error) {
	infoptr, err := runtimeGUI.Seach_WSL_Regedit_Info(name)
	if err != nil {
		if global.AppLogger != nil {
			global.AppLogger.Error("GetPath: 获取发行版 %s 路径失败: %s", name, err.Error())
		}
		return err.Error(), err
	}
	WSL_Regedit_Info[name] = *infoptr
	if global.AppLogger != nil {
		global.AppLogger.Info("GetPath: 成功获取发行版 %s 安装路径: %s", name, WSL_Regedit_Info[name].BasePath)
	}
	return WSL_Regedit_Info[name].BasePath, nil
}

// 获取WSL发行版运行信息
func (a *App) GetMetrics(name string) runtimeGUI.Metrics {
	Info := installWSL.WSLinfo{
		Linux_Version: name,
	}
	ptr, err := runtimeGUI.GetMetrics_Runtime(Info)
	if err != nil {
		if global.AppLogger != nil {
			global.AppLogger.Error("GetMetrics: 获取发行版 %s 运行信息失败: %s", name, err.Error())
		}
		return runtimeGUI.Metrics{}
	}
	return *ptr
}

// UninstallDistro 卸载发行版
func (a *App) UninstallDistro(name string) error {
	if global.AppLogger != nil {
		global.AppLogger.Info("UninstallDistro: 开始卸载发行版 %s", name)
	}
	Info := installWSL.WSLinfo{
		Linux_Version: name,
	}

	runtime.EventsEmit(a.ctx, "uninstall:progress", fmt.Sprintf("正在停止 %s 发行版", Info.Linux_Version))
	if err := a.StopDistro(name); err != nil {
		if global.AppLogger != nil {
			global.AppLogger.Error("UninstallDistro: 停止发行版 %s 失败: %s", name, err.Error())
		}
		return err
	}
	runtime.EventsEmit(a.ctx, "uninstall:progress", fmt.Sprintf("开始卸载 %s 发行版", Info.Linux_Version))
	if err := installWSL.UninstallWSL(a.ctx, Info); err != nil {
		if global.AppLogger != nil {
			global.AppLogger.Error("UninstallDistro: 卸载发行版 %s 失败: %s", name, err.Error())
		}
		return err
	}
	runtime.EventsEmit(a.ctx, "uninstall:progress", "success")
	if global.AppLogger != nil {
		global.AppLogger.Info("UninstallDistro: 发行版 %s 卸载成功", name)
	}
	return nil
}

// 检查管理员权限
func (a *App) CheckAdmin() bool {
	if run.GOOS == "windows" {
		return start.CheckAdmin() == nil
	}
	// 非 Windows 系统根据逻辑返回
	return false
}

// 检查WSL功能组件
func (a *App) CheckWSL() bool {
	if run.GOOS == "windows" {
		err := start.DetectWSL()
		if err != nil {
			if global.AppLogger != nil {
				global.AppLogger.Warning("CheckWSL: WSL功能检测失败: %s", err.Error())
			}
			return false
		}
		if global.AppLogger != nil {
			global.AppLogger.Info("CheckWSL: WSL功能检测正常")
		}
		return true
	}
	return false
}

// 迁移WSL系统函数
func (a *App) StartMigration(option MigrationOptions) error {
	global.AppLogger.Info("StartMigration: 启动迁移发行版 %s, 从 %s 迁移到 %s", option.DistroName, option.SourcePath, option.TargetPath)
	Info := installWSL.WSLinfo{
		Linux_Version: option.DistroName,
		Install_Path:  &installWSL.WSLpath{Path: option.TargetPath},
	}
	// 先读取默认用户配置
	// 预防先操作系统出现问题
	runtime.EventsEmit(a.ctx, "migration:progress", "保存用户配置中......")
	global.AppLogger.Info("StartMigration: 正在保存用户配置")
	time.Sleep(2 * time.Second)
	var user_cache string
	user, err := runtimeGUI.GetDefaultUser(Info)
	if err != nil {
		global.AppLogger.Warning("StartMigration: 在wsl.conf中找不到默认用户,将寻找发行版内部用户组")
		runtime.EventsEmit(a.ctx, "migration:progress", "在wsl.conf中找不到默认用户,将寻找发行版内部用户组")
		time.Sleep(2 * time.Second)
		// 获取WSL发行版内部用户,迁移系统时选择默认内部用户启动
		list, err := installWSL.GetWSLUsers(option.DistroName)
		if err != nil {
			runtime.EventsEmit(a.ctx, "migration:done", map[string]interface{}{
				"status": "failed",
				"error":  "发行版中无用户",
			})
			global.AppLogger.Error("StartMigration: 配置用户错误,错误信息: %s", err.Error())
			return err
		}
		global.AppLogger.Info("StartMigration: 正在等待用户选择用户")
		runtime.EventsEmit(a.ctx, "migration:users", list)
		// 堵塞机制,选择完成之后才继续往下执行
		userChan := make(chan string)
		runtime.EventsOn(a.ctx, "migration:select-user", func(optionalData ...interface{}) {
			if len(optionalData) > 0 {
				// 将 interface{} 转换为 string
				username, ok := optionalData[0].(string)
				if ok {
					userChan <- username
				}
			}
		})
		// 锁进程
		user_cache = <-userChan
	} else {
		user_cache = user
	}

	// 刷新Info
	Info = installWSL.WSLinfo{
		Linux_Version: option.DistroName,
		Install_Path:  &installWSL.WSLpath{Path: option.TargetPath},
		Auth:          &installWSL.WSLAuth{User: user_cache, Password: "0"},
	}
	global.AppLogger.Info("StartMigration: 正在关闭发行版 %s", Info.Linux_Version)
	time.Sleep(2 * time.Second)
	runtime.EventsEmit(a.ctx, "migration:progress", "关闭发行版中......")
	if err := a.StopDistro(Info.Linux_Version); err != nil {
		global.AppLogger.Error("StartMigration: 关闭发行版 %s 失败: %s", Info.Linux_Version, err.Error())
	}
	// if err := a.StopDistro(option.DistroName); err != nil {
	// 	a.appLogger.Error("关闭发行版失败: %s", err.Error())
	// 	return err
	// }
	global.AppLogger.Info("StartMigration: 发行版 %s 已关闭", Info.Linux_Version)
	runtime.EventsEmit(a.ctx, "migration:progress", "准备工作完成")
	time.Sleep(2 * time.Second)
	// 异步处理,防止堵塞
	go installWSL.MovingPathWSL(a.ctx, Info)
	// 已接收
	return nil
}

// 打开发行版内部目录
func (a *App) OpenDistroFolder(distroName string) error {
	if global.AppLogger != nil {
		global.AppLogger.Info("OpenDistroFolder: 正在打开发行版 %s 的文件夹", distroName)
	}
	Info := installWSL.WSLinfo{
		Linux_Version: distroName,
	}
	defaultUser, err := runtimeGUI.GetDefaultUser(Info)
	if err != nil {
		cmd := exec.Command("explorer.exe", fmt.Sprintf(`\\wsl$\%s\home`, distroName))
		if err := cmd.Start(); err != nil {
			if global.AppLogger != nil {
				global.AppLogger.Error("OpenDistroFolder: 打开发行版 %s 文件夹失败: %s", distroName, err.Error())
			}
			return err
		}
		return nil
	}
	// Windows下调用explorer
	cmd := exec.Command("explorer.exe", fmt.Sprintf(`\\wsl$\%s\home\%s`, distroName, defaultUser))
	if err := cmd.Start(); err != nil {
		if global.AppLogger != nil {
			global.AppLogger.Error("OpenDistroFolder: 打开发行版 %s 文件夹失败: %s", distroName, err.Error())
		}
		return err
	}
	if global.AppLogger != nil {
		global.AppLogger.Info("OpenDistroFolder: 已打开发行版 %s 文件夹", distroName)
	}
	return nil
}

// 启动发行版按钮
func (a *App) StartDistro(name string) {
	if global.AppLogger != nil {
		global.AppLogger.Info("正在启动发行版: %s", name)
	}

	Info := installWSL.WSLinfo{
		Linux_Version: name,
	}
	installWSL.Start_cmd(Info, "Start")

	if global.AppLogger != nil {
		global.AppLogger.Info("发行版 %s 启动命令已发送", name)
	}
}

// .wslconfig全局性能写入配置
func (a *App) SavePerformanceConfig(config setting.PerformanceConfig) error {
	if global.AppLogger != nil {
		global.AppLogger.Info("SavePerformanceConfig: 正在保存性能配置, 内存限制: %dGB, CPU核心数: %d", config.MemoryLimit, config.ProcessorCount)
	}
	if err := setting.Wriding_PerformanceConfig(config); err != nil {
		if global.AppLogger != nil {
			global.AppLogger.Error("SavePerformanceConfig: 保存性能配置失败: %s", err.Error())
		}
		return err
	}
	Info := installWSL.WSLinfo{
		Linux_Version: "",
	}
	installWSL.Start_cmd(Info, "ShutdownAll")
	if global.AppLogger != nil {
		global.AppLogger.Info("SavePerformanceConfig: 性能配置保存成功,已关闭所有WSL发行版")
	}
	return nil
}

// .wslconfig全局性能读取配置
func (a *App) GetPerformanceConfig() setting.PerformanceConfig {
	config := setting.Rading_PerformanceConfig()
	if global.AppLogger != nil {
		global.AppLogger.Info("GetPerformanceConfig: 已读取性能配置, 内存限制: %dGB, CPU核心数: %d", config.MemoryLimit, config.ProcessorCount)
	}
	return config
}

// GetLogs 读取日志文件内容（历史日志）
// 用于前端初始化时加载历史记录
func (a *App) GetLogs() ([]string, error) {
	if global.AppLogger == nil {
		return []string{}, nil
	}

	// 这里我们复用之前 appLogs 逻辑会更好，
	// 但为了解耦，我们直接读取文件
	// 由于 Logger 结构体没有暴露读取方法，我们在这里简单实现一次读取
	// 或者更简单：我们让前端调用 GetAppLogs（内存日志），这里返回文件路径或内容

	// 建议：前端先调用 GetAppLogs 获取内存日志
	// 然后如果需要完整历史，再调用这个

	// 为了简单，我们这里直接读取文件
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = "C:/Users/Public/AppData/Local"
	}
	logPath := fmt.Sprintf("%s/Easy-WSL-GUI/logs/app.log", localAppData)

	data, err := os.ReadFile(logPath)
	if err != nil {
		return []string{}, err
	}

	lines := strings.Split(string(data), "\n")
	// 只返回最后 500 行
	start := 0
	if len(lines) > 500 {
		start = len(lines) - 500
	}
	return lines[start:], nil
}

// PushLogsToFrontend 推送日志到前端
// 在 startup 中调用
func (a *App) PushLogsToFrontend() {
	if global.AppLogger == nil || a.ctx == nil {
		return
	}

	ch := global.AppLogger.GetChan()
	go func() {
		// 使用 for range 监听通道，更优雅
		for msg := range ch {
			// 使用 Wails Event 系统推送到前端
			runtime.EventsEmit(a.ctx, "log-event", msg)
		}
	}()
}

// 获取 WSL 版本
func (a *App) GetWSLVersion() string {
	Info := installWSL.WSLinfo{
		Linux_Version: "",
	}
	version := setting.GetOnlyWslVersion(Info)
	if global.AppLogger != nil {
		global.AppLogger.Info("GetWSLVersion: WSL版本: %s", version)
	}
	return version
}

// 显示详细版本信息
func (a *App) ShowWSLInfo() string {
	Info := installWSL.WSLinfo{
		Linux_Version: "",
	}

	line, err := installWSL.Start_cmd(Info, "Version")
	if err != nil {
		if global.AppLogger != nil {
			global.AppLogger.Error("ShowWSLInfo: 获取WSL版本信息失败: %s", err.Error())
		}
		return err.Error()
	}

	result := installWSL.Reduce_Unicode(line)
	if global.AppLogger != nil {
		global.AppLogger.Info("ShowWSLInfo: WSL版本信息: %s", result)
	}
	return result
}

// 停止发行版
func (a *App) StopDistro(name string) error {
	if global.AppLogger != nil {
		global.AppLogger.Info("StopDistro: 正在停止发行版 %s", name)
	}
	// 错误重试计数器
	var i uint8

	Info := installWSL.WSLinfo{
		Linux_Version: name,
	}
	// 确保完全关闭
	for {
		i++

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
		installWSL.Start_cmd(Info, "Stop")
		time.Sleep(3 * time.Second)
		if i >= 10 {
			if global.AppLogger != nil {
				global.AppLogger.Error("StopDistro: 暂停发行版 %s 出错,已重试%d次", name, i)
			}
			return errors.New("暂停发行版出错,请手动暂停发行版")
		}
	}
	if global.AppLogger != nil {
		global.AppLogger.Info("StopDistro: 发行版 %s 已停止", name)
	}
	return nil
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

// GetAPTSource 获取当前发行版的 APT 软件源
func (a *App) GetAPTSource(distroName string) string {
	if global.AppLogger != nil {
		global.AppLogger.Info("GetAPTSource: 正在获取发行版 %s 的APT软件源", distroName)
	}
	Info := installWSL.WSLinfo{
		Linux_Version: distroName,
	}
	version, err := setting.CheckCurrentAptSource(a.ctx, Info)
	if err != nil {
		if global.AppLogger != nil {
			global.AppLogger.Error("GetAPTSource: 获取发行版 %s 软件源失败: %s", distroName, err.Error())
		}
		return err.Error()
	}
	if global.AppLogger != nil {
		global.AppLogger.Info("GetAPTSource: 发行版 %s 当前软件源: %s", distroName, version)
	}
	return version
}

// 换源函数
func (a *App) ChangeAPTSource(distroName string, source string) string {
	if global.AppLogger != nil {
		global.AppLogger.Info("ChangeAPTSource: 正在为发行版 %s 切换软件源为 %s", distroName, source)
	}
	Info := installWSL.WSLinfo{
		Linux_Version: distroName,
	}
	result, err := setting.ChangeDistroSource(a.ctx, Info, source)
	if err != nil {
		if global.AppLogger != nil {
			global.AppLogger.Error("ChangeAPTSource: 发行版 %s 换源失败: %s", distroName, err.Error())
		}
		return err.Error()
	}
	if global.AppLogger != nil {
		global.AppLogger.Info("ChangeAPTSource: 发行版 %s 换源成功", distroName)
	}
	return result
}

// 启用 WSL 功能组件
func (a *App) EnableWSLFeature() {
	cmdStr := `dism.exe /online /enable-feature /featurename:Microsoft-Windows-Subsystem-Linux /all /norestart; dism.exe /online /enable-feature /featurename:VirtualMachinePlatform /all /norestart; wsl --update; echo "Done. Please Restart Computer."; pause`

	exec.Command("powershell", "start-process", "powershell", "-verb", "runas", "-argumentlist", fmt.Sprintf("'-c \"%s\"'", cmdStr)).Start()
}

// 获取发行版下载信息Json
func (a *App) GetDistroList() ([]networkGUI.DistroItem, error) {
	if global.AppLogger != nil {
		global.AppLogger.Info("GetDistroList: 正在获取发行版列表")
	}
	distros, err := networkGUI.GetDistroList()
	if err != nil {
		if global.AppLogger != nil {
			global.AppLogger.Error("GetDistroList: 获取发行版列表失败: %s", err.Error())
		}
		return nil, err
	}
	if global.AppLogger != nil {
		global.AppLogger.Info("GetDistroList: 成功获取 %d 个发行版", len(distros))
	}
	return distros, nil
}

// 获取最大内存使用 && CPU线程数
func (a *App) GetSystemSpecs() SystemSpecs {
	v, err := mem.VirtualMemory()
	if err != nil {
		if global.AppLogger != nil {
			global.AppLogger.Error("GetSystemSpecs: 获取系统规格失败: %s", err.Error())
		}
		return SystemSpecs{}
	}

	totalGB := int(v.Total / 1024 / 1024 / 1024)
	if global.AppLogger != nil {
		global.AppLogger.Info("GetSystemSpecs: 系统内存: %dGB, CPU逻辑核心数: %d", totalGB, run.NumCPU())
	}
	return SystemSpecs{

		TotalMemoryGB: totalGB,
		LogicalCores:  run.NumCPU(),
	}
}

// GetAppVersion 获取当前应用版本
func (a *App) GetAppVersion() string {
	return CurrentAppVersion
}

func (a *App) TriggerUpdateAlert() {
	info, err := networkGUI.CheckForUpdate(CurrentAppVersion, UpdateJsonUrl)
	if err != nil {
		return
	}

	if info != nil {
		updateData := UpdateInfo{
			Version:     info.Version,
			UpdateLog:   info.UpdateLog,
			ReleaseDate: info.ReleaseDate,
			Url:         info.Url,
		}
		// 发送更新信息
		runtime.EventsEmit(a.ctx, "new-version", updateData)
	}
}

// 获取发行版内部安装包
func (a *App) GetInstalledPackages(distroName string) ([]runtimeGUI.SoftwarePackage, error) {
	if global.AppLogger != nil {
		global.AppLogger.Info("正在获取发行版 %s 的软件包列表...", distroName)
	}

	package_json, err := runtimeGUI.GetInstalledPackages(distroName)
	if err != nil {
		if global.AppLogger != nil {
			global.AppLogger.Info("获取软件包列表失败: %v", err)
		}
		return nil, err
	}

	if global.AppLogger != nil {
		global.AppLogger.Info("成功获取 %d 个软件包", len(package_json))
	}
	return package_json, nil
}

// UninstallPackage 卸载指定的软件包
// 前端调用示例: UninstallPackage(distroName string, packageName string)
// 返回: error (如果卸载失败返回错误)
func (a *App) UninstallPackage(distroName string, packageName string) error {
	// 示例: wsl -d {distroName} -- sudo apt remove -y {packageName}
	fmt.Printf("Uninstalling package %s from %s\n", packageName, distroName)
	return nil
}

func (a *App) CheckDockerInstalled(distroName string) bool { return false }

func (a *App) SetDockerRegistryMirror(distroName string, mirrorUrl string) error { return nil }
