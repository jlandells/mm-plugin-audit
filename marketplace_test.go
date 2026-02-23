package main

import (
	"testing"
)

func TestMarketplacePlugin_Struct(t *testing.T) {
	mp := &MarketplacePlugin{
		Version:     "2.6.0",
		HomepageURL: "https://github.com/mattermost/mattermost-plugin-github",
	}
	if mp.Version != "2.6.0" {
		t.Errorf("expected version 2.6.0, got %s", mp.Version)
	}
	if mp.HomepageURL != "https://github.com/mattermost/mattermost-plugin-github" {
		t.Errorf("unexpected HomepageURL: %s", mp.HomepageURL)
	}
}
