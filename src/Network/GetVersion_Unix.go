//go:build !windows
// +build !windows

package network

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

// 获取发行版Json请求
func GetDistroList() ([]DistroItem, error) { return nil, nil }
