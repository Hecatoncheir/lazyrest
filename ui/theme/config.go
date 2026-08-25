package theme

import (
	"fmt"

	appcolor "github.com/Hecatoncheir/lazyrest/color"
	"github.com/gdamore/tcell/v2"
	"github.com/lucasb-eyer/go-colorful"
)

type Config struct {
	Preset               string `yaml:"preset"`
	Background           string `yaml:"background"`
	PanelBackground      string `yaml:"panel_background"`
	PanelFocus           string `yaml:"panel_focus"`
	Foreground           string `yaml:"foreground"`
	Muted                string `yaml:"muted"`
	Accent               string `yaml:"accent"`
	Border               string `yaml:"border"`
	BorderFocus          string `yaml:"border_focus"`
	SelectionBackground  string `yaml:"selection_background"`
	SelectionForeground  string `yaml:"selection_foreground"`
	Progress             string `yaml:"progress"`
	ProgressForeground   string `yaml:"progress_foreground"`
	Success              string `yaml:"success"`
	SuccessForeground    string `yaml:"success_foreground"`
	Failure              string `yaml:"failure"`
	FailureForeground    string `yaml:"failure_foreground"`
	Breadcrumb           string `yaml:"breadcrumb"`
	BreadcrumbForeground string `yaml:"breadcrumb_foreground"`
}

func DefaultConfig() Config {
	return Config{Preset: "gruvbox"}
}

func FromConfig(config Config) (Theme, error) {
	config, err := resolvePreset(config)
	if err != nil {
		return Theme{}, err
	}
	result := NewDefault()
	set := func(value string, targets ...*tcell.Color) error {
		if value == "" {
			return nil
		}
		if _, err := colorful.Hex(value); err != nil {
			return fmt.Errorf("invalid theme color %q: %w", value, err)
		}
		terminal := appcolor.Color(value).ToTerminal()
		for _, target := range targets {
			*target = terminal
		}
		return nil
	}
	settings := []struct {
		value   string
		targets []*tcell.Color
	}{
		{config.Background, []*tcell.Color{&result.Background}},
		{config.PanelBackground, []*tcell.Color{&result.Tree.Background, &result.Suites.Background, &result.Suites.SuiteBackground, &result.Suite.Background, &result.Producer.Background, &result.Footer.Background}},
		{config.PanelFocus, []*tcell.Color{&result.Tree.BackgroundFocus, &result.Suites.BackgroundFocus, &result.Suite.BackgroundFocus, &result.Producer.BackgroundFocus}},
		{config.Foreground, []*tcell.Color{&result.Tree.Node.Foreground, &result.Suite.Foreground, &result.Producer.Foreground}},
		{config.Muted, []*tcell.Color{&result.Tree.Title, &result.Suites.Title, &result.Suites.SuiteForeground, &result.Suite.Title, &result.Producer.Title, &result.Footer.Foreground, &result.Producer.Syntax.Punctuation, &result.Producer.Syntax.Comment}},
		{config.Accent, []*tcell.Color{&result.Tree.TitleFocus, &result.Tree.NodeDirectory.Foreground, &result.Suites.TitleFocus, &result.Suite.TitleFocus, &result.Producer.TitleFocus, &result.Producer.Syntax.Key}},
		{config.Border, []*tcell.Color{&result.Border, &result.Tree.Border, &result.Suites.Border, &result.Suite.Border, &result.Producer.Border}},
		{config.BorderFocus, []*tcell.Color{&result.Tree.BorderFocus, &result.Suites.BorderFocus, &result.Suite.BorderFocus, &result.Producer.BorderFocus}},
		{config.SelectionBackground, []*tcell.Color{&result.Suites.SuiteFocusBackground, &result.Footer.SelectedFileName.Background}},
		{config.SelectionForeground, []*tcell.Color{&result.Suites.SuiteFocusForeground, &result.Footer.SelectedFileName.Foreground}},
		{config.Progress, []*tcell.Color{&result.Footer.SuiteBackground, &result.Producer.Syntax.Number, &result.Producer.Syntax.Variable}},
		{config.ProgressForeground, []*tcell.Color{&result.Footer.SuiteForeground}},
		{config.Success, []*tcell.Color{&result.Footer.SuiteSuccess.Background, &result.Producer.Syntax.String}},
		{config.SuccessForeground, []*tcell.Color{&result.Footer.SuiteSuccess.Foreground}},
		{config.Failure, []*tcell.Color{&result.Footer.SuiteFailure.Background, &result.Producer.Syntax.Literal, &result.Producer.Syntax.Keyword}},
		{config.FailureForeground, []*tcell.Color{&result.Footer.SuiteFailure.Foreground}},
		{config.Breadcrumb, []*tcell.Color{&result.Footer.RootDirectoryPath.Background}},
		{config.BreadcrumbForeground, []*tcell.Color{&result.Footer.RootDirectoryPath.Foreground}},
	}
	for _, setting := range settings {
		if err := set(setting.value, setting.targets...); err != nil {
			return Theme{}, err
		}
	}
	result.Footer.RootDirectoryPath.ArrowBackground = result.Footer.Background
	result.Footer.RootDirectoryPath.ArrowForeground = result.Footer.RootDirectoryPath.Background
	result.Footer.SelectedFileName.ArrowBackground = result.Footer.Background
	result.Footer.SelectedFileName.ArrowForeground = result.Footer.SelectedFileName.Background
	return result, nil
}
