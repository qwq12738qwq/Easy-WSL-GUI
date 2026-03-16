//go:build !windows
// +build !windows

package runtimeGUI

type SoftwarePackage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Source  string `json:"source"`
}

func GetInstalledPackages(distroName string) ([]SoftwarePackage, error) {
	return []SoftwarePackage{}, nil
}
