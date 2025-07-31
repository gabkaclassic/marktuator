package md

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var testLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

func TestExtractLinks(t *testing.T) {
	content := []byte(`
# Sample Document

Here is a [relative link](doc2.md#section-2)

And here is an [absolute link](https://example.com)

And a [mailto link](mailto:test@example.com)
`)

	files := map[string][]byte{
		"doc1.md": content,
	}

	links := ExtractLinks(files, testLogger)

	if len(links) != 3 {
		t.Fatalf("expected 3 links, got %d", len(links))
	}

	tests := []struct {
		text       string
		url        string
		isRelative bool
		fragment   string
	}{
		{"relative link", "doc2.md#section-2", true, "section-2"},
		{"absolute link", "https://example.com", false, ""},
		{"mailto link", "mailto:test@example.com", false, ""},
	}

	for i, want := range tests {
		got := links[i]
		if got.Text != want.text || got.URL != want.url || got.IsRelative != want.isRelative || got.Fragment != want.fragment {
			t.Errorf("link %d: got %+v, want %+v", i, got, want)
		}
	}
}

func TestGenerateAnchor(t *testing.T) {
	cases := map[string]string{
		"Heading One":     "heading-one",
		"  Trim Me  ":     "trim-me",
		"Special!@#":      "special",
		"Tabs\tTabs":      "tabs-tabs",
		"UPPER case Text": "upper-case-text",
	}

	for input, expected := range cases {
		if got := generateAnchor(input); got != expected {
			t.Errorf("generateAnchor(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestHasMDHeader(t *testing.T) {
	content := []byte(`
# First Header
## Second Header
### Another one
`)

	tests := []struct {
		fragment string
		want     bool
	}{
		{"first-header", true},
		{"second-header", true},
		{"another-one", true},
		{"not-existing", false},
	}

	for _, tt := range tests {
		if got := hasMDHeader(tt.fragment, content, testLogger); got != tt.want {
			t.Errorf("hasMDHeader(%q) = %t, want %t", tt.fragment, got, tt.want)
		}
	}
}

func TestCheckRelativeLink(t *testing.T) {
	tempDir := t.TempDir()

	file1 := filepath.Join(tempDir, "doc1.md")
	file2 := filepath.Join(tempDir, "doc2.md")

	os.WriteFile(file1, []byte(`[Go to](doc2.md#section-two)`), 0644)
	os.WriteFile(file2, []byte(`## Section Two`), 0644)

	files := ReadMdFiles(tempDir, testLogger)
	links := ExtractLinks(files, testLogger)

	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}

	link := links[0]

	ok := CheckRelativeLink(link.URL, link.File, files, testLogger)
	if !ok {
		t.Errorf("expected link to be valid, got invalid")
	}
}

func TestReadMdFiles(t *testing.T) {
	tempDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tempDir, "test.md"), []byte("# Test"), 0644)
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	files := ReadMdFiles(tempDir, testLogger)

	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}

	for name := range files {
		if !strings.HasSuffix(name, "test.md") {
			t.Errorf("unexpected file name: %s", name)
		}
	}
}
