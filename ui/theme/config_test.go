package theme

import (
	"reflect"
	"testing"
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
