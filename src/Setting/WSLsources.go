//go:build windows
// +build windows

package setting

import (
	"Golang-WSL-GUI/src/Global"
	"Golang-WSL-GUI/src/installWSL"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// DistroVersion 表示从 /etc/os-release 中解析出来的简化发行版信息
//
// Name      对应 NAME 字段, 例如 "Ubuntu" / "Debian GNU/Linux"
// ID        对应 ID 字段,   例如 "ubuntu" / "debian" / "fedora"
// IDLike    对应 ID_LIKE 字段, 用于判断家族, 例如 "debian" / "rhel fedora"
// VersionID 对应 VERSION_ID 字段, 例如 "24.04" / "12"
// Codename  对应 VERSION_CODENAME / UBUNTU_CODENAME 等字段, 例如 "noble" / "bookworm"
type DistroVersion struct {
	Name      string
	ID        string
	IDLike    string
	VersionID string
	Codename  string
}

// 对前端暴露的软件源枚举值, 与前端和 app.go 保持一致
const (
	SourceOfficial = "official" // 官方源
	SourceAliyun   = "aliyun"   // 阿里云镜像
	SourceTuna     = "tsinghua" // 清华镜像
	SourceTencent  = "tencent"  // 腾讯云镜像
	SourceUSTC     = "ustc"     // 中科大镜像
	SourceUnknown  = "unknown"  // 无法识别当前源
)

// 内部使用的发行版家族枚举, 用于根据不同包管理器执行不同逻辑
type distroFamily int

const (
	familyUnknown distroFamily = iota
	familyDebian               // Debian / Ubuntu / Kali 等 APT 系
	familyRHEL                 // Fedora / AlmaLinux 等 DNF(YUM) 系
	familySUSE                 // openSUSE / SLE 等 Zypper 系
)

// runInWSL 在指定发行版内部执行一段 shell 命令
//
// 这里不走 installWSL.Start_cmd, 目的是:
//  1. 避免旧版本中 "DistroVersion"、"OldSources" 等 action 实现存在的引号问题
//  2. 保证本文件可以单独修复/演进, 符合 "只修改 WSLsources.go" 的要求
//
// 所有 WSL 内部操作 (读取 /etc/os-release、sources.list 以及 *.repo 文件等)
// 统一通过该函数完成, 并统一做编码规整 (Reduce_Unicode).
func runInWSL(info installWSL.WSLinfo, shell string) (string, error) {
	if info.Linux_Version == "" {
		return "", fmt.Errorf("Linux_Version 不能为空")
	}
	cmd := exec.Command(
		"wsl.exe",
		"-d", info.Linux_Version,
		"-u", "root",
		"--",
		"sh", "-c", shell,
	)
	out, err := cmd.CombinedOutput()
	return installWSL.Reduce_Unicode(out), err
}

// GetSimpleDistroInfo 读取 /etc/os-release, 提取发行版基础信息
//
// 相比旧实现, 不再依赖 installWSL.Start_cmd("DistroVersion"), 而是直接在
// 发行版内部执行 "cat /etc/os-release", 避免原有 action 中引号不闭合导致的异常。
func GetSimpleDistroInfo(info installWSL.WSLinfo) (DistroVersion, error) {
	var version DistroVersion

	content, err := runInWSL(info, "cat /etc/os-release 2>/dev/null")
	if err != nil {
		return version, err
	}

	for _, line := range strings.Split(content, "\n") {
		cleanLine := strings.TrimSpace(line)
		if cleanLine == "" || strings.HasPrefix(cleanLine, "#") {
			continue
		}

		parts := strings.SplitN(cleanLine, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := strings.Trim(parts[1], "\"")

		switch key {
		case "NAME":
			version.Name = value
		case "ID":
			version.ID = strings.ToLower(value)
		case "ID_LIKE":
			version.IDLike = strings.ToLower(value)
		case "VERSION_ID":
			version.VersionID = value
		case "VERSION_CODENAME", "UBUNTU_CODENAME":
			if version.Codename == "" {
				version.Codename = value
			}
		}
	}

	return version, nil
}

// detectFamily 根据 DistroVersion 判断发行版家族 (APT / DNF / Zypper)
func detectFamily(d DistroVersion) distroFamily {
	all := strings.ToLower(strings.Join([]string{d.ID, d.IDLike, d.Name}, " "))
	all = strings.TrimSpace(all)

	switch {
	case strings.Contains(all, "ubuntu"), strings.Contains(all, "debian"), strings.Contains(all, "kali"):
		return familyDebian
	case strings.Contains(all, "fedora"), strings.Contains(all, "rhel"), strings.Contains(all, "almalinux"), strings.Contains(all, "alma"):
		return familyRHEL
	case strings.Contains(all, "suse"), strings.Contains(all, "opensuse"):
		return familySUSE
	default:
		return familyUnknown
	}
}

// CheckCurrentAptSource 检测当前发行版的软件源
//
// 为了兼容旧代码, 名称仍然保留 "Apt", 但内部已经扩展支持:
//   - Ubuntu 20.04 / 22.04 / 24.04 / 25.04 / 26.04
//   - Debian
//   - Kali
//   - Fedora
//   - AlmaLinux
//   - openSUSE (Leap / Tumbleweed)
//
// 返回值统一为 SourceOfficial / SourceAliyun / SourceTuna / SourceUnknown
// 前端可以直接用该结果高亮当前源。
func CheckCurrentAptSource(_ context.Context, info installWSL.WSLinfo) (string, error) {
	if global.AppLogger != nil {
		global.AppLogger.Info("CheckCurrentAptSource: 正在检测发行版 %s 的软件源", info.Linux_Version)
	}

	distro, err := GetSimpleDistroInfo(info)
	if err != nil {
		if global.AppLogger != nil {
			global.AppLogger.Error("CheckCurrentAptSource: 获取发行版 %s 系统信息失败: %s", info.Linux_Version, err.Error())
		}
		return "", fmt.Errorf("获取系统信息失败: %v", err)
	}

	family := detectFamily(distro)
	if family == familyUnknown {
		if global.AppLogger != nil {
			global.AppLogger.Warning("CheckCurrentAptSource: 发行版 %s 家族未知", info.Linux_Version)
		}
		return SourceUnknown, nil
	}

	var content string
	// 分不同家族读取对应的软件源配置文件
	switch family {
	case familyDebian:
		// APT 系: 优先读取 Ubuntu 的 deb822 格式 ubuntu.sources,
		// 若不存在则回退到傳統 /etc/apt/sources.list
		if strings.ToLower(distro.ID) == "ubuntu" {
			content, err = runInWSL(info, "if [ -f /etc/apt/sources.list.d/ubuntu.sources ]; then cat /etc/apt/sources.list.d/ubuntu.sources; elif [ -f /etc/apt/sources.list ]; then cat /etc/apt/sources.list; fi; cat /etc/apt/sources.list.d/*.list 2>/dev/null")
		} else {
			content, err = runInWSL(info, "cat /etc/apt/sources.list 2>/dev/null; cat /etc/apt/sources.list.d/*.list 2>/dev/null")
		}
	case familyRHEL:
		// DNF(YUM) 系: 只提取 baseurl/mirrorlist, 便于判断镜像站
		content, err = runInWSL(info, "grep -hE '^(baseurl|mirrorlist)=' /etc/yum.repos.d/*.repo 2>/dev/null || true")
	case familySUSE:
		// Zypper 系: 读取所有 repo 文件中的 baseurl
		content, err = runInWSL(info, "grep -hE '^baseurl=' /etc/zypp/repos.d/*.repo 2>/dev/null || true")
	}

	if err != nil {
		// 读取失败视为未知源, 但不视为致命错误
		if global.AppLogger != nil {
			global.AppLogger.Warning("CheckCurrentAptSource: 读取发行版 %s 软件源配置失败", info.Linux_Version)
		}
		return SourceUnknown, nil
	}

	contentLower := strings.ToLower(content)
	trimmed := strings.TrimSpace(contentLower)

	// 优先判断是否为常见国内镜像
	switch {
	case strings.Contains(contentLower, "mirrors.aliyun.com"):
		if global.AppLogger != nil {
			global.AppLogger.Info("CheckCurrentAptSource: 发行版 %s 当前使用阿里云镜像", info.Linux_Version)
		}
		return SourceAliyun, nil
	case strings.Contains(contentLower, "mirrors.tuna.tsinghua.edu.cn"):
		if global.AppLogger != nil {
			global.AppLogger.Info("CheckCurrentAptSource: 发行版 %s 当前使用清华镜像", info.Linux_Version)
		}
		return SourceTuna, nil
	case strings.Contains(contentLower, "mirrors.cloud.tencent.com"), strings.Contains(contentLower, "mirrors.tencent.com"):
		if global.AppLogger != nil {
			global.AppLogger.Info("CheckCurrentAptSource: 发行版 %s 当前使用腾讯云镜像", info.Linux_Version)
		}
		return SourceTencent, nil
	case strings.Contains(contentLower, "mirrors.ustc.edu.cn"):
		if global.AppLogger != nil {
			global.AppLogger.Info("CheckCurrentAptSource: 发行版 %s 当前使用中科大镜像", info.Linux_Version)
		}
		return SourceUSTC, nil
	}

	// 其次根据不同家族的官方域名判断是否仍在使用官方源
	if trimmed == "" {
		// 文件为空时, 默认为官方源 (系统可能还在使用默认配置)
		if global.AppLogger != nil {
			global.AppLogger.Info("CheckCurrentAptSource: 发行版 %s 当前使用官方源", info.Linux_Version)
		}
		return SourceOfficial, nil
	}

	switch family {
	case familyDebian:
		if strings.Contains(contentLower, "archive.ubuntu.com") ||
			strings.Contains(contentLower, "security.ubuntu.com") ||
			strings.Contains(contentLower, "deb.debian.org") ||
			strings.Contains(contentLower, "security.debian.org") ||
			strings.Contains(contentLower, "http.kali.org") {
			if global.AppLogger != nil {
				global.AppLogger.Info("CheckCurrentAptSource: 发行版 %s 当前使用官方源", info.Linux_Version)
			}
			return SourceOfficial, nil
		}
	case familyRHEL:
		if strings.Contains(contentLower, "download.fedoraproject.org") ||
			strings.Contains(contentLower, "repo.almalinux.org") ||
			strings.Contains(contentLower, "mirrors.almalinux.org") {
			if global.AppLogger != nil {
				global.AppLogger.Info("CheckCurrentAptSource: 发行版 %s 当前使用官方源", info.Linux_Version)
			}
			return SourceOfficial, nil
		}
	case familySUSE:
		if strings.Contains(contentLower, "download.opensuse.org") {
			if global.AppLogger != nil {
				global.AppLogger.Info("CheckCurrentAptSource: 发行版 %s 当前使用官方源", info.Linux_Version)
			}
			return SourceOfficial, nil
		}
	}

	// 既不是已知官方域名, 也不是已有的镜像站, 统一视为 unknown
	if global.AppLogger != nil {
		global.AppLogger.Warning("CheckCurrentAptSource: 发行版 %s 当前软件源未知", info.Linux_Version)
	}
	return SourceUnknown, nil
}

// ChangeDistroSource 根据发行版与目标源枚举, 在 WSL 中执行换源操作
//
// 支持发行版:
//   - Debian / Ubuntu / Kali (APT)
//   - Fedora / AlmaLinux (DNF)
//   - openSUSE / SLE (Zypper)
//
// 设计原则:
//  1. 不删除原始源文件, 只做注释 + 追加新配置;
//  2. 针对 APT: 注释 /etc/apt/sources.list 中的 deb 行, 在
//     /etc/apt/sources.list.d 中写入新源;
//  3. 针对 DNF/Zypper: 为每个 repo 文件生成 .easywslgui.bak 备份, 并在
//     文件内注释原始 baseurl 行后追加新的镜像 baseurl;
//  4. 尽量使用通用变量 (如 $releasever/$basearch), 以兼容不同版本
//     (Ubuntu 20.04/22.04/24.04/25.04/26.04 等)。
//
// 返回值: "success" 表示切换成功, 其余情况返回具体错误信息。
func ChangeDistroSource(_ context.Context, info installWSL.WSLinfo, source string) (string, error) {
	if global.AppLogger != nil {
		global.AppLogger.Info("ChangeDistroSource: 正在为发行版 %s 切换软件源为 %s", info.Linux_Version, source)
	}

	if source != SourceOfficial && source != SourceAliyun && source != SourceTuna && source != SourceTencent && source != SourceUSTC {
		if global.AppLogger != nil {
			global.AppLogger.Error("ChangeDistroSource: 不支持的软件源类型: %s", source)
		}
		return "", fmt.Errorf("不支持的软件源类型: %s", source)
	}

	distro, err := GetSimpleDistroInfo(info)
	if err != nil {
		if global.AppLogger != nil {
			global.AppLogger.Error("ChangeDistroSource: 获取发行版 %s 系统信息失败: %s", info.Linux_Version, err.Error())
		}
		return "", fmt.Errorf("获取系统信息失败: %v", err)
	}

	family := detectFamily(distro)
	if family == familyUnknown {
		if global.AppLogger != nil {
			global.AppLogger.Error("ChangeDistroSource: 当前系统 %s 暂不支持换源", distro.Name)
		}
		return "", fmt.Errorf("当前系统暂不支持换源: %s", distro.Name)
	}

	switch family {
	case familyDebian:
		if err := changeDebianFamilySource(info, distro, source); err != nil {
			if global.AppLogger != nil {
				global.AppLogger.Error("ChangeDistroSource: Debian系换源失败: %s", err.Error())
			}
			return "", err
		}
	case familyRHEL:
		if err := changeYumLikeSource(info, source); err != nil {
			if global.AppLogger != nil {
				global.AppLogger.Error("ChangeDistroSource: DNF系换源失败: %s", err.Error())
			}
			return "", err
		}
	case familySUSE:
		if err := changeZypperSource(info, source); err != nil {
			if global.AppLogger != nil {
				global.AppLogger.Error("ChangeDistroSource: Zypper系换源失败: %s", err.Error())
			}
			return "", err
		}
	}

	if global.AppLogger != nil {
		global.AppLogger.Info("ChangeDistroSource: 发行版 %s 软件源切换成功", info.Linux_Version)
	}
	return "success", nil
}

// changeDebianFamilySource 处理 Debian/Ubuntu/Kali 等 APT 系发行版换源逻辑
func changeDebianFamilySource(info installWSL.WSLinfo, distro DistroVersion, source string) error {
	isUbuntu := strings.ToLower(distro.ID) == "ubuntu"
	usedDeb822 := false

	if isUbuntu {
		content, _ := runInWSL(info, "cat /etc/apt/sources.list 2>/dev/null")
		if strings.Contains(content, "# Ubuntu sources have moved to the /etc/apt/sources.list.d/ubuntu.sources") {
			if err := updateUbuntuDeb822Sources(info, source); err != nil {
				return fmt.Errorf("更新 Ubuntu deb822 源失败: %v", err)
			}
			usedDeb822 = true
		}
	}

	if usedDeb822 {
		_, _ = runInWSL(info, "rm -f /etc/apt/sources.list.d/easy-wsl-gui.list 2>/dev/null || true")
		_, _ = runInWSL(info, "(apt-get update || apt update) 2>/dev/null || true")
		return nil
	}

	// 1. 备份原始 /etc/apt/sources.list
	_, _ = runInWSL(info, "if [ -f /etc/apt/sources.list ]; then cp /etc/apt/sources.list /etc/apt/sources.list.easywslgui.bak 2>/dev/null || true; fi")

	// 2. 注释掉原有的 deb / deb-src 行, 保留原始内容
	_, _ = runInWSL(info, "if [ -f /etc/apt/sources.list ]; then sed -i 's/^\\s*deb /#&/; s/^\\s*deb-src /#&/' /etc/apt/sources.list 2>/dev/null || true; fi")

	// 3. 生成新的源列表写入 /etc/apt/sources.list.d/easy-wsl-gui.list
	lines := buildDebianFamilySourceLines(distro, source)
	if err := writeLinesToFile(info, "/etc/apt/sources.list.d/easy-wsl-gui.list", lines); err != nil {
		return fmt.Errorf("写入 APT 源文件失败: %v", err)
	}

	// 4. 刷新索引, 若失败也不视为致命错误, 交给前端提示
	_, _ = runInWSL(info, "(apt-get update || apt update) 2>/dev/null || true")
	return nil
}

// buildDebianFamilySourceLines 根据发行版 + 目标源枚举生成 APT 源配置
func buildDebianFamilySourceLines(distro DistroVersion, source string) []string {
	lines := []string{
		"# EASY-WSL-GUI: 自动生成的软件源配置",
	}

	id := strings.ToLower(distro.ID)
	codename := strings.ToLower(distro.Codename)
	if codename == "" {
		// 部分发行版没有提供 VERSION_CODENAME, 针对常见版本做兜底映射
		codename = fallbackCodename(id, distro.VersionID)
	}

	// 默认兜底: 若仍然未知, 使用 generic 分支, 由发行版自身解析
	if codename == "" {
		codename = "stable"
	}

	switch id {
	case "ubuntu":
		lines = append(lines, ubuntuSourceLines(codename, source)...)
	case "debian":
		lines = append(lines, debianSourceLines(codename, source)...)
	case "kali":
		lines = append(lines, kaliSourceLines(codename, source)...)
	default:
		// 未知但 APT 系 (例如 "linuxmint" 等), 使用 Debian 通用模板
		lines = append(lines, debianSourceLines(codename, source)...)
	}

	return lines
}

// fallbackCodename 针对常见发行版做 VERSION_ID -> codename 的兜底映射
func fallbackCodename(id, versionID string) string {
	id = strings.ToLower(id)
	versionID = strings.TrimSpace(versionID)

	switch id {
	case "ubuntu":
		// 覆盖题目中提到的几个 LTS/中期版本
		switch versionID {
		case "20.04":
			return "focal"
		case "22.04":
			return "jammy"
		case "24.04":
			return "noble"
		case "25.04":
			// 根据 Ubuntu 版本习惯, 25.04 为中期版本, 这里使用当前公开代号
			return "oracular"
		case "26.04":
			// 26.04 将会是下一个 LTS, 使用占位代号, 实际以系统为准
			return "plucky"
		}
	case "debian":
		// 仅做简单映射, 实际以系统 /etc/os-release 为准
		switch versionID {
		case "12":
			return "bookworm"
		case "11":
			return "bullseye"
		case "10":
			return "buster"
		}
	case "kali":
		// Kali 长期使用 rolling 分支
		return "kali-rolling"
	}

	return ""
}

// ubuntuSourceLines 构造 Ubuntu 系软件源
func ubuntuSourceLines(codename, source string) []string {
	base, security := ubuntuMirrorBase(source)

	return []string{
		fmt.Sprintf("deb %s %s main restricted universe multiverse", base, codename),
		fmt.Sprintf("deb %s %s-updates main restricted universe multiverse", base, codename),
		fmt.Sprintf("deb %s %s-backports main restricted universe multiverse", base, codename),
		fmt.Sprintf("deb %s %s-security main restricted universe multiverse", security, codename),
	}
}

func ubuntuMirrorBase(source string) (string, string) {
	switch source {
	case SourceAliyun:
		return "https://mirrors.aliyun.com/ubuntu", "https://mirrors.aliyun.com/ubuntu"
	case SourceTuna:
		return "https://mirrors.tuna.tsinghua.edu.cn/ubuntu", "https://mirrors.tuna.tsinghua.edu.cn/ubuntu"
	case SourceTencent:
		return "https://mirrors.cloud.tencent.com/ubuntu", "https://mirrors.cloud.tencent.com/ubuntu"
	case SourceUSTC:
		return "https://mirrors.ustc.edu.cn/ubuntu", "https://mirrors.ustc.edu.cn/ubuntu"
	default:
		return "http://archive.ubuntu.com/ubuntu", "http://security.ubuntu.com/ubuntu"
	}
}

// debianSourceLines 构造 Debian 系软件源
func debianSourceLines(codename, source string) []string {
	var base, security string

	switch source {
	case SourceAliyun:
		base = "https://mirrors.aliyun.com/debian"
		security = "https://mirrors.aliyun.com/debian-security"
	case SourceTuna:
		base = "https://mirrors.tuna.tsinghua.edu.cn/debian"
		security = "https://mirrors.tuna.tsinghua.edu.cn/debian-security"
	case SourceTencent:
		base = "https://mirrors.cloud.tencent.com/debian"
		security = "https://mirrors.cloud.tencent.com/debian-security"
	case SourceUSTC:
		base = "https://mirrors.ustc.edu.cn/debian"
		security = "https://mirrors.ustc.edu.cn/debian-security"
	default:
		base = "http://deb.debian.org/debian"
		security = "http://security.debian.org/debian-security"
	}

	return []string{
		fmt.Sprintf("deb %s %s main contrib non-free non-free-firmware", base, codename),
		fmt.Sprintf("deb %s %s-updates main contrib non-free non-free-firmware", base, codename),
		fmt.Sprintf("deb %s %s-security main contrib non-free non-free-firmware", security, codename),
	}
}

// kaliSourceLines 构造 Kali 软件源
func kaliSourceLines(codename, source string) []string {
	var base string

	switch source {
	case SourceAliyun:
		base = "https://mirrors.aliyun.com/kali"
	case SourceTuna:
		base = "https://mirrors.tuna.tsinghua.edu.cn/kali"
	case SourceTencent:
		base = "https://mirrors.cloud.tencent.com/kali"
	case SourceUSTC:
		base = "https://mirrors.ustc.edu.cn/kali"
	default:
		base = "http://http.kali.org/kali"
	}

	if codename == "" {
		codename = "kali-rolling"
	}

	return []string{
		fmt.Sprintf("deb %s %s main contrib non-free non-free-firmware", base, codename),
	}
}

func updateUbuntuDeb822Sources(info installWSL.WSLinfo, source string) error {
	var script string
	if source == SourceOfficial {
		script = `
if [ -f /etc/apt/sources.list.d/ubuntu.sources.easywslgui.bak ]; then
  cp /etc/apt/sources.list.d/ubuntu.sources.easywslgui.bak /etc/apt/sources.list.d/ubuntu.sources 2>/dev/null || true
fi
`
	} else {
		base, _ := ubuntuMirrorBase(source)
		script = fmt.Sprintf(`
if [ -f /etc/apt/sources.list.d/ubuntu.sources ]; then
  if [ ! -f /etc/apt/sources.list.d/ubuntu.sources.easywslgui.bak ]; then
    cp /etc/apt/sources.list.d/ubuntu.sources /etc/apt/sources.list.d/ubuntu.sources.easywslgui.bak 2>/dev/null || true
  fi
  awk -v mirror="%s" '
    /^URIs:/ { print "URIs: " mirror; next }
    { print }
  ' /etc/apt/sources.list.d/ubuntu.sources > /etc/apt/sources.list.d/ubuntu.sources.easywslgui.tmp && mv /etc/apt/sources.list.d/ubuntu.sources.easywslgui.tmp /etc/apt/sources.list.d/ubuntu.sources
fi
`, base)
	}

	if _, err := runInWSL(info, script); err != nil {
		return err
	}
	return nil
}

// writeLinesToFile 在 WSL 发行版内部写入指定文件 (覆盖写入, 不追加)
//
// 为保持简单和稳健, 采用多次 runInWSL 调用逐行写入:
//  1. 首先使用 printf ” > file 清空文件;
//  2. 然后对每一行执行一次 printf 'line\n' >> file;
//
// 这样可以避免复杂的 here-doc 转义问题, 同时保证即使中途失败也不会
// 破坏原有备份文件。
func writeLinesToFile(info installWSL.WSLinfo, filePath string, lines []string) error {
	// 清空目标文件
	if _, err := runInWSL(info, fmt.Sprintf("printf '' > %s", filePath)); err != nil {
		return err
	}

	for _, line := range lines {
		escaped := strings.ReplaceAll(line, "'", "'\\''")
		cmd := fmt.Sprintf("printf '%s\\n' >> %s", escaped, filePath)
		if _, err := runInWSL(info, cmd); err != nil {
			return err
		}
	}
	return nil
}

// changeYumLikeSource 处理 Fedora / AlmaLinux 等 DNF(YUM) 系换源
func changeYumLikeSource(info installWSL.WSLinfo, source string) error {
	var mirror string
	switch source {
	case SourceAliyun:
		mirror = "https://mirrors.aliyun.com"
	case SourceTuna:
		mirror = "https://mirrors.tuna.tsinghua.edu.cn"
	case SourceTencent:
		mirror = "https://mirrors.cloud.tencent.com"
	case SourceUSTC:
		mirror = "https://mirrors.ustc.edu.cn"
	default:
		// 切回官方: 通过恢复备份文件实现
		_, _ = runInWSL(info, "for f in /etc/yum.repos.d/*.repo.easywslgui.bak; do [ -f \"$f\" ] || continue; orig=${f%.easywslgui.bak}; cp \"$f\" \"$orig\" 2>/dev/null || true; done")
		_, _ = runInWSL(info, "(dnf clean all && dnf makecache -y) 2>/dev/null || true")
		return nil
	}

	// 备份原始 repo 文件, 并在文件内注释原始 baseurl 行后附加新的镜像地址
	script := fmt.Sprintf(`
for f in /etc/yum.repos.d/*.repo; do
  if [ -f "$f" ]; then
    cp "$f" "$f.easywslgui.bak" 2>/dev/null || true
    awk -v mirror="%s" '
      /^baseurl=/ {
        original=$0
        sub(/^baseurl=http(s)?:\/\/[^/]*\//, "", $0)
        path=$0
        print "# EASY-WSL-GUI: " original
        print "baseurl=" mirror "/" path
        next
      }
      { print }
    ' "$f" > "${f}.tmp" && mv "${f}.tmp" "$f"
  fi
done
`, mirror)

	if _, err := runInWSL(info, script); err != nil {
		return fmt.Errorf("更新 DNF 源失败: %v", err)
	}

	_, _ = runInWSL(info, "(dnf clean all && dnf makecache -y) 2>/dev/null || true")
	return nil
}

// changeZypperSource 处理 openSUSE / SLE 等 Zypper 系换源
func changeZypperSource(info installWSL.WSLinfo, source string) error {
	var mirror string
	switch source {
	case SourceAliyun:
		mirror = "https://mirrors.aliyun.com/opensuse"
	case SourceTuna:
		mirror = "https://mirrors.tuna.tsinghua.edu.cn/opensuse"
	case SourceTencent:
		mirror = "https://mirrors.cloud.tencent.com/opensuse"
	case SourceUSTC:
		mirror = "https://mirrors.ustc.edu.cn/opensuse"
	default:
		// 官方源恢复
		_, _ = runInWSL(info, "for f in /etc/zypp/repos.d/*.repo.easywslgui.bak; do [ -f \"$f\" ] || continue; orig=${f%.easywslgui.bak}; cp \"$f\" \"$orig\" 2>/dev/null || true; done")
		_, _ = runInWSL(info, "zypper --gpg-auto-import-keys refresh -f 2>/dev/null || true")
		return nil
	}

	script := fmt.Sprintf(`
for f in /etc/zypp/repos.d/*.repo; do
  if [ -f "$f" ]; then
    cp "$f" "$f.easywslgui.bak" 2>/dev/null || true
    awk -v mirror="%s" '
      /^baseurl=/ {
        original=$0
        sub(/^baseurl=http(s)?:\/\/[^/]*\//, "", $0)
        path=$0
        print "# EASY-WSL-GUI: " original
        print "baseurl=" mirror "/" path
        next
      }
      { print }
    ' "$f" > "${f}.tmp" && mv "${f}.tmp" "$f"
  fi
done
`, mirror)

	if _, err := runInWSL(info, script); err != nil {
		return fmt.Errorf("更新 Zypper 源失败: %v", err)
	}

	_, _ = runInWSL(info, "zypper --gpg-auto-import-keys refresh -f 2>/dev/null || true")
	return nil
}
