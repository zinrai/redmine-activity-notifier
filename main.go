package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/mmcdole/gofeed"
	"gopkg.in/yaml.v2"
)

type config struct {
	Interval  time.Duration `yaml:"interval"`
	SlackURL  string        `yaml:"slack_url"`
	AtomURL   string        `yaml:"atom_url"`
	BasicAuth *basicAuth    `yaml:"basic_auth"`
}

type basicAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type redmine2Slack struct {
	cfg *config
}

func newRedmine2Slack(configFile string) (*redmine2Slack, error) {
	cfg := &config{}

	// Get the absolute path of the executable
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	// Get the directory of the executable
	execDir := filepath.Dir(execPath)

	// Construct the absolute path of the config file
	configPath := filepath.Join(execDir, configFile)

	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	if err := yaml.Unmarshal(yamlFile, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &redmine2Slack{cfg: cfg}, nil
}

func (r *redmine2Slack) atomFeed(ctx context.Context) ([]map[string]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.cfg.AtomURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if r.cfg.BasicAuth != nil {
		req.SetBasicAuth(r.cfg.BasicAuth.Username, r.cfg.BasicAuth.Password)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	fp := gofeed.NewParser()
	feed, err := fp.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse atom feed: %w", err)
	}

	var result []map[string]string
	now := time.Now()
	for _, item := range feed.Items {
		updated := *item.UpdatedParsed
		if now.Sub(updated) < r.cfg.Interval {
			result = append(result, map[string]string{
				"title": item.Title,
				"name":  item.Author.Name,
				"link":  item.Link,
			})
		}
	}

	return result, nil
}

func (r *redmine2Slack) sendToSlack(ctx context.Context, payload map[string]string) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.cfg.SlackURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

func (r *redmine2Slack) run(ctx context.Context) error {
	result, err := r.atomFeed(ctx)
	if err != nil {
		return fmt.Errorf("failed to get atom feed: %w", err)
	}

	for i := len(result) - 1; i >= 0; i-- {
		elem := result[i]
		content := fmt.Sprintf("*%s*\n%s\n <%s>", elem["title"], elem["name"], elem["link"])

		payload := map[string]string{
			"text": content,
		}

		if err := r.sendToSlack(ctx, payload); err != nil {
			log.Printf("failed to send to slack: %v", err)
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}

func main() {
	configFile := "config.yaml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	ctx := context.Background()
	r2s, err := newRedmine2Slack(configFile)
	if err != nil {
		log.Fatalf("failed to create redmine2Slack: %v", err)
	}

	if err := r2s.run(ctx); err != nil {
		log.Fatalf("failed to run: %v", err)
	}
}
