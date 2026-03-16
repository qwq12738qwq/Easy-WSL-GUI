//go:build !windows
// +build !windows

package Logger

import (
	"os"
	"sync"
)

type Logger struct {
	file   *os.File
	logger interface{}
	mu     sync.Mutex
	ch     chan string
}

func NewLogger() (*Logger, error) {
	return nil, nil
}

func (l *Logger) Info(format string, args ...interface{}) {}

func (l *Logger) Warning(format string, args ...interface{}) {}

func (l *Logger) Error(format string, args ...interface{}) {}

func (l *Logger) GetChan() <-chan string {
	return nil
}

func (l *Logger) Close() {}

func loaddingLogFile(appName string) (*os.File, error) {
	return nil, nil
}
