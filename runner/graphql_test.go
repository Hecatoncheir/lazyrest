package runner

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	parser "github.com/Hecatoncheir/lazyrest/parser/http"
)

func graphQLRunner(t *testing.T, suite parser.HttpSuite, body string, sent *http.Request) Runner {
	t.Helper()
	return NewFromSuiteWithConfig(suite, Config{
		Client: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			*sent = *request
			payload, err := io.ReadAll(request.Body)
			if err != nil {
				return nil, err
			}
			sent.Body = io.NopCloser(strings.NewReader(string(payload)))
			return &http.Response{
				StatusCode:    http.StatusOK,
				Body:          io.NopCloser(strings.NewReader(body)),
				ContentLength: int64(len(body)),
				Header:        http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		})},
	})
}

func TestExecute_SendsGraphQLAsJSON(t *testing.T) {
	suite := parser.HttpSuite{
		Method:           http.MethodPost,
		Uri:              "https://example.test/graphql",
		Body:             "query GetUser($id: ID!) {\n  user(id: $id) { name }\n}",
		BodyType:         parser.BodyTypeGraphQL,
		GraphQLVariables: `{"id": "42"}`,
		GraphQLOperation: "GetUser",
		Header:           http.Header{},
	}
	var sent http.Request
	runner := graphQLRunner(t, suite, `{"data":{"user":{"name":"Ada"}}}`, &sent)

	response, err := runner.Execute(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if contentType := sent.Header.Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Fatalf("unexpected content type: %q", contentType)
	}

	payload, err := io.ReadAll(sent.Body)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Query         string         `json:"query"`
		Variables     map[string]any `json:"variables"`
		OperationName string         `json:"operationName"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("the request body is not JSON: %v (%s)", err, payload)
	}
	if decoded.Query != suite.Body {
		t.Errorf("unexpected query: %q", decoded.Query)
	}
	if decoded.Variables["id"] != "42" {
		t.Errorf("unexpected variables: %#v", decoded.Variables)
	}
	if decoded.OperationName != "GetUser" {
		t.Errorf("unexpected operation: %q", decoded.OperationName)
	}
	if !response.IsSuccessful() {
		t.Error("a response without errors was reported as failed")
	}
}

func TestExecute_KeepsRawGraphQLWhenDeclared(t *testing.T) {
	query := "query {\n  viewer { id }\n}"
	suite := parser.HttpSuite{
		Method:   http.MethodPost,
		Uri:      "https://example.test/graphql",
		Body:     query,
		BodyType: parser.BodyTypeGraphQL,
		Header:   http.Header{"Content-Type": []string{"application/graphql"}},
	}
	var sent http.Request
	runner := graphQLRunner(t, suite, "{}", &sent)

	if _, err := runner.Execute(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	payload, err := io.ReadAll(sent.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(payload) != query {
		t.Fatalf("the declared raw body was re-encoded: %q", payload)
	}
	if contentType := sent.Header.Get("Content-Type"); contentType != "application/graphql" {
		t.Fatalf("unexpected content type: %q", contentType)
	}
}

func TestExecute_ReportsGraphQLErrorsOnSuccessfulStatus(t *testing.T) {
	suite := parser.HttpSuite{
		Method:   http.MethodPost,
		Uri:      "https://example.test/graphql",
		Body:     "query {\n  viewer { id }\n}",
		BodyType: parser.BodyTypeGraphQL,
		Header:   http.Header{},
	}
	var sent http.Request
	body := `{"data":null,"errors":[{"message":"Field 'viewer' is not defined"},{"message":""}]}`
	runner := graphQLRunner(t, suite, body, &sent)

	response, err := runner.Execute(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", response.StatusCode)
	}
	if len(response.GraphQLErrors) != 2 {
		t.Fatalf("unexpected errors: %#v", response.GraphQLErrors)
	}
	if response.GraphQLErrors[0] != "Field 'viewer' is not defined" {
		t.Errorf("unexpected first error: %q", response.GraphQLErrors[0])
	}
	if response.GraphQLErrors[1] != "unspecified GraphQL error" {
		t.Errorf("an error without a message was dropped: %q", response.GraphQLErrors[1])
	}
	if response.IsSuccessful() {
		t.Error("a GraphQL error response was reported as successful")
	}
}

func TestExecute_LeavesNonGraphQLErrorFieldsAlone(t *testing.T) {
	suite := parser.HttpSuite{
		Method:   http.MethodGet,
		Uri:      "https://example.test/users",
		BodyType: "json",
		Header:   http.Header{},
	}
	var sent http.Request
	runner := graphQLRunner(t, suite, `{"errors":[{"message":"validation failed"}]}`, &sent)

	response, err := runner.Execute(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(response.GraphQLErrors) != 0 {
		t.Fatalf("a REST response was read as GraphQL: %#v", response.GraphQLErrors)
	}
	if !response.IsSuccessful() {
		t.Error("a successful REST response was reported as failed")
	}
}

func TestExecute_ReportsUnreadableGraphQLVariables(t *testing.T) {
	suite := parser.HttpSuite{
		Method:           http.MethodPost,
		Uri:              "https://example.test/graphql",
		Body:             "query {\n  viewer { id }\n}",
		BodyType:         parser.BodyTypeGraphQL,
		GraphQLVariables: "{not json}",
		Header:           http.Header{},
	}
	runner := NewFromSuiteWithConfig(suite, Config{})

	if _, err := runner.Execute(context.Background(), nil); err == nil || !strings.Contains(err.Error(), "GraphQL variables") {
		t.Fatalf("unreadable variables were not reported: %v", err)
	}
}
