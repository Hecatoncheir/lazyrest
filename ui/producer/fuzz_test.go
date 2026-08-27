package producer

import (
	"strings"
	"testing"
)

func FuzzShellQuote(f *testing.F) {
	for _, value := range []string{
		"",
		"plain",
		"it's quoted",
		"$HOME $(whoami) `date`",
		"line one\nline two",
		"пароль и пробелы",
	} {
		f.Add(value)
	}

	f.Fuzz(func(t *testing.T, value string) {
		quoted := shellQuote(value)
		if len(quoted) < 2 || quoted[0] != '\'' || quoted[len(quoted)-1] != '\'' {
			t.Fatalf("shell value is not enclosed in single quotes: %q", quoted)
		}
		decoded := strings.ReplaceAll(quoted[1:len(quoted)-1], `'"'"'`, `'`)
		if decoded != value {
			t.Fatalf("shell quoting did not round-trip: got %q, want %q", decoded, value)
		}
	})
}
