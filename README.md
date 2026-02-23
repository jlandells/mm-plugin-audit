# mm-plugin-audit

A command-line utility that compares the plugins installed on a Mattermost instance against the
latest versions available in the Mattermost Marketplace, producing a clear report of which plugins
have updates available, which are up to date, and which are bundled Mattermost plugins or
third-party/custom plugins not tracked in the Marketplace.

## Why You'd Use It

The Mattermost System Console shows installed plugins and their versions, but gives no indication
of whether a newer version is available. Checking each plugin manually against the Marketplace is
tedious and frequently skipped. In regulated environments — defence, government, financial
services — there is often a formal requirement to demonstrate that all software components are
kept up to date. `mm-plugin-audit` automates this check and produces output suitable for compliance
evidence.

## Installation

Download the pre-built binary for your platform from the
[Releases](https://github.com/jlandells/mm-plugin-audit/releases) page.

| Platform       | Filename                            |
|----------------|-------------------------------------|
| Linux (amd64)  | `mm-plugin-audit-linux-amd64`       |
| macOS (amd64)  | `mm-plugin-audit-darwin-amd64`      |
| macOS (arm64)  | `mm-plugin-audit-darwin-arm64`      |
| Windows (amd64)| `mm-plugin-audit-windows-amd64.exe` |

On Linux and macOS, make the binary executable after downloading:

```bash
chmod +x mm-plugin-audit-*
```

No other steps are required — the binary has no external dependencies.

## Authentication

The tool requires a connection to your Mattermost instance with **System Administrator**
privileges. Two authentication methods are supported.

### Personal Access Token (recommended)

Generate a Personal Access Token in **System Console > Integrations > Personal Access Tokens**
(or ask your admin to enable them). Pass it via the `--token` flag or `MM_TOKEN` environment
variable.

```bash
mm-plugin-audit --url https://mattermost.example.com --token YOUR_TOKEN
```

> **Note:** Personal Access Tokens may be disabled on some Mattermost instances. If so, use
> username/password authentication below.

### Username and Password

Pass your username via the `--username` flag or `MM_USERNAME` environment variable. You will be
prompted to enter your password interactively (the password is not displayed on screen).

```bash
mm-plugin-audit --url https://mattermost.example.com --username admin
Password:
```

For non-interactive or automation scenarios, set the `MM_PASSWORD` environment variable:

```bash
export MM_PASSWORD='your-password'
mm-plugin-audit --url https://mattermost.example.com --username admin
```

> **Note:** There is intentionally no `--password` flag. Passwords passed as CLI flags appear in
> shell history and process listings, which is a security risk.

## Usage

```
mm-plugin-audit [flags]
```

### Flag Reference

| Flag | Env Var | Type | Default | Description |
|------|---------|------|---------|-------------|
| `--url` | `MM_URL` | string | *(required)* | Mattermost server URL |
| `--token` | `MM_TOKEN` | string | *(empty)* | Personal Access Token |
| `--username` | `MM_USERNAME` | string | *(empty)* | Username for password auth |
| `--format` | *(none)* | string | `table` | Output format: `table`, `csv`, `json` |
| `--output` | *(none)* | string | *(stdout)* | Write output to this file path |
| `--outdated-only` | *(none)* | bool | `false` | Show only plugins with available updates (plus bundled and third-party) |
| `--verbose` / `-v` | *(none)* | bool | `false` | Enable verbose logging to stderr |
| `--version` | *(none)* | bool | `false` | Print version and exit |

## Examples

### Basic run with token auth

```bash
mm-plugin-audit --url https://mattermost.example.com --token YOUR_TOKEN
```

### Basic run with username/password auth

```bash
mm-plugin-audit --url https://mattermost.example.com --username admin
Password: ********
```

### Using environment variables

```bash
export MM_URL=https://mattermost.example.com
export MM_TOKEN=YOUR_TOKEN
mm-plugin-audit
```

### Writing output to a file

```bash
mm-plugin-audit --url https://mattermost.example.com --token YOUR_TOKEN \
  --format csv --output report.csv
```

### Show only outdated, bundled, and third-party plugins

```bash
mm-plugin-audit --url https://mattermost.example.com --token YOUR_TOKEN \
  --outdated-only
```

### JSON output piped to jq

```bash
mm-plugin-audit --url https://mattermost.example.com --token YOUR_TOKEN \
  --format json | jq '.plugins[] | select(.update_available == true)'
```

## Output Formats

Plugins are categorised into four groups, checked in strict priority order:

1. **Bundled** — plugins bundled with the Mattermost server distribution, identified by an
   exact built-in list of known plugin IDs (e.g. `com.mattermost.calls`, `playbooks`, `github`)
2. **Marketplace** — listed in the Mattermost Marketplace (queried via the server's proxy
   endpoint); version comparison is available for these plugins
3. **Mattermost Plugin** — official Mattermost plugins not bundled or in the Marketplace,
   identified by a `github.com/mattermost/` homepage URL in the plugin manifest
4. **Third-Party / Custom** — everything else; plugins developed in-house or by third parties

### Table (default)

Human-readable, aligned columns with separate sections for each category:

```
=== Marketplace Plugins (2) ===
NAME                    INSTALLED  LATEST   UPDATE?  STATUS
Confluence              1.3.0      1.4.0    YES ⚠    Enabled
WelcomeBot              1.2.0      1.2.0    No       Enabled

=== Mattermost Plugins (1) ===
NAME                    INSTALLED  STATUS
Google Calendar         1.1.0      Enabled

=== Bundled Mattermost Plugins (3) ===
NAME                    INSTALLED  STATUS
Calls                   1.10.0     Enabled
GitHub                  2.6.0      Enabled
Playbooks               2.6.2      Enabled

=== Third-Party / Custom Plugins (1) ===
NAME                    INSTALLED  STATUS
Pexip                   1.3.0      Enabled

Summary: 7 plugin(s) total — 2 marketplace (1 outdated, 1 up to date), 1 mattermost, 3 bundled, 1 third-party/custom — 7 enabled, 0 disabled
```

The UPDATE? column shows:
- **YES** — a newer version is available in the Marketplace
- **No** — you are running the latest Marketplace version (or newer)

### CSV

One row per plugin with a header row. Suitable for import into spreadsheets or processing with
other tools:

```csv
plugin_id,name,installed_version,latest_version,update_available,status,type,source,marketplace_url
com.mattermost.confluence,Confluence,1.3.0,1.4.0,true,enabled,both,marketplace,https://github.com/mattermost/mattermost-plugin-confluence
com.mattermost.gcal,Google Calendar,1.1.0,,unknown,enabled,both,mattermost-plugin,
com.mattermost.calls,Calls,1.10.0,,unknown,enabled,both,bundled,
com.pexip.meetings,Pexip,1.3.0,,unknown,enabled,server,third-party,
```

- `update_available`: `true`, `false`, or `unknown` (for non-Marketplace plugins)
- `source`: `marketplace`, `bundled`, `mattermost-plugin`, or `third-party`
- Empty string for fields not applicable to non-Marketplace plugins

### JSON

Structured output with a `plugins` array and `summary` object:

```json
{
  "plugins": [
    {
      "plugin_id": "com.mattermost.confluence",
      "name": "Confluence",
      "installed_version": "1.3.0",
      "latest_version": "1.4.0",
      "update_available": true,
      "status": "enabled",
      "type": "both",
      "source": "marketplace",
      "marketplace_url": "https://github.com/mattermost/mattermost-plugin-confluence"
    },
    {
      "plugin_id": "com.mattermost.gcal",
      "name": "Google Calendar",
      "installed_version": "1.1.0",
      "latest_version": "",
      "update_available": null,
      "status": "enabled",
      "type": "both",
      "source": "mattermost-plugin",
      "marketplace_url": ""
    },
    {
      "plugin_id": "com.mattermost.calls",
      "name": "Calls",
      "installed_version": "1.10.0",
      "latest_version": "",
      "update_available": null,
      "status": "enabled",
      "type": "both",
      "source": "bundled",
      "marketplace_url": ""
    },
    {
      "plugin_id": "com.pexip.meetings",
      "name": "Pexip",
      "installed_version": "1.3.0",
      "latest_version": "",
      "update_available": null,
      "status": "enabled",
      "type": "server",
      "source": "third-party",
      "marketplace_url": ""
    }
  ],
  "summary": {
    "total": 4,
    "marketplace": 1,
    "bundled": 1,
    "mattermost_plugin": 1,
    "third_party": 1,
    "outdated": 1,
    "up_to_date": 0,
    "unknown": 3,
    "enabled": 4,
    "disabled": 0
  }
}
```

- `update_available` is `true`, `false`, or `null` (for non-Marketplace plugins)
- `source` indicates how the plugin was classified
- The `summary` object provides aggregate counts for quick assessment

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Configuration error — missing URL, invalid authentication, bad flags |
| `2` | API error — Mattermost instance unreachable or unexpected response |
| `3` | Marketplace unreachable — cannot compare versions (common in air-gapped environments) |
| `4` | Output error — unable to write to the specified output file |

These codes allow the tool to be used reliably in scripts and CI/CD pipelines. For example, you
can check for exit code 3 specifically to handle the air-gapped case.

## Limitations

- **Air-gapped environments:** The tool queries the Mattermost Marketplace via your server's
  proxy endpoint (`/api/v4/plugins/marketplace`). If your server cannot reach the Marketplace,
  the proxy may return an error or a limited plugin list, which could affect version comparison
  for Marketplace plugins. Bundled, Mattermost, and third-party plugins will still be
  categorised correctly.
- **Bundled, Mattermost, and third-party plugins** cannot be checked for updates, as only
  Marketplace plugins have version comparison. Bundled plugins are identified using an exact
  built-in list of 14 known plugin IDs (e.g. `com.mattermost.calls`, `playbooks`, `github`,
  `com.github.manland.mattermost-plugin-gitlab`). Mattermost plugins are identified by a
  `github.com/mattermost/` homepage URL in the plugin manifest. Newly-released official
  plugins may initially appear as "Third-Party / Custom" until the bundled list is updated.
- **Read-only:** This tool does not install, update, enable, disable, or remove plugins. It only
  reports on the current state.
- **No compatibility check:** The tool does not assess whether a newer plugin version is compatible
  with your running Mattermost server version.

## Integration Testing

To test against a local Mattermost instance:

1. Start a local Mattermost instance (e.g. via Docker)
2. Create a System Administrator account and generate a Personal Access Token
3. Install one or more plugins via the System Console
4. Run the tool:
   ```bash
   ./mm-plugin-audit --url http://localhost:8065 --token YOUR_TOKEN -v
   ```
5. Verify the output matches the plugins you installed

## Contributing

We welcome contributions from the community! Whether it's a bug report, a feature suggestion,
or a pull request, your input is valuable to us. Please feel free to contribute in the
following ways:
- **Issues and Pull Requests**: For specific questions, issues, or suggestions for improvements,
  open an issue or a pull request in this repository.
- **Mattermost Community**: Join the discussion in the
  [Integrations and Apps](https://community.mattermost.com/core/channels/integrations) channel
  on the Mattermost Community server.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contact

For questions, feedback, or contributions regarding this project, please use the following methods:
- **Issues and Pull Requests**: For specific questions, issues, or suggestions for improvements,
  feel free to open an issue or a pull request in this repository.
- **Mattermost Community**: Join us in the Mattermost Community server, where we discuss all
  things related to extending Mattermost. You can find me in the channel
  [Integrations and Apps](https://community.mattermost.com/core/channels/integrations).
- **Social Media**: Follow and message me on Twitter, where I'm
  [@jlandells](https://twitter.com/jlandells).
