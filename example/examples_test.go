package example

import (
	"context"
	"path/filepath"
	"testing"

	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	parserhurl "github.com/Hecatoncheir/lazyrest/parser/hurl"
)

func TestHTTPExamplesParseWithoutDiagnostics(t *testing.T) {
	paths, err := filepath.Glob("*.http")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no HTTP examples found")
	}

	parser, err := parserhttp.NewParser()
	if err != nil {
		t.Fatal(err)
	}
	defer parser.Close()

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			result, err := parser.ParseFile(context.Background(), path)
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Suites) == 0 {
				t.Fatal("example contains no requests")
			}
			if len(result.Diagnostics) != 0 {
				t.Fatalf("example contains parser diagnostics: %+v", result.Diagnostics)
			}
		})
	}
}

func TestHurlExamplesAreRunnableSessions(t *testing.T) {
	paths, err := filepath.Glob("*.hurl")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no Hurl examples found")
	}

	parser, err := parserhurl.NewParser()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			suites, err := parser.GetSuitesFromFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if len(suites) == 0 {
				t.Fatalf("no entries were read from %s", path)
			}
			for index, suite := range suites {
				if !suite.IsHurl {
					t.Errorf("entry %d is not marked as Hurl: %+v", index, suite)
				}
				if suite.HurlEntry != index+1 {
					t.Errorf("entry %d is numbered %d", index, suite.HurlEntry)
				}
			}
		})
	}
}
