package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

var version = "dev"

// logOutput is the writer for verbose/error output (stderr). Exposed for testing.
var logOutput io.Writer = os.Stderr

func main() {
	os.Exit(run())
}

func run() int {
	// Define flags
	urlFlag := flag.String("url", "", "Mattermost server URL (or set MM_URL)")
	tokenFlag := flag.String("token", "", "Personal Access Token (or set MM_TOKEN)")
	usernameFlag := flag.String("username", "", "Username for password auth (or set MM_USERNAME)")
	formatFlag := flag.String("format", "table", "Output format: table, csv, json")
	outputFlag := flag.String("output", "", "Write output to file")
	outdatedOnly := flag.Bool("outdated-only", false, "Show only plugins with available updates (plus custom/private)")
	verbose := flag.Bool("verbose", false, "Enable verbose logging to stderr")
	showVersion := flag.Bool("version", false, "Print version and exit")

	// Short flags
	flag.BoolVar(verbose, "v", false, "Enable verbose logging to stderr")

	flag.Parse()

	if *showVersion {
		fmt.Fprintf(os.Stdout, "mm-plugin-audit %s\n", version)
		return ExitSuccess
	}

	// Resolve URL
	serverURL := resolveFlag(*urlFlag, "MM_URL")
	if serverURL == "" {
		fmt.Fprintln(os.Stderr, "error: server URL is required. Use --url or set the MM_URL environment variable.")
		return ExitConfigError
	}
	serverURL = strings.TrimRight(serverURL, "/")

	// Validate format
	format := strings.ToLower(*formatFlag)
	if format != "table" && format != "csv" && format != "json" {
		fmt.Fprintf(os.Stderr, "error: invalid format %q. Use table, csv, or json.\n", *formatFlag)
		return ExitConfigError
	}

	// Resolve authentication
	token := resolveFlag(*tokenFlag, "MM_TOKEN")
	username := resolveFlag(*usernameFlag, "MM_USERNAME")

	var password string
	if token == "" && username == "" {
		fmt.Fprintln(os.Stderr, "error: authentication required. Use --token (or MM_TOKEN) for token auth, or --username (or MM_USERNAME) for password auth.")
		return ExitConfigError
	}

	if token == "" && username != "" {
		// Need password
		if term.IsTerminal(int(os.Stdin.Fd())) {
			fmt.Fprint(os.Stderr, "Password: ")
			passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Fprintln(os.Stderr) // move to next line
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: failed to read password: %v\n", err)
				return ExitConfigError
			}
			password = string(passwordBytes)
		} else {
			password = os.Getenv("MM_PASSWORD")
			if password == "" {
				fmt.Fprintln(os.Stderr, "error: password required. Set MM_PASSWORD environment variable for non-interactive use.")
				return ExitConfigError
			}
		}
	}

	logf := verboseLogger(*verbose)

	// Create Mattermost client
	logf("Connecting to %s...", serverURL)
	mmClient, err := NewMMClient(ClientConfig{
		URL:      serverURL,
		Token:    token,
		Username: username,
		Password: password,
	})
	if err != nil {
		if cliErr, ok := err.(*CLIError); ok {
			fmt.Fprintln(os.Stderr, cliErr.Message)
			return cliErr.Code
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return ExitAPIError
	}

	// Run audit
	result, err := RunAudit(mmClient, AuditOptions{
		OutdatedOnly: *outdatedOnly,
		Verbose:      *verbose,
	}, logf)
	if err != nil {
		if cliErr, ok := err.(*CLIError); ok {
			fmt.Fprintln(os.Stderr, cliErr.Message)
			return cliErr.Code
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return ExitAPIError
	}

	// Determine output writer
	var w io.Writer = os.Stdout
	if *outputFlag != "" {
		f, err := os.Create(*outputFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: unable to write to %s (%v), falling back to stdout\n", *outputFlag, err)
		} else {
			defer f.Close()
			w = f
		}
	}

	// Write output
	if err := FormatOutput(w, result, format); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to write output: %v\n", err)
		return ExitOutputError
	}

	return ExitSuccess
}

// resolveFlag returns the flag value if set, otherwise falls back to the environment variable.
func resolveFlag(flagVal, envVar string) string {
	if flagVal != "" {
		return flagVal
	}
	return os.Getenv(envVar)
}
