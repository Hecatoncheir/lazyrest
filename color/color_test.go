package color

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestColorToRGB(t *testing.T) {
	cases := []struct {
		name           string
		color          Color
		red, green, bl int32
	}{
		{name: "black", color: "#000000"},
		{name: "white", color: "#ffffff", red: 255, green: 255, bl: 255},
		{name: "the gruvbox accent", color: "#83a598", red: 131, green: 165, bl: 152},
		{name: "the short form", color: "#f00", red: 255},
		// A value that is not a colour reads as black rather than failing, so
		// that a mistyped theme still renders.
		{name: "not a colour", color: "nonsense"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			red, green, blue := testCase.color.ToRGB()
			if red != testCase.red || green != testCase.green || blue != testCase.bl {
				t.Fatalf("got (%d, %d, %d), want (%d, %d, %d)", red, green, blue, testCase.red, testCase.green, testCase.bl)
			}
		})
	}
}

func TestColorToTerminal(t *testing.T) {
	if got, want := Color("#83a598").ToTerminal(), tcell.NewRGBColor(131, 165, 152); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}
