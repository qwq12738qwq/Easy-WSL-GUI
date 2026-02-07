//go:build !windows
// +build !windows

package runtimeGUI

import "Golang-WSL-GUI/src/installWSL"

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

type Regedit_WSL struct {
	BasePath    string
	RunOOBE     bool
	VhdFileName string
}

// 获取WSL基本状态,如Running/Stopping
func GetWSLallStatus() ([]*List, error) { return nil, nil }

// 读取WSL在注册表的信息
func Seach_WSL_Regedit_Info(wsl_name string) (*Regedit_WSL, error) { return nil, nil }

// 读取默认配置用户
func GetDefaultUser(Info installWSL.WSLinfo) (string, error) { return "", nil }

// 提取WSL版本
func GetOnlyWslVersion(Info installWSL.WSLinfo) string { return "" }

// 正在运行发行版状态
func GetMetrics_Runtime(Info installWSL.WSLinfo) (*Metrics, error) { return nil, nil }
