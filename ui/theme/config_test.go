package theme

import (
	"math"
	"reflect"
	"testing"

	appcolor "github.com/Hecatoncheir/lazyrest/color"
	"github.com/gdamore/tcell/v2"
	"github.com/lucasb-eyer/go-colorful"
)

func TestFromConfigOverridesSemanticColors(t *testing.T) {
	configured, err := FromConfig(Config{PanelBackground: "#010203", Success: "#040506"})
	if err != nil {
		t.Fatal(err)
	}
	defaults := NewDefault()
	if configured.Tree.Background == defaults.Tree.Background || configured.Producer.Background != configured.Tree.Background {
		t.Fatal("panel background was not applied consistently")
	}
	if configured.Footer.SuiteSuccess.Background == defaults.Footer.SuiteSuccess.Background {
		t.Fatal("success color was not applied")
	}
}

func TestFromConfigRejectsInvalidColor(t *testing.T) {
	if _, err := FromConfig(Config{Accent: "not-a-color"}); err == nil {
		t.Fatal("expected invalid color error")
	}
}

func TestBuiltInPresets(t *testing.T) {
	for _, name := range []string{"gruvbox", "catppuccin-mocha", "tokyo-night", "dracula", "nord", "monokai"} {
		if _, err := FromConfig(Config{Preset: name}); err != nil {
			t.Fatalf("preset %s failed: %v", name, err)
		}
	}
}

func TestDefaultThemeMatchesGruvboxPreset(t *testing.T) {
	configured, err := FromConfig(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	if defaults := NewDefault(); !reflect.DeepEqual(configured, defaults) {
		t.Fatal("default theme has drifted from the gruvbox preset")
	}
}

func TestBuiltInPresetsMeetContrastTargets(t *testing.T) {
	for name, preset := range presets {
		textPairs := []struct {
			role       string
			foreground string
			background string
		}{
			{role: "foreground", foreground: preset.Foreground, background: preset.PanelBackground},
			{role: "muted", foreground: preset.Muted, background: preset.PanelBackground},
			{role: "accent", foreground: preset.Accent, background: preset.PanelFocus},
			{role: "selection", foreground: preset.SelectionForeground, background: preset.SelectionBackground},
			{role: "progress", foreground: preset.ProgressForeground, background: preset.Progress},
			{role: "success", foreground: preset.SuccessForeground, background: preset.Success},
			{role: "failure", foreground: preset.FailureForeground, background: preset.Failure},
			{role: "breadcrumb", foreground: preset.BreadcrumbForeground, background: preset.Breadcrumb},
		}
		for _, pair := range textPairs {
			if ratio := contrastRatio(t, pair.foreground, pair.background); ratio < 4.5 {
				t.Errorf("preset %s %s contrast %.2f, want at least 4.5", name, pair.role, ratio)
			}
		}
		if ratio := contrastRatio(t, preset.BorderFocus, preset.PanelFocus); ratio < 3 {
			t.Errorf("preset %s focus border contrast %.2f, want at least 3", name, ratio)
		}
	}
}

func contrastRatio(t *testing.T, foreground, background string) float64 {
	t.Helper()
	foregroundColor, err := colorful.Hex(foreground)
	if err != nil {
		t.Fatal(err)
	}
	backgroundColor, err := colorful.Hex(background)
	if err != nil {
		t.Fatal(err)
	}
	foregroundLuminance := relativeLuminance(foregroundColor)
	backgroundLuminance := relativeLuminance(backgroundColor)
	if foregroundLuminance < backgroundLuminance {
		foregroundLuminance, backgroundLuminance = backgroundLuminance, foregroundLuminance
	}
	return (foregroundLuminance + 0.05) / (backgroundLuminance + 0.05)
}

func relativeLuminance(value colorful.Color) float64 {
	linear := func(component float64) float64 {
		if component <= 0.04045 {
			return component / 12.92
		}
		return math.Pow((component+0.055)/1.055, 2.4)
	}
	return 0.2126*linear(value.R) + 0.7152*linear(value.G) + 0.0722*linear(value.B)
}

func TestEmbeddedPresetsDefineEveryColor(t *testing.T) {
	for name, preset := range presets {
		value := reflect.ValueOf(preset)
		typeOfPreset := value.Type()
		for index := 0; index < value.NumField(); index++ {
			field := typeOfPreset.Field(index)
			if field.Name == "Preset" {
				continue
			}
			if value.Field(index).String() == "" {
				t.Errorf("preset %s is missing %s", name, field.Tag.Get("yaml"))
			}
		}
	}
}

func TestPresetAllowsColorOverrides(t *testing.T) {
	preset, err := FromConfig(Config{Preset: "nord"})
	if err != nil {
		t.Fatal(err)
	}
	overridden, err := FromConfig(Config{Preset: "nord", Accent: "#010203"})
	if err != nil {
		t.Fatal(err)
	}
	if overridden.Tree.TitleFocus == preset.Tree.TitleFocus {
		t.Fatal("accent override was not applied over preset")
	}
}

func TestFromConfigRejectsUnknownPreset(t *testing.T) {
	if _, err := FromConfig(Config{Preset: "unknown"}); err == nil {
		t.Fatal("expected unknown preset error")
	}
}

func TestSyntaxColorsFollowTheSemanticPalette(t *testing.T) {
	configured, err := FromConfig(Config{Preset: "dracula"})
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		role string
		got  tcell.Color
		want string
	}{
		{role: "string", got: configured.Syntax.String, want: "#50fa7b"},
		{role: "key", got: configured.Syntax.Key, want: "#8be9fd"},
		{role: "number", got: configured.Syntax.Number, want: "#f1fa8c"},
		{role: "literal", got: configured.Syntax.Literal, want: "#ff5555"},
		{role: "punctuation", got: configured.Syntax.Punctuation, want: "#c0c4d6"},
	}
	for _, testCase := range cases {
		if want := appcolor.Color(testCase.want).ToTerminal(); testCase.got != want {
			t.Errorf("%s took %v, want %v", testCase.role, testCase.got, want)
		}
	}
}

func TestEveryPresetDefinesSyntaxColors(t *testing.T) {
	for _, name := range PresetNames() {
		configured, err := FromConfig(Config{Preset: name})
		if err != nil {
			t.Fatal(err)
		}
		colors := configured.Syntax
		for role, color := range map[string]tcell.Color{
			"key": colors.Key, "string": colors.String, "number": colors.Number,
			"literal": colors.Literal, "keyword": colors.Keyword,
			"variable": colors.Variable, "punctuation": colors.Punctuation,
			"comment": colors.Comment,
		} {
			if !color.Valid() {
				t.Errorf("preset %s leaves %s unset", name, role)
			}
		}
	}
}
