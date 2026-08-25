package theme

import "testing"

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
