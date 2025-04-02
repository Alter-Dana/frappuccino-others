package utils

import (
	"log/slog"
	"os"
	"sync"
)

var (
	logger *slog.Logger
	once   sync.Once
)

// GetLogger returns a singleton instance of the logger
func GetLogger() *slog.Logger {
	once.Do(func() {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	})
	return logger
}
