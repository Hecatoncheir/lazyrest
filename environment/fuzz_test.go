package environment

import "testing"

func FuzzParseDotEnvValue(f *testing.F) {
	for _, value := range []string{
		"",
		"plain value",
		"plain # comment",
		"'single quoted # value' # comment",
		`"double quoted\nvalue"`,
		`"escaped \\"quote\\" and \\\\ slash"`,
		"'unterminated",
		`"value" unexpected`,
	} {
		f.Add(value)
	}

	f.Fuzz(func(t *testing.T, value string) {
		first, firstErr := parseDotEnvValue(value)
		second, secondErr := parseDotEnvValue(value)
		if first != second {
			t.Fatalf("dotenv parsing is not deterministic: %q != %q", first, second)
		}
		if (firstErr == nil) != (secondErr == nil) {
			t.Fatalf("dotenv parsing changed error state: %v != %v", firstErr, secondErr)
		}
		if firstErr != nil && firstErr.Error() != secondErr.Error() {
			t.Fatalf("dotenv parsing changed its error: %v != %v", firstErr, secondErr)
		}
	})
}
