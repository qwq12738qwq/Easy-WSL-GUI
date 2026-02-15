//go:build !windows
// +build !windows

package setting

import (
	"Golang-WSL-GUI/src/installWSL"
	"context"
)

func CheckCurrentAptSource(ctx context.Context, Info installWSL.WSLinfo) (string, error) {
	return "", nil
}

func ChangeDistroSource(ctx context.Context, Info installWSL.WSLinfo, sources string) (string, error) {
	return "", nil
}
