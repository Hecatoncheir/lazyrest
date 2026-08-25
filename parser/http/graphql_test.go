package http

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSplitGraphQLBody(t *testing.T) {
	cases := []struct {
		name          string
		body          string
		wantQuery     string
		wantVariables string
	}{
		{
			name:          "query with variables",
			body:          "query Example($id: ID!) {\n  user(id: $id) { name }\n}\n\n{\n  \"id\": \"42\"\n}\n",
			wantQuery:     "query Example($id: ID!) {\n  user(id: $id) { name }\n}",
			wantVariables: "{\n  \"id\": \"42\"\n}",
		},
		{
			name:      "query without variables",
			body:      "query {\n  viewer { id }\n}\n",
			wantQuery: "query {\n  viewer { id }\n}",
		},
		{
			name:      "anonymous query is not read as variables",
			body:      "{\n  viewer { id }\n}\n",
			wantQuery: "{\n  viewer { id }\n}",
		},
		{
			name:          "variables holding a nested object",
			body:          "query Example {\n  viewer { id }\n}\n\n{\n\"filter\":\n{\n\"a\": 1\n}\n}\n",
			wantQuery:     "query Example {\n  viewer { id }\n}",
			wantVariables: "{\n\"filter\":\n{\n\"a\": 1\n}\n}",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			query, variables := splitGraphQLBody(testCase.body)
			if query != testCase.wantQuery {
				t.Errorf("unexpected query: %q, want %q", query, testCase.wantQuery)
			}
			if variables != testCase.wantVariables {
				t.Errorf("unexpected variables: %q, want %q", variables, testCase.wantVariables)
			}
		})
	}
}

func TestGraphQLOperationName(t *testing.T) {
	cases := []struct {
		name  string
		query string
		want  string
	}{
		{name: "named query", query: "query GetUser($id: ID!) {\n  user(id: $id) { name }\n}", want: "GetUser"},
		{name: "named mutation", query: "mutation CreateUser {\n  createUser { id }\n}", want: "CreateUser"},
		{name: "anonymous query", query: "{\n  viewer { id }\n}"},
		{name: "several operations stay unnamed", query: "query A {\n  a\n}\nquery B {\n  b\n}"},
		{name: "a fragment is not an operation", query: "fragment Fields on User {\n  name\n}\nquery One {\n  a\n}", want: "One"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := graphQLOperationName(testCase.query); got != testCase.want {
				t.Errorf("unexpected operation: %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestParseFile_GraphQLRequest(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{
			name: "marked with the request type header",
			content: `# @name Fetch user
POST https://example.com/graphql
X-REQUEST-TYPE: GraphQL

query GetUser($id: ID!) {
  user(id: $id) { name }
}

{
  "id": "{{userId}}"
}
`,
		},
		{
			name: "recognized from the body",
			content: `# @name Fetch user
POST https://example.com/graphql
Accept: application/json

query GetUser($id: ID!) {
  user(id: $id) { name }
}

{
  "id": "{{userId}}"
}
`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			filePath := filepath.Join(t.TempDir(), "graphql.http")
			if err := os.WriteFile(filePath, []byte(testCase.content), 0644); err != nil {
				t.Fatal(err)
			}

			parser, err := NewParser()
			if err != nil {
				t.Fatal(err)
			}
			defer parser.Close()

			result, err := parser.ParseFileWithOptions(context.Background(), filePath, ParseOptions{
				Variables: map[string]string{"userId": "42"},
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Suites) != 1 {
				t.Fatalf("expected one request, got %d", len(result.Suites))
			}

			suite := result.Suites[0]
			if suite.BodyType != BodyTypeGraphQL {
				t.Fatalf("request was not recognized as GraphQL: %q", suite.BodyType)
			}
			if suite.Body != "query GetUser($id: ID!) {\n  user(id: $id) { name }\n}" {
				t.Errorf("unexpected query: %q", suite.Body)
			}
			if suite.GraphQLVariables != "{\n  \"id\": \"42\"\n}" {
				t.Errorf("variables were not split or substituted: %q", suite.GraphQLVariables)
			}
			if suite.GraphQLOperation != "GetUser" {
				t.Errorf("unexpected operation: %q", suite.GraphQLOperation)
			}
		})
	}
}
