//go:build !windows
// +build !windows

package network

type UpdateInfo struct {
	Version     string `json:"version"`
	UpdateLog   string `json:"updateLog"`
	ReleaseDate string `json:"releaseDate"`
	Url         string `json:"url"`
}

func CheckForUpdate(currentVersion string, jsonUrl string) (*UpdateInfo, error) { return nil, nil }
