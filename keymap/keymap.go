package keymap

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

type Action string

const (
	Help             Action = "help"
	Diagnostics      Action = "diagnostics"
	Quit             Action = "quit"
	FocusLeft        Action = "focus_left"
	FocusDown        Action = "focus_down"
	FocusUp          Action = "focus_up"
	FocusRight       Action = "focus_right"
	Open             Action = "open"
	Run              Action = "run"
	Back             Action = "back"
	Search           Action = "search"
	SearchFinish     Action = "search_finish"
	SearchNext       Action = "search_next"
	SearchPrevious   Action = "search_previous"
	Reload           Action = "reload"
	MoveDown         Action = "move_down"
	MoveUp           Action = "move_up"
	HalfPageDown     Action = "half_page_down"
	HalfPageUp       Action = "half_page_up"
	CenterView       Action = "center_view"
	ToggleBody       Action = "toggle_body"
	CopyResponseBody Action = "copy_response_body"
	CopyResponse     Action = "copy_response"
	SaveResponse     Action = "save_response"
	SaveFullResponse Action = "save_full_response"
	HistoryPrevious  Action = "history_previous"
	HistoryNext      Action = "history_next"
	CommandPalette   Action = "command_palette"
	ReloadConfig     Action = "reload_config"
)

var defaults = map[Action][]string{
	Help:             {"?"},
	Diagnostics:      {"d"},
	Quit:             {"q", "ctrl+c"},
	FocusLeft:        {"ctrl+h"},
	FocusDown:        {"ctrl+j"},
	FocusUp:          {"ctrl+k"},
	FocusRight:       {"ctrl+l"},
	Open:             {"enter", "l"},
	Run:              {"enter"},
	Back:             {"esc"},
	Search:           {"/"},
	SearchFinish:     {"enter", "esc"},
	SearchNext:       {"n"},
	SearchPrevious:   {"N"},
	Reload:           {"r"},
	MoveDown:         {"j"},
	MoveUp:           {"k"},
	HalfPageDown:     {"ctrl+d"},
	HalfPageUp:       {"ctrl+u"},
	CenterView:       {"zz"},
	ToggleBody:       {"p"},
	CopyResponseBody: {"y"},
	CopyResponse:     {"Y"},
	SaveResponse:     {"s"},
	SaveFullResponse: {"S"},
	HistoryPrevious:  {"["},
	HistoryNext:      {"]"},
	CommandPalette:   {":", "ctrl+p"},
	ReloadConfig:     {"ctrl+r"},
}

type Bindings struct {
	keys map[Action][]binding
}

type binding struct {
	name  string
	steps []key
}

type key struct {
	rune      rune
	code      tcell.Key
	modifiers tcell.ModMask
}

type SequenceMatch int

const (
	SequenceNoMatch SequenceMatch = iota
	SequencePrefix
	SequenceFull
)

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

	bindings := &Bindings{keys: make(map[Action][]binding, len(configured))}
	for action, names := range configured {
		for _, name := range names {
			parsed, err := parse(name)
			if err != nil {
				return nil, fmt.Errorf("keybinding %q: %w", action, err)
			}
			if len(parsed.steps) > 1 && action != CenterView {
				return nil, fmt.Errorf("keybinding %q: key sequences are only supported for %q", action, CenterView)
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
		if len(candidate.steps) == 1 && candidate.steps[0].matches(event) {
			return true
		}
	}
	return false
}

func (bindings *Bindings) MatchesSequence(action Action, events []*tcell.EventKey) SequenceMatch {
	if bindings == nil || len(events) == 0 {
		return SequenceNoMatch
	}
	matchedPrefix := false
	for _, candidate := range bindings.keys[action] {
		if len(events) > len(candidate.steps) {
			continue
		}
		matched := true
		for index, event := range events {
			if event == nil || !candidate.steps[index].matches(event) {
				matched = false
				break
			}
		}
		if !matched {
			continue
		}
		if len(events) == len(candidate.steps) {
			return SequenceFull
		}
		matchedPrefix = true
	}
	if matchedPrefix {
		return SequencePrefix
	}
	return SequenceNoMatch
}

func (bindings *Bindings) Describe(action Action) string {
	if bindings == nil {
		return ""
	}
	names := make([]string, 0, len(bindings.keys[action]))
	for _, binding := range bindings.keys[action] {
		names = append(names, binding.name)
	}
	return strings.Join(names, " / ")
}

func (bindings *Bindings) Map() map[string][]string {
	result := make(map[string][]string, len(bindings.keys))
	for action, bindings := range bindings.keys {
		for _, binding := range bindings {
			result[string(action)] = append(result[string(action)], binding.name)
		}
	}
	return result
}

func (bindings *Bindings) validateConflicts() error {
	global := []Action{
		Help, Diagnostics, Quit,
		FocusLeft, FocusDown, FocusUp, FocusRight,
		HalfPageDown, HalfPageUp, CenterView,
		CommandPalette, ReloadConfig,
	}
	contexts := []struct {
		name    string
		actions []Action
	}{
		{"files", append(append([]Action{}, global...), Open, Search, SearchNext, SearchPrevious, Reload)},
		{"suites", append(append([]Action{}, global...), Open, Back, Search, MoveDown, MoveUp)},
		{"suite", append(append([]Action{}, global...), Run, Back)},
		{"producer", append(append([]Action{}, global...), Back, Search, HistoryPrevious, HistoryNext, ToggleBody, CopyResponseBody, CopyResponse, SaveResponse, SaveFullResponse)},
		{"search", []Action{SearchFinish}},
		{"overlay", []Action{Quit, Back, CommandPalette, ReloadConfig, Help, Diagnostics, MoveDown, MoveUp, HalfPageDown, HalfPageUp, CenterView}},
	}
	for _, context := range contexts {
		used := map[string]struct {
			action  Action
			binding binding
		}{}
		for _, action := range context.actions {
			for _, candidate := range bindings.keys[action] {
				identity := candidate.identity()
				for _, previous := range used {
					if previous.action != action && (candidate.hasPrefix(previous.binding) || previous.binding.hasPrefix(candidate)) {
						actions := []string{string(previous.action), string(action)}
						sort.Strings(actions)
						return fmt.Errorf("key %q conflicts with %q assigned to %q and %q in %s context", candidate.name, previous.binding.name, actions[0], actions[1], context.name)
					}
				}
				used[identity] = struct {
					action  Action
					binding binding
				}{action: action, binding: candidate}
			}
		}
	}
	return nil
}

func (candidate binding) identity() string {
	identities := make([]string, 0, len(candidate.steps))
	for _, step := range candidate.steps {
		identities = append(identities, step.identity())
	}
	return strings.Join(identities, ";")
}

func (candidate binding) hasPrefix(prefix binding) bool {
	if len(prefix.steps) > len(candidate.steps) {
		return false
	}
	for index, step := range prefix.steps {
		if step.identity() != candidate.steps[index].identity() {
			return false
		}
	}
	return true
}

func (candidate key) identity() string {
	if candidate.rune != 0 {
		return fmt.Sprintf("r:%U:m:%d", candidate.rune, candidate.modifiers)
	}
	return fmt.Sprintf("k:%d", candidate.code)
}

func (candidate key) matches(event *tcell.EventKey) bool {
	if candidate.rune != 0 {
		return event.Key() == tcell.KeyRune && event.Rune() == candidate.rune && event.Modifiers()&tcell.ModCtrl == candidate.modifiers
	}
	if candidate.code == tcell.KeyCtrlH && event.Key() == tcell.KeyBackspace {
		return true
	}
	return event.Key() == candidate.code
}

func parse(value string) (binding, error) {
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) == 1 {
		r, _ := utf8.DecodeRuneInString(value)
		return binding{name: value, steps: []key{{rune: r}}}, nil
	}

	normalized := strings.ToLower(value)
	if strings.HasPrefix(normalized, "ctrl+") && len([]rune(normalized)) == 6 {
		letter := []rune(normalized)[5]
		if letter >= 'a' && letter <= 'z' {
			return binding{name: normalized, steps: []key{{code: tcell.Key(int(tcell.KeyCtrlA) + int(letter-'a'))}}}, nil
		}
		return binding{name: normalized, steps: []key{{rune: letter, modifiers: tcell.ModCtrl}}}, nil
	}
	if strings.HasPrefix(normalized, "ctrl+") {
		return binding{}, fmt.Errorf("unsupported key %q", value)
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
			return binding{name: normalized, steps: []key{{code: tcell.Key(int(tcell.KeyF1) + number - 1)}}}, nil
		}
	}
	if code, ok := special[normalized]; ok {
		return binding{name: normalized, steps: []key{{code: code}}}, nil
	}
	sequence := []rune(value)
	if len(sequence) > 1 {
		steps := make([]key, 0, len(sequence))
		for _, r := range sequence {
			if !unicode.IsPrint(r) || unicode.IsSpace(r) {
				return binding{}, fmt.Errorf("unsupported key %q", value)
			}
			steps = append(steps, key{rune: r})
		}
		return binding{name: value, steps: steps}, nil
	}
	return binding{}, fmt.Errorf("unsupported key %q", value)
}
