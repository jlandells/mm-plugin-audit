# PRD: mm-plugin-audit — Mattermost Plugin Version Auditor

**Version:** 1.0  
**Status:** Ready for Development  
**Language:** Go  
**Binary Name:** `mm-plugin-audit`

---

## 1. Overview

`mm-plugin-audit` is a standalone command-line utility that compares the plugins installed on a Mattermost instance against the latest versions available in the Mattermost Marketplace. It produces a report showing each plugin's installed version, the latest available version, whether an update is available, and whether the plugin is currently enabled. Custom or private plugins not listed in the Marketplace are reported separately.

---

## 2. Background & Problem Statement

The Mattermost System Console shows which plugins are installed and their current version numbers, but gives no indication of whether a newer version is available. Administrators must manually visit the Marketplace and check each plugin individually — a process that is tedious and frequently skipped.

In regulated environments (defence, government, financial services), there is often a formal requirement to demonstrate that all software components are kept up to date, particularly given that plugin updates frequently include security patches. Without a tool like this, compliance with that requirement is difficult to evidence.

---

## 3. Goals

- Compare installed plugin versions against the Mattermost Marketplace
- Clearly identify which plugins have updates available
- Identify plugins that are installed but not in the Marketplace (custom/private plugins)
- Report enabled/disabled status for each plugin
- Produce output suitable for compliance reporting (CSV, JSON) or quick review (table)

---

## 4. Non-Goals

- This tool does not install, update, enable, disable, or remove plugins — it is read-only
- This tool does not assess plugin compatibility with the running Mattermost version
- This tool does not assess the security posture of individual plugins
- This tool does not report on plugin configuration

---

## 5. Target Users

Mattermost System Administrators and IT Operations teams responsible for keeping instances patched and up to date, particularly in regulated or security-conscious environments.

---

## 6. User Stories

- As a System Administrator, I want to know which of my installed plugins have updates available so that I can plan maintenance accordingly.
- As a Security Officer, I want a report showing that all plugins are on their latest versions so that I can include it in our patch compliance evidence.
- As a System Administrator, I want to know which plugins are installed but not from the official Marketplace so that I can review custom plugins separately.
- As a System Administrator, I want to see which plugins are currently disabled so that I can decide whether to remove them.

---

## 7. Functional Requirements

### 7.1 Installed Plugin Inventory

- The tool MUST retrieve the full list of installed plugins from the Mattermost instance
- For each plugin, the tool MUST report:
  - Plugin ID
  - Plugin name (display name)
  - Installed version
  - Enabled or disabled status
  - Whether it is a server-side plugin, webapp plugin, or both

### 7.2 Marketplace Version Lookup

- The tool MUST query the Mattermost Marketplace API to retrieve the latest available version for each installed plugin
- The Marketplace API is public and requires no authentication
- Version comparison MUST use semantic versioning logic (i.e. `1.10.0 > 1.9.0`) — naive string comparison is not acceptable
- If a plugin is present in the Marketplace, the tool MUST report:
  - Latest available version
  - Whether an update is available (boolean)
  - Marketplace URL for the plugin (for convenience)

### 7.3 Custom / Private Plugins

- If an installed plugin is not found in the Marketplace, it MUST be flagged as `custom/private` and reported in a separate section (in table view) or with a distinct field value (in CSV/JSON)
- The tool MUST NOT treat a Marketplace lookup failure as an error for an individual plugin — it should be treated as "not in Marketplace"

### 7.4 Filtering

- `--outdated-only` MUST filter the output to show only plugins for which an update is available
- Plugins not in the Marketplace should still appear when `--outdated-only` is used, as they cannot be confirmed as up to date

### 7.5 Marketplace Connectivity

- In air-gapped or restricted network environments, the Marketplace API (`api.integrations.mattermost.com`) may not be reachable
- If the Marketplace API cannot be reached, the tool MUST NOT silently show all plugins as up to date
- It MUST exit with a clear error message explaining that the Marketplace is unreachable and suggesting the admin check network connectivity or consult the Marketplace manually
- This behaviour should be documented prominently in the README, as air-gapped environments are a common deployment pattern for this tool's target users

---

## 8. CLI Specification

### Usage

```
mm-plugin-audit [flags]
```

### Connection Flags (required)

| Flag | Environment Variable | Description |
|------|----------------------|-------------|
| `--url URL` | `MM_URL` | Mattermost server URL, e.g. `https://mattermost.example.com` |

### Authentication Flags

| Flag | Environment Variable | Description |
|------|----------------------|-------------|
| `--token TOKEN` | `MM_TOKEN` | Personal Access Token (preferred) |
| `--username USERNAME` | `MM_USERNAME` | Username for password-based auth |
| *(no flag)* | `MM_PASSWORD` | Password (env var only — never a CLI flag) |

Authentication resolution order:
1. `--token` / `MM_TOKEN`
2. `--username` + interactive password prompt (if terminal is interactive)
3. `--username` + `MM_PASSWORD` environment variable

### Operational Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--outdated-only` | `false` | Show only plugins with available updates (plus custom/private plugins) |
| `--format table\|csv\|json` | `table` | Output format |
| `--output FILE` | *(stdout)* | Write output to a file |
| `--verbose` / `-v` | `false` | Enable verbose logging to stderr |

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Configuration error (missing URL, invalid auth) |
| `2` | API error — Mattermost instance unreachable or unexpected response |
| `3` | Marketplace unreachable (network error contacting `api.integrations.mattermost.com`) |
| `4` | Output error (unable to write file) |

Note: exit code 3 is distinct so that scripts can detect and handle the air-gapped case separately.

---

## 9. Output Specification

### 9.1 Table Format

Two sections: Marketplace plugins, then Custom/Private plugins.

```
=== Marketplace Plugins (12) ===
NAME                    | INSTALLED | LATEST  | UPDATE?  | STATUS
GitHub                  | 2.1.4     | 2.2.0   | YES ⚠    | Enabled
Jira                    | 4.0.1     | 4.0.1   | No       | Enabled
Zoom                    | 1.6.0     | 1.6.0   | No       | Disabled
...

=== Custom / Private Plugins (2) ===
NAME                    | INSTALLED | STATUS
my-internal-plugin      | 1.0.3     | Enabled
legacy-integration      | 0.9.1     | Disabled
```

### 9.2 CSV Format

One row per plugin. Columns:

```
plugin_id, name, installed_version, latest_version, update_available, status, marketplace, marketplace_url
```

- `latest_version` — empty string if not in Marketplace
- `update_available` — `true`, `false`, or `unknown` (if not in Marketplace)
- `status` — `enabled` or `disabled`
- `marketplace` — `true` or `false`
- `marketplace_url` — URL if in Marketplace, empty string otherwise

### 9.3 JSON Format

```json
[
  {
    "plugin_id": "com.mattermost.plugin-github",
    "name": "GitHub",
    "installed_version": "2.1.4",
    "latest_version": "2.2.0",
    "update_available": true,
    "status": "enabled",
    "marketplace": true,
    "marketplace_url": "https://api.integrations.mattermost.com/..."
  },
  {
    "plugin_id": "com.example.my-plugin",
    "name": "my-internal-plugin",
    "installed_version": "1.0.3",
    "latest_version": "",
    "update_available": false,
    "status": "enabled",
    "marketplace": false,
    "marketplace_url": ""
  }
]
```

---

## 10. Authentication Detail

The token or user account MUST have **System Administrator** role to access the plugin management API endpoints.

Password handling:
- Interactive terminal: prompt with echo suppressed via `golang.org/x/term`
- Non-interactive: use `MM_PASSWORD` environment variable
- Never accept password as a CLI flag

---

## 11. API Endpoints Used

| Endpoint | Purpose |
|----------|---------|
| `GET /api/v4/plugins` | List all installed plugins and their status |
| `GET https://api.integrations.mattermost.com/api/v1/latest?plugin_id={id}` | Look up latest Marketplace version for a plugin |

Note: the Marketplace API endpoint format should be verified against current Mattermost Marketplace API documentation before implementation, as it may change.

---

## 12. Error Handling

- Missing `--url` / `MM_URL`: exit code 1 with clear message
- Authentication failure (401): exit code 1 with clear message
- Permission denied (403): exit code 1, message should explicitly state that a System Administrator account is required
- Mattermost instance unreachable: exit code 2 with clear message
- Marketplace API unreachable: exit code 3 with message: `Warning: Mattermost Marketplace is unreachable. Version comparison is not possible. If you are in an air-gapped environment, please consult the Marketplace manually.`
- Output file write error: exit code 4 with clear message

---

## 13. Testing Requirements

- Unit tests for semantic version comparison logic (must correctly handle cases like `1.10.0 > 1.9.0`)
- Unit tests for Marketplace-not-reachable handling
- Unit tests for custom/private plugin detection
- Unit tests for CSV and JSON output formatting
- Mock API responses for both the Mattermost instance and the Marketplace API

---

## 14. Out of Scope

- Installing, updating, or removing plugins
- Assessing plugin compatibility with the running Mattermost server version
- Checking plugin configuration or health

---

## 15. Acceptance Criteria

- [ ] Running with valid credentials lists all installed plugins with version information
- [ ] Plugins with available updates are clearly flagged
- [ ] `--outdated-only` returns only plugins with available updates and custom/private plugins
- [ ] Custom/private plugins (not in Marketplace) are listed separately and not treated as errors
- [ ] Version comparison correctly identifies `1.10.0` as newer than `1.9.0`
- [ ] If Marketplace is unreachable, a clear error is shown and the tool exits with code 3
- [ ] `--format csv --output plugins.csv` produces a valid CSV
- [ ] `--format json` produces valid, `jq`-parseable JSON
- [ ] All errors go to stderr; all data output goes to stdout
- [ ] Binary runs on Linux (amd64), macOS (arm64 and amd64), and Windows (amd64) without dependencies