package logger_test

import (
	"log/slog"
	"marktuator/pkg/logger"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLogger_ToFile(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := logger.LoggerConfig{
		OutputToFile: true,
		FilePath:     logPath,
		Level:        slog.LevelInfo,
		UseJSON:      false,
	}

	log, err := logger.GetLogger(cfg)
	assert.NoError(t, err)

	log.Info("File log test")

	data, err := os.ReadFile(logPath)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "File log test")
}

func TestGetLogger_FileError(t *testing.T) {
	cfg := logger.LoggerConfig{
		OutputToFile: true,
		FilePath:     "/invalid_path/logfile.log",
		Level:        slog.LevelInfo,
		UseJSON:      false,
	}

	log, err := logger.GetLogger(cfg)
	assert.Nil(t, log)
	assert.Error(t, err)
}
