package http

import (
	"context"
	nethttp "net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func loginStore() *ResponseStore {
	store := &ResponseStore{}
	store.Record(HttpSuite{Name: "login"}, ResponseValue{
		Body: `{"token": "abc123", "user": {"id": 42, "roles": ["admin", "dev"]}, "ok": true}`,
		Header: nethttp.Header{
			"X-Session": []string{"s-1"},
			"Location":  []string{"https://example.com/next"},
		},
	})
	return store
}

func TestResolveResponseReferences(t *testing.T) {
	cases := []struct {
		name      string
		reference string
		want      string
		secret    bool
	}{
		{name: "a string member", reference: "{{login.response.body.$.token}}", want: "abc123", secret: true},
		{name: "a nested number", reference: "{{login.response.body.$.user.id}}", want: "42"},
		{name: "an array index", reference: "{{login.response.body.$.user.roles[0]}}", want: "admin"},
		{name: "a boolean", reference: "{{login.response.body.$.ok}}", want: "true"},
		{name: "a whole object", reference: "{{login.response.body.$.user.roles}}", want: `["admin","dev"]`},
		{name: "the whole body", reference: "{{login.response.body}}", want: `{"token": "abc123", "user": {"id": 42, "roles": ["admin", "dev"]}, "ok": true}`, secret: true},
		{name: "a header", reference: "{{login.response.headers.X-Session}}", want: "s-1", secret: true},
		{name: "a header by another case", reference: "{{login.response.headers.x-session}}", want: "s-1", secret: true},
		{name: "spaces inside the braces", reference: "{{ login.response.body.$.token }}", want: "abc123", secret: true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			suite := HttpSuite{Uri: "https://example.com/" + testCase.reference, Header: nethttp.Header{}}
			unresolved := ResolveResponseReferences(&suite, loginStore())
			if len(unresolved) != 0 {
				t.Fatalf("unexpected failures: %v", unresolved)
			}
			if suite.Uri != "https://example.com/"+testCase.want {
				t.Errorf("got %q, want %q", suite.Uri, "https://example.com/"+testCase.want)
			}
			if marked := slices.Contains(suite.SecretValues, testCase.want); marked != testCase.secret {
				t.Errorf("secret classification for %q is %v, want %v: %#v", testCase.reference, marked, testCase.secret, suite.SecretValues)
			}
		})
	}
}

func TestResolveResponseReferences_ReachesEveryPartOfARequest(t *testing.T) {
	suite := HttpSuite{
		Uri:              "https://example.com/{{login.response.body.$.user.id}}",
		Body:             `{"token": "{{login.response.body.$.token}}"}`,
		GraphQLVariables: `{"id": "{{login.response.body.$.user.id}}"}`,
		Header:           nethttp.Header{"Authorization": []string{"Bearer {{login.response.body.$.token}}"}},
	}

	if unresolved := ResolveResponseReferences(&suite, loginStore()); len(unresolved) != 0 {
		t.Fatalf("unexpected failures: %v", unresolved)
	}
	if suite.Uri != "https://example.com/42" {
		t.Errorf("the uri was not resolved: %q", suite.Uri)
	}
	if suite.Body != `{"token": "abc123"}` {
		t.Errorf("the body was not resolved: %q", suite.Body)
	}
	if suite.GraphQLVariables != `{"id": "42"}` {
		t.Errorf("the variables were not resolved: %q", suite.GraphQLVariables)
	}
	if got := suite.Header.Get("Authorization"); got != "Bearer abc123" {
		t.Errorf("the header was not resolved: %q", got)
	}
	if !slices.Contains(suite.SecretValues, "abc123") {
		t.Fatalf("runtime token was not marked as secret: %#v", suite.SecretValues)
	}
}

func TestResolveResponseReferences_TreatsAnyValueInASensitiveHeaderAsSecret(t *testing.T) {
	suite := HttpSuite{
		Header: nethttp.Header{"X-Api-Key": []string{"{{login.response.body.$.user.id}}"}},
	}
	if unresolved := ResolveResponseReferences(&suite, loginStore()); len(unresolved) != 0 {
		t.Fatalf("unexpected failures: %v", unresolved)
	}
	if !slices.Contains(suite.SecretValues, "42") {
		t.Fatalf("value placed in a sensitive header was not protected: %#v", suite.SecretValues)
	}
}

func TestResolveResponseReferences_ReportsWhatItCannotResolve(t *testing.T) {
	cases := []struct {
		name      string
		reference string
		contains  string
	}{
		{name: "a request that has not run", reference: "{{signup.response.body.$.token}}", contains: `"signup" has not been run yet`},
		{name: "a member that is absent", reference: "{{login.response.body.$.missing}}", contains: `no "missing"`},
		{name: "an index out of range", reference: "{{login.response.body.$.user.roles[9]}}", contains: "out of range"},
		{name: "a header that is absent", reference: "{{login.response.headers.X-Absent}}", contains: "no \"X-Absent\" header"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			suite := HttpSuite{Uri: testCase.reference, Header: nethttp.Header{}}
			unresolved := ResolveResponseReferences(&suite, loginStore())
			if len(unresolved) != 1 {
				t.Fatalf("expected one failure, got %v", unresolved)
			}
			if !strings.Contains(unresolved[0], testCase.contains) {
				t.Errorf("got %q, want it to mention %q", unresolved[0], testCase.contains)
			}
			// The reference stays, so the request shows what it waited for.
			if suite.Uri != testCase.reference {
				t.Errorf("the reference was replaced anyway: %q", suite.Uri)
			}
		})
	}
}

func TestResolveResponseReferences_ReportsABodyThatIsNotJSON(t *testing.T) {
	store := ResponseStore{}
	store.Record(HttpSuite{Name: "login"}, ResponseValue{Body: "plain text", Header: nethttp.Header{}})
	suite := HttpSuite{Uri: "{{login.response.body.$.token}}", Header: nethttp.Header{}}

	unresolved := ResolveResponseReferences(&suite, &store)
	if len(unresolved) != 1 || !strings.Contains(unresolved[0], "not JSON") {
		t.Fatalf("unexpected failures: %v", unresolved)
	}
}

func TestResolveResponseReferences_ScopesRequestNamesToTheirSourceFile(t *testing.T) {
	store := ResponseStore{}
	store.Record(HttpSuite{SourceFilePath: "first.http", Name: "login"}, ResponseValue{
		Body:   `{"token":"first-token"}`,
		Header: nethttp.Header{},
	})
	store.Record(HttpSuite{SourceFilePath: "second.http", Name: "login"}, ResponseValue{
		Body:   `{"token":"second-token"}`,
		Header: nethttp.Header{},
	})

	first := HttpSuite{
		SourceFilePath: "first.http",
		Uri:            "https://example.test/{{login.response.body.$.token}}",
		Header:         nethttp.Header{},
	}
	if unresolved := ResolveResponseReferences(&first, &store); len(unresolved) != 0 {
		t.Fatalf("first file did not find its response: %v", unresolved)
	}
	if first.Uri != "https://example.test/first-token" {
		t.Fatalf("first file used the wrong response: %q", first.Uri)
	}

	second := HttpSuite{
		SourceFilePath: "second.http",
		Uri:            "https://example.test/{{login.response.body.$.token}}",
		Header:         nethttp.Header{},
	}
	if unresolved := ResolveResponseReferences(&second, &store); len(unresolved) != 0 {
		t.Fatalf("second file did not find its response: %v", unresolved)
	}
	if second.Uri != "https://example.test/second-token" {
		t.Fatalf("second file used the wrong response: %q", second.Uri)
	}

	third := HttpSuite{
		SourceFilePath: "third.http",
		Uri:            "https://example.test/{{login.response.body.$.token}}",
		Header:         nethttp.Header{},
	}
	unresolved := ResolveResponseReferences(&third, &store)
	if len(unresolved) != 1 || !strings.Contains(unresolved[0], `"login" has not been run yet`) {
		t.Fatalf("third file reused another file's response: %v", unresolved)
	}
}

func TestParseFile_DoesNotReportAResponseReferenceAsUndefined(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "chain.http")
	content := `# @name login
POST https://example.com/auth

# @name profile
GET https://example.com/me
Authorization: Bearer {{login.response.body.$.token}}
X-Session: {{login.response.headers.X-Session}}
`
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	parser, err := NewParser()
	if err != nil {
		t.Fatal(err)
	}
	defer parser.Close()

	result, err := parser.ParseFile(context.Background(), filePath)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("a response reference was reported as a missing variable: %+v", result.Diagnostics)
	}
	if len(result.Suites) != 2 {
		t.Fatalf("expected two requests, got %d", len(result.Suites))
	}
	// The reference survives parsing untouched, to be filled in when it runs.
	if got := result.Suites[1].Header.Get("Authorization"); got != "Bearer {{login.response.body.$.token}}" {
		t.Errorf("the reference did not survive parsing: %q", got)
	}
}
