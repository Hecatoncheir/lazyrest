package color

import (
	"github.com/gdamore/tcell/v2"
	"github.com/lucasb-eyer/go-colorful"
)

// Color is a colour written as hexadecimal RGB, such as "#83a598". A value that
// does not parse reads as black, so a mistyped theme still renders.
type Color string

// ToRGB returns the three channels of the colour, each from 0 to 255.
func (color Color) ToRGB() (r, g, b int32) {
	parsed, err := colorful.Hex(string(color))
	if err != nil {
		return 0, 0, 0
	}
	red, green, blue := parsed.RGB255()
	return int32(red), int32(green), int32(blue)
}

func (color Color) ToTerminal() tcell.Color {
	return tcell.NewRGBColor(color.ToRGB())
}
