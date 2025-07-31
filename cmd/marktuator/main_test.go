package main

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gabkaclassic/marktuator/pkg/md"
	"github.com/gabkaclassic/marktuator/pkg/url_validator"
)

var testLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

type mockRoundTripper struct {
	statusCodes map[string]int
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	code := m.statusCodes[req.URL.String()]
	if code == 0 {
		code = 404
	}
	return &http.Response{
		StatusCode: code,
		Body:       http.NoBody,
		Request:    req,
	}, nil
}

func TestCheckLinks_RelativeLinks(t *testing.T) {
	tmp := t.TempDir()

	mdFile1 := filepath.Join(tmp, "doc1.md")
	mdFile2 := filepath.Join(tmp, "doc2.md")

	err := os.WriteFile(mdFile1, []byte(`[OK](doc2.md#section-ok) [FAIL](doc2.md#not-there)`), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(mdFile2, []byte(`## Section OK`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	files := md.ReadMdFiles(tmp, testLogger)
	links := md.ExtractLinks(files, testLogger)

	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(links))
	}

	cfg := url_validator.LinksValidatorConfig{
		AllowedStatuses: url_validator.PrepareAllowedStatuses(200),
		Timeout:         2 * time.Second,
	}

	results := checkLinks(links, http.Client{}, cfg, files, testLogger)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	for _, r := range results {
		if strings.Contains(r.link.Fragment, "not-there") && r.ok {
			t.Errorf("expected broken link to fail, got OK")
		}
		if strings.Contains(r.link.Fragment, "section-ok") && !r.ok {
			t.Errorf("expected valid link to succeed, got FAIL")
		}
	}
}

func TestCheckLinks_AbsoluteLinks(t *testing.T) {
	links := []md.Link{
		{
			File:       "dummy.md",
			Text:       "Valid link",
			URL:        "https://example.com",
			IsRelative: false,
		},
		{
			File:       "dummy.md",
			Text:       "Broken link",
			URL:        "https://doesnotexist.example",
			IsRelative: false,
		},
	}

	mockClient := http.Client{
		Transport: &mockRoundTripper{
			statusCodes: map[string]int{
				"https://example.com":          200,
				"https://doesnotexist.example": 404,
			},
		},
		Timeout: 1 * time.Second,
	}

	cfg := url_validator.LinksValidatorConfig{
		AllowedStatuses: url_validator.PrepareAllowedStatuses(200),
		Timeout:         1 * time.Second,
	}

	results := checkLinks(links, mockClient, cfg, nil, testLogger)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	for _, r := range results {
		if r.link.URL == "https://example.com" && !r.ok {
			t.Errorf("expected 200 OK link to pass, got fail")
		}
		if r.link.URL == "https://doesnotexist.example" && r.ok {
			t.Errorf("expected 404 link to fail, got success")
		}
	}
}
