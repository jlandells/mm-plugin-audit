package main

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/mod/semver"
)

// Plugin source categories.
const (
	SourceMarketplace = "marketplace"
	SourceBundled     = "bundled"
	SourceMattermost  = "mattermost-plugin"
	SourceThirdParty  = "third-party"
)

// bundledPlugins lists the exact plugin IDs that are bundled with Mattermost.
var bundledPlugins = map[string]bool{
	"mattermost-ai":                               true, // mattermost-plugin-agents
	"focalboard":                                   true, // mattermost-plugin-boards
	"com.mattermost.calls":                         true, // mattermost-plugin-calls
	"com.mattermost.plugin-channel-export":         true, // mattermost-plugin-channel-export
	"github":                                       true, // mattermost-plugin-github
	"com.github.manland.mattermost-plugin-gitlab":  true, // mattermost-plugin-gitlab
	"jira":                                         true, // mattermost-plugin-jira
	"com.mattermost.mattermost-plugin-metrics":     true, // mattermost-plugin-metrics
	"com.mattermost.mscalendar":                    true, // mattermost-plugin-mscalendar
	"com.mattermost.msteamsmeetings":               true, // mattermost-plugin-msteams-meetings
	"playbooks":                                    true, // mattermost-plugin-playbooks
	"mattermost-plugin-servicenow":                 true, // mattermost-plugin-servicenow
	"com.mattermost.user-survey":                   true, // mattermost-plugin-user-survey
	"zoom":                                         true, // mattermost-plugin-zoom
}

// PluginReport holds the audit data for a single plugin.
type PluginReport struct {
	PluginID         string `json:"plugin_id"`
	Name             string `json:"name"`
	InstalledVersion string `json:"installed_version"`
	LatestVersion    string `json:"latest_version"`
	UpdateAvailable  string `json:"-"`
	UpdateAvailJSON  *bool  `json:"update_available"`
	Status           string `json:"status"`
	Source           string `json:"source"`
	MarketplaceURL   string `json:"marketplace_url"`
	PluginType       string `json:"type"`
}

// AuditSummary holds aggregate statistics for the audit.
type AuditSummary struct {
	Total            int `json:"total"`
	Marketplace      int `json:"marketplace"`
	Bundled          int `json:"bundled"`
	MattermostPlugin int `json:"mattermost_plugin"`
	ThirdParty       int `json:"third_party"`
	Outdated         int `json:"outdated"`
	UpToDate         int `json:"up_to_date"`
	Unknown          int `json:"unknown"`
	Enabled          int `json:"enabled"`
	Disabled         int `json:"disabled"`
}

// AuditResult holds the full audit output.
type AuditResult struct {
	Plugins []PluginReport `json:"plugins"`
	Summary AuditSummary   `json:"summary"`
}

// AuditOptions controls the behaviour of RunAudit.
type AuditOptions struct {
	OutdatedOnly bool
	Verbose      bool
}

// NormalizeVersion prepends "v" if missing, as required by golang.org/x/mod/semver.
func NormalizeVersion(v string) string {
	if v == "" {
		return ""
	}
	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}

// CompareVersions compares two version strings using semantic versioning.
// Returns -1 if installed < latest, 0 if equal, 1 if installed > latest.
// Falls back to string equality if either version isn't valid semver.
func CompareVersions(installed, latest string) int {
	ni := NormalizeVersion(installed)
	nl := NormalizeVersion(latest)

	if !semver.IsValid(ni) || !semver.IsValid(nl) {
		if installed == latest {
			return 0
		}
		return -1
	}

	return semver.Compare(ni, nl)
}

// DeterminePluginType returns the plugin type based on which components are present.
func DeterminePluginType(hasServer, hasWebapp bool) string {
	if hasServer && hasWebapp {
		return "both"
	}
	if hasServer {
		return "server"
	}
	if hasWebapp {
		return "webapp"
	}
	return "unknown"
}

// ClassifyPluginSource determines the source category for a plugin using
// strict 4-tier priority (checked in sequence, first match wins):
//  1. Plugin ID in bundledPlugins map -> Bundled
//  2. Plugin found in marketplace API response -> Marketplace
//  3. Plugin's homepage_url contains github.com/mattermost/ (not mattermost-community/) -> Mattermost Plugin
//  4. None of the above -> Third-party/Custom
func ClassifyPluginSource(pluginID string, inMarketplace bool, homepageURL string) string {
	if bundledPlugins[pluginID] {
		return SourceBundled
	}
	if inMarketplace {
		return SourceMarketplace
	}
	if strings.Contains(homepageURL, "github.com/mattermost/") &&
		!strings.Contains(homepageURL, "github.com/mattermost-community/") {
		return SourceMattermost
	}
	return SourceThirdParty
}

// InstalledPlugin is a simplified representation of a plugin from the Mattermost API.
type InstalledPlugin struct {
	ID          string
	Name        string
	Version     string
	HomepageURL string
	Status      string // "enabled" or "disabled"
	HasServer   bool
	HasWebapp   bool
}

// RunAudit fetches installed plugins, queries the Marketplace, and produces an AuditResult.
func RunAudit(mmClient MattermostClient, opts AuditOptions, logf func(string, ...interface{})) (*AuditResult, error) {
	logf("Fetching installed plugins from Mattermost instance...")
	installed, err := mmClient.GetPlugins()
	if err != nil {
		return nil, err
	}

	logf("Found %d installed plugin(s)", len(installed))

	logf("Fetching Marketplace catalogue...")
	mpCatalogue, err := mmClient.GetMarketplacePlugins()
	if err != nil {
		return nil, err
	}
	logf("Marketplace catalogue contains %d plugin(s)", len(mpCatalogue))

	var reports []PluginReport

	for _, p := range installed {
		report := PluginReport{
			PluginID:         p.ID,
			Name:             p.Name,
			InstalledVersion: p.Version,
			Status:           p.Status,
			PluginType:       DeterminePluginType(p.HasServer, p.HasWebapp),
		}

		mpPlugin, inMarketplace := mpCatalogue[p.ID]
		report.Source = ClassifyPluginSource(p.ID, inMarketplace, p.HomepageURL)

		if report.Source == SourceMarketplace {
			report.LatestVersion = mpPlugin.Version
			report.MarketplaceURL = mpPlugin.HomepageURL

			cmp := CompareVersions(p.Version, mpPlugin.Version)
			if cmp < 0 {
				report.UpdateAvailable = "true"
				b := true
				report.UpdateAvailJSON = &b
			} else {
				report.UpdateAvailable = "false"
				b := false
				report.UpdateAvailJSON = &b
			}
		} else {
			report.UpdateAvailable = "unknown"
			report.UpdateAvailJSON = nil
		}

		reports = append(reports, report)
	}

	// Filter if --outdated-only
	if opts.OutdatedOnly {
		var filtered []PluginReport
		for _, r := range reports {
			if r.UpdateAvailable == "true" || r.Source != SourceMarketplace {
				filtered = append(filtered, r)
			}
		}
		reports = filtered
	}

	// Sort: marketplace first, then mattermost, then bundled, then third-party — alphabetically within each group
	sourceOrder := map[string]int{
		SourceMarketplace: 0,
		SourceMattermost:  1,
		SourceBundled:     2,
		SourceThirdParty:  3,
	}
	sort.Slice(reports, func(i, j int) bool {
		oi, oj := sourceOrder[reports[i].Source], sourceOrder[reports[j].Source]
		if oi != oj {
			return oi < oj
		}
		return strings.ToLower(reports[i].Name) < strings.ToLower(reports[j].Name)
	})

	// Compute summary
	summary := AuditSummary{}
	for _, r := range reports {
		summary.Total++
		switch r.Source {
		case SourceMarketplace:
			summary.Marketplace++
			if r.UpdateAvailable == "true" {
				summary.Outdated++
			} else {
				summary.UpToDate++
			}
		case SourceBundled:
			summary.Bundled++
			summary.Unknown++
		case SourceMattermost:
			summary.MattermostPlugin++
			summary.Unknown++
		case SourceThirdParty:
			summary.ThirdParty++
			summary.Unknown++
		}
		if r.Status == "enabled" {
			summary.Enabled++
		} else {
			summary.Disabled++
		}
	}

	return &AuditResult{
		Plugins: reports,
		Summary: summary,
	}, nil
}

// verboseLogger returns a logging function that prints to stderr when verbose is true.
func verboseLogger(verbose bool) func(string, ...interface{}) {
	return func(format string, args ...interface{}) {
		if verbose {
			fmt.Fprintf(logOutput, "[verbose] "+format+"\n", args...)
		}
	}
}
