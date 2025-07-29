package logger

import (
	"io"
	"log/slog"
	"os"
)

type LoggerConfig struct {
	OutputToFile bool
	FilePath     string
	Level        slog.Level
	UseJSON      bool
}

func GetLogger(cfg LoggerConfig) (*slog.Logger, error) {
	var output io.Writer

	if cfg.OutputToFile {

		if cfg.FilePath == "" {
			cfg.FilePath = "./marktuator.log"
		}

		file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		output = file
	} else {
		output = os.Stdout
	}

	var handler slog.Handler
	if cfg.UseJSON {
		handler = slog.NewJSONHandler(output, &slog.HandlerOptions{
			Level: cfg.Level,
		})
	} else {
		handler = slog.NewTextHandler(output, &slog.HandlerOptions{
			Level: cfg.Level,
		})
	}

	return slog.New(handler), nil
}
