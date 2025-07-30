package validator

import (
	"log/slog"
	"net/http"
	"time"
)

type LinksValidatorConfig struct {
	AllowedStatuses map[int]struct{}
	Timeout         time.Duration
}

func PrepareAllowedStatuses(statuses ...int) map[int]struct{} {

	preparedStatuses := make(map[int]struct{})

	for _, status := range statuses {
		preparedStatuses[status] = struct{}{}
	}

	return preparedStatuses
}

func GetClient(config LinksValidatorConfig) http.Client {
	client := http.Client{
		Timeout: config.Timeout,
	}

	return client
}

func CheckLink(url string, client http.Client, config LinksValidatorConfig, log *slog.Logger) bool {

	log.Debug("Check URL", slog.String("url", url))
	resp, err := client.Get(url)

	if err != nil {
		log.Debug("Error while check URL", slog.String("url", url), slog.String("error", err.Error()))
		return false
	}
	defer resp.Body.Close()
	log.Debug("Sucess request for URL check", slog.String("url", url), slog.String("status", resp.Status))
	_, ok := config.AllowedStatuses[resp.StatusCode]

	return ok
}
