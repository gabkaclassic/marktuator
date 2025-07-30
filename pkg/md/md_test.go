package md

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractLinks(t *testing.T) {
	files := map[string][]byte{
		"test1.md": []byte(`# Title

This is a [test link](https://example.com).
Another [link](https://golang.org) is here.`),
		"test2.md": []byte(`No links here.`),
	}

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	links := ExtractLinks(files, log)

	assert.Len(t, links, 2)

	assert.Contains(t, links, Link{
		File: "test1.md",
		Text: "test link",
		URL:  "https://example.com",
	})

	assert.Contains(t, links, Link{
		File: "test1.md",
		Text: "link",
		URL:  "https://golang.org",
	})
}

func TestReadMdFiles(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "a.md")
	file2 := filepath.Join(tmpDir, "b.md")
	nonMdFile := filepath.Join(tmpDir, "not-txt")

	os.WriteFile(file1, []byte("# Hello [Go](https://golang.org)"), 0644)
	os.WriteFile(file2, []byte("Some [text](https://example.com) here"), 0644)
	os.WriteFile(nonMdFile, []byte("Should still be read"), 0644)

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	files := ReadMdFiles(tmpDir, log)

	assert.Len(t, files, 3)

	assert.Contains(t, files, file1)
	assert.Contains(t, files, file2)
	assert.Contains(t, files, nonMdFile)

	assert.True(t, strings.Contains(string(files[file1]), "Go"))
}
