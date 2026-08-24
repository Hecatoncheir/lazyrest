package ui

import (
	"testing"

	"github.com/Hecatoncheir/lazyrest/finder"
	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
)

func TestModelSnapshotDoesNotExposeMutableState(t *testing.T) {
	model := NewModel("/workspace", "development")
	model.update(func(state *State) {
		state.Directory = finder.Directory{
			Name:     "workspace",
			Files:    []finder.File{{Name: "request.http"}},
			Warnings: []string{"warning"},
		}
		state.Suites = []parserhttp.HttpSuite{{
			Name:         "Request",
			Header:       map[string]string{"Accept": "application/json"},
			SecretValues: []string{"secret"},
		}}
	})

	snapshot := model.Snapshot()
	snapshot.Directory.Files[0].Name = "changed.http"
	snapshot.Directory.Warnings[0] = "changed"
	snapshot.Suites[0].Header["Accept"] = "text/plain"
	snapshot.Suites[0].SecretValues[0] = "changed"

	current := model.Snapshot()
	if current.Directory.Files[0].Name != "request.http" || current.Directory.Warnings[0] != "warning" {
		t.Fatalf("directory state was mutated through a snapshot: %+v", current.Directory)
	}
	if current.Suites[0].Header["Accept"] != "application/json" || current.Suites[0].SecretValues[0] != "secret" {
		t.Fatalf("suite state was mutated through a snapshot: %+v", current.Suites[0])
	}
}
