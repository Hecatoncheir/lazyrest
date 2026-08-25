package http

import (
	nethttp "net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var variablePattern = regexp.MustCompile(`\{\{\s*([A-Za-z_][A-Za-z0-9_.-]*)\s*\}\}`)

type VariableResolution struct {
	Missing []string
	Cycles  []string
}

type variableResolver struct {
	variables map[string]string
	cache     map[string]string
	missing   map[string]struct{}
	cycles    map[string]struct{}
}

func parseVariableDeclaration(content string) (string, string, bool) {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "@")
	name, value, found := strings.Cut(content, "=")
	if !found {
		return "", "", false
	}
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)
	if name == "" {
		return "", "", false
	}
	if unquoted, err := strconv.Unquote(value); err == nil {
		value = unquoted
	}
	return name, value, true
}

func newVariableResolver(variables map[string]string) *variableResolver {
	return &variableResolver{
		variables: variables,
		cache:     make(map[string]string),
		missing:   make(map[string]struct{}),
		cycles:    make(map[string]struct{}),
	}
}

func (resolver *variableResolver) resolveText(value string, stack []string) string {
	return variablePattern.ReplaceAllStringFunc(value, func(match string) string {
		parts := variablePattern.FindStringSubmatch(match)
		return resolver.resolveName(parts[1], stack)
	})
}

func (resolver *variableResolver) resolveName(name string, stack []string) string {
	if value, ok := resolver.cache[name]; ok {
		return value
	}
	for index, current := range stack {
		if current == name {
			cycle := append(slices.Clone(stack[index:]), name)
			resolver.cycles[strings.Join(cycle, " -> ")] = struct{}{}
			return "{{" + name + "}}"
		}
	}
	raw, ok := resolver.variables[name]
	if !ok {
		resolver.missing[name] = struct{}{}
		return "{{" + name + "}}"
	}
	value := resolver.resolveText(raw, append(stack, name))
	resolver.cache[name] = value
	return value
}

func resolveSuiteVariables(suite *HttpSuite, variables map[string]string) VariableResolution {
	resolver := newVariableResolver(variables)
	suite.Uri = resolver.resolveText(suite.Uri, nil)
	suite.Body = resolver.resolveText(suite.Body, nil)
	resolvedHeaders := make(nethttp.Header, len(suite.Header))
	for key, values := range suite.Header {
		resolvedKey := resolver.resolveText(key, nil)
		for _, value := range values {
			resolvedHeaders.Add(resolvedKey, resolver.resolveText(value, nil))
		}
	}
	suite.Header = resolvedHeaders

	result := VariableResolution{
		Missing: make([]string, 0, len(resolver.missing)),
		Cycles:  make([]string, 0, len(resolver.cycles)),
	}
	for name := range resolver.missing {
		result.Missing = append(result.Missing, name)
	}
	for cycle := range resolver.cycles {
		result.Cycles = append(result.Cycles, cycle)
	}
	slices.Sort(result.Missing)
	slices.Sort(result.Cycles)
	return result
}

func resolveSecretVariables(names []string, variables map[string]string) []string {
	resolver := newVariableResolver(variables)
	values := make(map[string]struct{})
	for _, name := range names {
		value := resolver.resolveName(name, nil)
		if value != "" && !variablePattern.MatchString(value) {
			values[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	slices.Sort(result)
	return result
}

// ResolveVariables returns the variables with their nested {{references}}
// resolved, which is the form other tools need them in.
func ResolveVariables(variables map[string]string) map[string]string {
	resolver := newVariableResolver(variables)
	resolved := make(map[string]string, len(variables))
	for name := range variables {
		resolved[name] = resolver.resolveName(name, nil)
	}
	return resolved
}

// ResolveSecretValues returns the resolved values of the variables named as
// secret. Those are the values that have to be redacted from output.
func ResolveSecretValues(options ParseOptions) []string {
	return resolveSecretVariables(options.SecretVariables, options.Variables)
}
