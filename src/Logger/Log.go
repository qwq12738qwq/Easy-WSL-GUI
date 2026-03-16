//go:build windows
// +build windows

package Logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger 日志记录器结构体
// 包含文件指针、互斥锁、日志通道
type Logger struct {
	file   *os.File
	logger *log.Logger // 自定义 logger，支持自定义前缀
	mu     sync.Mutex
	ch     chan string
}

// NewLogger 创建日志记录器
// 步骤：
// 1. 调用 loaddingLogFile("Easy-WSL-GUI") 获取已打开的日志文件
// 2. 创建自定义 logger，设置前缀格式为 "[时间戳] [级别]"
// 3. 初始化带缓冲的通道，用于实时推送日志给前端
func NewLogger() (*Logger, error) {
	// 步骤1: 调用已存在的 loaddingLogFile 函数获取文件
	// 参数 "Easy-WSL-GUI" 是应用名称，会自动创建目录: %APPDATA%/Easy-WSL-GUI/logs/log.txt
	file_pth, err := loaddingLogFile("Easy-WSL-GUI")
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %v", err)
	}

	// 步骤2: 创建自定义 logger
	// 参数: 输出文件, 前缀(留空，我们在日志方法中自定义), 日志选项(同时输出到文件和控制台)
	// log.LstdFlags 添加标准时间戳，我们后面自己添加更详细的
	logger := log.New(file_pth, "", log.LstdFlags|log.Lshortfile)

	// 步骤3: 初始化通道，缓冲大小 100，防止阻塞
	return &Logger{
		file:   file_pth,
		logger: logger,
		ch:     make(chan string, 100),
	}, nil
}

// getTimestamp 获取当前时间戳字符串
// 格式: 2024-01-01 12:00:00
func getTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// formatMessage 格式化日志消息
// 输入: 级别, 格式, 参数
// 输出: [时间] [级别] 消息内容
func formatMessage(level, format string, args ...interface{}) string {
	msg := fmt.Sprintf(format, args...)
	return fmt.Sprintf("[%s] [%s] %s", getTimestamp(), level, msg)
}

// Info 写入 Info 级别日志
// 用途: 记录一般信息，如操作成功、状态变更等
// 1. 格式化消息，添加时间戳和 [INFO] 标记
// 2. 加锁保证线程安全
// 3. 写入文件和通道
func (l *Logger) Info(format string, args ...interface{}) {
	msg := formatMessage("INFO", format, args...)

	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Println(msg)

	// 推送到通道供前端实时读取
	select {
	case l.ch <- msg:
	default:
		// 通道满则丢弃，防止阻塞主线程
	}

	// 打印到控制台
	fmt.Println(msg)
}

// Warning 写入 Warning 级别日志
// 用途: 记录警告信息，如配置缺失、预期外的行为等
// 实现同 Info，只是级别标记为 [WARNING]
func (l *Logger) Warning(format string, args ...interface{}) {
	msg := formatMessage("WARNING", format, args...)

	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Println(msg)

	select {
	case l.ch <- msg:
	default:
	}

	fmt.Println(msg)
}

// Error 写入 Error 级别日志
// 用途: 记录错误信息，如操作失败、异常等
// 实现同 Info，只是级别标记为 [ERROR]
func (l *Logger) Error(format string, args ...interface{}) {
	msg := formatMessage("ERROR", format, args...)

	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Println(msg)

	select {
	case l.ch <- msg:
	default:
	}

	fmt.Println(msg)
}

// GetChan 获取日志通道
// 用途: 供外部（如 app.go）监听实时日志
// 返回: 只读通道，外部用 for range 读取
func (l *Logger) GetChan() <-chan string {
	return l.ch
}

// Close 关闭日志记录器
// 用途: 程序退出时调用，释放资源
// 1. 加锁
// 2. 关闭文件
// 3. 关闭通道
func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		l.file.Close()
	}
	close(l.ch)
}

// 读取日志文件(带文件存在检测)
// 此函数由用户提供，不做任何修改
// 用途: 获取日志文件句柄，自动创建目录和文件
// 参数: appName - 应用名称，用于构建路径
// 返回: *os.File - 打开的文件指针, error - 错误信息
func loaddingLogFile(appName string) (*os.File, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	logDir := filepath.Join(userConfigDir, appName, "logs")

	// 创建目录
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}
	// 拼接文件名路径
	logFilePath := filepath.Join(logDir, "log.txt")
	file_pth, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}

	return file_pth, nil
}
