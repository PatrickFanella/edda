package main

import "testing"

func TestParseConfigPathUsesDefaultWhenFlagMissing(t *testing.T) {
	defaultPath := "/tmp/default.yaml"

	got, err := parseConfigPath(nil, defaultPath)
	if err != nil {
		t.Fatalf("parseConfigPath() error = %v", err)
	}
	if got != defaultPath {
		t.Fatalf("expected default path %q, got %q", defaultPath, got)
	}
}

func TestParseConfigPathUsesFlagOverride(t *testing.T) {
	defaultPath := "/tmp/default.yaml"
	overridePath := "/tmp/override.yaml"

	got, err := parseConfigPath([]string{"--config", overridePath}, defaultPath)
	if err != nil {
		t.Fatalf("parseConfigPath() error = %v", err)
	}
	if got != overridePath {
		t.Fatalf("expected override path %q, got %q", overridePath, got)
	}
}

func TestParseConfigPathReturnsErrorForUnknownFlag(t *testing.T) {
	if _, err := parseConfigPath([]string{"--unknown"}, ""); err == nil {
		t.Fatal("expected error for unknown flag, got nil")
	}
}
