package keymap

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

type Action string

const (
	Help            Action = "help"
	Diagnostics     Action = "diagnostics"
	Quit            Action = "quit"
	FocusLeft       Action = "focus_left"
	FocusDown       Action = "focus_down"
	FocusUp         Action = "focus_up"
	FocusRight      Action = "focus_right"
	Open            Action = "open"
	Run             Action = "run"
	Back            Action = "back"
	Search          Action = "search"
	SearchFinish    Action = "search_finish"
	SearchNext      Action = "search_next"
	SearchPrevious  Action = "search_previous"
	Reload          Action = "reload"
	MoveDown        Action = "move_down"
	MoveUp          Action = "move_up"
	ToggleBody      Action = "toggle_body"
	HistoryPrevious Action = "history_previous"
	HistoryNext     Action = "history_next"
	CommandPalette  Action = "command_palette"
	ReloadConfig    Action = "reload_config"
)

var defaults = map[Action][]string{
	Help:            {"?"},
	Diagnostics:     {"d"},
	Quit:            {"q", "ctrl+c"},
	FocusLeft:       {"ctrl+h"},
	FocusDown:       {"ctrl+j"},
	FocusUp:         {"ctrl+k"},
	FocusRight:      {"ctrl+l"},
	Open:            {"enter"},
	Run:             {"enter"},
	Back:            {"esc"},
	Search:          {"/"},
	SearchFinish:    {"enter", "esc"},
	SearchNext:      {"n"},
	SearchPrevious:  {"N"},
	Reload:          {"r"},
	MoveDown:        {"j"},
	MoveUp:          {"k"},
	ToggleBody:      {"p"},
	HistoryPrevious: {"["},
	HistoryNext:     {"]"},
	CommandPalette:  {":", "ctrl+p"},
	ReloadConfig:    {"ctrl+r"},
}

type Bindings struct {
	keys map[Action][]key
}

type key struct {
	name string
	rune rune
	code tcell.Key
}

func New(overrides map[string][]string) (*Bindings, error) {
	configured := make(map[Action][]string, len(defaults))
	for action, keys := range defaults {
		configured[action] = append([]string(nil), keys...)
	}
	for name, keys := range overrides {
		action := Action(name)
		if _, ok := defaults[action]; !ok {
			return nil, fmt.Errorf("unknown keybinding action %q", name)
		}
		if len(keys) == 0 {
			return nil, fmt.Errorf("keybinding %q must contain at least one key", name)
		}
		configured[action] = append([]string(nil), keys...)
	}

	bindings := &Bindings{keys: make(map[Action][]key, len(configured))}
	for action, names := range configured {
		for _, name := range names {
			parsed, err := parse(name)
			if err != nil {
				return nil, fmt.Errorf("keybinding %q: %w", action, err)
			}
			bindings.keys[action] = append(bindings.keys[action], parsed)
		}
	}
	if err := bindings.validateConflicts(); err != nil {
		return nil, err
	}
	return bindings, nil
}

func Default() *Bindings {
	bindings, err := New(nil)
	if err != nil {
		panic(err)
	}
	return bindings
}

func (bindings *Bindings) Matches(action Action, event *tcell.EventKey) bool {
	if bindings == nil || event == nil {
		return false
	}
	for _, candidate := range bindings.keys[action] {
		if candidate.matches(event) {
			return true
		}
	}
	return false
}

func (bindings *Bindings) Describe(action Action) string {
	if bindings == nil {
		return ""
	}
	names := make([]string, 0, len(bindings.keys[action]))
	for _, key := range bindings.keys[action] {
		names = append(names, key.name)
	}
	return strings.Join(names, " / ")
}

func (bindings *Bindings) Map() map[string][]string {
	result := make(map[string][]string, len(bindings.keys))
	for action, keys := range bindings.keys {
		for _, key := range keys {
			result[string(action)] = append(result[string(action)], key.name)
		}
	}
	return result
}

func (bindings *Bindings) validateConflicts() error {
	global := []Action{Help, Diagnostics, Quit, FocusLeft, FocusDown, FocusUp, FocusRight, CommandPalette, ReloadConfig}
	contexts := []struct {
		name    string
		actions []Action
	}{
		{"files", append(append([]Action{}, global...), Open, Search, SearchNext, SearchPrevious, Reload)},
		{"suites", append(append([]Action{}, global...), Open, Back, Search, MoveDown, MoveUp)},
		{"suite", append(append([]Action{}, global...), Run, Back)},
		{"producer", append(append([]Action{}, global...), Back, Search, HistoryPrevious, HistoryNext, ToggleBody)},
		{"search", []Action{SearchFinish}},
		{"overlay", []Action{Quit, Back, CommandPalette, ReloadConfig, Help, Diagnostics, MoveDown, MoveUp}},
	}
	for _, context := range contexts {
		used := map[string]Action{}
		for _, action := range context.actions {
			for _, candidate := range bindings.keys[action] {
				identity := candidate.identity()
				if previous, ok := used[identity]; ok && previous != action {
					actions := []string{string(previous), string(action)}
					sort.Strings(actions)
					return fmt.Errorf("key %q is assigned to both %q and %q in %s context", candidate.name, actions[0], actions[1], context.name)
				}
				used[identity] = action
			}
		}
	}
	return nil
}

func (candidate key) identity() string {
	if candidate.rune != 0 {
		return fmt.Sprintf("r:%U", candidate.rune)
	}
	return fmt.Sprintf("k:%d", candidate.code)
}

func (candidate key) matches(event *tcell.EventKey) bool {
	if candidate.rune != 0 {
		return event.Key() == tcell.KeyRune && event.Rune() == candidate.rune && event.Modifiers()&tcell.ModCtrl == 0
	}
	if candidate.code == tcell.KeyCtrlH && event.Key() == tcell.KeyBackspace {
		return true
	}
	return event.Key() == candidate.code
}

func parse(value string) (key, error) {
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) == 1 {
		r, _ := utf8.DecodeRuneInString(value)
		return key{name: value, rune: r}, nil
	}

	normalized := strings.ToLower(value)
	if strings.HasPrefix(normalized, "ctrl+") && len([]rune(normalized)) == 6 {
		letter := []rune(normalized)[5]
		if letter >= 'a' && letter <= 'z' {
			return key{name: normalized, code: tcell.Key(int(tcell.KeyCtrlA) + int(letter-'a'))}, nil
		}
	}
	special := map[string]tcell.Key{
		"enter": tcell.KeyEnter, "esc": tcell.KeyEsc, "escape": tcell.KeyEsc,
		"backspace": tcell.KeyBackspace, "tab": tcell.KeyTab,
		"up": tcell.KeyUp, "down": tcell.KeyDown, "left": tcell.KeyLeft, "right": tcell.KeyRight,
		"home": tcell.KeyHome, "end": tcell.KeyEnd, "pgup": tcell.KeyPgUp, "pgdn": tcell.KeyPgDn,
	}
	if strings.HasPrefix(normalized, "f") {
		number, err := strconv.Atoi(strings.TrimPrefix(normalized, "f"))
		if err == nil && number >= 1 && number <= 12 {
			return key{name: normalized, code: tcell.Key(int(tcell.KeyF1) + number - 1)}, nil
		}
	}
	if code, ok := special[normalized]; ok {
		return key{name: normalized, code: code}, nil
	}
	return key{}, fmt.Errorf("unsupported key %q", value)
}
