package hurl

import (
	"bufio"
	"lazyrest/parser/http"
	"os"
	"strings"
)

type Parser struct{}

type parserState int

const (
	stateNone parserState = iota
	stateHeaders
	stateBody
)

func NewParser() (Parser, error) {
	return Parser{}, nil
}

func (p Parser) GetSuitesFromFile(filePath string) ([]http.HttpSuite, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var suites []http.HttpSuite
	var currentSuite *http.HttpSuite
	state := stateNone

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if isMethod(trimmedLine) {
			if currentSuite != nil {
				suites = append(suites, *currentSuite)
			}
			parts := strings.SplitN(trimmedLine, " ", 2)
			method := parts[0]
			uri := ""
			if len(parts) > 1 {
				uri = parts[1]
			}
			newSuite := http.NewHttpSuite()
			newSuite.Method = method
			newSuite.Uri = uri
			currentSuite = &newSuite
			state = stateHeaders
			continue
		}

		if currentSuite == nil {
			continue
		}

		if state == stateHeaders {
			if trimmedLine == "" {
				state = stateBody
				continue
			}
			// Check for header "Key: Value"
			if strings.Contains(line, ":") {
				parts := strings.SplitN(line, ":", 2)
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				currentSuite.Header[key] = val
			} else {
				// If no colon, it's probably an error or start of body, but let's be safe and switch to body
				state = stateBody
				if trimmedLine != "" {
					currentSuite.Body = line
				}
			}
		} else {
			// stateBody
			if currentSuite.Body != "" {
				currentSuite.Body += "\n" + line
			} else {
				currentSuite.Body = line
			}
		}
	}

	if currentSuite != nil {
		suites = append(suites, *currentSuite)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return suites, nil
}

func isMethod(line string) bool {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
	for _, m := range methods {
		if strings.HasPrefix(line, m+" ") || line == m {
			return true
		}
	}
	return false
}
