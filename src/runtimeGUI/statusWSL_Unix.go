//go:build !windows
// +build !windows

package runtimeGUI

type List struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Version string `json:"version"`
}

func GetWSLallStatus() ([]*List, error) { return nil, nil }
