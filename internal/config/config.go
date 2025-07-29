package config

import (
	"flag"
	"log/slog"
	"marktuator/pkg/logger"
	"marktuator/pkg/validator"
	"os"
	"strconv"
	"strings"
	"time"
)

type AppConfig struct {
	Validator  validator.LinksValidatorConfig
	Logger     logger.LoggerConfig
	TargetPath string
}

func ParseConfig() AppConfig {
	var cfg AppConfig

	timeout := flag.Int("timeout", 3, "Timeout in seconds for HTTP requests")
	statuses := flag.String("status", "200", "Comma-separated list of allowed HTTP status codes")

	logFile := flag.String("log", "", "Path to log file (default: stdout)")
	logLevel := flag.String("level", "info", "Log level (debug, info, warn, error)")
	useJSON := flag.Bool("json", false, "Use JSON log format")

	targetPath := flag.String("path", "", "Path to file or directory (required)")

	flag.Parse()

	cfg.Validator = ParseValidatorConfig(*timeout, *statuses)

	cfg.Logger = ParseLoggerConfig(*logFile, *logLevel, *useJSON)

	if *targetPath == "" {
		slog.Error("Target path is required")
		flag.Usage()
		os.Exit(1)
	}
	cfg.TargetPath = *targetPath

	return cfg
}

func ParseValidatorConfig(timeout int, statusStr string) validator.LinksValidatorConfig {
	statuses := strings.Split(statusStr, ",")
	allowedStatuses := make([]int, 0, len(statuses))

	for _, s := range statuses {
		status, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			slog.Error("Invalid status code", "code", s, "error", err)
			os.Exit(1)
		}
		allowedStatuses = append(allowedStatuses, status)
	}

	return validator.LinksValidatorConfig{
		AllowedStatuses: validator.PrepareAllowedStatuses(allowedStatuses...),
		Timeout:         time.Duration(timeout) * time.Second,
	}
}

func ParseLoggerConfig(logFile, logLevel string, useJSON bool) logger.LoggerConfig {
	var level slog.Level
	switch strings.ToLower(logLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	return logger.LoggerConfig{
		OutputToFile: logFile != "",
		FilePath:     logFile,
		Level:        level,
		UseJSON:      useJSON,
	}
}
