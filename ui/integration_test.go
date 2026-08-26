package ui

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Hecatoncheir/lazyrest/environment"
	"github.com/Hecatoncheir/lazyrest/finder"
	"github.com/Hecatoncheir/lazyrest/locale"
	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
	"github.com/Hecatoncheir/lazyrest/ui/tree"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestTUIRemainsInteractiveDuringBackgroundStartup(t *testing.T) {
	root := t.TempDir()
	application := BuildApplication(root, Config{Environment: environment.Config{Name: "test"}})
	releaseScan := make(chan struct{})
	application.scanFiles = func(ctx context.Context) tree.ScanResult {
		select {
		case <-releaseScan:
			return tree.ScanResult{Directory: finder.Directory{Name: filepath.Base(root), Path: root}}
		case <-ctx.Done():
			return tree.ScanResult{Err: ctx.Err()}
		}
	}
	application.loadEnvironment = func(string, environment.Config) (environment.Environment, error) {
		return environment.Environment{Name: "test", Values: map[string]string{}}, nil
	}

	screen, _ := runTestApplication(t, application)
	application.Start()
	waitFor(t, "startup loading state", func() bool {
		return application.Model.Snapshot().Files.Phase == PhaseLoading
	})
	waitForScreenText(t, application, screen, " Loading")

	screen.InjectKey(tcell.KeyRune, '?', tcell.ModNone)
	waitFor(t, "help during startup", func() bool {
		return application.Model.Snapshot().Overlay == OverlayHelp &&
			strings.Contains(applicationText(application, screen), "Open or close this help")
	})

	screen.InjectKey(tcell.KeyRune, '?', tcell.ModNone)
	close(releaseScan)
	waitFor(t, "completed startup", func() bool {
		state := application.Model.Snapshot()
		return state.Files.Phase == PhaseReady && state.Startup.Phase == PhaseReady
	})
}

func TestTUIUsesConfiguredLanguage(t *testing.T) {
	translator, err := locale.New("ru", nil)
	if err != nil {
		t.Fatal(err)
	}
	application := BuildApplication(t.TempDir(), Config{Locale: translator})
	screen, _ := runTestApplication(t, application)
	waitFor(t, "Russian interface", func() bool {
		text := applicationText(application, screen)
		return strings.Contains(text, "Файлы") && strings.Contains(text, "Запросы") && strings.Contains(text, "Результат")
	})
}

func TestTUIDiagnosticsAndHelpWorkflow(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "requests.http")
	request := "# @name Missing variable\nGET https://{{missing}}/users\n"
	if err := os.WriteFile(filePath, []byte(request), 0o600); err != nil {
		t.Fatal(err)
	}

	application := BuildApplication(root, Config{})
	screen, _ := runTestApplication(t, application)
	application.Start()
	waitFor(t, "file discovery", func() bool {
		return application.Model.Snapshot().Files.Phase == PhaseReady
	})

	application.Element.QueueUpdateDraw(func() {
		treeView := application.HttpFilesTree.Element.(*tview.TreeView)
		node := findFileReference(treeView.GetRoot(), filePath)
		if node == nil {
			t.Errorf("request file %q is missing from the tree", filePath)
			return
		}
		treeView.SetCurrentNode(node)
		application.Element.SetFocus(treeView)
	})
	screen.InjectKey(tcell.KeyRune, 'l', tcell.ModNone)
	waitFor(t, "l file selection", func() bool {
		state := application.Model.Snapshot()
		return state.SelectedFile != nil && state.SelectedFile.Path == filePath &&
			application.Element.GetFocus() == application.Suites.Element
	})
	waitFor(t, "parser diagnostics", func() bool {
		state := application.Model.Snapshot()
		return state.Parser.Phase == PhaseReady && len(state.Diagnostics) > 0
	})

	screen.InjectKey(tcell.KeyRune, 'd', tcell.ModNone)
	waitFor(t, "diagnostics window", func() bool {
		return application.Model.Snapshot().Overlay == OverlayDiagnostics &&
			strings.Contains(applicationText(application, screen), "undefined variable: missing")
	})

	screen.InjectKey(tcell.KeyRune, '?', tcell.ModNone)
	waitFor(t, "help window", func() bool {
		return application.Model.Snapshot().Overlay == OverlayHelp &&
			strings.Contains(applicationText(application, screen), "Move focus left")
	})

	screen.InjectKey(tcell.KeyEsc, 0, tcell.ModNone)
	waitFor(t, "closed overlay", func() bool {
		return application.Model.Snapshot().Overlay == OverlayNone
	})
}

func TestTUITreeCtrlLMovesFocusWithoutOpeningFile(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "requests.http")
	if err := os.WriteFile(filePath, []byte("GET https://example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	application := BuildApplication(root, Config{})
	screen, _ := runTestApplication(t, application)
	application.Start()
	waitFor(t, "file discovery", func() bool {
		return application.Model.Snapshot().Files.Phase == PhaseReady
	})

	application.Element.QueueUpdateDraw(func() {
		treeView := application.HttpFilesTree.Element.(*tview.TreeView)
		node := findFileReference(treeView.GetRoot(), filePath)
		if node == nil {
			t.Errorf("request file %q is missing from the tree", filePath)
			return
		}
		treeView.SetCurrentNode(node)
		application.Element.SetFocus(treeView)
	})

	screen.InjectKey(tcell.KeyCtrlL, 0, tcell.ModCtrl)
	waitFor(t, "Ctrl+l focus change", func() bool {
		return application.Element.GetFocus() == application.Suites.Element
	})
	if selected := application.Model.Snapshot().SelectedFile; selected != nil {
		t.Fatalf("Ctrl+l opened a file while moving focus: %+v", *selected)
	}
}

func TestTUIProducerAnimatesProgressWhileWaiting(t *testing.T) {
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseRequest := func() {
		releaseOnce.Do(func() { close(release) })
	}
	client := &http.Client{Transport: uiRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		select {
		case <-release:
		case <-request.Context().Done():
			return nil, request.Context().Err()
		}
		return &http.Response{
			StatusCode:    http.StatusOK,
			Header:        http.Header{"Content-Type": []string{"application/json"}},
			Body:          io.NopCloser(strings.NewReader(`{"ok":true}`)),
			ContentLength: 11,
			Proto:         "HTTP/1.1",
		}, nil
	})}
	defer releaseRequest()

	application := BuildApplication(t.TempDir(), Config{
		Runner: runner.Config{Client: client, Timeout: 10 * time.Second},
	})
	screen, _ := runTestApplication(t, application)
	application.Element.QueueUpdateDraw(func() {
		onSuiteRun(application)(parserhttp.HttpSuite{
			Name:   "Slow request",
			Method: http.MethodGet,
			Uri:    "https://example.test/slow",
			Header: http.Header{},
		})
	})

	waitForScreenText(t, application, screen, "Running request")
	waitForScreenText(t, application, screen, "====>")
	waitForScreenText(t, application, screen, " Running")
	firstFrame := applicationText(application, screen)
	waitFor(t, "animated progress bar", func() bool {
		text := applicationText(application, screen)
		return strings.Contains(text, "Running request") && text != firstFrame
	})

	releaseRequest()
	waitFor(t, "completed request", func() bool {
		request := application.Model.Snapshot().Request
		return request.Phase == PhaseReady && request.Outcome == OutcomeSuccess
	})
	waitForScreenText(t, application, screen, " Success")
}

func TestTUIProducerCopiesAndSavesTheCurrentResponse(t *testing.T) {
	root := t.TempDir()
	rawBody := `{"ok":true,"items":[1,2]}`
	client := &http.Client{Transport: uiRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode:    http.StatusOK,
			Header:        http.Header{"Content-Type": []string{"application/json"}},
			Body:          io.NopCloser(strings.NewReader(rawBody)),
			ContentLength: int64(len(rawBody)),
			Proto:         "HTTP/1.1",
		}, nil
	})}
	application := BuildApplication(root, Config{Runner: runner.Config{Client: client, Timeout: 10 * time.Second}})
	screen, _ := runTestApplication(t, application)
	application.Element.QueueUpdateDraw(func() {
		onSuiteRun(application)(parserhttp.HttpSuite{
			Name:   "List items",
			Method: http.MethodGet,
			Uri:    "https://example.test/items",
			Header: http.Header{},
		})
	})
	waitForScreenText(t, application, screen, `"items"`)

	screen.InjectKey(tcell.KeyRune, 'y', tcell.ModNone)
	waitFor(t, "pretty response body in clipboard", func() bool {
		return strings.Contains(applicationClipboard(application, screen), "\n  \"items\": [")
	})
	if copied := applicationClipboard(application, screen); strings.Contains(copied, "Response:") || strings.Contains(copied, "[green]") {
		t.Fatalf("clipboard contains response-pane markup: %q", copied)
	}

	screen.InjectKey(tcell.KeyRune, 'p', tcell.ModNone)
	screen.InjectKey(tcell.KeyRune, 'y', tcell.ModNone)
	waitFor(t, "raw response body in clipboard", func() bool {
		return applicationClipboard(application, screen) == rawBody
	})

	screen.InjectKey(tcell.KeyRune, 'Y', tcell.ModNone)
	waitFor(t, "full response in clipboard", func() bool {
		copied := applicationClipboard(application, screen)
		return strings.HasPrefix(copied, "HTTP/1.1 200 OK\n") &&
			strings.Contains(copied, "Content-Type: application/json\n\n"+rawBody)
	})

	savedPath := filepath.Join(root, "exports", "items.json")
	screen.InjectKey(tcell.KeyRune, 's', tcell.ModNone)
	waitFor(t, "save response dialog", func() bool {
		return application.Model.CurrentOverlay() == OverlaySaveResponse
	})
	setInputText(t, application, application.SaveResponse, savedPath)
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
	waitFor(t, "saved response body", func() bool {
		contents, err := os.ReadFile(savedPath)
		return err == nil && string(contents) == rawBody
	})
	info, err := os.Stat(savedPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("saved response permissions: %o", info.Mode().Perm())
	}

	if err := os.WriteFile(savedPath, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	screen.InjectKey(tcell.KeyRune, 's', tcell.ModNone)
	waitFor(t, "second save response dialog", func() bool {
		return application.Model.CurrentOverlay() == OverlaySaveResponse
	})
	setInputText(t, application, application.SaveResponse, savedPath)
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
	waitFor(t, "overwrite confirmation", func() bool {
		return strings.Contains(applicationText(application, screen), "press Enter again")
	})
	contents, err := os.ReadFile(savedPath)
	if err != nil || string(contents) != "old" {
		t.Fatalf("first Enter overwrote the response: %q, %v", contents, err)
	}
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
	waitFor(t, "confirmed response overwrite", func() bool {
		contents, err := os.ReadFile(savedPath)
		return err == nil && string(contents) == rawBody
	})
}

func setInputText(t *testing.T, application *Application, input *tview.InputField, text string) {
	t.Helper()
	done := make(chan struct{})
	application.Element.QueueUpdateDraw(func() {
		input.SetText(text)
		close(done)
	})
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out setting input text")
	}
}

type uiRoundTripFunc func(*http.Request) (*http.Response, error)

func (roundTrip uiRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTrip(request)
}

func runTestApplication(t *testing.T, application *Application) (tcell.SimulationScreen, <-chan error) {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	application.Element.SetScreen(screen)
	screen.SetSize(120, 36)
	done := make(chan error, 1)
	go func() {
		done <- application.Element.Run()
	}()
	waitFor(t, "initial draw", func() bool {
		return strings.TrimSpace(applicationText(application, screen)) != ""
	})
	t.Cleanup(func() {
		application.stopFooterProgress()
		application.Producer.CancelActive()
		application.Suites.CancelLoad()
		application.HttpFilesTree.CancelReload()
		application.Element.Stop()
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("TUI stopped with an error: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("TUI did not stop")
		}
	})
	return screen, done
}

func waitFor(t *testing.T, description string, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", description)
}

func waitForScreenText(t *testing.T, application *Application, screen tcell.SimulationScreen, expected string) {
	t.Helper()
	var content string
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		content = applicationText(application, screen)
		if strings.Contains(content, expected) {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for screen text %q:\n%s", expected, content)
}

func simulationText(screen tcell.SimulationScreen) string {
	cells, width, height := screen.GetContents()
	var output strings.Builder
	for row := 0; row < height; row++ {
		for column := 0; column < width; column++ {
			cell := cells[row*width+column]
			if len(cell.Runes) == 0 {
				output.WriteRune(' ')
				continue
			}
			output.WriteRune(cell.Runes[0])
		}
		output.WriteByte('\n')
	}
	return output.String()
}

func applicationText(application *Application, screen tcell.SimulationScreen) string {
	var content string
	application.Element.QueueUpdate(func() {
		content = simulationText(screen)
	})
	return content
}

func applicationClipboard(application *Application, screen tcell.SimulationScreen) string {
	var content string
	application.Element.QueueUpdate(func() {
		content = string(screen.GetClipboardData())
	})
	return content
}

func findFileReference(node *tview.TreeNode, path string) *tview.TreeNode {
	if node == nil {
		return nil
	}
	if file, ok := node.GetReference().(finder.File); ok && file.Path == path {
		return node
	}
	for _, child := range node.GetChildren() {
		if match := findFileReference(child, path); match != nil {
			return match
		}
	}
	return nil
}

func TestTUIChainsRequestsThroughAnEarlierResponse(t *testing.T) {
	var sentAuthorization string
	sourceFilePath := filepath.Join(t.TempDir(), "requests.http")
	client := &http.Client{Transport: uiRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		if strings.HasSuffix(request.URL.Path, "/auth") {
			return &http.Response{
				StatusCode:    http.StatusOK,
				Header:        http.Header{"Content-Type": []string{"application/json"}},
				Body:          io.NopCloser(strings.NewReader(`{"token":"abc123"}`)),
				ContentLength: 18,
				Proto:         "HTTP/1.1",
			}, nil
		}
		sentAuthorization = request.Header.Get("Authorization")
		return &http.Response{
			StatusCode:    http.StatusOK,
			Header:        http.Header{"Content-Type": []string{"application/json"}},
			Body:          io.NopCloser(strings.NewReader(`{"me":"ok"}`)),
			ContentLength: 11,
			Proto:         "HTTP/1.1",
		}, nil
	})}

	application := BuildApplication(t.TempDir(), Config{
		Runner: runner.Config{Client: client, Timeout: 10 * time.Second},
	})
	screen, _ := runTestApplication(t, application)

	login := parserhttp.HttpSuite{
		Name:           "login",
		Method:         http.MethodPost,
		Uri:            "https://example.test/auth",
		Header:         http.Header{},
		SourceFilePath: sourceFilePath,
	}
	application.Element.QueueUpdateDraw(func() { onSuiteRun(application)(login) })
	waitForScreenText(t, application, screen, "abc123")

	profile := parserhttp.HttpSuite{
		Name:           "profile",
		Method:         http.MethodGet,
		Uri:            "https://example.test/me",
		Header:         http.Header{"Authorization": []string{"Bearer {{login.response.body.$.token}}"}},
		SourceFilePath: sourceFilePath,
	}
	application.Element.QueueUpdateDraw(func() { onSuiteRun(application)(profile) })
	waitForScreenText(t, application, screen, `"me"`)

	if sentAuthorization != "Bearer abc123" {
		t.Fatalf("the captured token did not reach the request: %q", sentAuthorization)
	}
}

func TestTUIDoesNotResolveAResponseReferenceFromAnotherFile(t *testing.T) {
	var requests atomic.Int32
	client := &http.Client{Transport: uiRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Path != "/auth" {
			t.Errorf("request from another file was sent: %s", request.URL.Path)
			return nil, errors.New("must not be reached")
		}
		return &http.Response{
			StatusCode:    http.StatusOK,
			Header:        http.Header{"Content-Type": []string{"application/json"}},
			Body:          io.NopCloser(strings.NewReader(`{"token":"first-token"}`)),
			ContentLength: 23,
			Proto:         "HTTP/1.1",
		}, nil
	})}

	application := BuildApplication(t.TempDir(), Config{
		Runner: runner.Config{Client: client, Timeout: 10 * time.Second},
	})
	screen, _ := runTestApplication(t, application)

	login := parserhttp.HttpSuite{
		Name:           "login",
		Method:         http.MethodPost,
		Uri:            "https://example.test/auth",
		Header:         http.Header{},
		SourceFilePath: "first.http",
	}
	application.Element.QueueUpdateDraw(func() { onSuiteRun(application)(login) })
	waitForScreenText(t, application, screen, "first-token")

	profile := parserhttp.HttpSuite{
		Name:           "profile",
		Method:         http.MethodGet,
		Uri:            "https://example.test/me",
		Header:         http.Header{"Authorization": []string{"Bearer {{login.response.body.$.token}}"}},
		SourceFilePath: "second.http",
	}
	application.Element.QueueUpdateDraw(func() { onSuiteRun(application)(profile) })
	waitForScreenText(t, application, screen, `"login" has not`)
	waitFor(t, "cross-file reference failure", func() bool {
		state := application.Model.Snapshot()
		return state.Request.Phase == PhaseFailed && state.Request.Outcome == OutcomeFailure
	})

	if got := requests.Load(); got != 1 {
		t.Fatalf("unexpected number of sent requests: got %d, want 1", got)
	}
}

func TestTUIReportsAReferenceToARequestThatHasNotRun(t *testing.T) {
	client := &http.Client{Transport: uiRoundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Error("a request was sent although its reference could not be resolved")
		return nil, errors.New("must not be reached")
	})}

	application := BuildApplication(t.TempDir(), Config{
		Runner: runner.Config{Client: client, Timeout: 10 * time.Second},
	})
	screen, _ := runTestApplication(t, application)

	suite := parserhttp.HttpSuite{
		Name:   "profile",
		Method: http.MethodGet,
		Uri:    "https://example.test/me",
		Header: http.Header{"Authorization": []string{"Bearer {{login.response.body.$.token}}"}},
	}
	application.Element.QueueUpdateDraw(func() { onSuiteRun(application)(suite) })
	// The message wraps in the pane, so only the part before the wrap is
	// matched.
	waitForScreenText(t, application, screen, `"login" has not`)

	// Nothing was sent, but the run still has to settle: a request left in the
	// running state keeps the footer animating for ever.
	waitFor(t, "the run to be reported as finished", func() bool {
		state := application.Model.Snapshot()
		return state.Request.Phase != PhaseLoading && state.Request.Outcome == OutcomeFailure
	})
}
