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
	"strconv"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"
)

// 安装WSL发行版信息
type WSLinfo struct {
	Linux_Version   string
	Install_Path    *WSLpath
	Auth            *WSLAuth
	DownloadThreads *WSLDownload
}

type WSLpath struct {
	Path string
}

type WSLAuth struct {
	User     string
	Password string
}

type WSLDownload struct {
	DownloadThreads int
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
	"openSUSE-Tumbleweed": {
		URL:    "https://github.com/openSUSE/WSL-instarball/releases/download/v20260106.0/openSUSE-Tumbleweed-20260103.x86_64-1.224-Build1.224.wsl",
		Sha256: "0x394be699da2821b331355f3541e237aa3aa00bc4068f33283d68303d8336d484",
	},
	"openSUSE-Leap-16.0": {
		URL:    "https://github.com/openSUSE/WSL-instarball/releases/download/v20251001.0/openSUSE-Leap-16.0-16.0.x86_64-22.57-Build22.57.wsl",
		Sha256: "0x0d1faa095153beee0a9b5089b0f9aa3d2aec95e2cdcffdeeff84dd54c48b8393",
	},
	"AlmaLinux-8": {
		URL:    "https://github.com/AlmaLinux/wsl-images/releases/download/v8.10.20250415.0/AlmaLinux-8.10_x64_20250415.0.wsl",
		Sha256: "34c3bc6d3ac693968737c65db52b67f68b8c1a6f8b024450819841a967f59a3d",
	},
	"AlmaLinux-9": {
		URL:    "https://github.com/AlmaLinux/wsl-images/releases/download/v9.7.20251119.0/AlmaLinux-9.7_x64_20251119.0.wsl",
		Sha256: "0a6588f4f723fcb3edbc37dd3e3e13be8ffe0a5027e47513e3d4d2a4451794e7",
	},
	"AlmaLinux-Kitten-10": {
		URL:    "https://github.com/AlmaLinux/wsl-images/releases/download/v10-kitten.20251030.0/AlmaLinux-Kitten-10_x64_20251030.0.wsl",
		Sha256: "d765d65076b041f3a67ba60edc37d056eeab2a260aed8e077684e05b78ecd9f5",
	},
	"AlmaLinux-10": {
		URL:    "https://github.com/AlmaLinux/wsl-images/releases/download/v10.1.20251124.0/AlmaLinux-10.1_x64_20251124.0.wsl",
		Sha256: "24e8fa286a4081979d97e83a227fb89f332bcf731fe4b422679a3b455ab0be37",
	},
	"OpenSUSE-Leap-16.0": {
		URL:    "https://github.com/openSUSE/WSL-instarball/releases/download/v20251001.0/openSUSE-Leap-16.0-16.0.x86_64-22.57-Build22.57.wsl",
		Sha256: "0x0d1faa095153beee0a9b5089b0f9aa3d2aec95e2cdcffdeeff84dd54c48b8393",
	},
	"OpenSUSE-Tumbleweed": {
		URL:    "https://github.com/openSUSE/WSL-instarball/releases/download/v20260106.0/openSUSE-Tumbleweed-20260103.x86_64-1.224-Build1.224.wsl",
		Sha256: "0x394be699da2821b331355f3541e237aa3aa00bc4068f33283d68303d8336d484",
	},
	"SUSE-Linux-Enterprise-16.0": {
		URL:    "https://github.com/SUSE/WSL-instarball/releases/download/v20251201.0/SUSE-Linux-Enterprise-16.0-16.0.x86_64-1.9-Build1.9.wsl",
		Sha256: "0xf0fc07ed3543d3dc24cfb35b4194bbecf98485cefdd720c521034ac1c54bffd3",
	},
	"SUSE-Linux-Enterprise-15-SP7": {
		URL:    "https://github.com/SUSE/WSL-instarball/releases/download/v20251201.0/SUSE-Linux-Enterprise-15-SP7-15.7.x86_64-30.1-Build30.1.wsl",
		Sha256: "0x60924e13286ed15bdcf9069e3a24d3394fb858954de3bdfcb1ea576900b81b2e",
	},
}

// 删除BOM,控制字符,中文(保留空格,英文符号,换行回车)
func Reduce_Unicode(by_stream []byte) string {
	// 创建一个新的切片，容量设为 rawBytes 的长度
	finalBytes := make([]byte, 0, len(by_stream))

	for _, b := range by_stream {
		// 跳过 0x00, BOM(0xFE, 0xFF),并且只保留 ASCII 英文范围
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
			"--export", Info.Linux_Version, FilePath_string(Info),
		), nil
	case "Shutdown":
		return exec.Command(
			"wsl.exe",
			"-t", Info.Linux_Version,
		), nil
	case "Import":
		return exec.Command(
			"wsl.exe",
			"--import", Info.Linux_Version, Info.Install_Path.Path, FilePath_string(Info),
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
			"wsl.exe", "-d", Info.Linux_Version, "--",
			"sh", "-c",
			fmt.Sprintf(
				"useradd -m -s /bin/bash %s ",
				Info.Auth.User,
			),
		), nil
	case "ConfigPasswd":
		return exec.Command(
			"wsl.exe", "-d", Info.Linux_Version, "--",
			"sh", "-c",
			fmt.Sprintf(
				"echo '%s:%s' | chpasswd ",
				Info.Auth.User, Info.Auth.Password,
			),
		), nil
	case "ConfigSudo":
		return exec.Command(
			"wsl.exe", "-d", Info.Linux_Version, "--",
			"sh", "-c",
			fmt.Sprintf(
				"usermod -aG sudo %s",
				Info.Auth.User),
		), nil
	case "Default":
		return exec.Command(
			"wsl.exe", "-d", Info.Linux_Version, "--",
			"sh", "-c",
			fmt.Sprintf(
				`printf "\n[user]\ndefault=%s\n" >> /etc/wsl.conf`,
				Info.Auth.User),
		), nil
	case "Stop":
		return exec.Command(
			"wsl.exe", "--terminate", Info.Linux_Version, "true",
		), nil
	case "Start":
		return exec.Command(
			"wsl.exe", "-d", Info.Linux_Version,
		), nil
	case "Version":
		return exec.Command(
			"wsl.exe", "--version",
		), nil
	case "ShutdownAll":
		return exec.Command(
			"wsl.exe", "--shutdown",
		), nil
	case "CPUpercent":
		return exec.Command(
			"wsl.exe", "-d", Info.Linux_Version, "sh", "-c", "head -n 1 /proc/stat",
		), nil
	case "useMem":
		return exec.Command(
			"wsl.exe", "-d", Info.Linux_Version, "sh", "-c", "grep -E 'MemTotal|MemAvailable' /proc/meminfo",
		), nil
	default:
		return nil, errors.New("输入行为状态未注册")
	}
}

// 启动命令函数,将输出转为字节
func Start_cmd(Info WSLinfo, action string) ([]byte, error) {
	cmd, _ := Init_Admin_PowerShell(Info, action)
	// 缓冲区
	var rawBuf bytes.Buffer
	cmd.Stdout = &rawBuf
	cmd.Stderr = &rawBuf

	// 隐藏控制台窗口
	cmd.SysProcAttr = &windows.SysProcAttr{HideWindow: true}

	err := cmd.Run()

	return rawBuf.Bytes(), err
}

// 拼接路径字符串
func FilePath_string(Info WSLinfo) string {
	fileName := fmt.Sprintf(`\%s`, Info.Linux_Version)
	// 根据DownloadThreads是否是空指针判断是安装还是迁移
	if Info.DownloadThreads != nil {
		// 文件名拼凑
		if strings.Contains(WSLdownloadMap[Info.Linux_Version].URL, ".wsl") {
			fileName += ".wsl"
		}
	} else {
		// 迁移后缀
		fileName += ".tar"
	}

	return filepath.Join(Info.Install_Path.Path, fileName)

}

func WSL2_Downloader(ctx context.Context, Info WSLinfo) error {

	// 1. 创建临时文件
	// _, err := os.Create("wsl_log.txt")
	// if err != nil {
	// 	runtime.EventsEmit(ctx, "wsl-error", "无法创建日志文件")
	// 	return
	// }
	line, err := Start_cmd(Info, "Check")
	if err != nil {
		runtime.EventsEmit(ctx, "wsl-error", fmt.Sprintf("在检查步骤出错,出错代码: %s", err))
	}
	if runcode := parseWSLMessage(ctx, Reduce_Unicode(line), Info); runcode == 2 {
		return errors.New("发行版已存在")
	} else if runcode == 3 {
		runtime.EventsEmit(ctx, "wsl-output", fmt.Sprintf("%s 发行版已安装,开始配置用户", Info.Linux_Version))
		return errors.New("发行版存在,但未配置默认用户")
	}

	fullpath := FilePath_string(Info)

	// 创建目标目录
	if os.MkdirAll(filepath.Dir(fullpath), 0755) != nil {
		runtime.EventsEmit(ctx, "wsl-error", fmt.Sprintf("在路径 %s 创建安装文件失败", fullpath))
		return errors.New("创建文件失败")
	}

	// 发起请求
	resp, err := http.Get(WSLdownloadMap[Info.Linux_Version].URL)
	if err != nil {
		runtime.EventsEmit(ctx, "wsl-error", fmt.Sprintf("下载 %s 发行版失败,请检查网络连接", Info.Linux_Version))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		runtime.EventsEmit(ctx, "wsl-error", fmt.Sprintf("下载失败,网络错误码: %s ", strconv.Itoa(resp.StatusCode)))
		return errors.New("网络错误")
	}

	// 获取文件总大小 (Content-Length)
	totalSize := resp.ContentLength

	// fileInfo, err := os.Stat(fullpath)
	// if err == nil {
	// 	if fileInfo.Size() == totalSize {
	// 		runtime.EventsEmit(ctx, "wsl-output", "检测到本地文件存在且完整，跳过下载")
	// 		time.Sleep(2 * time.Second)
	// 	} else {
	// 		runtime.EventsEmit(ctx, "wsl-output", "检测到本地文件损坏,正在重新下载")
	// 		os.Remove(fullpath)
	// 		time.Sleep(2 * time.Second)
	// 	}
	// } else {

	// }
	// 创建本地文件
	out, err := os.Create(fullpath)
	if err != nil {

	}
	defer out.Close()
	// 哈希计算器
	hasher := sha256.New()
	// 数据流写入校验
	mw := io.MultiWriter(out, hasher)

	// 循环读取并计算进度
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
		os.Remove(Info.Install_Path.Path)
		runtime.EventsEmit(ctx, "wsl-error", "Sha256校验失败,请重新执行")
		return errors.New("Sha256校验失败")
	}

	runtime.EventsEmit(ctx, "wsl-output", fmt.Sprintf("下载完成,准备安装 %s", Info.Linux_Version))

	return nil
}

func parseWSLMessage(ctx context.Context, l string, Info WSLinfo) int {
	l = strings.ToLower(l)
	ToL_Version := strings.ToLower(Info.Linux_Version)
	switch {
	case strings.Contains(l, "requireselevation"):
		runtime.EventsEmit(ctx, "wsl-error", "需要权限执行WSL安装命令,检查是否给予权限")
		return 1
	case strings.Contains(l, ToL_Version):
		line, _ := Start_cmd(Info, "SeachUser")
		if strings.Contains(Reduce_Unicode(line), "default") {
			runtime.EventsEmit(ctx, "wsl-error", fmt.Sprintf("该发行版 %s 已经安装在Windows上", Info.Linux_Version))
			return 2
		} else {
			return 3
		}
	case strings.Contains(l, "invalid"):
		runtime.EventsEmit(ctx, "wsl-error", fmt.Sprintf("用户名 %s 不符合规范,请重新设置", Info.Auth.User))
		return 4
	case strings.Contains(l, "short"):
		runtime.EventsEmit(ctx, "wsl-error", fmt.Sprintf("密码 %s 设置太短", Info.Auth.Password))
		return 5
	case strings.Contains(l, "dictionary"):
		runtime.EventsEmit(ctx, "wsl-error", fmt.Sprintf("密码 %s 不符合字典规范,请重新设置", Info.Auth.Password))
		return 6
	default:
		return -1
	}
	// runtime.EventsEmit(ctx, "wsl-error", exitCode)
}

func WSL2_Installer(ctx context.Context, Info WSLinfo) error {
	runtime.EventsEmit(ctx, "wsl-output", fmt.Sprintf("正在解压安装发行版 %s ", Info.Linux_Version))
	time.Sleep(2 * time.Second)
	line, err := Start_cmd(Info, "Import")

	if err != nil {
		runtime.EventsEmit(ctx, "wsl-error", fmt.Sprintf("解压安装出现错误: %s ,报错信息: %s", err.Error(), Reduce_Unicode(line)))
		return err
	}
	runtime.EventsEmit(ctx, "wsl-output", fmt.Sprintf("安装发行版 %s 成功", Info.Linux_Version))
	return nil

}

// 配置用户名,密码函数
func WSL2_Setting_User(ctx context.Context, Info WSLinfo) error {
	line, _ := Start_cmd(Info, "ConfigUser")
	if parseWSLMessage(ctx, Reduce_Unicode(line), Info) != -1 {
		time.Sleep(2 * time.Second)
		return errors.New("用户名配置错误")
	}

	line, _ = Start_cmd(Info, "ConfigPasswd")

	if parseWSLMessage(ctx, Reduce_Unicode(line), Info) != -1 {
		time.Sleep(2 * time.Second)
		return errors.New("密码配置错误")
	}

	line, _ = Start_cmd(Info, "ConfigSudo")

	if parseWSLMessage(ctx, Reduce_Unicode(line), Info) != -1 {
		time.Sleep(2 * time.Second)
		return errors.New("无法配置用户Sudo权限")
	}

	line, _ = Start_cmd(Info, "Default")

	if parseWSLMessage(ctx, Reduce_Unicode(line), Info) != -1 {
		time.Sleep(2 * time.Second)
		return errors.New("无法配置默认用户")
	}

	line, err := Start_cmd(Info, "Stop")
	if err != nil {
		runtime.EventsEmit(ctx, "wsl-error", fmt.Sprintf("暂停发行版出现错误: %s ", err))
	}
	// 循环检测wsl发行版是否关停

	time.Sleep(2 * time.Second)

	return nil
}

// 迁移发行版
func MovingPathWSL(ctx context.Context, Info WSLinfo) error {
	runtime.EventsEmit(ctx, "migration:progress", "正在导出发行版......")
	// 导出
	line, err := Start_cmd(Info, "Export")
	if err != nil {
		runtime.EventsEmit(ctx, "migration:done", map[string]interface{}{
			"status": "failed",
			"error":  fmt.Sprintf("导出出现问题: %s", Reduce_Unicode(line)),
		})
		os.Remove(FilePath_string(Info))
		return err
	}
	// 卸载
	runtime.EventsEmit(ctx, "migration:progress", "正在卸载发行版......")
	line, err = Start_cmd(Info, "Uninstall")
	if err != nil {
		runtime.EventsEmit(ctx, "migration:done", map[string]interface{}{
			"status": "failed",
			"error":  fmt.Sprintf("卸载出现问题: %s", Reduce_Unicode(line)),
		})
		os.Remove(FilePath_string(Info))
		return err
	}
	//导入
	runtime.EventsEmit(ctx, "migration:progress", "正在迁移发行版......")
	line, err = Start_cmd(Info, "Import")
	if err != nil {
		runtime.EventsEmit(ctx, "migration:done", map[string]interface{}{
			"status": "failed",
			"error":  fmt.Sprintf("导入出现问题: %s", Reduce_Unicode(line)),
		})
		os.Remove(FilePath_string(Info))
		return err
	}
	// 大致处理下
	Start_cmd(Info, "Start")
	time.Sleep(10 * time.Second)
	// 配置用户
	runtime.EventsEmit(ctx, "migration:progress", "正在还原用户配置......")
	line, err = Start_cmd(Info, "ConfigUser")
	if err != nil {
		runtime.EventsEmit(ctx, "migration:done", map[string]interface{}{
			"status": "failed",
			"error":  fmt.Sprintf("设置账户出现问题: %s", Reduce_Unicode(line)),
		})
		return err
	}
	// 发送完成信息
	runtime.EventsEmit(ctx, "migration:done", map[string]interface{}{
		"status": "success",
	})
	return nil
}

func UninstallWSL(ctx context.Context, Info WSLinfo) error {
	line, err := Start_cmd(Info, "Uninstall")
	if err != nil {
		runtime.EventsEmit(ctx, "uninstall:failed", fmt.Sprintf("卸载 %s 发行版失败: %s", Info.Linux_Version, Reduce_Unicode(line)))
		return err
	}
	return nil
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
