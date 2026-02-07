//go:build !windows
// +build !windows

package setting

import "Golang-WSL-GUI/src/installWSL"

func GetOnlyWslVersion(Info installWSL.WSLinfo) string { return "" }
