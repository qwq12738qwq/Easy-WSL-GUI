//go:build !windows
// +build !windows

package installWSL

type WSLGroup struct {
	Name  string
	GID   int
	Users []string
}

func GetWSLUserGroups(distroName string) ([]WSLGroup, error) { return nil, nil }
