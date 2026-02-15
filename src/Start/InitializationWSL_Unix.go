//go:build !windows
// +build !windows

package start

// 输出错误
func ShowFatalError(err error) {}

// 详细检测wsl
func DetectWSL() error { return nil }

// 启动时生成.wslconfig文件
func EnsureWslConfigExists() error { return nil }

// 检测管理员
func CheckAdmin() error { return nil }
