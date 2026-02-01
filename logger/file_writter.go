package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// RotateFileWriter automatically rotates the log file daily.
type RotateFileWriter struct {
	LogDir   string
	LogType  string
	file     *os.File
	currDate string
	mu       sync.Mutex
}

// Write implements io.Writer.
func (w *RotateFileWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	today := time.Now().Format("2006-01-02")
	if w.file == nil || today != w.currDate {
		if err := w.rotate(today); err != nil {
			return 0, err
		}
	}

	return w.file.Write(p)
}

// rotate closes the old file and opens a new one for today's date.
func (w *RotateFileWriter) rotate(date string) error {
	if w.file != nil {
		_ = w.file.Close()
	}

	logPath := filepath.Join(w.LogDir, fmt.Sprintf("%s.log", date))

	if err := os.MkdirAll(w.LogDir, filePermission); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	w.file = file
	w.currDate = date
	return nil
}
