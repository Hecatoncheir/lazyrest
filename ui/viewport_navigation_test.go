package ui

import (
	"fmt"
	"testing"

	"github.com/Hecatoncheir/lazyrest/ui/suite"
	"github.com/Hecatoncheir/lazyrest/ui/suites"
	uitree "github.com/Hecatoncheir/lazyrest/ui/tree"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestHalfPageNavigationMovesSuitesByHalfTheVisibleItems(t *testing.T) {
	application := newNavigationTestApplication()
	list := populatedList(20)
	list.SetRect(0, 0, 40, 10)
	application.Suites = &suites.Suites{Element: list}
	application.Element.SetFocus(list)
	handler := onInputCallback(application)

	if returned := handler(tcell.NewEventKey(tcell.KeyCtrlD, 0, tcell.ModCtrl)); returned != nil {
		t.Fatal("Ctrl+d was not consumed")
	}
	if current := list.GetCurrentItem(); current != 2 {
		t.Fatalf("Ctrl+d moved to item %d, want 2", current)
	}
	if offset, _ := list.GetOffset(); offset != 2 {
		t.Fatalf("Ctrl+d scrolled to item offset %d, want 2", offset)
	}
	if returned := handler(tcell.NewEventKey(tcell.KeyCtrlU, 0, tcell.ModCtrl)); returned != nil {
		t.Fatal("Ctrl+u was not consumed")
	}
	if current := list.GetCurrentItem(); current != 0 {
		t.Fatalf("Ctrl+u moved to item %d, want 0", current)
	}
	if offset, _ := list.GetOffset(); offset != 0 {
		t.Fatalf("Ctrl+u scrolled to item offset %d, want 0", offset)
	}
}

func TestHalfPageNavigationScrollsText(t *testing.T) {
	application := newNavigationTestApplication()
	view := tview.NewTextView()
	view.SetRect(0, 0, 40, 10)
	view.ScrollTo(20, 0)
	application.Suite = &suite.Suite{Element: view}
	application.Element.SetFocus(view)

	if returned := onInputCallback(application)(tcell.NewEventKey(tcell.KeyCtrlU, 0, tcell.ModCtrl)); returned != nil {
		t.Fatal("Ctrl+u was not consumed")
	}
	row, _ := view.GetScrollOffset()
	if row != 15 {
		t.Fatalf("Ctrl+u scrolled to row %d, want 15", row)
	}
}

func TestHalfPageNavigationMovesTreeSelection(t *testing.T) {
	application := newNavigationTestApplication()
	treeView, children := populatedTree(20)
	treeView.SetRect(0, 0, 40, 10)
	application.HttpFilesTree = &uitree.Tree{Element: treeView}
	application.Element.SetFocus(treeView)

	if returned := onInputCallback(application)(tcell.NewEventKey(tcell.KeyCtrlD, 0, tcell.ModCtrl)); returned != nil {
		t.Fatal("Ctrl+d was not consumed")
	}
	if current := treeView.GetCurrentNode(); current != children[4] {
		t.Fatalf("Ctrl+d selected %q, want %q", current.GetText(), children[4].GetText())
	}
}

func TestZZCentersCurrentListItem(t *testing.T) {
	application := newNavigationTestApplication()
	list := populatedList(30).ShowSecondaryText(false)
	list.SetRect(0, 0, 40, 10)
	list.SetCurrentItem(12)
	application.Element.SetFocus(list)
	handler := onInputCallback(application)
	z := tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModNone)

	if returned := handler(z); returned != nil {
		t.Fatal("first z was not consumed")
	}
	if offset, _ := list.GetOffset(); offset != 0 {
		t.Fatalf("first z changed the list offset to %d", offset)
	}
	if returned := handler(z); returned != nil {
		t.Fatal("second z was not consumed")
	}
	if offset, _ := list.GetOffset(); offset != 7 {
		t.Fatalf("zz centered with offset %d, want 7", offset)
	}
	if current := list.GetCurrentItem(); current != 12 {
		t.Fatalf("zz changed the current item to %d", current)
	}
}

func TestZZCentersTreeWithoutChangingSelection(t *testing.T) {
	application := newNavigationTestApplication()
	treeView, children := populatedTree(30)
	treeView.SetRect(0, 0, 40, 10)
	treeView.SetCurrentNode(children[20])
	application.HttpFilesTree = &uitree.Tree{Element: treeView}
	application.Element.SetFocus(treeView)

	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(40, 10)
	treeView.Draw(screen)

	handler := onInputCallback(application)
	z := tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModNone)
	handler(z)
	handler(z)
	treeView.Draw(screen)

	if current := treeView.GetCurrentNode(); current != children[20] {
		t.Fatalf("zz changed the current tree node to %q", current.GetText())
	}
	if offset := treeView.GetScrollOffset(); offset != 16 {
		t.Fatalf("zz centered the tree with offset %d, want 16", offset)
	}
}

func TestZZCentersTextScrollAnchor(t *testing.T) {
	application := newNavigationTestApplication()
	view := tview.NewTextView()
	view.SetRect(0, 0, 40, 10)
	view.ScrollTo(20, 0)
	application.Element.SetFocus(view)
	handler := onInputCallback(application)
	z := tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModNone)

	handler(z)
	handler(z)
	row, _ := view.GetScrollOffset()
	if row != 15 {
		t.Fatalf("zz centered the scroll anchor at row %d, want 15", row)
	}
}

func TestViewportSequenceMismatchProcessesTheCurrentKey(t *testing.T) {
	application := newNavigationTestApplication()
	list := populatedList(10).ShowSecondaryText(false)
	application.Element.SetFocus(list)
	handler := onInputCallback(application)

	z := tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModNone)
	if returned := handler(z); returned != nil {
		t.Fatal("sequence prefix was not consumed")
	}
	j := tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone)
	if returned := handler(j); returned != j {
		t.Fatal("key after an incomplete sequence was not passed through")
	}
	if len(application.pendingViewKeys) != 0 {
		t.Fatal("incomplete sequence was not cleared")
	}
}

func TestViewportSequenceDoesNotCaptureInputFields(t *testing.T) {
	application := newNavigationTestApplication()
	input := tview.NewInputField()
	application.Element.SetFocus(input)
	z := tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModNone)

	if returned := onInputCallback(application)(z); returned != z {
		t.Fatal("z was captured while an input field was focused")
	}
}

func populatedList(count int) *tview.List {
	list := tview.NewList()
	for index := 0; index < count; index++ {
		list.AddItem(fmt.Sprintf("item %d", index), "preview", 0, nil)
	}
	return list
}

func populatedTree(count int) (*tview.TreeView, []*tview.TreeNode) {
	root := tview.NewTreeNode("root").SetExpanded(true)
	children := make([]*tview.TreeNode, 0, count)
	for index := 0; index < count; index++ {
		child := tview.NewTreeNode(fmt.Sprintf("node %d", index))
		children = append(children, child)
		root.AddChild(child)
	}
	return tview.NewTreeView().SetRoot(root).SetCurrentNode(root), children
}
