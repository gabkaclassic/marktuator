package url_validator

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPrepareAllowedStatuses(t *testing.T) {
	statuses := PrepareAllowedStatuses(200, 201, 202)

	assert.Contains(t, statuses, 200)
	assert.Contains(t, statuses, 201)
	assert.Contains(t, statuses, 202)
	assert.NotContains(t, statuses, 404)
}

func TestGetClient(t *testing.T) {
	config := LinksValidatorConfig{
		Timeout: 2 * time.Second,
	}
	client := GetClient(config)

	assert.Equal(t, 2*time.Second, client.Timeout)
}

func TestCheckLink_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	config := LinksValidatorConfig{
		AllowedStatuses: PrepareAllowedStatuses(200),
		Timeout:         2 * time.Second,
	}
	client := GetClient(config)

	ok := CheckLink(ts.URL, client, config, log)

	assert.True(t, ok)
}

func TestCheckLink_NotAllowedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	config := LinksValidatorConfig{
		AllowedStatuses: PrepareAllowedStatuses(200, 201),
		Timeout:         2 * time.Second,
	}
	client := GetClient(config)

	ok := CheckLink(ts.URL, client, config, log)

	assert.False(t, ok)
}

func TestCheckLink_ConnectionError(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	config := LinksValidatorConfig{
		AllowedStatuses: PrepareAllowedStatuses(200),
		Timeout:         1 * time.Second,
	}
	client := GetClient(config)

	ok := CheckLink("http://localhost:0123456789", client, config, log)

	assert.False(t, ok)
}
