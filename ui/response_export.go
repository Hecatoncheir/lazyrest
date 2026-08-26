package ui

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Hecatoncheir/lazyrest/ui/footer"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (application *Application) copyResponse(full bool) {
	exported, ok := application.Producer.CurrentResponse()
	if !ok {
		application.showExportError(application.config.Locale.Text("no_response_to_export"))
		return
	}
	payload := exported.Body
	message := "response_body_copied"
	if full {
		payload = exported.Full
		message = "response_copied"
	}
	if application.screen == nil {
		application.showExportError(application.config.Locale.Text("clipboard_unavailable"))
		return
	}
	application.screen.SetClipboard([]byte(payload))
	application.stopFooterProgress()
	application.Footer.UpdateIndicatorState(footer.IndicatorSuccess)
	status := application.config.Locale.Format(message, len([]byte(payload)))
	if exported.Truncated {
		status += " — " + application.config.Locale.Text("body_truncated")
	}
	application.Footer.UpdateStatus(status)
}

func (application *Application) buildSaveResponseInput() {
	translator := application.config.Locale
	input := tview.NewInputField().
		SetLabel(translator.Text("path") + ": ")
	input.SetBorder(true).
		SetTitle(translator.Text("save_response"))
	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			application.saveResponse()
		}
	})
	input.SetChangedFunc(func(_ string) {
		application.saveOverwritePath = ""
		input.SetTitle(application.config.Locale.Text("save_response"))
	})
	application.SaveResponse = input
	application.applySaveResponseTheme()
}

func (application *Application) applySaveResponseTheme() {
	if application.SaveResponse == nil {
		return
	}
	uiTheme := application.theme.Suites
	application.SaveResponse.
		SetLabelColor(uiTheme.SuiteForeground).
		SetFieldTextColor(uiTheme.SuiteFocusForeground).
		SetFieldBackgroundColor(uiTheme.SuiteFocusBackground)
	application.SaveResponse.SetBackgroundColor(uiTheme.BackgroundFocus)
	application.SaveResponse.SetBorderColor(uiTheme.BorderFocus)
	application.SaveResponse.SetTitleColor(uiTheme.TitleFocus)
}

func (application *Application) openSaveResponse() {
	exported, ok := application.Producer.CurrentResponse()
	if !ok {
		if application.Model.CurrentOverlay() != OverlayNone {
			application.closeOverlay()
		}
		application.showExportError(application.config.Locale.Text("no_response_to_export"))
		return
	}
	application.pendingExport = &exported
	application.saveOverwritePath = ""
	root := application.Model.Snapshot().RootDirectoryPath
	application.SaveResponse.SetText(filepath.Join(root, exported.SuggestedFileName))
	application.SaveResponse.SetTitle(application.config.Locale.Text("save_response"))
	application.openOverlay(OverlaySaveResponse)
}

func (application *Application) saveResponse() {
	if application.pendingExport == nil {
		application.closeOverlay()
		application.showExportError(application.config.Locale.Text("no_response_to_export"))
		return
	}
	path := strings.TrimSpace(application.SaveResponse.GetText())
	if path == "" {
		application.showExportError(application.config.Locale.Text("response_path_required"))
		return
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(application.Model.Snapshot().RootDirectoryPath, path)
	}
	path = filepath.Clean(path)
	_, statErr := os.Lstat(path)
	exists := statErr == nil
	if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		application.showExportError(application.config.Locale.Format("save_response_error", statErr))
		return
	}
	if exists && application.saveOverwritePath != path {
		application.saveOverwritePath = path
		application.SaveResponse.SetTitle(application.config.Locale.Text("confirm_response_overwrite"))
		application.stopFooterProgress()
		application.Footer.UpdateIndicatorState(footer.IndicatorDefault)
		application.Footer.UpdateStatus(application.config.Locale.Text("confirm_response_overwrite"))
		return
	}

	exported := *application.pendingExport
	application.pendingExport = nil
	application.saveOverwritePath = ""
	application.closeOverlay()
	application.stopFooterProgress()
	application.Footer.UpdateIndicatorState(footer.IndicatorDefault)
	application.Footer.UpdateStatus(application.config.Locale.Text("saving_response"))

	go func() {
		err := writeResponseFile(path, []byte(exported.RawBody), exists)
		application.Element.QueueUpdateDraw(func() {
			if err != nil {
				application.showExportError(application.config.Locale.Format("save_response_error", err))
				return
			}
			application.Footer.UpdateIndicatorState(footer.IndicatorSuccess)
			status := application.config.Locale.Format("response_saved", len([]byte(exported.RawBody)), path)
			if exported.Truncated {
				status += " — " + application.config.Locale.Text("body_truncated")
			}
			application.Footer.UpdateStatus(status)
		})
	}()
}

func (application *Application) showExportError(message string) {
	application.stopFooterProgress()
	application.Footer.UpdateIndicatorState(footer.IndicatorFailure)
	application.Footer.UpdateStatus(message)
}

func writeResponseFile(path string, contents []byte, overwrite bool) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("create response directory: %w", err)
	}
	if overwrite {
		return replaceResponseFile(path, contents)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("open response file: %w", err)
	}
	removeIncomplete := true
	defer func() {
		if removeIncomplete {
			_ = os.Remove(path)
		}
	}()
	if err := writePrivateResponse(file, contents); err != nil {
		return err
	}
	removeIncomplete = false
	return nil
}

func replaceResponseFile(path string, contents []byte) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".lazyrest-response-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary response file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := writePrivateResponse(temporary, contents); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace response file: %w", err)
	}
	return nil
}

func writePrivateResponse(file *os.File, contents []byte) error {
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return fmt.Errorf("secure response file: %w", err)
	}
	written, writeErr := file.Write(contents)
	if writeErr == nil && written != len(contents) {
		writeErr = io.ErrShortWrite
	}
	if writeErr == nil {
		writeErr = file.Sync()
	}
	closeErr := file.Close()
	if writeErr != nil {
		return fmt.Errorf("write response file: %w", writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close response file: %w", closeErr)
	}
	return nil
}
