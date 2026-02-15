//go:build windows
// +build windows

package network

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type DistroVersion struct {
	Label  string `json:"label"`  // 显示名称 (如 "Ubuntu-24.04")
	Value  string `json:"value"`  // 内部标识 (如 "Ubuntu-24.04")
	Url    string `json:"url"`    // 下载链接
	Sha256 string `json:"sha256"` // 校验和
}

type DistroItem struct {
	Id       int             `json:"id"`       // 无所谓,不冲突就行
	Name     string          `json:"name"`     // 发行版大类 (如 "Ubuntu")
	Desc     string          `json:"desc"`     // 描述
	State    string          `json:"state"`    // "online" 或 "offline"
	ImgName  string          `json:"img_name"` // 图标文件名 (不含后缀)
	Versions []DistroVersion `json:"versions"` // 版本列表
}

var json_url string = "https://gitee.com/MrLi114514/wsl_package/raw/master/Package/DistributionInfo.json"

func GetDistroList() ([]DistroItem, error) {

	client := &http.Client{Timeout: 10 * time.Second}
	// 5次重试
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		// 构造请求
		req, err := http.NewRequest("GET", json_url, nil)
		if err != nil {
			return nil, err
		}
		// if token != "" {
		// 	req.Header.Add("Authorization", "token "+token)
		// }

		// 发起请求
		resp, err := client.Do(req)

		// 网络重试
		if err == nil {
			defer resp.Body.Close()
		}
		//  5xx或429时才重试
		if resp.StatusCode == http.StatusOK {
			var distros []DistroItem
			if err := json.NewDecoder(resp.Body).Decode(&distros); err != nil {
				return nil, err
			}
			return distros, nil
		}

		// 404直接跳出
		if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != 429 {
			break
		}

		// 等待一段时间再重试 (指数退避)
		waitTime := time.Duration(1<<i) * time.Second
		time.Sleep(waitTime)
	}

	return nil, errors.New("未知错误")
}
