package main

import (
	"log/slog"
	"marktuator/pkg/logger"
	"marktuator/pkg/md"
	"marktuator/pkg/validator"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetupLogger(t *testing.T) {
	cfg := logger.LoggerConfig{
		OutputToFile: false,
		FilePath:     "",
		Level:        slog.LevelDebug,
		UseJSON:      false,
	}

	log := setupLogger(cfg)
	assert.NotNil(t, log)

	cfg.UseJSON = true
	log = setupLogger(cfg)
	assert.NotNil(t, log)

	tmpFile, err := os.CreateTemp(t.TempDir(), "logfile*.log")
	assert.NoError(t, err)
	defer tmpFile.Close()

	cfg = logger.LoggerConfig{
		OutputToFile: true,
		FilePath:     tmpFile.Name(),
		Level:        slog.LevelInfo,
		UseJSON:      false,
	}

	log = setupLogger(cfg)
	assert.NotNil(t, log)

	log.Info("Test log entry")

	content, err := os.ReadFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.Contains(t, string(content), "Test log entry")
}

func TestCheckLinks_ReturnsResults(t *testing.T) {
	okServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer okServer.Close()

	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer failServer.Close()

	links := []md.Link{
		{File: "test.md", Text: "ok", URL: okServer.URL},
		{File: "test.md", Text: "fail", URL: failServer.URL},
	}

	cfg := validator.LinksValidatorConfig{
		AllowedStatuses: validator.PrepareAllowedStatuses(200),
		Timeout:         2 * time.Second,
	}
	client := validator.GetClient(cfg)
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	results := checkLinks(links, client, cfg, log)

	assert.Len(t, results, 2)

	for _, result := range results {
		switch result.link.URL {
		case okServer.URL:
			assert.True(t, result.ok, "Expected OK for %s", result.link.URL)
		case failServer.URL:
			assert.False(t, result.ok, "Expected FAIL for %s", result.link.URL)
		default:
			t.Errorf("Unexpected URL: %s", result.link.URL)
		}
	}
}
