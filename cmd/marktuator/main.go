package main

import (
	"fmt"
	"log/slog"
	"marktuator/internal/config"
	"marktuator/pkg/logger"
	"marktuator/pkg/md"
	"marktuator/pkg/validator"
	"net/http"
	"sync"
)

func main() {
	cfg := config.ParseConfig()
	log := setupLogger(cfg.Logger)
	log.Debug("Start marktuator")

	log.Debug("Read md files form", slog.String("filepath", cfg.TargetPath))
	content := md.ReadMdFiles(cfg.TargetPath, log)

	log.Debug("Extract links from MD files")
	listLinks := md.ExtractLinks(content, log)

	log.Debug("Create HTTP client for validator")
	client := validator.GetClient(cfg.Validator)

	log.Debug("Check links for available")
	CheckLinks(listLinks, client, cfg.Validator, log)

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

func CheckLinks(linksList []md.Link, client http.Client, cfg validator.LinksValidatorConfig, log *slog.Logger) {

	var resultsWg sync.WaitGroup
	var handlerWg sync.WaitGroup

	var results = make(chan CheckResult, len(linksList))

	handlerWg.Add(1)
	go func() {
		defer handlerWg.Done()
		for result := range results {
			if result.ok {
				log.Debug("Link available:", slog.Any("link", result.link))
			} else {
				fmt.Printf("Link unavailable: %s\n", result.link)
				log.Info("Link unavailable:", slog.Any("link", result.link))
			}
		}
	}()

	for _, link := range linksList {
		resultsWg.Add(1)
		go func(l *md.Link) {
			defer resultsWg.Done()
			results <- CheckResult{
				ok:   validator.CheckLink(l.URL, client, cfg, log),
				link: l,
			}
		}(&link)
	}

	resultsWg.Wait()
	close(results)
	handlerWg.Wait()

	log.Info("All links checked")
}
