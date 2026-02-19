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
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// 安装WSL发行版信息
type WSLinfo struct {
	Linux_Version   string
	Install_Path    *WSLpath
	Auth            *WSLAuth
	DownloadThreads *WSLDownload
	DownloadInfo    *Download_WSL
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

// 删除BOM,控制字符,中文(保留空格,英文符号,换行回车)
func Reduce_Unicode(by_stream []byte) string {
	var result string
	// UTF8 & UTF16LE验证
	switch {
	case hasUTF16BOM(by_stream):
		decoded, err := io.ReadAll(transform.NewReader(bytes.NewReader(by_stream), unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()))
		if err == nil {
			result = string(decoded)
			break
		}
		fallthrough
	case tryUTF16LE(by_stream):
		decoded, err := io.ReadAll(transform.NewReader(bytes.NewReader(by_stream), unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()))
		if err == nil {
			result = string(decoded)
			break
		}
		fallthrough
	default:
		if utf8.Valid(by_stream) {
			result = string(by_stream)
		} else {
			result = string(by_stream)
		}
	}

	// 去除不可见的 BOM 头、回车符 \r 和多余空格
	result = strings.ReplaceAll(result, "\uFEFF", "") // 去除 UTF-16 BOM
	result = strings.ReplaceAll(result, "\r", "")     // 统一换行符
	result = strings.TrimSpace(result)                // 去除首尾空白

	return result
}

func hasUTF16BOM(b []byte) bool {
	return len(b) >= 2 && ((b[0] == 0xFF && b[1] == 0xFE) || (b[0] == 0xFE && b[1] == 0xFF))
}

func tryUTF16LE(b []byte) bool {
	if len(b) < 4 {
		return false
	}
	zeros := 0
	odd := 0
	for i := 1; i < len(b); i += 2 {
		odd++
		if b[i] == 0x00 {
			zeros++
		}
	}
	return odd > 0 && float64(zeros)/float64(odd) > 0.3
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
			"wsl.exe", "-d", Info.Linux_Version, "-u", "root", "--",
			"sh", "-c",
			fmt.Sprintf(
				"useradd -m -s /bin/bash %s ",
				Info.Auth.User,
			),
		), nil
	case "ConfigPasswd":
		return exec.Command(
			"wsl.exe", "-d", Info.Linux_Version, "-u", "root", "--",
			"sh", "-c",
			fmt.Sprintf(
				"echo '%s:%s' | chpasswd ",
				Info.Auth.User, Info.Auth.Password,
			),
		), nil
	case "ConfigSudo":
		if Info.Linux_Version == "Arch" {
			return exec.Command(
				"wsl.exe", "-d", Info.Linux_Version, "-u", "root", "--",
				"sh", "-c",
				fmt.Sprintf(
					"usermod -aG wheel  %s",
					Info.Auth.User),
			), nil
		}
		return exec.Command(
			"wsl.exe", "-d", Info.Linux_Version, "-u", "root", "--",
			"sh", "-c",
			fmt.Sprintf(
				"usermod -aG sudo %s",
				Info.Auth.User),
		), nil
	case "Default":
		return exec.Command(
			"wsl.exe", "-d", Info.Linux_Version, "-u", "root", "--",
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
	case "DistroVersion":
		return exec.Command(
			"wsl.exe", "-d", Info.Linux_Version, "-u", "root", "sh", "-c", `cat /etc/os-release | grep -E "^(NAME|VERSION_ID)=`,
		), nil
	case "OldSources":
		return exec.Command(
			"wsl.exe", "-d", Info.Linux_Version, "-u", "root", "sh", "-c", `cat /etc/apt/sources.list`,
		), nil
	case "Sources":
		return exec.Command(
			"wsl.exe", "-d", Info.Linux_Version, "-u", "root", "sh", "-c", `cat /etc/os-release | grep -E "^(NAME|VERSION_ID)=`,
		), nil
	case "UserGroups":
		return exec.Command(
			"wsl.exe", "-d", Info.Linux_Version, "-u", "root", "--",
			"sh", "-c",
			"cat /etc/group",
		), nil
	case "UserList":
		return exec.Command(
			"wsl.exe", "-d", Info.Linux_Version, "-u", "root", "--",
			"sh", "-c",
			"cat /etc/passwd",
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
	// 根据DownloadThreads和DownloadInfo是否是空指针判断是安装还是迁移
	if Info.DownloadInfo != nil && Info.DownloadThreads != nil {
		// 文件名拼凑
		if strings.Contains(Info.DownloadInfo.URL, ".wsl") {
			fileName += ".wsl"
		}
	} else {
		// 迁移后缀
		fileName += ".tar"
	}

	return filepath.Join(Info.Install_Path.Path, fileName)

}

// 多线程下载+重试
func WSL2_Downloader(ctx context.Context, Info WSLinfo) error {

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

	// 发起初始请求以获取长度
	resp, err := http.Get(Info.DownloadInfo.URL)
	if err != nil {
		runtime.EventsEmit(ctx, "wsl-error", fmt.Sprintf("下载 %s 发行版失败,请检查网络连接", Info.Linux_Version))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		runtime.EventsEmit(ctx, "wsl-error", fmt.Sprintf("下载失败,网络错误码: %s ", strconv.Itoa(resp.StatusCode)))
		return errors.New("网络错误")
	}

	totalSize := resp.ContentLength

	out, err := os.Create(fullpath)
	if err != nil {
		runtime.EventsEmit(ctx, "wsl-error", "创建本地文件失败")
		return err
	}
	defer out.Close()

	hasher := sha256.New()
	var downloaded int64

	// 并行下载(分片 Range 请求) + 重试3次，失败取消下载
	if totalSize <= 0 {
		runtime.EventsEmit(ctx, "wsl-error", "无法获取镜像大小，取消下载")
		return errors.New("未知长度")
	}

	threads := 4
	if Info.DownloadThreads != nil && Info.DownloadThreads.DownloadThreads > 0 {
		threads = Info.DownloadThreads.DownloadThreads
	}

	if err := out.Truncate(totalSize); err != nil {
		runtime.EventsEmit(ctx, "wsl-error", "预分配文件空间失败")
		return err
	}

	downloadCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	client := &http.Client{}
	url := Info.DownloadInfo.URL

	var wg sync.WaitGroup
	errCh := make(chan error, threads)

	// 进度上报协程
	doneProgress := make(chan struct{})
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-doneProgress:
				return
			case <-ticker.C:
				if totalSize > 0 {
					percent := float64(atomic.LoadInt64(&downloaded)) / float64(totalSize) * 100
					runtime.EventsEmit(ctx, "wsl-output", fmt.Sprintf("正在下载镜像: %.2f%%", percent))
				}
			}
		}
	}()

	chunkSize := totalSize / int64(threads)
	for i := 0; i < threads; i++ {
		start := int64(i) * chunkSize
		end := start + chunkSize - 1
		if i == threads-1 {
			end = totalSize - 1
		}

		wg.Add(1)
		go func(s, e int64) {
			defer wg.Done()
			retries := 0
			for {
				req, _ := http.NewRequestWithContext(downloadCtx, "GET", url, nil)
				req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", s, e))
				r, reqErr := client.Do(req)
				if reqErr != nil {
					retries++
					if retries >= 3 {
						errCh <- reqErr
						cancel()
						return
					}
					time.Sleep(500 * time.Millisecond)
					continue
				}

				// 206 Partial Content 是 Range 成功；允许特例：完整文件且是首片
				if r.StatusCode != http.StatusPartialContent && !(r.StatusCode == http.StatusOK && s == 0 && e == totalSize-1) {
					r.Body.Close()
					retries++
					if retries >= 3 {
						errCh <- fmt.Errorf("HTTP状态码: %d", r.StatusCode)
						cancel()
						return
					}
					time.Sleep(500 * time.Millisecond)
					continue
				}

				// 读取并写入到指定偏移
				offset := s
				buf := make([]byte, 64*1024)
				for {
					n, re := r.Body.Read(buf)
					if n > 0 {
						if _, we := out.WriteAt(buf[:n], offset); we != nil {
							r.Body.Close()
							errCh <- we
							cancel()
							return
						}
						atomic.AddInt64(&downloaded, int64(n))
						offset += int64(n)
					}
					if re != nil {
						r.Body.Close()
						if re == io.EOF {
							break
						}
						retries++
						if retries >= 3 {
							errCh <- re
							cancel()
							return
						}
						time.Sleep(500 * time.Millisecond)
						// 重试会重新发起请求
						continue
					}
				}
				// 正常完成
				break
			}
		}(start, end)
	}

	wg.Wait()
	close(doneProgress)

	select {
	case e := <-errCh:
		os.Remove(fullpath)
		runtime.EventsEmit(ctx, "wsl-error", fmt.Sprintf("下载失败并已取消: %s", e))
		return e
	default:
	}

	// 下载完成后进行哈希计算
	hasher.Reset()
	if _, err := out.Seek(0, 0); err == nil {
		if _, err := io.Copy(hasher, out); err != nil {
			runtime.EventsEmit(ctx, "wsl-error", "计算哈希失败")
			return err
		}
	} else {
		f, ferr := os.Open(fullpath)
		if ferr != nil {
			runtime.EventsEmit(ctx, "wsl-error", "打开文件计算哈希失败")
			return ferr
		}
		defer f.Close()
		if _, err := io.Copy(hasher, f); err != nil {
			runtime.EventsEmit(ctx, "wsl-error", "计算哈希失败")
			return err
		}
	}

	actualSha256 := hex.EncodeToString(hasher.Sum(nil))

	// 校验比较
	if actualSha256 != Info.DownloadInfo.Sha256 {
		// 如果校验失败，删除残缺文件
		os.Remove(fullpath)
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
	runtime.EventsEmit(ctx, "wsl-output", fmt.Sprintf("安装发行版 %s 完成", Info.Linux_Version))
	time.Sleep(2 * time.Second)
	runtime.EventsEmit(ctx, "wsl-output", "正在删除下载残留......")
	os.Remove(FilePath_string(Info))
	time.Sleep(2 * time.Second)
	return nil

}

// 配置用户名,密码函数
func WSL2_Setting_User(ctx context.Context, Info WSLinfo) error {
	runtime.EventsEmit(ctx, "wsl-output", "正在配置用户......")
	line, _ := Start_cmd(Info, "ConfigUser")
	if parseWSLMessage(ctx, Reduce_Unicode(line), Info) != -1 {
		time.Sleep(2 * time.Second)
		return errors.New("用户名配置错误")
	}
	time.Sleep(2 * time.Second)
	runtime.EventsEmit(ctx, "wsl-output", "正在配置密码......")
	line, _ = Start_cmd(Info, "ConfigPasswd")

	if parseWSLMessage(ctx, Reduce_Unicode(line), Info) != -1 {
		time.Sleep(2 * time.Second)
		return errors.New("密码配置错误")
	}
	time.Sleep(2 * time.Second)
	runtime.EventsEmit(ctx, "wsl-output", "正在配置权限......")
	line, _ = Start_cmd(Info, "ConfigSudo")

	if parseWSLMessage(ctx, Reduce_Unicode(line), Info) != -1 {
		time.Sleep(2 * time.Second)
		return errors.New("无法配置用户Sudo权限")
	}
	time.Sleep(2 * time.Second)
	runtime.EventsEmit(ctx, "wsl-output", "正在设置默认账户......")
	line, _ = Start_cmd(Info, "Default")

	if parseWSLMessage(ctx, Reduce_Unicode(line), Info) != -1 {
		time.Sleep(2 * time.Second)
		return errors.New("无法配置默认用户")
	}
	time.Sleep(2 * time.Second)
	line, err := Start_cmd(Info, "Stop")
	if err != nil {
		runtime.EventsEmit(ctx, "wsl-error", fmt.Sprintf("暂停发行版出现错误: %s ", err))
	}

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
	time.Sleep(2 * time.Second)
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
	time.Sleep(2 * time.Second)
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
	line, err = Start_cmd(Info, "Default")
	if err != nil {
		runtime.EventsEmit(ctx, "migration:done", map[string]interface{}{
			"status": "failed",
			"error":  fmt.Sprintf("设置账户出现问题: %s", Reduce_Unicode(line)),
		})
		return err
	}
	runtime.EventsEmit(ctx, "migration:progress", "删除迁移残留......")
	os.Remove(FilePath_string(Info))
	time.Sleep(3 * time.Second)
	Start_cmd(Info, "Stop")
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
