//go:build windows
// +build windows

package runtimeGUI

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	setting "Golang-WSL-GUI/src/Setting"
	"Golang-WSL-GUI/src/installWSL"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

type DiskBytes struct {
	Used  int64
	Total int64
}

type Regedit_WSL struct {
	BasePath    string
	RunOOBE     bool
	VhdFileName string
}

type Metrics struct {
	CPU        string `json:"cpu"`        // 例如: "15%"
	MemUsed    string `json:"memUsed"`    // 例如: "1.2 GB"
	MemTotal   string `json:"memTotal"`   // 例如: "8.0 GB"
	UsedBytes  int64  `json:"usedBytes"`  // 磁盘已用字节 (用于前端 formatBytes 计算)
	TotalBytes int64  `json:"totalBytes"` // 磁盘总字节
	Disk       string `json:"disk"`       // 磁盘百分比字符串 (兼容旧逻辑)
}

type List struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Version string `json:"version"`
}

var List_Slice []*List

func GetWSLallStatus() ([]*List, error) {
	currentList, err := WSLsrtatus()
	if err != nil {
		return nil, err
	}
	List_Slice = currentList
	if len(List_Slice) == 0 {
		// 这里可以选择返回空切片而不是报错，让前端显示"暂无数据"
		return []*List{}, nil
	}
	return List_Slice, nil
}

// 检测发行版是否在运行
func WSLsrtatus() ([]*List, error) {
	cmd := exec.Command("wsl.exe", "--list", "--verbose")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	content := installWSL.Reduce_Unicode(out)
	lines := strings.Split(content, "\n")

	var currentList []*List

	for _, line := range lines[1:] { // 跳过标题行
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSpace(line)

		fields := strings.Fields(line)
		if len(fields) >= 3 {
			currentList = append(currentList, &List{
				Name:    fields[0],
				Status:  fields[1],             // 存入 "Running" 或 "Stopped"
				Version: fields[len(fields)-1], // 获取版本号
			})
		}
	}

	return currentList, nil
}

// GetMetrics 返回单个发行版的详细数据
func GetMetrics_Runtime(Info installWSL.WSLinfo) (*Metrics, error) {
	memtotal := setting.Rading_PerformanceConfig()
	diskptr := getFileSize(Info)
	return &Metrics{
		CPU:        fmt.Sprintf(`%.1f%%`, GetCpuUsageSingleShot(Info)),
		MemUsed:    fmt.Sprintf(`%.1fGB`, GetDistroMemUsage(Info)),
		MemTotal:   fmt.Sprintf(`%dGB`, memtotal.MemoryLimit),
		UsedBytes:  diskptr.Used,
		TotalBytes: diskptr.Total,
	}, nil
}

// WSL占用 / 剩余总空间
func getFileSize(Info installWSL.WSLinfo) *DiskBytes {
	regeditptr, _ := Seach_WSL_Regedit_Info(Info.Linux_Version)

	filepath := fmt.Sprintf(`%s\%s`, regeditptr.BasePath, regeditptr.VhdFileName)
	fileInfo, _ := os.Stat(filepath)
	use := fileInfo.Size()

	var total uint64
	pathPtr, _ := windows.UTF16PtrFromString(regeditptr.BasePath)
	windows.GetDiskFreeSpaceEx(pathPtr, nil, &total, nil)
	return &DiskBytes{
		Used:  use,
		Total: int64(total),
	}
}

// 内存占用
func GetDistroMemUsage(Info installWSL.WSLinfo) float64 {
	// memory.current 记录了当前该发行版所在控制组消耗的内存字节数
	// 这是 Linux 内部最底层的统计方式
	out, err := installWSL.Start_cmd(Info, "useMem")
	if err != nil {
		return 0.0
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return 0.0
	}

	// 提取数字部分
	var total, avail float64
	fmt.Sscanf(lines[0], "MemTotal: %f", &total)
	fmt.Sscanf(lines[1], "MemAvailable: %f", &avail)

	if total == 0 {
		return 0.0
	}

	// 计算已用空间 (KB -> GB)
	usedGB := (total - avail) / 1024.0 / 1024.0
	return usedGB
}

// CPU占用
func GetCpuUsageSingleShot(Info installWSL.WSLinfo) float64 {
	// 定义内部获取快照的辅助闭包
	getSnap := func() (idle, total uint64) {
		// 使用 sh -c 配合 head 确保读取最快
		out, err := installWSL.Start_cmd(Info, "CPUpercent")
		if err != nil {
			return 0, 0
		}

		fields := strings.Fields(installWSL.Reduce_Unicode(out))
		if len(fields) < 5 {
			return 0, 0
		}

		for i := 1; i < len(fields); i++ {
			val, _ := strconv.ParseUint(fields[i], 10, 64)
			total += val
			if i == 4 { // 第 4 个索引位 (从 1 开始计) 是 idle
				idle = val
			}
		}
		return
	}

	// 第一次采样
	i1, t1 := getSnap()

	// 睡一秒
	time.Sleep(1 * time.Second)

	// 第二次采样
	i2, t2 := getSnap()

	// 计算差值
	totalDiff := t2 - t1
	idleDiff := i2 - i1

	return float64(totalDiff-idleDiff) / float64(totalDiff) * 100
}

func Seach_WSL_Regedit_Info(wsl_name string) (*Regedit_WSL, error) {
	rootPath := `Software\Microsoft\Windows\CurrentVersion\Lxss`

	// 打开 Lxss
	k, err := registry.OpenKey(registry.CURRENT_USER, rootPath, registry.READ)
	if err != nil {
		return nil, errors.New("无法打开注册表,检查权限")
	}
	defer k.Close()

	// 2. 获取所有GUID
	subKeys, err := k.ReadSubKeyNames(-1)
	if err != nil {
		return nil, err
	}

	for _, guid := range subKeys {
		// 逐个打开GUID比对
		skPath := rootPath + `\` + guid
		sk, err := registry.OpenKey(registry.CURRENT_USER, skPath, registry.QUERY_VALUE)
		if err != nil {
			continue
		}

		name, _, _ := sk.GetStringValue("DistributionName")
		if name == wsl_name {
			basePath, _, _ := sk.GetStringValue("BasePath")
			vhdFile, _, _ := sk.GetStringValue("VhdFileName")
			oobeVal, _, _ := sk.GetIntegerValue("RunOOBE")
			sk.Close()
			return &Regedit_WSL{
				BasePath:    basePath,
				VhdFileName: vhdFile,
				RunOOBE:     oobeVal != 0,
			}, nil

		}
		sk.Close()
	}

	return nil, errors.New("在注册表未找到发行版")
}

func GetDefaultUser(Info installWSL.WSLinfo) (string, error) {
	line, err := installWSL.Start_cmd(Info, "SeachUser")
	if err != nil {
		return installWSL.Reduce_Unicode(line), err
	}

	cleanStr := installWSL.Reduce_Unicode(line)

	// 解析逻辑
	lines := strings.Split(cleanStr, "\n")
	var isUserSection bool

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 匹配段落标签 [user]
		if strings.EqualFold(line, "[user]") {
			isUserSection = true
			continue
		}

		// 如果在 [user] 段下找到了 default=xxx
		if isUserSection {
			// 如果遇到下一个段落 [xxx]，说明 user 段结束了
			if strings.HasPrefix(line, "[") {
				break
			}

			if strings.Contains(line, "=") {
				parts := strings.SplitN(line, "=", 2)
				if strings.EqualFold(strings.TrimSpace(parts[0]), "default") {
					return strings.TrimSpace(parts[1]), nil
				}
			}
		}
	}

	return "未发现配置默认用户", errors.New("未发现配置默认用户")
}
