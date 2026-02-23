package main

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestClassifyAPIError_Unauthorized(t *testing.T) {
	resp := &model.Response{StatusCode: 401}
	err := classifyAPIError("https://mm.example.com", resp, nil)
	if err.Code != ExitConfigError {
		t.Errorf("expected exit code %d for 401, got %d", ExitConfigError, err.Code)
	}
}

func TestClassifyAPIError_Forbidden(t *testing.T) {
	resp := &model.Response{StatusCode: 403}
	err := classifyAPIError("https://mm.example.com", resp, nil)
	if err.Code != ExitConfigError {
		t.Errorf("expected exit code %d for 403, got %d", ExitConfigError, err.Code)
	}
	if err.Message != "error: permission denied. This operation requires a System Administrator account." {
		t.Errorf("unexpected message: %s", err.Message)
	}
}

func TestClassifyAPIError_ServerError(t *testing.T) {
	resp := &model.Response{StatusCode: 500}
	err := classifyAPIError("https://mm.example.com", resp, nil)
	if err.Code != ExitAPIError {
		t.Errorf("expected exit code %d for 500, got %d", ExitAPIError, err.Code)
	}
}

func TestClassifyAPIError_ConnectionFailure(t *testing.T) {
	err := classifyAPIError("https://mm.example.com", nil, nil)
	if err.Code != ExitAPIError {
		t.Errorf("expected exit code %d for connection failure, got %d", ExitAPIError, err.Code)
	}
}

func TestClassifyAPIError_NotFound(t *testing.T) {
	resp := &model.Response{StatusCode: 404}
	err := classifyAPIError("https://mm.example.com", resp, nil)
	if err.Code != ExitAPIError {
		t.Errorf("expected exit code %d for 404, got %d", ExitAPIError, err.Code)
	}
}

func TestClassifyAPIError_NoServerURL(t *testing.T) {
	err := classifyAPIError("", nil, nil)
	if err.Code != ExitAPIError {
		t.Errorf("expected exit code %d, got %d", ExitAPIError, err.Code)
	}
	if err.Message != "error: unexpected API error." {
		t.Errorf("unexpected message: %s", err.Message)
	}
}
