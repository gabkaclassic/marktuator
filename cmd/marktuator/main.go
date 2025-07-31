package main

import (
	"fmt"
	"github.com/gabkaclassic/marktuator/internal/config"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gabkaclassic/marktuator/pkg/logger"
	"github.com/gabkaclassic/marktuator/pkg/md"
	"github.com/gabkaclassic/marktuator/pkg/url_validator"
)

func main() {
	cfg := config.ParseConfig()
	log := setupLogger(cfg.Logger)
	log.Debug("Start marktuator")

	log.Debug("Read md files form", slog.String("filepath", cfg.TargetPath))
	content := md.ReadMdFiles(cfg.TargetPath, log)

	log.Debug("Extract links from MD files")
	listLinks := md.ExtractLinks(content, log)

	log.Debug("Create HTTP client for url_validator")
	client := url_validator.GetClient(cfg.Validator)

	log.Debug("Check links for available")
	checkLinks(listLinks, client, cfg.Validator, content, log)

	log.Debug("Marktuator finished")
}

func setupLogger(cfg logger.LoggerConfig) *slog.Logger {

	log, err := logger.GetLogger(cfg)

	if err != nil {
		panic(err)
	}

	return log
}

type CheckResult struct {
	link *md.Link
	ok   bool
}

func checkLinks(
	linksList []md.Link,
	client http.Client,
	cfg url_validator.LinksValidatorConfig,
	files map[string][]byte,
	log *slog.Logger,
) []CheckResult {
	var resultsWg sync.WaitGroup
	resultsCh := make(chan CheckResult, len(linksList))

	for _, link := range linksList {
		resultsWg.Add(1)
		go func(l md.Link) {
			defer resultsWg.Done()
			var ok bool
			if l.IsRelative {
				ok = md.CheckRelativeLink(l.URL, l.File, files, log)
			} else {
				ok = url_validator.CheckLink(l.URL, client, cfg, log)
			}
			resultsCh <- CheckResult{link: &l, ok: ok}
		}(link)
	}

	resultsWg.Wait()
	close(resultsCh)

	var results []CheckResult
	for result := range resultsCh {
		if result.ok {
			log.Debug("Link available:", slog.Any("link", result.link))
		} else {
			fmt.Printf("Link unavailable: %s\n", result.link)
			log.Info("Link unavailable:", slog.Any("link", result.link))
		}
		results = append(results, result)
	}

	log.Info("All links checked")
	return results
}
