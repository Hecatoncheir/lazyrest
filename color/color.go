package color

import (
	"github.com/gdamore/tcell/v2"
	"github.com/lucasb-eyer/go-colorful"
)

type Color string

func (color Color) ToRGB() (r, g, b int32) {
	colorFromString, _ := colorful.Hex(string(color))
	red, green, blue, _ := colorFromString.RGBA()
	return int32(red), int32(green), int32(blue)
}

func (color Color) ToTerminal() tcell.Color {
	// tcell.NewHexColor(0xdda15e)
	terminalColor := tcell.NewRGBColor(color.ToRGB())
	return terminalColor
}
