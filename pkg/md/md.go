package md

import (
	"fmt"
	"io/fs"
	"log/slog"
	urls "net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type Link struct {
	File       string
	Text       string
	URL        string
	IsRelative bool
	Fragment   string
}

func (link Link) String() string {
	return fmt.Sprintf("[%s](%s%s) in file %s (relative: %t)", link.Text, link.URL, link.Fragment, link.File, link.IsRelative)
}

func ExtractLinks(files map[string][]byte, log *slog.Logger) []Link {
	md := goldmark.New()
	links := make([]Link, 0)
	log.Debug("Start parsing files")
	for file, content := range files {
		log.Debug("Creating new parser for file", slog.String("path", file))

		doc := md.Parser().Parse(text.NewReader(content))

		ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if entering {
				if link, ok := n.(*ast.Link); ok {
					var url = string(link.Destination)
					log.Debug("Extract text from link", slog.String("path", file), slog.String("url", url))
					var text = extractText(link, content)

					parsedUrl, err := urls.Parse(url)

					if err != nil {
						slog.Debug("Invalid URL found, skip", slog.String("error", err.Error()), slog.String("url", url))
						return ast.WalkContinue, nil
					}

					isRelative := !parsedUrl.IsAbs() && !strings.HasPrefix(url, "mailto:") && url != ""

					fragment := parsedUrl.Fragment

					newLink := Link{
						File:       file,
						Text:       text,
						URL:        url,
						IsRelative: isRelative,
						Fragment:   fragment,
					}
					links = append(
						links,
						newLink,
					)
					log.Debug("Found new link in file", slog.Any("link", newLink))
				}
			}
			return ast.WalkContinue, nil
		})
	}
	log.Debug("Parsing files finished")

	return links
}

func extractText(node ast.Node, content []byte) string {
	var sb strings.Builder

	var extract func(n ast.Node)
	extract = func(n ast.Node) {
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			switch c := child.(type) {
			case *ast.Text:
				sb.Write(c.Segment.Value(content))
			default:
				extract(child)
			}
		}
	}

	extract(node)
	return sb.String()
}

func ReadMdFiles(path string, log *slog.Logger) map[string][]byte {

	filesContent := make(map[string][]byte)

	filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			log.Debug("Reading directory", slog.String("path", path))
			return nil
		} else {

			content, err := os.ReadFile(path)

			if err != nil {
				log.Error("File reading error", slog.String("path", path), slog.String("error", err.Error()))
				return err
			}

			filesContent[path] = content
		}

		return nil
	})
	return filesContent
}

func CheckRelativeLink(relativeUrl string, path string, files map[string][]byte, log *slog.Logger) bool {

	u, err := urls.Parse(relativeUrl)
	if err != nil {
		log.Debug("Invalid relative URL", slog.String("url", relativeUrl), slog.String("error", err.Error()))
		return false
	}

	baseDir := filepath.Dir(path)
	targetPath := filepath.Join(baseDir, u.Path)

	content, exists := files[targetPath]

	if !exists {
		log.Info("File for relative link is not found", slog.String("path", targetPath))
		return false
	}

	if u.Fragment == "" {
		log.Debug("Fragment for relative link not found", slog.Any("link", relativeUrl), slog.String("path", targetPath))
		return true
	}

	found := hasMDHeader(u.Fragment, content, log)

	if !found {
		log.Info("Fragment for relative link is not found", slog.String("fragment", u.Fragment), slog.String("path", path), slog.String("url", relativeUrl))
	}

	return found
}

func hasMDHeader(fragment string, content []byte, log *slog.Logger) bool {
	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(content))

	found := false

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if heading, ok := n.(*ast.Heading); ok {
				text := extractText(heading, content)
				anchor := generateAnchor(text)

				if anchor == fragment {
					found = true
					return ast.WalkStop, nil
				}
			}
		}
		return ast.WalkContinue, nil
	})

	return found
}

func generateAnchor(text string) string {
	text = strings.ToLower(text)
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, " ", "-")
	text = strings.ReplaceAll(text, "\t", "-")

	var sb strings.Builder
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
