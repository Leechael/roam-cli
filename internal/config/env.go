package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	defaultBaseURL = "https://api.roamresearch.com/api/graph"
	defaultTimeout = 10
)

type Config struct {
	Token          string
	Graph          string
	BaseURL        string
	TimeoutSeconds int
}

func New(token, graph, baseURL string, timeoutSeconds int) (*Config, error) {
	if token == "" {
		token = os.Getenv("ROAM_API_TOKEN")
	}
	if graph == "" {
		graph = os.Getenv("ROAM_API_GRAPH")
	}
	if baseURL == "" {
		baseURL = os.Getenv("ROAM_API_BASE_URL")
	}
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if timeoutSeconds <= 0 {
		timeoutSeconds = timeoutFromEnv()
	}
	if token == "" || graph == "" {
		var missing []string
		if token == "" {
			missing = append(missing, "ROAM_API_TOKEN")
		}
		if graph == "" {
			missing = append(missing, "ROAM_API_GRAPH")
		}
		return nil, fmt.Errorf("%s not set. Run \"roam-cli help configuration\" for setup instructions", strings.Join(missing, " and "))
	}

	return &Config{
		Token:          token,
		Graph:          graph,
		BaseURL:        strings.TrimRight(baseURL, "/"),
		TimeoutSeconds: timeoutSeconds,
	}, nil
}

func timeoutFromEnv() int {
	v := os.Getenv("ROAM_TIMEOUT_SECONDS")
	if v == "" {
		return defaultTimeout
	}
	i, err := strconv.Atoi(v)
	if err != nil || i <= 0 {
		return defaultTimeout
	}
	return i
}
