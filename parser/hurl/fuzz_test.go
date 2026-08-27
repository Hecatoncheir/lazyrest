package hurl

import (
	"reflect"
	"testing"
)

func FuzzSplitEntries(f *testing.F) {
	for _, source := range []string{
		"GET https://example.test\n",
		"# @name login\nPOST https://example.test/login\n{\"name\":\"demo\"}\nHTTP 200\n[Asserts]\nstatus == 200\n",
		"GET https://example.test/one\nHTTP 200\n\nPATCH https://example.test/two\nHTTP 204\n",
		"  GET https://example.test/body-line\nDELETE /items/1\n",
		"not a hurl request\r\n",
	} {
		f.Add(source)
	}

	f.Fuzz(func(t *testing.T, source string) {
		first := splitEntries(source)
		second := splitEntries(source)
		if !reflect.DeepEqual(first, second) {
			t.Fatal("Hurl entry splitting is not deterministic")
		}
		for index, current := range first {
			if current.Number != index+1 {
				t.Fatalf("unexpected Hurl entry number: got %d, want %d", current.Number, index+1)
			}
			if current.Method == "" || current.Uri == "" || current.Text == "" {
				t.Fatalf("incomplete Hurl entry: %+v", current)
			}
		}
	})
}
