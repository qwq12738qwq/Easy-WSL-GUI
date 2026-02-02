//go:build windows
// +build windows

package runtimeGUI

import (
	"errors"
	"os/exec"
	"strings"
	"syscall"

	"Golang-WSL-GUI/src/installWSL"
)

type Metrics struct {
	CPU      string `json:"cpu"`      // 例如: "12.5%"
	MemUsed  string `json:"memUsed"`  // 例如: "1.2GB"
	MemTotal string `json:"memTotal"` // 例如: "8.0GB"
	Disk     string `json:"disk"`     // 例如: "15.4GB"
}

type List struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Version string `json:"version"`
}

var List_Slice []*List

func GetWSLallStatus() ([]*List, error) {
	list_name, err := ListWSLAll()
	if err != nil {
		return nil, err
	}
	// 遍历切片
	for _, name := range list_name {
		if name == "" {
			return nil, errors.New("未安装WSL发行版")
		}
		if name == "" {
			continue // 跳过空行，而不是直接报错返回
		}
		// 运行状态函数
		var Status string
		if WSLstatus(name) {
			Status = "Running"
		} else {
			Status = "Stopped"
		}
		cache := &List{
			Name:    name,
			Status:  Status,
			Version: "2",
		}
		List_Slice = append(List_Slice, cache)
	}
	if len(List_Slice) == 0 {
		// 这里可以选择返回空切片而不是报错，让前端显示"暂无数据"
		return []*List{}, nil
	}
	return List_Slice, nil
}

// 获取已存在发行版的名字
func ListWSLAll() ([]string, error) {
	cmd := exec.Command("wsl.exe", "--list", "--quiet")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(installWSL.Reduce_Unicode(out)), "\n")
	return lines, nil
}

// 检测发行版是否在运行
func WSLstatus(name string) bool {
	cmd := exec.Command("wsl.exe", "-d", name, "--", "true")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := cmd.Run()
	return err == nil
}

// GetMetrics 返回单个发行版的详细数据
func GetMetrics(name string) (*Metrics, error) {

	// 这里是你的逻辑：执行脚本或调用 API 获取数据
	return &Metrics{
		CPU:      "15%",
		MemUsed:  "2.1GB",
		MemTotal: "16.0GB",
		Disk:     "24.5GB",
	}, nil
}
