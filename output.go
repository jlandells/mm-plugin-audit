package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// FormatOutput writes the audit result in the specified format.
func FormatOutput(w io.Writer, result *AuditResult, format string) error {
	switch strings.ToLower(format) {
	case "table":
		return formatTable(w, result)
	case "csv":
		return formatCSV(w, result)
	case "json":
		return formatJSON(w, result)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func formatTable(w io.Writer, result *AuditResult) error {
	// Separate plugins by source
	var marketplace, mattermost, bundled, thirdParty []PluginReport
	for _, p := range result.Plugins {
		switch p.Source {
		case SourceMarketplace:
			marketplace = append(marketplace, p)
		case SourceMattermost:
			mattermost = append(mattermost, p)
		case SourceBundled:
			bundled = append(bundled, p)
		default:
			thirdParty = append(thirdParty, p)
		}
	}

	// Marketplace plugins section
	fmt.Fprintf(w, "=== Marketplace Plugins (%d) ===\n", len(marketplace))
	if len(marketplace) > 0 {
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "NAME\tINSTALLED\tLATEST\tUPDATE?\tSTATUS")
		for _, p := range marketplace {
			updateStr := updateIndicator(p)
			statusStr := capitalizeStatus(p.Status)
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", p.Name, p.InstalledVersion, p.LatestVersion, updateStr, statusStr)
		}
		tw.Flush()
	} else {
		fmt.Fprintln(w, "(none)")
	}

	fmt.Fprintln(w)

	// Mattermost plugins section
	fmt.Fprintf(w, "=== Mattermost Plugins (%d) ===\n", len(mattermost))
	if len(mattermost) > 0 {
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "NAME\tINSTALLED\tSTATUS")
		for _, p := range mattermost {
			statusStr := capitalizeStatus(p.Status)
			fmt.Fprintf(tw, "%s\t%s\t%s\n", p.Name, p.InstalledVersion, statusStr)
		}
		tw.Flush()
	} else {
		fmt.Fprintln(w, "(none)")
	}

	fmt.Fprintln(w)

	// Bundled plugins section
	fmt.Fprintf(w, "=== Bundled Mattermost Plugins (%d) ===\n", len(bundled))
	if len(bundled) > 0 {
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "NAME\tINSTALLED\tSTATUS")
		for _, p := range bundled {
			statusStr := capitalizeStatus(p.Status)
			fmt.Fprintf(tw, "%s\t%s\t%s\n", p.Name, p.InstalledVersion, statusStr)
		}
		tw.Flush()
	} else {
		fmt.Fprintln(w, "(none)")
	}

	fmt.Fprintln(w)

	// Third-party plugins section
	fmt.Fprintf(w, "=== Third-Party / Custom Plugins (%d) ===\n", len(thirdParty))
	if len(thirdParty) > 0 {
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "NAME\tINSTALLED\tSTATUS")
		for _, p := range thirdParty {
			statusStr := capitalizeStatus(p.Status)
			fmt.Fprintf(tw, "%s\t%s\t%s\n", p.Name, p.InstalledVersion, statusStr)
		}
		tw.Flush()
	} else {
		fmt.Fprintln(w, "(none)")
	}

	fmt.Fprintln(w)

	// Summary
	fmt.Fprintf(w, "Summary: %d plugin(s) total — %d marketplace (%d outdated, %d up to date), %d mattermost, %d bundled, %d third-party/custom — %d enabled, %d disabled\n",
		result.Summary.Total,
		result.Summary.Marketplace,
		result.Summary.Outdated,
		result.Summary.UpToDate,
		result.Summary.MattermostPlugin,
		result.Summary.Bundled,
		result.Summary.ThirdParty,
		result.Summary.Enabled,
		result.Summary.Disabled,
	)

	return nil
}

// updateIndicator returns a human-readable string for the UPDATE? column.
func updateIndicator(p PluginReport) string {
	if p.UpdateAvailable == "true" {
		return "YES ⚠"
	}
	return "No"
}

func formatCSV(w io.Writer, result *AuditResult) error {
	cw := csv.NewWriter(w)

	// Header
	if err := cw.Write([]string{
		"plugin_id", "name", "installed_version", "latest_version",
		"update_available", "status", "type", "source", "marketplace_url",
	}); err != nil {
		return err
	}

	for _, p := range result.Plugins {
		if err := cw.Write([]string{
			p.PluginID,
			p.Name,
			p.InstalledVersion,
			p.LatestVersion,
			p.UpdateAvailable,
			p.Status,
			p.PluginType,
			p.Source,
			p.MarketplaceURL,
		}); err != nil {
			return err
		}
	}

	cw.Flush()
	return cw.Error()
}

// jsonOutput is the JSON-specific output structure with summary at top level.
type jsonOutput struct {
	Plugins []jsonPlugin `json:"plugins"`
	Summary AuditSummary `json:"summary"`
}

type jsonPlugin struct {
	PluginID         string `json:"plugin_id"`
	Name             string `json:"name"`
	InstalledVersion string `json:"installed_version"`
	LatestVersion    string `json:"latest_version"`
	UpdateAvailable  *bool  `json:"update_available"`
	Status           string `json:"status"`
	PluginType       string `json:"type"`
	Source           string `json:"source"`
	MarketplaceURL   string `json:"marketplace_url"`
}

func formatJSON(w io.Writer, result *AuditResult) error {
	plugins := make([]jsonPlugin, 0, len(result.Plugins))
	for _, p := range result.Plugins {
		jp := jsonPlugin{
			PluginID:         p.PluginID,
			Name:             p.Name,
			InstalledVersion: p.InstalledVersion,
			LatestVersion:    p.LatestVersion,
			UpdateAvailable:  p.UpdateAvailJSON,
			Status:           p.Status,
			PluginType:       p.PluginType,
			Source:           p.Source,
			MarketplaceURL:   p.MarketplaceURL,
		}
		plugins = append(plugins, jp)
	}

	out := jsonOutput{
		Plugins: plugins,
		Summary: result.Summary,
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w)
	return err
}

func capitalizeStatus(s string) string {
	if s == "enabled" {
		return "Enabled"
	}
	if s == "disabled" {
		return "Disabled"
	}
	return s
}
