package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/locale"
	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/gdamore/tcell/v2"
)

func (application *Application) buildCapturedResponsesOverlay() {
	application.Captured = application.newOverlayView("")
	application.Captured.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if application.config.Keybindings.Matches(keymap.ClearCaptured, event) {
			application.clearCapturedResponses()
			return nil
		}
		return event
	})
	application.refreshCapturedResponses()
}

func (application *Application) refreshCapturedResponses() {
	if application.Captured == nil || application.Producer == nil {
		return
	}
	captures := application.Producer.CapturedResponses()
	translator := application.config.Locale
	application.Captured.SetText(renderCapturedResponses(captures, application.Model.Snapshot().RootDirectoryPath, translator))
	application.Captured.ScrollToBeginning()
	application.Captured.SetTitle(fmt.Sprintf(
		"%s (%d) — %s %s · q/Esc %s",
		translator.Text("captured_responses"),
		len(captures),
		application.config.Keybindings.Describe(keymap.ClearCaptured),
		translator.Text("clear_captured_responses"),
		translator.Text("close"),
	))
}

func (application *Application) clearCapturedResponses() {
	if application.Producer == nil {
		return
	}
	count := application.Producer.ClearCapturedResponses()
	application.refreshCapturedResponses()
	if application.Footer != nil {
		application.Footer.UpdateStatus(application.config.Locale.Format("captured_responses_cleared", count))
	}
}

func renderCapturedResponses(captures []parserhttp.CapturedResponse, rootDirectory string, translator *locale.Translator) string {
	if len(captures) == 0 {
		return translator.Text("no_captured_responses")
	}
	var output strings.Builder
	currentFile := ""
	for _, capture := range captures {
		file := displayCapturedSource(rootDirectory, capture.SourceFilePath, translator)
		if file != currentFile {
			if output.Len() > 0 {
				output.WriteString("\n")
			}
			output.WriteString(file)
			output.WriteString("\n")
			currentFile = file
		}
		output.WriteString("  ")
		output.WriteString(singleLine(capture.Name))
		output.WriteString(" — ")
		if capture.Status != "" {
			output.WriteString(singleLine(capture.Status))
			output.WriteString(" · ")
		}
		output.WriteString(translator.Format("captured_response_details", capture.HeaderCount, capture.BodyBytes))
		output.WriteString("\n")
	}
	return strings.TrimSuffix(output.String(), "\n")
}

func displayCapturedSource(rootDirectory, sourceFilePath string, translator *locale.Translator) string {
	if sourceFilePath == "" {
		return translator.Text("unknown_file")
	}
	root, rootErr := filepath.Abs(rootDirectory)
	source, sourceErr := filepath.Abs(sourceFilePath)
	if rootErr == nil && sourceErr == nil {
		if relative, err := filepath.Rel(root, source); err == nil && relative != "." && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return singleLine(relative)
		}
	}
	return singleLine(sourceFilePath)
}

func singleLine(value string) string {
	return strings.Map(func(character rune) rune {
		if unicode.IsControl(character) {
			return ' '
		}
		return character
	}, strings.TrimSpace(value))
}
