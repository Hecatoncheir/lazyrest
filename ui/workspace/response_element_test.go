package workspace

import (
	"testing"

	"github.com/rivo/tview"
)

func buildWorkspace(t *testing.T) (*Workspace, tview.Primitive) {
	t.Helper()
	widget := New()
	producerElement := tview.NewTextView()
	widget.Element = tview.NewFlex().SetDirection(tview.FlexRowCSS)
	widget.suitesArea = tview.NewFlex()
	widget.treeElement = tview.NewTextView()
	widget.suitesElement = tview.NewTextView()
	widget.suiteElement = tview.NewTextView()
	widget.producerElement = producerElement
	widget.mode = layoutWide
	widget.focus = focusFiles
	widget.render()
	return widget, producerElement
}

func TestSetResponseElementSwapsTheSlot(t *testing.T) {
	widget, producerElement := buildWorkspace(t)
	streamElement := tview.NewTextView()

	widget.SetResponseElement(streamElement)
	if widget.ResponseElement() != streamElement {
		t.Fatal("the response slot was not swapped")
	}

	widget.SetResponseElement(producerElement)
	if widget.ResponseElement() != producerElement {
		t.Fatal("the response slot did not swap back")
	}
}

func TestSetResponseElementIgnoresNil(t *testing.T) {
	widget, producerElement := buildWorkspace(t)
	widget.SetResponseElement(nil)
	if widget.ResponseElement() != producerElement {
		t.Fatal("a nil element emptied the response slot")
	}
}

// Focus detection compares against whatever fills the slot, so it has to keep
// working after a swap.
func TestResizeTracksFocusAfterASwap(t *testing.T) {
	widget, _ := buildWorkspace(t)
	streamElement := tview.NewTextView()
	widget.SetResponseElement(streamElement)

	widget.Resize(160, streamElement)
	if widget.focus != focusResponse {
		t.Fatalf("focus = %v, want focusResponse after focusing the swapped element", widget.focus)
	}
}

// Every layout mode has to keep placing the slot, not just the wide one.
func TestSwappedElementSurvivesEveryLayoutMode(t *testing.T) {
	widget, _ := buildWorkspace(t)
	streamElement := tview.NewTextView()
	widget.SetResponseElement(streamElement)

	for _, width := range []int{60, 100, 160} {
		widget.Resize(width, streamElement)
		if widget.ResponseElement() != streamElement {
			t.Fatalf("width %d dropped the swapped element", width)
		}
	}
}
