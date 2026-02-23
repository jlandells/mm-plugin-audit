package main

import "fmt"

// Exit codes
const (
	ExitSuccess          = 0 // Success
	ExitConfigError      = 1 // Missing URL, invalid auth, bad flags
	ExitAPIError         = 2 // Mattermost instance unreachable or unexpected response
	ExitMarketplaceError = 3 // Marketplace API unreachable (air-gapped)
	ExitOutputError      = 4 // Unable to write output file
)

// CLIError wraps an error with an exit code for structured error handling.
type CLIError struct {
	Code    int
	Message string
	Err     error
}

func (e *CLIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *CLIError) Unwrap() error {
	return e.Err
}

func configError(msg string, err error) *CLIError {
	return &CLIError{Code: ExitConfigError, Message: msg, Err: err}
}

func apiError(msg string, err error) *CLIError {
	return &CLIError{Code: ExitAPIError, Message: msg, Err: err}
}

func marketplaceError(msg string, err error) *CLIError {
	return &CLIError{Code: ExitMarketplaceError, Message: msg, Err: err}
}

func outputError(msg string, err error) *CLIError {
	return &CLIError{Code: ExitOutputError, Message: msg, Err: err}
}
