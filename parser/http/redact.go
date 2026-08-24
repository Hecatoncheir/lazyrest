package http

import (
	"slices"
	"strings"
)

func RedactSecrets(value string, secrets []string) string {
	ordered := slices.Clone(secrets)
	slices.SortFunc(ordered, func(left, right string) int { return len(right) - len(left) })
	for _, secret := range ordered {
		if secret != "" {
			value = strings.ReplaceAll(value, secret, "<redacted>")
		}
	}
	return value
}

func (suite HttpSuite) Redact(value string) string {
	return RedactSecrets(value, suite.SecretValues)
}
