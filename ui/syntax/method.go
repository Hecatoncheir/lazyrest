package syntax

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

// MethodPalette colours an HTTP method by what the method does, so that a list
// of requests can be read at a glance.
type MethodPalette struct {
	Read   tcell.Color
	Create tcell.Color
	Update tcell.Color
	Delete tcell.Color
	Other  tcell.Color
}

// Method returns the markup of an HTTP method. The text is escaped, so the
// caller must not escape it again.
func Method(method string, palette MethodPalette) string {
	if method == "" {
		return ""
	}
	tag := colorTag(methodColor(method, palette))
	if tag == "" {
		return escape(method)
	}
	return tag + escape(method) + "[-]"
}

func methodColor(method string, palette MethodPalette) tcell.Color {
	switch strings.ToUpper(method) {
	case "GET", "HEAD", "OPTIONS", "TRACE":
		return palette.Read
	case "POST":
		return palette.Create
	case "PUT", "PATCH":
		return palette.Update
	case "DELETE":
		return palette.Delete
	}
	return palette.Other
}
