//go:build !windows
// +build !windows

package start

// 输出错误
func ShowFatalError(err error) {}

// 详细检测wsl
func DetectWSL() error { return nil }
