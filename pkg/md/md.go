package md

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type Link struct {
	File string
	Text string
	URL  string
}

func LinkToString(link *Link) string {
	return fmt.Sprintf("[%s](%s) in file %s", link.Text, link.URL, link.File)
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
					log.Debug("Found new link in file", slog.String("path", file), slog.String("url", url), slog.String("text", text))

					links = append(
						links,
						Link{
							File: file,
							Text: text,
							URL:  url,
						},
					)
				}
			}
			return ast.WalkContinue, nil
		})
	}
	log.Debug("Parsing files finished")

	return links
}

func extractText(link *ast.Link, content []byte) string {
	var text string

	for child := link.FirstChild(); child != nil; child = child.NextSibling() {
		if txt, ok := child.(*ast.Text); ok {
			text += string(txt.Value(content))
		}
	}

	return text
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
