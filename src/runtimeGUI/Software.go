//go:build windows
// +build windows

package runtimeGUI

import (
	global "Golang-WSL-GUI/src/Global"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type SoftwarePackage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Source  string `json:"source"`
}

// 读取已安装软件包
func GetInstalledPackages(distroName string) ([]SoftwarePackage, error) {
	filePath := fmt.Sprintf(`\\wsl$\%s\var\lib\dpkg\status`, distroName)

	// if _, err := os.Stat(filePath); os.IsNotExist(err) {
	// 	return []SoftwarePackage{}, nil
	// }
	file, err := os.Open(filePath)
	if err != nil {
		// 更换成Powershell,使用cat /var/lib/dpkg/status读取
		return nil, errors.New("打开WSL文件失败")
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	global.AppLogger.Info("已读取到软件列表", string(content))

	var packages []SoftwarePackage
	var currentPkg SoftwarePackage

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// 空格结束上一个包读取
		if strings.TrimSpace(line) == "" {
			if currentPkg.Name != "" {
				currentPkg.Source = "dpkg/status" // status文件不含源信息，统一标记
				packages = append(packages, currentPkg)
				currentPkg = SoftwarePackage{} // 重置
			}
			continue
		}

		// 忽略以空格开头的行
		// 比如 Description 内部的列表项，这些行不需要拆分
		if strings.HasPrefix(line, " ") {
			continue
		}

		// 拆分 Key: Value
		if strings.HasPrefix(line, "Package: ") {
			currentPkg.Name = strings.TrimPrefix(line, "Package: ")
		} else if strings.HasPrefix(line, "Version: ") {
			currentPkg.Version = strings.TrimPrefix(line, "Version: ")
		}
	}

	// 处理文件最后没有空行的情况
	if currentPkg.Name != "" {
		currentPkg.Source = "dpkg/status"
		packages = append(packages, currentPkg)
	}

	// 转换成 JSON 字符串
	jsonData, _ := json.MarshalIndent(packages, "", "    ")
	fmt.Println(string(jsonData))

	return packages, nil
}
