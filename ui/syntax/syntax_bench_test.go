package syntax

import (
	"fmt"
	"strings"
	"testing"
)

func benchmarkBody(size int) string {
	var builder strings.Builder
	builder.WriteString("{\n  \"items\": [\n")
	for index := 0; builder.Len() < size; index++ {
		fmt.Fprintf(&builder, "    {\"id\": %d, \"name\": \"item %d\", \"active\": true, \"score\": 12.5},\n", index, index)
	}
	builder.WriteString("    null\n  ]\n}")
	return builder.String()
}

func BenchmarkHighlightJSON(b *testing.B) {
	for _, size := range []int{4 << 10, 64 << 10, 250 << 10} {
		body := benchmarkBody(size)
		palette := testPalette()
		b.Run(fmt.Sprintf("%dKiB", len(body)/1024), func(b *testing.B) {
			b.SetBytes(int64(len(body)))
			for b.Loop() {
				Highlight(body, LanguageJSON, palette)
			}
		})
	}
}
