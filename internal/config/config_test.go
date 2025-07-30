package config_test

import (
	"log/slog"
	"marktuator/internal/config"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseValidatorConfig_Valid(t *testing.T) {
	validatorCfg := config.ParseValidatorConfig(5, "200, 201, 302")

	assert.Equal(t, 5*time.Second, validatorCfg.Timeout)
	assert.Contains(t, validatorCfg.AllowedStatuses, 200)
	assert.Contains(t, validatorCfg.AllowedStatuses, 201)
	assert.Contains(t, validatorCfg.AllowedStatuses, 302)
}

func TestParseLoggerConfig(t *testing.T) {
	cfg := config.ParseLoggerConfig("", "debug", true)
	assert.False(t, cfg.OutputToFile)
	assert.Equal(t, slog.LevelDebug, cfg.Level)
	assert.True(t, cfg.UseJSON)

	cfg = config.ParseLoggerConfig("log.txt", "warn", false)
	assert.True(t, cfg.OutputToFile)
	assert.Equal(t, "log.txt", cfg.FilePath)
	assert.Equal(t, slog.LevelWarn, cfg.Level)
	assert.False(t, cfg.UseJSON)

	cfg = config.ParseLoggerConfig("", "invalid-level", false)
	assert.Equal(t, slog.LevelInfo, cfg.Level)
}

func TestParseConfig_Success(t *testing.T) {
	os.Args = []string{
		"cmd",
		"-timeout=10",
		"-status=200,404",
		"-log=log.txt",
		"-level=error",
		"-json=true",
		"-path=/some/path",
	}

	cfg := config.ParseConfig()

	assert.Equal(t, 10*time.Second, cfg.Validator.Timeout)
	assert.Contains(t, cfg.Validator.AllowedStatuses, 200)
	assert.Contains(t, cfg.Validator.AllowedStatuses, 404)

	assert.True(t, cfg.Logger.OutputToFile)
	assert.Equal(t, "log.txt", cfg.Logger.FilePath)
	assert.Equal(t, slog.LevelError, cfg.Logger.Level)
	assert.True(t, cfg.Logger.UseJSON)

	assert.Equal(t, "/some/path", cfg.TargetPath)
}

func TestHelperMissingPath(t *testing.T) {
	if os.Getenv("TEST_MISSING_PATH") != "1" {
		return
	}

	os.Args = []string{
		"cmd",
		"-timeout=3",
		"-status=200",
	}

	config.ParseConfig()
}
