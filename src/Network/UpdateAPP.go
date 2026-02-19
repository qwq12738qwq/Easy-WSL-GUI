//go:build windows
// +build windows

package network

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// UpdateInfo 对应 version.json 的结构
type UpdateInfo struct {
	Version     string `json:"version"`
	UpdateLog   string `json:"updateLog"`
	ReleaseDate string `json:"releaseDate"`
	Url         string `json:"url"`
}

// CheckForUpdate 检查更新
// jsonUrl: 远程 version.json 的地址
func CheckForUpdate(currentVersion string, jsonUrl string) (*UpdateInfo, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(jsonUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var remoteInfo UpdateInfo
	err = json.Unmarshal(body, &remoteInfo)
	if err != nil {
		return nil, err
	}

	// 如果远程版本大于当前版本，返回更新信息
	if compareVersion(remoteInfo.Version, currentVersion) > 0 {
		return &remoteInfo, nil
	}

	return nil, nil
}

// compareVersion 版本号比对
// return 1 if v1 > v2
// return 0 if v1 == v2
// return -1 if v1 < v2
func compareVersion(v1, v2 string) int {
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	// 分离主版本号和预发布版本号 (e.g. "1.0.0-beta.1" -> Main:"1.0.0", Pre:"beta.1")
	var v1Main, v1Pre string
	if i := strings.Index(v1, "-"); i != -1 {
		v1Main = v1[:i]
		v1Pre = v1[i+1:]
	} else {
		v1Main = v1
	}

	var v2Main, v2Pre string
	if i := strings.Index(v2, "-"); i != -1 {
		v2Main = v2[:i]
		v2Pre = v2[i+1:]
	} else {
		v2Main = v2
	}

	// 1. 比较主版本号 (Core Version)
	parts1 := strings.Split(v1Main, ".")
	parts2 := strings.Split(v2Main, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		val1 := 0
		if i < len(parts1) {
			val1, _ = strconv.Atoi(parts1[i])
		}

		val2 := 0
		if i < len(parts2) {
			val2, _ = strconv.Atoi(parts2[i])
		}

		if val1 > val2 {
			return 1
		}
		if val1 < val2 {
			return -1
		}
	}

	// 2. 主版本号相同，比较预发布版本号 (Pre-release)
	// 规则: 正式版 (无后缀) > 预发布版 (有后缀)
	if v1Pre == "" && v2Pre != "" {
		return 1
	}
	if v1Pre != "" && v2Pre == "" {
		return -1
	}
	if v1Pre == "" && v2Pre == "" {
		return 0
	}

	// 3. 都有后缀，简单比较 (实际 SemVer 规则更复杂，这里做简单 ASCII 比较)
	// e.g. "beta" > "alpha" -> true (beta is newer than alpha)
	// e.g. "rc" > "beta" -> true
	if v1Pre > v2Pre {
		return 1
	}
	if v1Pre < v2Pre {
		return -1
	}

	return 0
}
