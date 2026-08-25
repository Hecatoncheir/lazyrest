package http

import (
	"context"
	"maps"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// requestBoundaries are the node types that start something new and therefore
// end the text region belonging to the preceding request.
var requestBoundaries = map[string]struct{}{
	"request":              {},
	"comment":              {},
	"variable_declaration": {},
}

// maxRecoveryDepth bounds how many times a request region is re-parsed after
// the grammar folds the rest of the document into an ERROR node.
const maxRecoveryDepth = 8

func getSuites(ctx context.Context, source []byte, tree *sitter.Tree, options ParseOptions) ([]HttpSuite, []Diagnostic) {
	return collectSuites(ctx, source, tree, options, 0)
}

func collectSuites(ctx context.Context, source []byte, tree *sitter.Tree, options ParseOptions, depth int) ([]HttpSuite, []Diagnostic) {
	suites := []HttpSuite{}
	diagnostics := []Diagnostic{}
	pendingName := ""
	variables := maps.Clone(options.Variables)
	if variables == nil {
		variables = make(map[string]string)
	}

	rootNode := tree.RootNode()
	consumedUntil := uint32(0)

	for i := 0; i < int(rootNode.ChildCount()); i++ {
		node := rootNode.Child(i)
		if node.StartByte() < consumedUntil {
			// Already read as part of the preceding request region.
			continue
		}
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
			regionEnd := requestRegionEnd(rootNode, i, uint32(len(source)))
			consumedUntil = regionEnd
			regionText, remainder := clipRequestRegion(string(source[node.StartByte():regionEnd]))
			suite := getSuite(source, node, regionText)
			if pendingName != "" {
				suite.Name = pendingName
				pendingName = ""
			}
			resolution := resolveSuiteVariables(&suite, variables)
			suite.SecretValues = resolveSecretVariables(options.SecretVariables, variables)
			for _, name := range resolution.Missing {
				diagnostics = append(diagnostics, newDiagnostic(node, "undefined variable: "+name))
			}
			for _, cycle := range resolution.Cycles {
				diagnostics = append(diagnostics, newDiagnostic(node, "cyclic variable reference: "+cycle))
			}
			if suite.Name == "" {
				suite.Name = strings.TrimSpace(suite.Method + " " + suite.Uri)
			}
			suites = append(suites, suite)
			if remainder != "" && !holdsOnlyComments(remainder) {
				lineOffset := int(node.StartPoint().Row) + strings.Count(regionText, "\n")
				recovered, recoveredDiagnostics := recoverRemainder(ctx, remainder, options, variables, depth, lineOffset)
				suites = append(suites, recovered...)
				diagnostics = append(diagnostics, recoveredDiagnostics...)
			}
		case "ERROR":
			diagnostics = append(diagnostics, newDiagnostic(node, "unrecognized content"))
		}
	}

	return suites, diagnostics
}

// requestRegionEnd returns the byte offset at which the request at index ends.
// Everything up to the next boundary node belongs to the request, including the
// ERROR nodes the grammar produces for bodies and for header values it cannot
// read.
func requestRegionEnd(rootNode *sitter.Node, index int, sourceLength uint32) uint32 {
	for next := index + 1; next < int(rootNode.ChildCount()); next++ {
		sibling := rootNode.Child(next)
		if _, boundary := requestBoundaries[sibling.Type()]; boundary {
			return sibling.StartByte()
		}
	}
	return sourceLength
}

// recoverRemainder re-parses the text the grammar folded into the ERROR node of
// a request it could not read. Without this the requests inside that text would
// be missing from the document.
func recoverRemainder(ctx context.Context, text string, options ParseOptions, variables map[string]string, depth, lineOffset int) ([]HttpSuite, []Diagnostic) {
	unrecognized := []Diagnostic{{Line: lineOffset + 1, Column: 1, Message: "unrecognized content after the request body"}}
	if depth >= maxRecoveryDepth {
		return nil, unrecognized
	}

	nestedParser := getParser()
	defer nestedParser.Close()
	source := []byte(text)
	tree, err := getTree(ctx, source, nestedParser)
	if err != nil {
		return nil, unrecognized
	}
	defer tree.Close()

	nestedOptions := ParseOptions{Variables: variables, SecretVariables: options.SecretVariables}
	suites, diagnostics := collectSuites(ctx, source, tree, nestedOptions, depth+1)
	if len(suites) == 0 {
		return nil, unrecognized
	}
	for index := range diagnostics {
		diagnostics[index].Line += lineOffset
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
