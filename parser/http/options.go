package http

type ParseOptions struct {
	Variables       map[string]string
	SecretVariables []string

	// baseDirectory resolves the relative paths of external bodies. It is set
	// from the file being parsed.
	baseDirectory string
}
