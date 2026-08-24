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
	}
	if !reflect.DeepEqual(result.Values, want) {
		t.Fatalf("unexpected variables: got %#v, want %#v", result.Values, want)
	}
	if !reflect.DeepEqual(result.SecretVariables, []string{"token"}) {
		t.Fatalf("unexpected secret variables: %#v", result.SecretVariables)
	}
}

func TestLoadWithoutSelectedProfileDoesNotReadFiles(t *testing.T) {
	result, err := Load(t.TempDir(), Config{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "" || len(result.Values) != 0 {
		t.Fatalf("unexpected empty environment: %+v", result)
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
