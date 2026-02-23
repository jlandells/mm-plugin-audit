package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
)

// MattermostClient defines the interface for interacting with the Mattermost API.
type MattermostClient interface {
	GetPlugins() ([]InstalledPlugin, error)
	GetMarketplacePlugins() (map[string]*MarketplacePlugin, error)
}

// MMClient wraps model.Client4 and implements MattermostClient.
type MMClient struct {
	client *model.Client4
}

// ClientConfig holds the configuration for connecting to a Mattermost instance.
type ClientConfig struct {
	URL      string
	Token    string
	Username string
	Password string
}

// NewMMClient creates a new Mattermost client and authenticates.
func NewMMClient(cfg ClientConfig) (*MMClient, error) {
	serverURL := strings.TrimRight(cfg.URL, "/")
	client := model.NewAPIv4Client(serverURL)

	if cfg.Token != "" {
		client.SetToken(cfg.Token)

		// Validate the token by making a test call
		_, resp, err := client.GetPlugins(context.Background())
		if err != nil {
			return nil, classifyAPIError(serverURL, resp, err)
		}

		return &MMClient{client: client}, nil
	}

	if cfg.Username != "" {
		user, resp, err := client.Login(context.Background(), cfg.Username, cfg.Password)
		if err != nil {
			return nil, classifyAPIError(serverURL, resp, err)
		}
		_ = user
		return &MMClient{client: client}, nil
	}

	return nil, configError(
		"error: authentication required. Use --token (or MM_TOKEN) for token auth, or --username (or MM_USERNAME) for password auth.",
		nil,
	)
}

// GetPlugins retrieves all installed plugins from the Mattermost instance.
func (c *MMClient) GetPlugins() ([]InstalledPlugin, error) {
	pluginsResp, resp, err := c.client.GetPlugins(context.Background())
	if err != nil {
		return nil, classifyAPIError("", resp, err)
	}

	var plugins []InstalledPlugin

	for _, p := range pluginsResp.Active {
		plugins = append(plugins, InstalledPlugin{
			ID:          p.Id,
			Name:        p.Name,
			Version:     p.Version,
			HomepageURL: p.HomepageURL,
			Status:      "enabled",
			HasServer:   p.Server != nil,
			HasWebapp:   p.Webapp != nil,
		})
	}

	for _, p := range pluginsResp.Inactive {
		plugins = append(plugins, InstalledPlugin{
			ID:          p.Id,
			Name:        p.Name,
			Version:     p.Version,
			HomepageURL: p.HomepageURL,
			Status:      "disabled",
			HasServer:   p.Server != nil,
			HasWebapp:   p.Webapp != nil,
		})
	}

	return plugins, nil
}

// GetMarketplacePlugins fetches the Marketplace catalogue via the server's proxy endpoint.
func (c *MMClient) GetMarketplacePlugins() (map[string]*MarketplacePlugin, error) {
	result := make(map[string]*MarketplacePlugin)

	page := 0
	perPage := 200
	for {
		filter := &model.MarketplacePluginFilter{
			Page:    page,
			PerPage: perPage,
		}
		plugins, resp, err := c.client.GetMarketplacePlugins(context.Background(), filter)
		if err != nil {
			return nil, classifyAPIError("", resp, err)
		}

		for _, p := range plugins {
			if p.Manifest != nil {
				result[p.Manifest.Id] = &MarketplacePlugin{
					Version:     p.Manifest.Version,
					HomepageURL: p.HomepageURL,
				}
			}
		}

		if len(plugins) < perPage {
			break
		}
		page++
	}

	return result, nil
}

// classifyAPIError maps Mattermost API errors to appropriate CLIError types.
func classifyAPIError(serverURL string, resp *model.Response, err error) *CLIError {
	if resp != nil {
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return configError("error: authentication failed. Check your token or credentials.", err)
		case http.StatusForbidden:
			return configError("error: permission denied. This operation requires a System Administrator account.", err)
		case http.StatusNotFound:
			return apiError(fmt.Sprintf("error: API endpoint not found on %s. Check the server URL.", serverURL), err)
		}
		if resp.StatusCode >= 500 {
			return apiError(
				fmt.Sprintf("error: the Mattermost server returned an unexpected error (HTTP %d). Check server logs for details.", resp.StatusCode),
				err,
			)
		}
	}

	if serverURL != "" {
		return apiError(fmt.Sprintf("error: unable to connect to %s. Check the URL and network connectivity.", serverURL), err)
	}
	return apiError("error: unexpected API error.", err)
}
