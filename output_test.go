package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
)

func sampleResult() *AuditResult {
	trueVal := true
	falseVal := false
	return &AuditResult{
		Plugins: []PluginReport{
			{
				PluginID:         "com.mattermost.confluence",
				Name:             "Confluence",
				InstalledVersion: "1.3.0",
				LatestVersion:    "1.4.0",
				UpdateAvailable:  "true",
				UpdateAvailJSON:  &trueVal,
				Status:           "enabled",
				Source:           SourceMarketplace,
				MarketplaceURL:   "https://github.com/mattermost/mattermost-plugin-confluence",
				PluginType:       "both",
			},
			{
				PluginID:         "com.mattermost.welcomebot",
				Name:             "WelcomeBot",
				InstalledVersion: "1.2.0",
				LatestVersion:    "1.2.0",
				UpdateAvailable:  "false",
				UpdateAvailJSON:  &falseVal,
				Status:           "enabled",
				Source:           SourceMarketplace,
				MarketplaceURL:   "https://github.com/mattermost/mattermost-plugin-welcomebot",
				PluginType:       "server",
			},
			{
				PluginID:         "com.mattermost.gcal",
				Name:             "Google Calendar",
				InstalledVersion: "1.1.0",
				LatestVersion:    "",
				UpdateAvailable:  "unknown",
				UpdateAvailJSON:  nil,
				Status:           "enabled",
				Source:           SourceMattermost,
				MarketplaceURL:   "",
				PluginType:       "both",
			},
			{
				PluginID:         "com.mattermost.calls",
				Name:             "Calls",
				InstalledVersion: "1.10.0",
				LatestVersion:    "",
				UpdateAvailable:  "unknown",
				UpdateAvailJSON:  nil,
				Status:           "enabled",
				Source:           SourceBundled,
				MarketplaceURL:   "",
				PluginType:       "both",
			},
			{
				PluginID:         "com.pexip.meetings",
				Name:             "Pexip",
				InstalledVersion: "1.3.0",
				LatestVersion:    "",
				UpdateAvailable:  "unknown",
				UpdateAvailJSON:  nil,
				Status:           "disabled",
				Source:           SourceThirdParty,
				MarketplaceURL:   "",
				PluginType:       "server",
			},
		},
		Summary: AuditSummary{
			Total:            5,
			Marketplace:      2,
			Bundled:          1,
			MattermostPlugin: 1,
			ThirdParty:       1,
			Outdated:         1,
			UpToDate:         1,
			Unknown:          3,
			Enabled:          4,
			Disabled:         1,
		},
	}
}

func TestFormatTable(t *testing.T) {
	var buf bytes.Buffer
	result := sampleResult()

	err := FormatOutput(&buf, result, "table")
	if err != nil {
		t.Fatalf("FormatOutput() returned error: %v", err)
	}

	output := buf.String()

	// Check four sections exist
	if !strings.Contains(output, "=== Marketplace Plugins (2) ===") {
		t.Error("missing marketplace plugins section header")
	}
	if !strings.Contains(output, "=== Mattermost Plugins (1) ===") {
		t.Error("missing mattermost plugins section header")
	}
	if !strings.Contains(output, "=== Bundled Mattermost Plugins (1) ===") {
		t.Error("missing bundled plugins section header")
	}
	if !strings.Contains(output, "=== Third-Party / Custom Plugins (1) ===") {
		t.Error("missing third-party plugins section header")
	}

	// Check header columns
	if !strings.Contains(output, "NAME") || !strings.Contains(output, "INSTALLED") || !strings.Contains(output, "LATEST") {
		t.Error("missing table headers")
	}

	// Check plugin names appear
	if !strings.Contains(output, "Confluence") {
		t.Error("missing Confluence plugin")
	}
	if !strings.Contains(output, "WelcomeBot") {
		t.Error("missing WelcomeBot plugin")
	}
	if !strings.Contains(output, "Google Calendar") {
		t.Error("missing Google Calendar plugin")
	}
	if !strings.Contains(output, "Calls") {
		t.Error("missing Calls plugin")
	}
	if !strings.Contains(output, "Pexip") {
		t.Error("missing Pexip plugin")
	}

	// Check update indicator
	if !strings.Contains(output, "YES") {
		t.Error("missing YES update indicator for Confluence")
	}

	// Check summary
	if !strings.Contains(output, "5 plugin(s) total") {
		t.Error("missing summary line")
	}
	if !strings.Contains(output, "1 mattermost") {
		t.Error("missing mattermost count in summary")
	}
	if !strings.Contains(output, "1 bundled") {
		t.Error("missing bundled count in summary")
	}
	if !strings.Contains(output, "1 third-party/custom") {
		t.Error("missing third-party count in summary")
	}
}

func TestFormatCSV(t *testing.T) {
	var buf bytes.Buffer
	result := sampleResult()

	err := FormatOutput(&buf, result, "csv")
	if err != nil {
		t.Fatalf("FormatOutput() returned error: %v", err)
	}

	// Verify CSV is parseable
	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("CSV is not parseable: %v", err)
	}

	// Header + 5 data rows
	if len(records) != 6 {
		t.Errorf("expected 6 rows (header + 5 data), got %d", len(records))
	}

	// Check header
	expectedHeaders := []string{"plugin_id", "name", "installed_version", "latest_version",
		"update_available", "status", "type", "source", "marketplace_url"}
	if len(records[0]) != len(expectedHeaders) {
		t.Errorf("expected %d columns, got %d", len(expectedHeaders), len(records[0]))
	}
	for i, h := range expectedHeaders {
		if records[0][i] != h {
			t.Errorf("header[%d] = %q, want %q", i, records[0][i], h)
		}
	}

	// Check first data row (Confluence — marketplace, outdated)
	if records[1][0] != "com.mattermost.confluence" {
		t.Errorf("expected plugin_id com.mattermost.confluence, got %s", records[1][0])
	}
	if records[1][4] != "true" {
		t.Errorf("expected update_available true, got %s", records[1][4])
	}
	if records[1][7] != "marketplace" {
		t.Errorf("expected source marketplace, got %s", records[1][7])
	}

	// Check mattermost plugin row (Google Calendar)
	if records[3][7] != "mattermost-plugin" {
		t.Errorf("expected source mattermost-plugin for Google Calendar, got %s", records[3][7])
	}

	// Check bundled plugin row (Calls)
	if records[4][7] != "bundled" {
		t.Errorf("expected source bundled for Calls, got %s", records[4][7])
	}

	// Check third-party plugin row (Pexip)
	if records[5][7] != "third-party" {
		t.Errorf("expected source third-party for Pexip, got %s", records[5][7])
	}
	if records[5][4] != "unknown" {
		t.Errorf("expected update_available unknown for Pexip, got %s", records[5][4])
	}
}

func TestFormatJSON(t *testing.T) {
	var buf bytes.Buffer
	result := sampleResult()

	err := FormatOutput(&buf, result, "json")
	if err != nil {
		t.Fatalf("FormatOutput() returned error: %v", err)
	}

	// Verify JSON is parseable
	var parsed jsonOutput
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("JSON is not parseable: %v", err)
	}

	// Check structure
	if len(parsed.Plugins) != 5 {
		t.Errorf("expected 5 plugins, got %d", len(parsed.Plugins))
	}

	// Check summary
	if parsed.Summary.Total != 5 {
		t.Errorf("expected total 5, got %d", parsed.Summary.Total)
	}
	if parsed.Summary.Outdated != 1 {
		t.Errorf("expected outdated 1, got %d", parsed.Summary.Outdated)
	}
	if parsed.Summary.Bundled != 1 {
		t.Errorf("expected bundled 1, got %d", parsed.Summary.Bundled)
	}
	if parsed.Summary.MattermostPlugin != 1 {
		t.Errorf("expected mattermost_plugin 1, got %d", parsed.Summary.MattermostPlugin)
	}
	if parsed.Summary.ThirdParty != 1 {
		t.Errorf("expected third_party 1, got %d", parsed.Summary.ThirdParty)
	}

	// Check marketplace plugin
	confluence := parsed.Plugins[0]
	if confluence.UpdateAvailable == nil || *confluence.UpdateAvailable != true {
		t.Error("expected Confluence update_available to be true")
	}
	if confluence.Source != "marketplace" {
		t.Errorf("expected Confluence source marketplace, got %s", confluence.Source)
	}

	// Check mattermost plugin
	gcal := parsed.Plugins[2]
	if gcal.UpdateAvailable != nil {
		t.Error("expected Google Calendar update_available to be null")
	}
	if gcal.Source != "mattermost-plugin" {
		t.Errorf("expected Google Calendar source mattermost-plugin, got %s", gcal.Source)
	}

	// Check bundled plugin
	calls := parsed.Plugins[3]
	if calls.UpdateAvailable != nil {
		t.Error("expected Calls update_available to be null")
	}
	if calls.Source != "bundled" {
		t.Errorf("expected Calls source bundled, got %s", calls.Source)
	}

	// Check third-party plugin
	pexip := parsed.Plugins[4]
	if pexip.Source != "third-party" {
		t.Errorf("expected Pexip source third-party, got %s", pexip.Source)
	}
}

func TestFormatJSON_ValidStructure(t *testing.T) {
	var buf bytes.Buffer
	result := sampleResult()

	err := FormatOutput(&buf, result, "json")
	if err != nil {
		t.Fatalf("FormatOutput() returned error: %v", err)
	}

	// Verify it's valid JSON by unmarshalling into a generic map
	var raw map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	// Must have top-level "plugins" and "summary" keys
	if _, ok := raw["plugins"]; !ok {
		t.Error("missing top-level 'plugins' key")
	}
	if _, ok := raw["summary"]; !ok {
		t.Error("missing top-level 'summary' key")
	}
}

func TestFormatOutput_UnknownFormat(t *testing.T) {
	var buf bytes.Buffer
	result := sampleResult()

	err := FormatOutput(&buf, result, "xml")
	if err == nil {
		t.Error("expected error for unknown format")
	}
}

func TestFormatTable_EmptyResult(t *testing.T) {
	var buf bytes.Buffer
	result := &AuditResult{
		Plugins: []PluginReport{},
		Summary: AuditSummary{},
	}

	err := FormatOutput(&buf, result, "table")
	if err != nil {
		t.Fatalf("FormatOutput() returned error: %v", err)
	}

	output := buf.String()
	if strings.Count(output, "(none)") != 4 {
		t.Errorf("expected (none) for all four empty sections, got %d", strings.Count(output, "(none)"))
	}
}

func TestFormatCSV_EmptyResult(t *testing.T) {
	var buf bytes.Buffer
	result := &AuditResult{
		Plugins: []PluginReport{},
		Summary: AuditSummary{},
	}

	err := FormatOutput(&buf, result, "csv")
	if err != nil {
		t.Fatalf("FormatOutput() returned error: %v", err)
	}

	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("CSV is not parseable: %v", err)
	}

	if len(records) != 1 {
		t.Errorf("expected 1 row (header only), got %d", len(records))
	}
}

func TestFormatJSON_EmptyResult(t *testing.T) {
	var buf bytes.Buffer
	result := &AuditResult{
		Plugins: []PluginReport{},
		Summary: AuditSummary{},
	}

	err := FormatOutput(&buf, result, "json")
	if err != nil {
		t.Fatalf("FormatOutput() returned error: %v", err)
	}

	var parsed jsonOutput
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("JSON is not parseable: %v", err)
	}

	if len(parsed.Plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(parsed.Plugins))
	}
}

func TestUpdateIndicator(t *testing.T) {
	tests := []struct {
		name   string
		report PluginReport
		expect string
	}{
		{
			"update available",
			PluginReport{UpdateAvailable: "true", InstalledVersion: "1.0.0", LatestVersion: "2.0.0"},
			"YES ⚠",
		},
		{
			"up to date",
			PluginReport{UpdateAvailable: "false", InstalledVersion: "1.0.0", LatestVersion: "1.0.0"},
			"No",
		},
		{
			"installed newer than marketplace",
			PluginReport{UpdateAvailable: "false", InstalledVersion: "1.11.0", LatestVersion: "1.10.0"},
			"No",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := updateIndicator(tt.report)
			if result != tt.expect {
				t.Errorf("updateIndicator() = %q, want %q", result, tt.expect)
			}
		})
	}
}
