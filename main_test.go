package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunGenerateAndPrintConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	generated := filepath.Join(home, "generated.yml")
	var output bytes.Buffer
	if err := run([]string{"--config", generated, "--generate-config"}, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "created") {
		t.Fatalf("unexpected generate output: %q", output.String())
	}
	output.Reset()
	if err := run([]string{"--config", generated, "--print-config", t.TempDir()}, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "keybindings:") || !strings.Contains(output.String(), "theme:") {
		t.Fatalf("resolved config was not printed: %q", output.String())
	}
}

func TestRunValidatesProjectConfigAndConflicts(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	root := t.TempDir()
	projectConfig := filepath.Join(root, ".lazyrest.yml")
	if err := os.WriteFile(projectConfig, []byte("language: ru\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := run([]string{"--validate-config", root}, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "configuration is valid") {
		t.Fatalf("unexpected validation output: %q", output.String())
	}
	conflict := "keybindings:\n  help: ['x']\n  reload: ['x']\n"
	if err := os.WriteFile(projectConfig, []byte(conflict), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"--validate-config", root}, &output); err == nil {
		t.Fatal("expected conflicting project bindings to fail validation")
	}
}

func TestRunPrintsVersion(t *testing.T) {
	var output strings.Builder
	if err := run([]string{"--version"}, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(output.String(), "lazyrest ") {
		t.Fatalf("unexpected version output: %q", output.String())
	}
	if strings.TrimSpace(output.String()) == "lazyrest" {
		t.Fatalf("the version is empty: %q", output.String())
	}
}

func TestRunTreatsHelpAsSuccess(t *testing.T) {
	var output strings.Builder
	if err := run([]string{"-h"}, &output); err != nil {
		t.Fatalf("asking for help failed: %v", err)
	}
	if !strings.Contains(output.String(), "-timeout") {
		t.Fatalf("usage was not printed: %q", output.String())
	}
}

func TestVersionTextIsNeverEmpty(t *testing.T) {
	if versionText() == "" {
		t.Fatal("the version is empty")
	}
}
