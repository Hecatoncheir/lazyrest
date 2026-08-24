package http

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

func getSuites(source []byte, tree *sitter.Tree) ([]HttpSuite, []Diagnostic) {
	suites := []HttpSuite{}
	diagnostics := []Diagnostic{}
	pendingName := ""
	variables := make(map[string]string)

	rootNode := tree.RootNode()

	for i := 0; i < int(rootNode.ChildCount()); i++ {
		node := rootNode.Child(i)
		nodeType := node.Type()

		switch nodeType {
		case "variable_declaration":
			name, value, ok := parseVariableDeclaration(node.Content(source))
			if !ok {
				diagnostics = append(diagnostics, newDiagnostic(node, "invalid variable declaration"))
				continue
			}
			variables[name] = value
		case "comment":
			if name := getNameFromComment(node.Content(source)); name != "" {
				pendingName = name
			}
		case "request":
			suite, err := getSuite(source, node)
			if err != nil {
				diagnostics = append(diagnostics, newDiagnostic(node, err.Error()))
				continue
			}
			if pendingName != "" {
				suite.Name = pendingName
				pendingName = ""
			}
			for _, name := range resolveSuiteVariables(&suite, variables) {
				diagnostics = append(diagnostics, newDiagnostic(node, "undefined variable: "+name))
			}
			if suite.Name == "" {
				suite.Name = strings.TrimSpace(suite.Method + " " + suite.Uri)
			}
			if node.HasError() {
				diagnostics = append(diagnostics, newDiagnostic(node, "request contains syntax that could not be parsed completely"))
			}
			suites = append(suites, suite)
		case "ERROR":
			if len(suites) > 0 && suites[len(suites)-1].Body == "" {
				suites[len(suites)-1].Body = strings.TrimSpace(node.Content(source))
				diagnostics = append(diagnostics, newDiagnostic(node, "unrecognized content was treated as the previous request body"))
			} else {
				diagnostics = append(diagnostics, newDiagnostic(node, "unrecognized content"))
			}
		}
	}

	return suites, diagnostics
}

func newDiagnostic(node *sitter.Node, message string) Diagnostic {
	point := node.StartPoint()
	return Diagnostic{
		Line:    int(point.Row) + 1,
		Column:  int(point.Column) + 1,
		Message: message,
	}
}

func getNameFromComment(comment string) string {
	comment = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(comment), "#"))
	if strings.HasPrefix(comment, "@name ") {
		return strings.TrimSpace(strings.TrimPrefix(comment, "@name "))
	}
	for _, prefix := range []string{"@suite(", "@test("} {
		if !strings.HasPrefix(comment, prefix) || !strings.HasSuffix(comment, ")") {
			continue
		}
		name := strings.TrimSuffix(strings.TrimPrefix(comment, prefix), ")")
		return strings.Trim(strings.TrimSpace(name), "\"'")
	}
	return ""
}
