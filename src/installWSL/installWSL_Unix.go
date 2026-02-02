//go:build !windows
// +build !windows

package installWSL

import "context"

// 安装WSL发行版信息
type WSLinfo struct {
	Linux_Version string
	Install_Path  string
	User          string
	Password      string
}

// WSL2发行包下载函数
func WSL2_Downloader(ctx context.Context, Info WSLinfo) error { return nil }

// 移动发行版函数
func MovingPathWSL(path string) {}

// 去空格,去中文裁剪
func Reduce_Unicode(by_stream []byte) string { return "" }

// 卸载WSL发行版函数
func UninstallWSL(name string) {}

// WSL发行版安装函数
func WSL2_Installer(ctx context.Context, Info WSLinfo) {}

// WSL发行版配置用户
func WSL2_Setting_User(ctx context.Context, Info WSLinfo) {}
