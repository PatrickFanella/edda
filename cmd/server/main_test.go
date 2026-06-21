package main

import "testing"

func TestParseConfigPath(t *testing.T) {
	t.Parallel()

	defaultPath := "/tmp/default.yaml"
	got, err := parseConfigPath(nil, defaultPath)
	if err != nil {
		t.Fatalf("parseConfigPath() error = %v", err)
	}
	if got != defaultPath {
		t.Fatalf("parseConfigPath() = %q, want %q", got, defaultPath)
	}

	overridePath := "/tmp/override.yaml"
	got, err = parseConfigPath([]string{"--config", overridePath}, defaultPath)
	if err != nil {
		t.Fatalf("parseConfigPath() error = %v", err)
	}
	if got != overridePath {
		t.Fatalf("parseConfigPath() = %q, want %q", got, overridePath)
	}
}

func TestParseConfigPathUnknownFlag(t *testing.T) {
	t.Parallel()

	if _, err := parseConfigPath([]string{"--unknown"}, ""); err == nil {
		t.Fatal("parseConfigPath() expected error for unknown flag, got nil")
	}
}
