//go:build !windows
// +build !windows

package installWSL

import (
	"context"
)

// 安装WSL发行版信息
type WSLinfo struct {
	Linux_Version   string
	Install_Path    *WSLpath
	Auth            *WSLAuth
	DownloadThreads *WSLDownload
}

type WSLpath struct {
	Path string
}

type WSLAuth struct {
	User     string
	Password string
}

type WSLDownload struct {
	DownloadThreads int
}

// WSL2发行包下载函数
func WSL2_Downloader(ctx context.Context, Info WSLinfo) error { return nil }

// 移动发行版函数
func MovingPathWSL(ctx context.Context, Info WSLinfo) {}

// 去空格,去中文裁剪
func Reduce_Unicode(by_stream []byte) string { return "" }

// 卸载WSL发行版函数
func UninstallWSL(ctx context.Context, Info WSLinfo) error { return nil }

// WSL发行版安装函数
func WSL2_Installer(ctx context.Context, Info WSLinfo) error { return nil }

// WSL发行版配置用户
func WSL2_Setting_User(ctx context.Context, Info WSLinfo) error { return nil }

// 启动命令函数
func Start_cmd(Info WSLinfo, action string) ([]byte, error) { return nil, nil }
