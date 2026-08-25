package runner

import (
	"encoding/json"
	"fmt"
	"strings"

	parser "github.com/Hecatoncheir/lazyrest/parser/http"
)

// encodeGraphQL builds the JSON document a GraphQL server expects, with the
// query held in the body of the suite.
func encodeGraphQL(suite parser.HttpSuite) (string, error) {
	payload := map[string]any{"query": suite.Body}
	if suite.GraphQLVariables != "" {
		var variables map[string]any
		if err := json.Unmarshal([]byte(suite.GraphQLVariables), &variables); err != nil {
			return "", fmt.Errorf("parse GraphQL variables: %w", err)
		}
		payload["variables"] = variables
	}
	if suite.GraphQLOperation != "" {
		payload["operationName"] = suite.GraphQLOperation
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode GraphQL request: %w", err)
	}
	return string(encoded), nil
}

// graphQLErrorMessages reports the errors of a GraphQL response. GraphQL
// answers with 200 even when the operation failed, so the body is the only
// place that failure appears.
func graphQLErrorMessages(body []byte) []string {
	var payload struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}

	messages := make([]string, 0, len(payload.Errors))
	for _, item := range payload.Errors {
		message := strings.TrimSpace(item.Message)
		if message == "" {
			message = "unspecified GraphQL error"
		}
		messages = append(messages, message)
	}
	if len(messages) == 0 {
		return nil
	}
	return messages
}
