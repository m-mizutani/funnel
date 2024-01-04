package utils

import (
	"log/slog"
	"os"
	"sync"
)

var (
	logger       = slog.Default()
	loggerMutex  sync.Mutex
	loggerLocked bool
)

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))
}

func Logger() *slog.Logger {
	return logger
}

func FinalizeLogger(handler slog.Handler) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	if loggerLocked {
		logger.Warn("Logger is already locked, but try to update handler")
		return
	}
	logger = slog.New(handler)
	loggerLocked = true
}

func ErrLog(err error) slog.Attr { return slog.Any("error", err) }
