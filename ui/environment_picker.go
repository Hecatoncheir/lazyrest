package ui

import (
	"maps"
	"slices"

	"github.com/Hecatoncheir/lazyrest/environment"
	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/ui/footer"
	"github.com/rivo/tview"
)

func (application *Application) buildEnvironmentPicker() {
	picker := tview.NewList().ShowSecondaryText(false)
	picker.SetBorder(true).SetTitleAlign(tview.AlignCenter)
	application.applyCommandPaletteTheme(picker)
	application.EnvironmentPicker = picker
	application.refreshEnvironmentPicker()
}

func (application *Application) refreshEnvironmentPicker() {
	if application.EnvironmentPicker == nil || application.Model == nil {
		return
	}
	picker := application.EnvironmentPicker
	picker.Clear()
	translator := application.config.Locale
	picker.SetTitle(translator.Text("choose_environment"))
	picker.AddItem(translator.Text("base_environment"), "", 0, func() {
		application.selectEnvironment("")
	})

	root := application.Model.Snapshot().RootDirectoryPath
	names, err := environment.Names(root, application.config.Environment)
	if err != nil {
		picker.AddItem(translator.Format("environment_list_error", err), "", 0, nil)
	}
	current := application.Model.Snapshot().EnvironmentName
	selected := 0
	for _, profile := range names {
		name := profile
		picker.AddItem(name, "", 0, func() { application.selectEnvironment(name) })
		if name == current {
			selected = picker.GetItemCount() - 1
		}
	}
	picker.SetCurrentItem(selected)
}

func (application *Application) selectEnvironment(name string) {
	application.closeOverlay()
	application.Producer.CancelActive()
	application.Model.update(func(state *State) {
		state.Request = TaskState{}
		state.Environment = TaskState{Phase: PhaseLoading}
	})
	application.stopFooterProgress()
	application.Footer.UpdateIndicatorState(footer.IndicatorDefault)
	application.Footer.UpdateStatus(application.config.Locale.Text("loading_environment"))
	root := application.Model.Snapshot().RootDirectoryPath
	config := application.config.Environment
	config.Name = name
	loadID := application.startEnvironmentLoad()

	go func() {
		selected, err := application.loadMergedEnvironment(root, config, name)
		application.Element.QueueUpdateDraw(func() {
			if !application.isCurrentEnvironmentLoad(loadID) {
				return
			}
			if err != nil {
				application.Model.update(func(state *State) {
					state.Environment = TaskState{Phase: PhaseFailed, Error: err.Error()}
				})
				application.refreshDiagnostics()
				application.Footer.UpdateIndicatorState(footer.IndicatorFailure)
				application.Footer.UpdateStatus(application.config.Locale.Format("environment_change_error", err))
				return
			}
			if err := application.Producer.ResetEnvironmentSession(); err != nil {
				application.Model.update(func(state *State) {
					state.Environment = TaskState{Phase: PhaseFailed, Error: err.Error()}
				})
				application.refreshDiagnostics()
				application.Footer.UpdateIndicatorState(footer.IndicatorFailure)
				application.Footer.UpdateStatus(application.config.Locale.Format("environment_change_error", err))
				return
			}
			application.config.Environment = config
			application.config.EnvironmentName = selected.Name
			application.Suites.SetParseOptions(parserhttp.ParseOptions{
				Variables:       selected.Values,
				SecretVariables: selected.SecretVariables,
			})
			application.Model.update(func(state *State) {
				state.EnvironmentName = selected.Name
				state.Environment = TaskState{Phase: PhaseReady}
			})
			application.refreshDiagnostics()
			application.Footer.UpdateEnvironment(selected.Name)
			application.Footer.UpdateIndicatorState(footer.IndicatorSuccess)
			displayName := selected.Name
			if displayName == "" {
				displayName = application.config.Locale.Text("base_environment")
			}
			application.Footer.UpdateStatus(application.config.Locale.Format("environment_changed", displayName))
			if file := application.Model.Snapshot().SelectedFile; file != nil {
				onSelectFileCallback(application)(*file)
			}
		})
	}()
}

func (application *Application) loadMergedEnvironment(root string, config environment.Config, initialName string) (environment.Environment, error) {
	selected := environment.Environment{
		Name:            initialName,
		Values:          maps.Clone(application.config.ParseOptions.Variables),
		SecretVariables: append([]string(nil), application.config.ParseOptions.SecretVariables...),
	}
	if selected.Values == nil {
		selected.Values = map[string]string{}
	}
	loaded, err := application.loadEnvironment(root, config)
	if err != nil {
		return environment.Environment{}, err
	}
	secret := make(map[string]struct{}, len(selected.SecretVariables)+len(loaded.SecretVariables))
	for _, variable := range selected.SecretVariables {
		secret[variable] = struct{}{}
	}
	loadedSecrets := make(map[string]struct{}, len(loaded.SecretVariables))
	for _, variable := range loaded.SecretVariables {
		loadedSecrets[variable] = struct{}{}
		secret[variable] = struct{}{}
	}
	for key, value := range loaded.Values {
		selected.Values[key] = value
		if _, isSecret := loadedSecrets[key]; !isSecret {
			delete(secret, key)
		}
	}
	selected.SecretVariables = make([]string, 0, len(secret))
	for variable := range secret {
		selected.SecretVariables = append(selected.SecretVariables, variable)
	}
	slices.Sort(selected.SecretVariables)
	if loaded.Name != "" {
		selected.Name = loaded.Name
	}
	return selected, nil
}
