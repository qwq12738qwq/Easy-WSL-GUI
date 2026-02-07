//go:build windows
// +build windows

package setting

import (
	"Golang-WSL-GUI/src/installWSL"
	"strings"
)

// 提取WSL版本
func GetOnlyWslVersion(Info installWSL.WSLinfo) string {
	// 执行命令
	out, err := installWSL.Start_cmd(Info, "Version")
	if err != nil {
		return err.Error()
	}
	lines := strings.Split(installWSL.Reduce_Unicode(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "WSL") && !strings.Contains(line, "WSLg") {
			fields := strings.Fields(line)
			for _, f := range fields {
				// 检查是否包含点号且第一个字符是数字
				if strings.Contains(f, ".") && f[0] >= '0' && f[0] <= '9' {
					return f
				}
			}
		}
	}
	return "Unknown"
}
