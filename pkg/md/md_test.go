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

func TestExtractLinks_Relative(t *testing.T) {
	content := []byte("See [intro](readme.md#start) and [guide](guide.md)")
	files := map[string][]byte{
		"test.md": content,
	}

	links := ExtractLinks(files, testLogger)

	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(links))
	}

	if !links[0].IsRelative || links[0].Fragment != "start" || links[0].URL != "readme.md#start" && links[0].Text != "intro" {
		t.Errorf("unexpected first link: %+v", links[0])
	}

	if !links[1].IsRelative || links[1].Fragment != "" || links[1].URL != "guide.md" {
		t.Errorf("unexpected second link: %+v", links[1])
	}
}

func TestExtractLinks_Absolute(t *testing.T) {
	content := []byte("External: [Google](https://google.com)")
	files := map[string][]byte{
		"doc.md": content,
	}

	links := ExtractLinks(files, testLogger)

	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}

	link := links[0]
	if link.URL != "https://google.com" || link.IsRelative {
		t.Errorf("expected absolute link, got %+v", link)
	}
}

func TestExtractLinks_Mailto(t *testing.T) {
	content := []byte("[Email me](mailto:hello@example.com)")
	files := map[string][]byte{
		"contact.md": content,
	}

	links := ExtractLinks(files, testLogger)

	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}

	link := links[0]
	if link.URL != "mailto:hello@example.com" || link.IsRelative {
		t.Errorf("expected mailto link to be absolute, got %+v", link)
	}
}

func TestExtractLinks_InvalidURL(t *testing.T) {
	content := []byte("[bad](::invalid::)")
	files := map[string][]byte{
		"invalid.md": content,
	}

	links := ExtractLinks(files, testLogger)

	if len(links) != 0 {
		t.Fatalf("expected 0 links, got %d", len(links))
	}
}

func TestExtractLinks_EmptyURL(t *testing.T) {
	content := []byte("[empty]()")
	files := map[string][]byte{
		"empty.md": content,
	}

	links := ExtractLinks(files, testLogger)

	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}

	link := links[0]
	if link.URL != "" {
		t.Errorf("expected empty relative link, got %+v", link)
	}
}

func TestExtractLinks_MultipleFiles(t *testing.T) {
	files := map[string][]byte{
		"file1.md": []byte("[One](a.md)"),
		"file2.md": []byte("[Two](b.md#frag)"),
	}

	links := ExtractLinks(files, testLogger)

	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(links))
	}

	if links[0].File == links[1].File {
		t.Errorf("expected different source files, got same: %s", links[0].File)
	}
}

func TestExtractLinks_TextExtraction(t *testing.T) {
	content := []byte("[**Bold Link**](link.md)")
	files := map[string][]byte{
		"text.md": content,
	}

	links := ExtractLinks(files, testLogger)

	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if !strings.Contains(links[0].Text, "Bold Link") {
		t.Errorf("expected link text to include 'Bold Link', got: %s", links[0].Text)
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

func TestCheckRelativeLink_ValidWithFragment(t *testing.T) {
	dir := t.TempDir()

	targetFile := filepath.Join(dir, "doc2.md")
	originFile := filepath.Join(dir, "doc1.md")

	err := os.WriteFile(targetFile, []byte("## My Section"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	files := map[string][]byte{
		targetFile: []byte("## My Section"),
	}

	ok := CheckRelativeLink("doc2.md#my-section", originFile, files, testLogger)
	if !ok {
		t.Errorf("expected valid link to succeed")
	}
}

func TestCheckRelativeLink_ValidNoFragment(t *testing.T) {
	dir := t.TempDir()

	targetFile := filepath.Join(dir, "doc2.md")
	originFile := filepath.Join(dir, "doc1.md")

	files := map[string][]byte{
		targetFile: []byte("Some content"),
	}

	ok := CheckRelativeLink("doc2.md", originFile, files, testLogger)
	if !ok {
		t.Errorf("expected valid link with no fragment to succeed")
	}
}

func TestCheckRelativeLink_FileNotFound(t *testing.T) {
	dir := t.TempDir()

	originFile := filepath.Join(dir, "doc1.md")
	files := map[string][]byte{} // no files

	ok := CheckRelativeLink("doc2.md#header", originFile, files, testLogger)
	if ok {
		t.Errorf("expected link to fail due to missing file")
	}
}

func TestCheckRelativeLink_FragmentNotFound(t *testing.T) {
	dir := t.TempDir()

	targetFile := filepath.Join(dir, "doc2.md")
	originFile := filepath.Join(dir, "doc1.md")

	files := map[string][]byte{
		targetFile: []byte("## Some Other Section"),
	}

	ok := CheckRelativeLink("doc2.md#missing", originFile, files, testLogger)
	if ok {
		t.Errorf("expected link to fail due to missing fragment")
	}
}

func TestCheckRelativeLink_InvalidURL(t *testing.T) {
	dir := t.TempDir()

	originFile := filepath.Join(dir, "doc1.md")

	files := map[string][]byte{}

	ok := CheckRelativeLink("://badurl", originFile, files, testLogger)
	if ok {
		t.Errorf("expected link to fail due to invalid URL")
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

func TestReadMdFiles_Basic(t *testing.T) {
	tmp := t.TempDir()

	path := filepath.Join(tmp, "test.md")
	err := os.WriteFile(path, []byte("# Hello"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	files := ReadMdFiles(tmp, testLogger)

	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}

	content, ok := files[path]
	if !ok || !strings.Contains(string(content), "Hello") {
		t.Errorf("file content incorrect or missing")
	}
}

func TestReadMdFiles_SkipDirectory(t *testing.T) {
	tmp := t.TempDir()

	_ = os.Mkdir(filepath.Join(tmp, "subdir"), 0755)
	err := os.WriteFile(filepath.Join(tmp, "file.md"), []byte("# Doc"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	files := ReadMdFiles(tmp, testLogger)

	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}
}

func TestReadMdFiles_ErrorReadingFile(t *testing.T) {
	tmp := t.TempDir()

	badFile := filepath.Join(tmp, "bad.md")

	if err := os.WriteFile(badFile, []byte("x"), 0000); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(badFile, 0644)

	files := ReadMdFiles(tmp, testLogger)

	_, ok := files[badFile]
	if ok {
		t.Errorf("expected bad file to be skipped due to read error")
	}
}
