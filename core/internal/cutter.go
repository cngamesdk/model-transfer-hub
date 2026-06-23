package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Cutter 日志切割器
type Cutter struct {
	level        string   // 日志级别
	director     string   // 日志文件夹
	file         *os.File // 文件句柄
	mutex        sync.Mutex
	maxSize      int64 // 最大文件大小（字节）
	currentSize  int64 // 当前文件大小
	retentionDay int   // 保留天数
}

type CutterOption func(*Cutter)

// WithMaxSize 设置最大文件大小
func WithMaxSize(size int64) CutterOption {
	return func(c *Cutter) {
		c.maxSize = size
	}
}

// NewCutter 创建日志切割器
func NewCutter(director, level string, retentionDay int, opts ...CutterOption) *Cutter {
	c := &Cutter{
		director:     director,
		level:        level,
		maxSize:      100 * 1024 * 1024, // 默认100MB
		retentionDay: retentionDay,
	}

	for _, opt := range opts {
		opt(c)
	}

	// 清理旧日志
	go c.cleanOldLogs()

	return c
}

// Write 实现io.Writer接口
func (c *Cutter) Write(p []byte) (n int, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 检查是否需要切割
	if c.file != nil && c.currentSize+int64(len(p)) > c.maxSize {
		c.file.Close()
		c.file = nil
	}

	// 打开或创建文件
	if c.file == nil {
		if err := c.openFile(); err != nil {
			return 0, err
		}
	}

	// 写入数据
	n, err = c.file.Write(p)
	c.currentSize += int64(n)

	return n, err
}

// openFile 打开日志文件
func (c *Cutter) openFile() error {
	// 基础文件名
	baseName := filepath.Join(c.director, c.level+".log")

	// 检查文件是否存在及大小
	info, err := os.Stat(baseName)
	if err == nil && info.Size() < c.maxSize {
		// 文件存在且未超过大小限制，追加写入
		c.file, err = os.OpenFile(baseName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		c.currentSize = info.Size()
		return nil
	}

	// 文件不存在或已超过大小限制，需要轮转
	if err == nil {
		// 重命名现有文件
		timestamp := time.Now().Format("150405")
		newName := filepath.Join(c.director, fmt.Sprintf("%s_%s.log", c.level, timestamp))
		os.Rename(baseName, newName)
	}

	// 创建新文件
	c.file, err = os.OpenFile(baseName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	c.currentSize = 0
	return nil
}

// cleanOldLogs 清理旧日志
func (c *Cutter) cleanOldLogs() {
	if c.retentionDay <= 0 {
		return
	}

	// 获取日志根目录
	rootDir := filepath.Dir(c.director)

	// 遍历日志目录
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return
	}

	cutoffDate := time.Now().AddDate(0, 0, -c.retentionDay)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// 尝试解析目录名为日期
		dirDate, err := time.Parse("2006-01-02", entry.Name())
		if err != nil {
			continue
		}

		// 删除过期日志
		if dirDate.Before(cutoffDate) {
			dirPath := filepath.Join(rootDir, entry.Name())
			os.RemoveAll(dirPath)
			fmt.Printf("清理过期日志目录: %s\n", dirPath)
		}
	}
}

// Sync 实现zapcore.WriteSyncer接口
func (c *Cutter) Sync() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.file != nil {
		return c.file.Sync()
	}
	return nil
}
