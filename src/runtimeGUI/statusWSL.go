//go:build windows
// +build windows

package runtimeGUI

import (
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
				Status:  fields[1],             // 直接存入 "Running" 或 "Stopped"
				Version: fields[len(fields)-1], // 获取最后一列的版本号
			})
		}
	}

	return currentList, nil
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
