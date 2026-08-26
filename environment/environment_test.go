package environment

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadMergesPublicAndPrivateProfiles(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, DefaultDotEnvFile), []byte("host=dotenv.example.com\ndotenv_only=private-base\ntoken=dotenv-token\n"), 0600); err != nil {
		t.Fatal(err)
	}
	public := `{"development":{"host":"api.example.com","client":{"version":2},"debug":true,"token":"public"}}`
	private := `{"development":{"token":"private-secret"}}`
	if err := os.WriteFile(filepath.Join(root, DefaultPublicFile), []byte(public), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, DefaultPrivateFile), []byte(private), 0600); err != nil {
		t.Fatal(err)
	}

	result, err := Load(root, Config{Name: "development"})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"host":           "api.example.com",
		"client.version": "2",
		"debug":          "true",
		"token":          "private-secret",
		"dotenv_only":    "private-base",
	}
	if !reflect.DeepEqual(result.Values, want) {
		t.Fatalf("unexpected variables: got %#v, want %#v", result.Values, want)
	}
	if !reflect.DeepEqual(result.SecretVariables, []string{"dotenv_only", "token"}) {
		t.Fatalf("unexpected secret variables: %#v", result.SecretVariables)
	}
}

func TestLoadWithoutSelectedProfileDoesNotRequireFiles(t *testing.T) {
	result, err := Load(t.TempDir(), Config{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "" || len(result.Values) != 0 {
		t.Fatalf("unexpected empty environment: %+v", result)
	}
}

func TestLoadDotEnvWithoutSelectedProfile(t *testing.T) {
	root := t.TempDir()
	contents := "\ufeff# project values\n" +
		"export HOST=api.example.com\n" +
		"TOKEN='hash # value' # private\n" +
		"MESSAGE=\"hello\\nworld\"\n" +
		"URL=https://example.test/#fragment\n" +
		"SPACED=value # comment\n" +
		"EMPTY=\n"
	if err := os.WriteFile(filepath.Join(root, DefaultDotEnvFile), []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}

	result, err := Load(root, Config{})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"HOST":    "api.example.com",
		"TOKEN":   "hash # value",
		"MESSAGE": "hello\nworld",
		"URL":     "https://example.test/#fragment",
		"SPACED":  "value",
		"EMPTY":   "",
	}
	if !reflect.DeepEqual(result.Values, want) {
		t.Fatalf("unexpected dotenv variables: got %#v, want %#v", result.Values, want)
	}
	wantSecrets := []string{"EMPTY", "HOST", "MESSAGE", "SPACED", "TOKEN", "URL"}
	if !reflect.DeepEqual(result.SecretVariables, wantSecrets) {
		t.Fatalf("unexpected dotenv secrets: got %#v, want %#v", result.SecretVariables, wantSecrets)
	}
}

func TestLoadDotEnvReportsLineErrors(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, DefaultDotEnvFile)
	if err := os.WriteFile(path, []byte("GOOD=value\nBROKEN\n"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := Load(root, Config{})
	if err == nil || !strings.Contains(err.Error(), path+":2") || !strings.Contains(err.Error(), "NAME=value") {
		t.Fatalf("unexpected dotenv error: %v", err)
	}
}

func TestLoadUsesCustomDotEnvFile(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".env.lazyrest"), []byte("TOKEN=custom\n"), 0600); err != nil {
		t.Fatal(err)
	}

	result, err := Load(root, Config{DotEnvFile: ".env.lazyrest"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Values["TOKEN"] != "custom" {
		t.Fatalf("custom dotenv file was not loaded: %#v", result.Values)
	}
}

func TestLoadRejectsMissingProfile(t *testing.T) {
	_, err := Load(t.TempDir(), Config{Name: "missing"})
	if err == nil || !strings.Contains(err.Error(), `environment "missing" was not found`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadRejectsInvalidVariableName(t *testing.T) {
	root := t.TempDir()
	content := `{"development":{"invalid name":"value"}}`
	if err := os.WriteFile(filepath.Join(root, DefaultPublicFile), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(root, Config{Name: "development"})
	if err == nil || !strings.Contains(err.Error(), "variable name") {
		t.Fatalf("unexpected error: %v", err)
	}
}
