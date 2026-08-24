package ui

import (
	"sync"

	"github.com/Hecatoncheir/lazyrest/finder"
	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
)

type Phase uint8

const (
	PhaseIdle Phase = iota
	PhaseLoading
	PhaseReady
	PhaseFailed
)

type TaskState struct {
	Phase       Phase
	Error       string
	Current     int64
	Total       int64
	HasProgress bool
	Outcome     Outcome
}

type Outcome uint8

const (
	OutcomeNone Outcome = iota
	OutcomeSuccess
	OutcomeFailure
)

type Overlay uint8

const (
	OverlayNone Overlay = iota
	OverlayDiagnostics
	OverlayHelp
)

type State struct {
	RootDirectoryPath string
	EnvironmentName   string
	Startup           TaskState
	Files             TaskState
	Parser            TaskState
	Request           TaskState
	Directory         finder.Directory
	SelectedFile      *finder.File
	Suites            []parserhttp.HttpSuite
	SelectedSuite     *parserhttp.HttpSuite
	Diagnostics       []parserhttp.Diagnostic
	Overlay           Overlay
}

type Model struct {
	mutex sync.RWMutex
	state State
}

func NewModel(rootDirectoryPath, environmentName string) *Model {
	return &Model{state: State{
		RootDirectoryPath: rootDirectoryPath,
		EnvironmentName:   environmentName,
	}}
}

func (model *Model) Snapshot() State {
	model.mutex.RLock()
	defer model.mutex.RUnlock()
	return cloneState(model.state)
}

func (model *Model) CurrentOverlay() Overlay {
	model.mutex.RLock()
	defer model.mutex.RUnlock()
	return model.state.Overlay
}

func (model *Model) update(update func(*State)) {
	model.mutex.Lock()
	defer model.mutex.Unlock()
	update(&model.state)
}

func cloneState(state State) State {
	cloned := state
	cloned.Directory = cloneDirectory(state.Directory)
	if state.SelectedFile != nil {
		selectedFile := *state.SelectedFile
		cloned.SelectedFile = &selectedFile
	}
	cloned.Suites = make([]parserhttp.HttpSuite, len(state.Suites))
	for index, suite := range state.Suites {
		cloned.Suites[index] = cloneSuite(suite)
	}
	if state.SelectedSuite != nil {
		selectedSuite := cloneSuite(*state.SelectedSuite)
		cloned.SelectedSuite = &selectedSuite
	}
	cloned.Diagnostics = append([]parserhttp.Diagnostic(nil), state.Diagnostics...)
	return cloned
}

func cloneDirectory(directory finder.Directory) finder.Directory {
	cloned := directory
	cloned.Files = append([]finder.File(nil), directory.Files...)
	cloned.Warnings = append([]string(nil), directory.Warnings...)
	cloned.Directories = make([]finder.Directory, len(directory.Directories))
	for index, child := range directory.Directories {
		cloned.Directories[index] = cloneDirectory(child)
	}
	return cloned
}

func cloneSuite(suite parserhttp.HttpSuite) parserhttp.HttpSuite {
	cloned := suite
	cloned.Header = make(map[string]string, len(suite.Header))
	for key, value := range suite.Header {
		cloned.Header[key] = value
	}
	cloned.SecretValues = append([]string(nil), suite.SecretValues...)
	return cloned
}

func directoryContainsFile(directory finder.Directory, path string) bool {
	for _, file := range directory.Files {
		if file.Path == path {
			return true
		}
	}
	for _, child := range directory.Directories {
		if directoryContainsFile(child, path) {
			return true
		}
	}
	return false
}
