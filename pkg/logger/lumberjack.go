package logger

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Logger struct {
	Filename   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
	size       int64
	file       *os.File
	mu         sync.Mutex
}

func (l *Logger) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file == nil {
		if err := l.openExistingOrNew(len(p)); err != nil {
			return 0, err
		}
	}

	if l.size+int64(len(p)) >= int64(l.MaxSize)*1024*1024 {
		if err := l.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = l.file.Write(p)
	l.size += int64(n)
	return n, err
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.close()
}

func (l *Logger) close() error {
	if l.file == nil {
		return nil
	}
	err := l.file.Close()
	l.file = nil
	return err
}

func (l *Logger) openExistingOrNew(writeLen int) error {
	l.close()
	if err := os.MkdirAll(filepath.Dir(l.Filename), 0755); err != nil {
		return err
	}

	info, err := os.Stat(l.Filename)
	if os.IsNotExist(err) {
		return l.openNew()
	}
	if err != nil {
		return err
	}

	if info.Size()+int64(writeLen) >= int64(l.MaxSize)*1024*1024 {
		return l.rotate()
	}

	file, err := os.OpenFile(l.Filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return l.openNew()
	}
	l.file = file
	l.size = info.Size()
	return nil
}

func (l *Logger) openNew() error {
	if err := os.MkdirAll(filepath.Dir(l.Filename), 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(l.Filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	l.file = file
	l.size = 0
	return nil
}

func (l *Logger) rotate() error {
	if err := l.close(); err != nil {
		return err
	}

	if err := l.backup(); err != nil {
		return err
	}

	if err := l.cleanupOldLogs(); err != nil {
		return err
	}

	return l.openNew()
}

func (l *Logger) backup() error {
	_, filename := filepath.Split(l.Filename)
	timestamp := time.Now().Format("2006-01-02T15-04-05.000")
	backupName := filename + "." + timestamp

	backupPath := filepath.Join(filepath.Dir(l.Filename), backupName)
	return os.Rename(l.Filename, backupPath)
}

func (l *Logger) cleanupOldLogs() error {
	if l.MaxAge <= 0 && l.MaxBackups <= 0 {
		return nil
	}

	files, err := filepath.Glob(l.Filename + ".*")
	if err != nil {
		return err
	}

	var backups []os.FileInfo
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		backups = append(backups, info)
	}

	if l.MaxAge > 0 {
		cutoff := time.Now().AddDate(0, 0, -l.MaxAge)
		for _, b := range backups {
			if b.ModTime().Before(cutoff) {
				os.Remove(filepath.Join(filepath.Dir(l.Filename), b.Name()))
			}
		}
	}

	return nil
}

var _ io.WriteCloser = (*Logger)(nil)
