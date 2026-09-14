package example

import (
	"context"
	"path/filepath"
	"strings"
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

func TestSocketExamplesParseAsRawStreams(t *testing.T) {
	paths, err := filepath.Glob("*.socket")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no socket examples found")
	}

	parser, err := parserhttp.NewParser()
	if err != nil {
		t.Fatal(err)
	}
	defer parser.Close()

	for _, path := range paths {
		result, err := parser.ParseFileWithOptions(context.Background(), path, parserhttp.ParseOptions{})
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if len(result.Diagnostics) != 0 {
			t.Errorf("%s: %v", path, result.Diagnostics)
		}
		if len(result.Suites) == 0 {
			t.Fatalf("%s: no requests parsed", path)
		}
		for _, suite := range result.Suites {
			if suite.Transport != parserhttp.TransportTCP {
				t.Errorf("%s: %q transport = %v, want tcp", path, suite.Name, suite.Transport)
			}
			// The body is the opening message. An example whose body is never
			// sent would promise what does not happen.
			if strings.TrimSpace(suite.Body) == "" {
				t.Errorf("%s: %q has no opening message", path, suite.Name)
			}
		}
	}
}

// A WebSocket lives in a .http file, so the variable in its URL has to be
// substituted before the transport is derived from the scheme.
func TestHTTPExamplesDeriveTheWebSocketTransport(t *testing.T) {
	parser, err := parserhttp.NewParser()
	if err != nil {
		t.Fatal(err)
	}
	defer parser.Close()

	result, err := parser.ParseFileWithOptions(context.Background(), "streams.http", parserhttp.ParseOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Suites) == 0 {
		t.Fatal("no requests parsed")
	}
	for _, suite := range result.Suites {
		if suite.Transport != parserhttp.TransportWebSocket {
			t.Errorf("%q transport = %v, want websocket (uri %q)", suite.Name, suite.Transport, suite.Uri)
		}
		if strings.TrimSpace(suite.Body) == "" {
			t.Errorf("%q has no opening message", suite.Name)
		}
	}
}

// A STOMP frame ends on a null byte, which a request file spells \0. Losing
// that suffix would leave the examples parsing cleanly and still malformed.
func TestStompExamplesEndOnANullEscape(t *testing.T) {
	parser, err := parserhttp.NewParser()
	if err != nil {
		t.Fatal(err)
	}
	defer parser.Close()

	found := 0
	for _, path := range []string{"stomp.socket", "streams.http"} {
		result, err := parser.ParseFileWithOptions(context.Background(), path, parserhttp.ParseOptions{})
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		for _, suite := range result.Suites {
			if !strings.HasPrefix(suite.Body, "CONNECT") {
				continue
			}
			found++
			if !strings.HasSuffix(suite.Body, `\0`) {
				t.Errorf("%s: %q does not end on a null escape: %q", path, suite.Name, suite.Body)
			}
			// The blank line between the headers and the body is part of the
			// frame, not formatting.
			if !strings.Contains(suite.Body, "\n\n") {
				t.Errorf("%s: %q lost the blank line that ends its headers", path, suite.Name)
			}
		}
	}
	if found != 2 {
		t.Fatalf("found %d STOMP examples, want one per transport", found)
	}
}
