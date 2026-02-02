//go:build windows
// +build windows

package installWSL

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"
)

// 安装WSL发行版信息
type WSLinfo struct {
	Linux_Version string
	Install_Path  string
	User          string
	Password      string
}

type Download_WSL struct {
	URL    string
	Sha256 string
}

// 下载map映射表
var WSLdownloadMap = map[string]Download_WSL{
	// "Ubuntu-24.04.02": {
	// 	URL:    "https://old-releases.ubuntu.com/releases/24.04.2/ubuntu-24.04.2-wsl-amd64.wsl",
	// 	Sha256: "5d1eea52103166f1c460dc012ed325c6eb31d2ce16ef6a00ffdfda8e99e12f43",
	// },
	"Ubuntu-24.04": {
		URL:    "https://releases.ubuntu.com/24.04/ubuntu-24.04.3-wsl-amd64.wsl",
		Sha256: "c74833a55e525b1e99e1541509c566bb3e32bdb53bf27ea3347174364a57f47c",
	},
	"Ubuntu-25.04": {
		URL:    "https://releases.ubuntu.com/25.04/ubuntu-25.04-wsl-amd64.wsl",
		Sha256: "91f3e836698719846191821300bd21f321811abcef6f448bf8d7d8f8517b2743",
	},
	"Ubuntu-25.10": {
		URL:    "https://releases.ubuntu.com/25.10/ubuntu-25.10-wsl-amd64.wsl",
		Sha256: "05299da14668ed5e1ddb49b92618725c5e6b55fca5bd163e314c227803af27e1",
	},
	"Ubuntu-26.04": {
		URL:    "https://releases.ubuntu.com/26.04-snapshot1/ubuntu-26.04-wsl-amd64.wsl",
		Sha256: "c77c9e8a5b0255cd02f5edbcd612663976d995980107ce116b2b61f9244cce79",
	},
	"Debian": {
		URL:    "https://salsa.debian.org/debian/WSL/-/jobs/7949331/artifacts/raw/Debian_WSL_AMD64_v1.22.0.0.wsl",
		Sha256: "543123ccc5f838e63dac81634fb0223dc8dcaa78fdb981387d625feb1ed168c7",
	},
	"Kali": {
		URL:    "https://kali.download/wsl-images/kali-2025.4/kali-linux-2025.4-wsl-rootfs-amd64.wsl",
		Sha256: "86aba7bb3d74d313e349f9f50d3f6119ee3b1491072920d063f17ce9b3f706ab",
	},
	"Arch": {
		URL:    "https://fastly.mirror.pkgbuild.com/wsl/2026.01.01.156076/archlinux-2026.01.01.156076.wsl",
		Sha256: "e3820c60df62edc22df29c9c16d2205512d95c1b086232a9b7bc3960542036d4",
	},
	"Fedora": {
		URL:    "https://download.fedoraproject.org/pub/fedora/linux/releases/43/Container/x86_64/images/Fedora-WSL-Base-43-1.6.x86_64.wsl",
		Sha256: "220780af9cf225e9645313b4c7b0457a26a38a53285eb203b2ab6188d54d5b82",
	},
}

// 删除BOM,控制字符,中文(保留空格,英文符号,换行回车)
func Reduce_Unicode(by_stream []byte) string {
	// 创建一个新的切片，容量设为 rawBytes 的长度
	finalBytes := make([]byte, 0, len(by_stream))

	for _, b := range by_stream {
		// 直接在这里过滤：跳过 0x00, BOM(0xFE, 0xFF)
		// 并且只保留 ASCII 英文范围
		if (b >= 32 && b <= 126) || b == 10 {
			finalBytes = append(finalBytes, b)
		}
	}
	return string(finalBytes)
}

// 初始化命令,Install为安装,Uninstall为卸载,传入要操作发行版的具体名字
func Init_Admin_PowerShell(Info WSLinfo, Action string) (*exec.Cmd, error) {
	// Install_Command := fmt.Sprintf(`wsl --install -d %s`, Linux_Version)
	switch Action {
	case "Check":
		return exec.Command(
			"wsl.exe",
			"-l",
			"-q",
		), nil
	case "Uninstall":
		return exec.Command(
			"wsl.exe",
			"--unregister", Info.Linux_Version,
		), nil
	case "Export":
		return exec.Command(
			"wsl.exe",
			"--export", Info.Linux_Version, Info.Install_Path,
		), nil
	case "Shutdown":
		return exec.Command(
			"wsl.exe",
			"-t", Info.Linux_Version,
		), nil
	case "Moving":
		return exec.Command(
			"",
		), nil
	case "Import":
		return exec.Command(
			"wsl.exe",
			"--import", Info.Linux_Version, Info.Install_Path, FilePath_string(Info),
			"--version", "2",
		), nil
	case "SeachUser":
		return exec.Command(
			"wsl.exe", "-d", Info.Linux_Version, "--",
			"sh", "-c",
			"grep -E '' /etc/wsl.conf",
		), nil
	case "ConfigUser":
		return exec.Command(
			"wsl", "-d", Info.Linux_Version, "--",
			"bash", "-c",
			fmt.Sprintf(
				"useradd -m -s /bin/bash %s && echo '%s:%s' | chpasswd && usermod -aG sudo %s",
				Info.User,
				Info.User, Info.Password,
				Info.User),
		), nil
	default:
		return exec.Command(""), errors.New("输入行为状态未注册")
	}
}

// 启动命令函数,将输出转为字节
func Start_cmd(Info WSLinfo, action string) []byte {
	cmd, _ := Init_Admin_PowerShell(Info, action)
	// 缓冲区
	var rawBuf bytes.Buffer
	cmd.Stdout = &rawBuf
	cmd.Stderr = &rawBuf

	// 隐藏控制台窗口
	cmd.SysProcAttr = &windows.SysProcAttr{HideWindow: true}

	cmd.Run()

	return rawBuf.Bytes()
}

// 拼接路径字符串
func FilePath_string(Info WSLinfo) string {
	fileName := fmt.Sprintf(`\%s`, Info.Linux_Version)
	// 文件名拼凑
	if strings.Contains(WSLdownloadMap[Info.Linux_Version].URL, ".wsl") {
		fileName += ".wsl"
	}

	return filepath.Join(Info.Install_Path, fileName)

}

func WSL2_Downloader(ctx context.Context, Info WSLinfo) error {

	// 1. 创建临时文件
	// _, err := os.Create("wsl_log.txt")
	// if err != nil {
	// 	runtime.EventsEmit(ctx, "wsl-error", "无法创建日志文件")
	// 	return
	// }
	if runcode := parseWSLMessage(ctx, Reduce_Unicode(Start_cmd(Info, "Check")), Info); runcode == 2 {
		return errors.New("发行版已存在")
	} else if runcode == 6 {
		runtime.EventsEmit(ctx, "wsl-output", fmt.Sprintf("%s 发行版已安装,开始配置用户", Info.Linux_Version))
		return errors.New("发行版存在,但未配置默认用户")
	}

	fullpath := FilePath_string(Info)

	// 创建目标目录
	if os.MkdirAll(filepath.Dir(fullpath), 0755) != nil {
		runtime.EventsEmit(ctx, "wsl-error", fmt.Sprintf("在路径 %s 创建安装文件失败", fullpath))
		return errors.New("创建文件失败")
	}

	// 2. 发起请求
	resp, err := http.Get(WSLdownloadMap[Info.Linux_Version].URL)
	if err != nil {
		runtime.EventsEmit(ctx, "wsl-error", fmt.Sprintf("下载 %s 发行版失败,请检查网络连接", Info.Linux_Version))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
	}

	// 3. 获取文件总大小 (Content-Length)
	totalSize := resp.ContentLength

	// 4. 创建本地文件
	out, err := os.Create(fullpath)
	if err != nil {

	}
	defer out.Close()

	// 哈希计算器
	hasher := sha256.New()
	// 数据流写入校验
	mw := io.MultiWriter(out, hasher)

	// 5. 循环读取并计算进度
	buffer := make([]byte, 64*1024) // 64KB 缓冲区
	var downloaded int64
	lastEmit := time.Now()

	for {
		n, readErr := resp.Body.Read(buffer)
		if n > 0 {
			// 写入文件
			_, writeErr := mw.Write(buffer[:n])
			if writeErr != nil {
				runtime.EventsEmit(ctx, "wsl-error", "写入文件失败")
				return errors.New("写入文件失败")
			}
			downloaded += int64(n)

			// 计算百分比并推送给前端关键词 'download'
			if totalSize > 0 {
				// 日志传输限流
				if time.Since(lastEmit) > 500*time.Millisecond {
					percent := float64(downloaded) / float64(totalSize) * 100
					runtime.EventsEmit(ctx, "wsl-output", fmt.Sprintf("正在下载镜像: %.2f%%", percent))
					lastEmit = time.Now()
				}
			}
		}

		if readErr != nil {
			if readErr == io.EOF {
				break // 下载完成
			}

		}
	}
	// 计算最终的哈希值
	actualSha256 := hex.EncodeToString(hasher.Sum(nil))

	// 校验比较
	if actualSha256 != WSLdownloadMap[Info.Linux_Version].Sha256 {
		// 如果校验失败，删除残缺文件
		os.Remove(Info.Install_Path)
		runtime.EventsEmit(ctx, "wsl-output", "Sha256校验失败,请重新执行")
	}

	runtime.EventsEmit(ctx, "wsl-output", fmt.Sprintf("下载完成,准备安装 %s", Info.Linux_Version))

	return nil
}

func parseWSLexitCode(ctx context.Context, code int, Version string) {
	// 将 int 转为无符号 32 位再转十六进制字符串，方便匹配（如 0x8004032D）
	hexCode := fmt.Sprintf("0x%08X", uint32(code))

	switch hexCode {
	case "0x00000000":
		runtime.EventsEmit(ctx, "wsl-output", fmt.Sprintf(`发行版 %s 安装成功`, Version))

	case "0x8004032D": // ERROR_ALREADY_EXISTS
		runtime.EventsEmit(ctx, "wsl-error", fmt.Sprintf("该发行版 %s 已经安装在Windows上", Version))

	case "0x80070005": // E_ACCESSDENIED
		runtime.EventsEmit(ctx, "wsl-error", "权限不足：请尝试以管理员身份运行")

	case "0x80370102": // 虚拟化未启用
		runtime.EventsEmit(ctx, "wsl-error", "虚拟机平台功能未开启或 BIOS 虚拟化被禁用")

	case "0x800701BC": // 需要更新内核
		runtime.EventsEmit(ctx, "wsl-error", "WSL 2 需要更新其内核组件")

	case "0x8007019E": // 未安装子系统功能
		runtime.EventsEmit(ctx, "wsl-error", "Windows 子系统功能 (Optional Feature) 未启用")

	case "0x8024402C", "0x80072EE2": // 网络相关
		runtime.EventsEmit(ctx, "wsl-error", "下载安装包失败，请检查网络连接")

	default:
		// 兜底：发送未知的十六进制错误码，方便后续排查
		runtime.EventsEmit(ctx, "wsl-error", "WSL 执行失败，错误代码: "+hexCode)
	}
}

func parseWSLMessage(ctx context.Context, l string, Info WSLinfo) int {
	l = strings.ToLower(l)
	ToL_Version := strings.ToLower(Info.Linux_Version)
	switch {
	case strings.Contains(l, "requireselevation"):
		runtime.EventsEmit(ctx, "wsl-error", "需要权限执行WSL安装命令,检查是否给予权限")
		return 1
	case strings.Contains(l, ToL_Version):
		if strings.Contains(Reduce_Unicode(Start_cmd(Info, "SeachUser")), "default") {
			runtime.EventsEmit(ctx, "wsl-error", fmt.Sprintf("该发行版 %s 已经安装在Windows上", Info.Linux_Version))
			return 2
		} else {
			return 3
		}
	case strings.Contains(l, "invaliduser"):
		runtime.EventsEmit(ctx, "wsl-error", fmt.Sprintf("用户名 %s 不符合规范,请重新设置", Info.User))
		return 4
	case strings.Contains(l, "istooshort"):
		runtime.EventsEmit(ctx, "wsl-error", fmt.Sprintf("密码 %s 设置太短", Info.Password))
		return 5
	case strings.Contains(l, "dictionarycheck"):
		runtime.EventsEmit(ctx, "wsl-error", fmt.Sprintf("密码 %s 不符合字典规范,请重新设置", Info.Password))
		return 6
	default:
		return -1
	}
	// runtime.EventsEmit(ctx, "wsl-error", exitCode)
}

func WSL2_Installer(ctx context.Context, Info WSLinfo) {
	runtime.EventsEmit(ctx, "wsl-output", fmt.Sprintf("正在解压安装发行版 %s ", Info.Linux_Version))
	cmd, _ := Init_Admin_PowerShell(Info, "Import")
	var rawBuf bytes.Buffer
	cmd.Stdout = &rawBuf
	cmd.Stderr = &rawBuf

	cmd.SysProcAttr = &windows.SysProcAttr{HideWindow: true}

	if err := cmd.Run(); err != nil {
		runtime.EventsEmit(ctx, "wsl-error", fmt.Sprintf("解压安装出现错误: %s ", err))
	} else {
		runtime.EventsEmit(ctx, "wsl-output", fmt.Sprintf("安装发行版 %s 成功", Info.Linux_Version))
	}

}

// 配置用户名,密码函数
func WSL2_Setting_User(ctx context.Context, Info WSLinfo) {
	Start_cmd(Info, "ConfigUser")
}

func MovingPathWSL(path string) {

}

func UninstallWSL(name string) {
	// cmd, _ := Init_Admin_PowerShell(name, "Uninstall")

	// // 创建临时日志文件
	// _, err := os.Create("wsl_log_Uninstall.txt")
	// if err != nil {
	// 	return
	// }

	// // 缓冲区
	// var rawBuf bytes.Buffer
	// cmd.Stdout = &rawBuf
	// cmd.Stderr = &rawBuf

	// cmd.SysProcAttr = &windows.SysProcAttr{HideWindow: true}

	// err = cmd.Run()

	// _ = os.WriteFile("wsl_log.txt", rawBuf.Bytes(), 0644)

}
