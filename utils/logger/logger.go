package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
)

func New(isDebug bool, logPath string) (*zerolog.Logger, error) {
	logLevel := zerolog.InfoLevel
	if isDebug {
		logLevel = zerolog.TraceLevel
	}

	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	runLogFile, _ := os.OpenFile(
		logPath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)

	zerolog.SetGlobalLevel(logLevel)
	multi := zerolog.MultiLevelWriter(zerolog.ConsoleWriter{Out: os.Stdout}, runLogFile)
	logger := zerolog.New(multi).With().Timestamp().Logger()

	return &logger, nil
}
