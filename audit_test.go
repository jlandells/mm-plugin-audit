package main

import (
	"testing"
)

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"empty string", "", ""},
		{"already has v prefix", "v1.2.3", "v1.2.3"},
		{"missing v prefix", "1.2.3", "v1.2.3"},
		{"prerelease", "1.2.3-rc1", "v1.2.3-rc1"},
		{"already has v with prerelease", "v1.2.3-rc1", "v1.2.3-rc1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeVersion(tt.input)
			if result != tt.expect {
				t.Errorf("NormalizeVersion(%q) = %q, want %q", tt.input, result, tt.expect)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name      string
		installed string
		latest    string
		expect    int
	}{
		{"equal versions", "1.0.0", "1.0.0", 0},
		{"equal with v prefix", "v1.0.0", "v1.0.0", 0},
		{"mixed v prefix", "1.0.0", "v1.0.0", 0},
		{"patch update available", "1.0.0", "1.0.1", -1},
		{"minor update available", "1.0.0", "1.1.0", -1},
		{"major update available", "1.0.0", "2.0.0", -1},
		{"installed is newer", "2.0.0", "1.0.0", 1},
		{"semver gotcha: 1.10.0 > 1.9.0", "1.9.0", "1.10.0", -1},
		{"semver gotcha reverse: 1.10.0 > 1.9.0", "1.10.0", "1.9.0", 1},
		{"installed newer than marketplace", "1.11.0", "1.10.0", 1},
		{"prerelease vs release", "1.0.0-rc1", "1.0.0", -1},
		{"both invalid semver but equal", "abc", "abc", 0},
		{"both invalid semver but different", "abc", "def", -1},
		{"installed invalid, latest valid", "abc", "1.0.0", -1},
		{"installed valid, latest invalid", "1.0.0", "abc", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareVersions(tt.installed, tt.latest)
			if result != tt.expect {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d",
					tt.installed, tt.latest, result, tt.expect)
			}
		})
	}
}

func TestDeterminePluginType(t *testing.T) {
	tests := []struct {
		name      string
		hasServer bool
		hasWebapp bool
		expect    string
	}{
		{"both components", true, true, "both"},
		{"server only", true, false, "server"},
		{"webapp only", false, true, "webapp"},
		{"neither", false, false, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DeterminePluginType(tt.hasServer, tt.hasWebapp)
			if result != tt.expect {
				t.Errorf("DeterminePluginType(%v, %v) = %q, want %q",
					tt.hasServer, tt.hasWebapp, result, tt.expect)
			}
		})
	}
}

func TestClassifyPluginSource(t *testing.T) {
	tests := []struct {
		name          string
		pluginID      string
		inMarketplace bool
		homepageURL   string
		expect        string
	}{
		// Tier 1: Bundled plugins (always win, regardless of marketplace presence)
		{"bundled - calls in marketplace", "com.mattermost.calls", true, "", SourceBundled},
		{"bundled - calls not in marketplace", "com.mattermost.calls", false, "", SourceBundled},
		{"bundled - github", "github", true, "", SourceBundled},
		{"bundled - github not in marketplace", "github", false, "", SourceBundled},
		{"bundled - jira", "jira", false, "", SourceBundled},
		{"bundled - playbooks", "playbooks", false, "", SourceBundled},
		{"bundled - focalboard", "focalboard", false, "", SourceBundled},
		{"bundled - zoom", "zoom", false, "", SourceBundled},
		{"bundled - gitlab", "com.github.manland.mattermost-plugin-gitlab", false, "", SourceBundled},
		{"bundled - mattermost-ai", "mattermost-ai", false, "", SourceBundled},
		{"bundled - channel-export", "com.mattermost.plugin-channel-export", false, "", SourceBundled},
		{"bundled - metrics", "com.mattermost.mattermost-plugin-metrics", false, "", SourceBundled},
		{"bundled - mscalendar", "com.mattermost.mscalendar", false, "", SourceBundled},
		{"bundled - msteamsmeetings", "com.mattermost.msteamsmeetings", false, "", SourceBundled},
		{"bundled - servicenow", "mattermost-plugin-servicenow", false, "", SourceBundled},
		{"bundled - user-survey", "com.mattermost.user-survey", false, "", SourceBundled},

		// Tier 2: Marketplace plugins (not bundled, in marketplace)
		{"marketplace plugin", "com.mattermost.confluence", true, "", SourceMarketplace},
		{"marketplace with homepage", "com.mattermost.welcomebot", true, "https://github.com/mattermost/mattermost-plugin-welcomebot", SourceMarketplace},

		// Tier 3: Mattermost plugins (not bundled, not in marketplace, mattermost homepage)
		{"mattermost plugin by homepage", "com.mattermost.gcal", false, "https://github.com/mattermost/mattermost-plugin-gcal", SourceMattermost},
		{"mattermost plugin different path", "some.plugin", false, "https://github.com/mattermost/some-plugin", SourceMattermost},

		// mattermost-community homepage should NOT match as mattermost
		{"community plugin not mattermost", "community.plugin", false, "https://github.com/mattermost-community/some-plugin", SourceThirdParty},

		// Tier 4: Third-party plugins
		{"third-party plugin", "com.pexip.meetings", false, "", SourceThirdParty},
		{"third-party with generic ID", "my-custom-plugin", false, "", SourceThirdParty},
		{"third-party with non-mattermost homepage", "some.plugin", false, "https://github.com/someorg/some-plugin", SourceThirdParty},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClassifyPluginSource(tt.pluginID, tt.inMarketplace, tt.homepageURL)
			if result != tt.expect {
				t.Errorf("ClassifyPluginSource(%q, %v, %q) = %q, want %q",
					tt.pluginID, tt.inMarketplace, tt.homepageURL, result, tt.expect)
			}
		})
	}
}

// mockMMClient implements MattermostClient for testing.
type mockMMClient struct {
	plugins   []InstalledPlugin
	err       error
	mpPlugins map[string]*MarketplacePlugin
	mpErr     error
}

func (m *mockMMClient) GetPlugins() ([]InstalledPlugin, error) {
	return m.plugins, m.err
}

func (m *mockMMClient) GetMarketplacePlugins() (map[string]*MarketplacePlugin, error) {
	if m.mpErr != nil {
		return nil, m.mpErr
	}
	if m.mpPlugins == nil {
		return make(map[string]*MarketplacePlugin), nil
	}
	return m.mpPlugins, nil
}

func noopLogger(format string, args ...interface{}) {}

func TestRunAudit_AllUpToDate(t *testing.T) {
	mm := &mockMMClient{
		plugins: []InstalledPlugin{
			{ID: "com.mattermost.confluence", Name: "Confluence", Version: "1.4.0", Status: "enabled", HasServer: true, HasWebapp: true},
			{ID: "com.mattermost.welcomebot", Name: "WelcomeBot", Version: "1.2.0", Status: "enabled", HasServer: true, HasWebapp: false},
		},
		mpPlugins: map[string]*MarketplacePlugin{
			"com.mattermost.confluence": {Version: "1.4.0", HomepageURL: "https://github.com/mattermost/mattermost-plugin-confluence"},
			"com.mattermost.welcomebot": {Version: "1.2.0", HomepageURL: "https://github.com/mattermost/mattermost-plugin-welcomebot"},
		},
	}

	result, err := RunAudit(mm, AuditOptions{}, noopLogger)
	if err != nil {
		t.Fatalf("RunAudit() returned error: %v", err)
	}

	if result.Summary.Total != 2 {
		t.Errorf("expected 2 total plugins, got %d", result.Summary.Total)
	}
	if result.Summary.Outdated != 0 {
		t.Errorf("expected 0 outdated, got %d", result.Summary.Outdated)
	}
	if result.Summary.UpToDate != 2 {
		t.Errorf("expected 2 up to date, got %d", result.Summary.UpToDate)
	}
	if result.Summary.Marketplace != 2 {
		t.Errorf("expected 2 marketplace, got %d", result.Summary.Marketplace)
	}
}

func TestRunAudit_SomeOutdated(t *testing.T) {
	mm := &mockMMClient{
		plugins: []InstalledPlugin{
			{ID: "com.mattermost.confluence", Name: "Confluence", Version: "1.3.0", Status: "enabled", HasServer: true, HasWebapp: true},
			{ID: "com.mattermost.welcomebot", Name: "WelcomeBot", Version: "1.2.0", Status: "enabled", HasServer: true, HasWebapp: false},
			{ID: "some-third-party", Name: "Custom Plugin", Version: "1.0.0", Status: "disabled", HasServer: true, HasWebapp: true},
		},
		mpPlugins: map[string]*MarketplacePlugin{
			"com.mattermost.confluence": {Version: "1.4.0", HomepageURL: "https://github.com/mattermost/mattermost-plugin-confluence"},
			"com.mattermost.welcomebot": {Version: "1.2.0", HomepageURL: "https://github.com/mattermost/mattermost-plugin-welcomebot"},
		},
	}

	result, err := RunAudit(mm, AuditOptions{}, noopLogger)
	if err != nil {
		t.Fatalf("RunAudit() returned error: %v", err)
	}

	if result.Summary.Total != 3 {
		t.Errorf("expected 3 total plugins, got %d", result.Summary.Total)
	}
	if result.Summary.Outdated != 1 {
		t.Errorf("expected 1 outdated, got %d", result.Summary.Outdated)
	}
	if result.Summary.UpToDate != 1 {
		t.Errorf("expected 1 up to date, got %d", result.Summary.UpToDate)
	}
	if result.Summary.Disabled != 1 {
		t.Errorf("expected 1 disabled, got %d", result.Summary.Disabled)
	}
}

func TestRunAudit_FourWayCategorization(t *testing.T) {
	mm := &mockMMClient{
		plugins: []InstalledPlugin{
			{ID: "com.mattermost.confluence", Name: "Confluence", Version: "1.4.0", Status: "enabled", HasServer: true, HasWebapp: true},
			{ID: "com.mattermost.calls", Name: "Calls", Version: "1.10.0", Status: "enabled", HasServer: true, HasWebapp: true},
			{ID: "com.mattermost.gcal", Name: "Google Calendar", Version: "1.1.0", Status: "enabled", HasServer: true, HasWebapp: true, HomepageURL: "https://github.com/mattermost/mattermost-plugin-gcal"},
			{ID: "com.pexip.meetings", Name: "Pexip", Version: "1.3.0", Status: "enabled", HasServer: true, HasWebapp: false},
		},
		mpPlugins: map[string]*MarketplacePlugin{
			"com.mattermost.confluence": {Version: "1.4.0", HomepageURL: "https://github.com/mattermost/mattermost-plugin-confluence"},
		},
	}

	result, err := RunAudit(mm, AuditOptions{}, noopLogger)
	if err != nil {
		t.Fatalf("RunAudit() returned error: %v", err)
	}

	if result.Summary.Marketplace != 1 {
		t.Errorf("expected 1 marketplace plugin, got %d", result.Summary.Marketplace)
	}
	if result.Summary.Bundled != 1 {
		t.Errorf("expected 1 bundled plugin, got %d", result.Summary.Bundled)
	}
	if result.Summary.MattermostPlugin != 1 {
		t.Errorf("expected 1 mattermost plugin, got %d", result.Summary.MattermostPlugin)
	}
	if result.Summary.ThirdParty != 1 {
		t.Errorf("expected 1 third-party plugin, got %d", result.Summary.ThirdParty)
	}

	// Check sources are correct
	for _, p := range result.Plugins {
		switch p.PluginID {
		case "com.mattermost.confluence":
			if p.Source != SourceMarketplace {
				t.Errorf("Confluence should be marketplace, got %s", p.Source)
			}
		case "com.mattermost.calls":
			if p.Source != SourceBundled {
				t.Errorf("Calls should be bundled, got %s", p.Source)
			}
		case "com.mattermost.gcal":
			if p.Source != SourceMattermost {
				t.Errorf("Google Calendar should be mattermost-plugin, got %s", p.Source)
			}
		case "com.pexip.meetings":
			if p.Source != SourceThirdParty {
				t.Errorf("Pexip should be third-party, got %s", p.Source)
			}
		}
	}
}

func TestRunAudit_InstalledNewerThanMarketplace(t *testing.T) {
	mm := &mockMMClient{
		plugins: []InstalledPlugin{
			{ID: "com.mattermost.confluence", Name: "Confluence", Version: "1.5.0", Status: "enabled", HasServer: true, HasWebapp: true},
		},
		mpPlugins: map[string]*MarketplacePlugin{
			"com.mattermost.confluence": {Version: "1.4.0", HomepageURL: "https://github.com/mattermost/mattermost-plugin-confluence"},
		},
	}

	result, err := RunAudit(mm, AuditOptions{}, noopLogger)
	if err != nil {
		t.Fatalf("RunAudit() returned error: %v", err)
	}

	if result.Summary.Outdated != 0 {
		t.Errorf("expected 0 outdated when installed is newer, got %d", result.Summary.Outdated)
	}
	if result.Summary.UpToDate != 1 {
		t.Errorf("expected 1 up to date when installed is newer, got %d", result.Summary.UpToDate)
	}
	if result.Plugins[0].UpdateAvailable != "false" {
		t.Errorf("expected update_available false when installed is newer, got %s", result.Plugins[0].UpdateAvailable)
	}
}

func TestRunAudit_OutdatedOnlyFilter(t *testing.T) {
	mm := &mockMMClient{
		plugins: []InstalledPlugin{
			{ID: "com.mattermost.confluence", Name: "Confluence", Version: "1.3.0", Status: "enabled", HasServer: true, HasWebapp: true},
			{ID: "com.mattermost.welcomebot", Name: "WelcomeBot", Version: "1.2.0", Status: "enabled", HasServer: true, HasWebapp: false},
			{ID: "com.mattermost.calls", Name: "Calls", Version: "1.10.0", Status: "enabled", HasServer: true, HasWebapp: true},
			{ID: "com.pexip.meetings", Name: "Pexip", Version: "1.3.0", Status: "enabled", HasServer: true, HasWebapp: false},
		},
		mpPlugins: map[string]*MarketplacePlugin{
			"com.mattermost.confluence": {Version: "1.4.0", HomepageURL: "https://github.com/mattermost/mattermost-plugin-confluence"},
			"com.mattermost.welcomebot": {Version: "1.2.0", HomepageURL: "https://github.com/mattermost/mattermost-plugin-welcomebot"},
		},
	}

	result, err := RunAudit(mm, AuditOptions{OutdatedOnly: true}, noopLogger)
	if err != nil {
		t.Fatalf("RunAudit() returned error: %v", err)
	}

	// Should include: Confluence (outdated), Calls (bundled), Pexip (third-party)
	// Should exclude: WelcomeBot (marketplace + up to date)
	if result.Summary.Total != 3 {
		t.Errorf("expected 3 plugins with --outdated-only, got %d", result.Summary.Total)
	}

	for _, p := range result.Plugins {
		if p.PluginID == "com.mattermost.welcomebot" {
			t.Error("WelcomeBot should not appear in --outdated-only output since it's up to date")
		}
	}
}

func TestRunAudit_MarketplaceUnreachable(t *testing.T) {
	mm := &mockMMClient{
		plugins: []InstalledPlugin{
			{ID: "com.mattermost.confluence", Name: "Confluence", Version: "1.3.0", Status: "enabled", HasServer: true, HasWebapp: true},
		},
		mpErr: apiError("marketplace unreachable", nil),
	}

	_, err := RunAudit(mm, AuditOptions{}, noopLogger)
	if err == nil {
		t.Fatal("expected error when marketplace is unreachable")
	}

	cliErr, ok := err.(*CLIError)
	if !ok {
		t.Fatalf("expected *CLIError, got %T", err)
	}
	if cliErr.Code != ExitAPIError {
		t.Errorf("expected exit code %d, got %d", ExitAPIError, cliErr.Code)
	}
}

func TestRunAudit_MMClientError(t *testing.T) {
	mm := &mockMMClient{
		err: apiError("connection failed", nil),
	}

	_, err := RunAudit(mm, AuditOptions{}, noopLogger)
	if err == nil {
		t.Fatal("expected error when MM client fails")
	}
}

func TestRunAudit_NoPlugins(t *testing.T) {
	mm := &mockMMClient{
		plugins: []InstalledPlugin{},
	}

	result, err := RunAudit(mm, AuditOptions{}, noopLogger)
	if err != nil {
		t.Fatalf("RunAudit() returned error: %v", err)
	}

	if result.Summary.Total != 0 {
		t.Errorf("expected 0 total plugins, got %d", result.Summary.Total)
	}
}

func TestRunAudit_SortOrder(t *testing.T) {
	mm := &mockMMClient{
		plugins: []InstalledPlugin{
			{ID: "com.example.zebra", Name: "Zebra Plugin", Version: "1.0.0", Status: "enabled", HasServer: true, HasWebapp: false},
			{ID: "com.mattermost.confluence", Name: "Confluence", Version: "1.4.0", Status: "enabled", HasServer: true, HasWebapp: true},
			{ID: "com.mattermost.calls", Name: "Calls", Version: "1.10.0", Status: "enabled", HasServer: true, HasWebapp: true},
			{ID: "com.mattermost.gcal", Name: "Google Calendar", Version: "1.1.0", Status: "enabled", HasServer: true, HasWebapp: true, HomepageURL: "https://github.com/mattermost/mattermost-plugin-gcal"},
			{ID: "com.mattermost.welcomebot", Name: "WelcomeBot", Version: "1.2.0", Status: "enabled", HasServer: true, HasWebapp: true},
		},
		mpPlugins: map[string]*MarketplacePlugin{
			"com.mattermost.confluence": {Version: "1.4.0", HomepageURL: "https://github.com/mattermost/mattermost-plugin-confluence"},
			"com.mattermost.welcomebot": {Version: "1.2.0", HomepageURL: "https://github.com/mattermost/mattermost-plugin-welcomebot"},
		},
	}

	result, err := RunAudit(mm, AuditOptions{}, noopLogger)
	if err != nil {
		t.Fatalf("RunAudit() returned error: %v", err)
	}

	if len(result.Plugins) != 5 {
		t.Fatalf("expected 5 plugins, got %d", len(result.Plugins))
	}

	// Expected order: Marketplace (Confluence, WelcomeBot), Mattermost (Google Calendar), Bundled (Calls), Third-party (Zebra)
	expected := []struct {
		name   string
		source string
	}{
		{"Confluence", SourceMarketplace},
		{"WelcomeBot", SourceMarketplace},
		{"Google Calendar", SourceMattermost},
		{"Calls", SourceBundled},
		{"Zebra Plugin", SourceThirdParty},
	}

	for i, exp := range expected {
		if result.Plugins[i].Name != exp.name {
			t.Errorf("position %d: expected %s, got %s", i, exp.name, result.Plugins[i].Name)
		}
		if result.Plugins[i].Source != exp.source {
			t.Errorf("position %d (%s): expected source %s, got %s", i, result.Plugins[i].Name, exp.source, result.Plugins[i].Source)
		}
	}
}

func TestRunAudit_BundledPluginOverridesMarketplace(t *testing.T) {
	// Even if a bundled plugin appears in the marketplace, it should be classified as bundled
	mm := &mockMMClient{
		plugins: []InstalledPlugin{
			{ID: "com.mattermost.calls", Name: "Calls", Version: "1.10.0", Status: "enabled", HasServer: true, HasWebapp: true},
			{ID: "github", Name: "GitHub", Version: "2.6.0", Status: "enabled", HasServer: true, HasWebapp: true},
		},
		mpPlugins: map[string]*MarketplacePlugin{
			"com.mattermost.calls": {Version: "1.10.0", HomepageURL: "https://github.com/mattermost/mattermost-plugin-calls"},
			"github":               {Version: "2.6.0", HomepageURL: "https://github.com/mattermost/mattermost-plugin-github"},
		},
	}

	result, err := RunAudit(mm, AuditOptions{}, noopLogger)
	if err != nil {
		t.Fatalf("RunAudit() returned error: %v", err)
	}

	if result.Summary.Bundled != 2 {
		t.Errorf("expected 2 bundled plugins, got %d", result.Summary.Bundled)
	}
	if result.Summary.Marketplace != 0 {
		t.Errorf("expected 0 marketplace plugins, got %d", result.Summary.Marketplace)
	}

	for _, p := range result.Plugins {
		if p.Source != SourceBundled {
			t.Errorf("plugin %s should be bundled, got %s", p.PluginID, p.Source)
		}
	}
}
