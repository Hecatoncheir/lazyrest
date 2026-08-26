package ui

import (
	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (application *Application) handleViewportInput(event *tcell.EventKey) bool {
	if application == nil || application.Element == nil || event == nil {
		return false
	}
	focused := application.Element.GetFocus()
	if !isScrollablePrimitive(focused) {
		application.resetViewportSequence()
		return false
	}

	bindings := application.config.Keybindings
	if bindings == nil {
		bindings = keymap.Default()
	}
	if bindings.Matches(keymap.HalfPageDown, event) {
		application.resetViewportSequence()
		moveHalfPage(application, focused, 1)
		return true
	}
	if bindings.Matches(keymap.HalfPageUp, event) {
		application.resetViewportSequence()
		moveHalfPage(application, focused, -1)
		return true
	}

	if application.pendingViewFocus != focused {
		application.resetViewportSequence()
	}
	sequence := append(append([]*tcell.EventKey(nil), application.pendingViewKeys...), event)
	switch bindings.MatchesSequence(keymap.CenterView, sequence) {
	case keymap.SequenceFull:
		application.resetViewportSequence()
		centerViewport(application, focused)
		return true
	case keymap.SequencePrefix:
		application.pendingViewKeys = sequence
		application.pendingViewFocus = focused
		return true
	}

	if len(application.pendingViewKeys) > 0 {
		application.resetViewportSequence()
		switch bindings.MatchesSequence(keymap.CenterView, []*tcell.EventKey{event}) {
		case keymap.SequenceFull:
			centerViewport(application, focused)
			return true
		case keymap.SequencePrefix:
			application.pendingViewKeys = []*tcell.EventKey{event}
			application.pendingViewFocus = focused
			return true
		}
	}
	return false
}

func (application *Application) resetViewportSequence() {
	application.pendingViewKeys = nil
	application.pendingViewFocus = nil
}

func isScrollablePrimitive(primitive tview.Primitive) bool {
	switch primitive.(type) {
	case *tview.TreeView, *tview.List, *tview.TextView:
		return true
	default:
		return false
	}
}

func moveHalfPage(application *Application, primitive tview.Primitive, direction int) {
	step := halfPageSize(application, primitive)
	switch element := primitive.(type) {
	case *tview.TreeView:
		element.Move(direction * step)
	case *tview.List:
		itemOffset, horizontal := element.GetOffset()
		element.SetCurrentItem(element.GetCurrentItem() + direction*step)
		itemOffset += direction * step
		if itemOffset < 0 {
			itemOffset = 0
		}
		maximumOffset := element.GetItemCount() - pageSize(application, primitive)
		if maximumOffset < 0 {
			maximumOffset = 0
		}
		if itemOffset > maximumOffset {
			itemOffset = maximumOffset
		}
		element.SetOffset(itemOffset, horizontal)
	case *tview.TextView:
		row, column := element.GetScrollOffset()
		row += direction * step
		if row < 0 {
			row = 0
		}
		element.ScrollTo(row, column)
	}
}

func centerViewport(application *Application, primitive tview.Primitive) {
	step := halfPageSize(application, primitive)
	switch element := primitive.(type) {
	case *tview.TreeView:
		centerTreeView(element, step)
	case *tview.List:
		_, horizontal := element.GetOffset()
		offset := element.GetCurrentItem() - step
		if offset < 0 {
			offset = 0
		}
		element.SetOffset(offset, horizontal)
	case *tview.TextView:
		row, column := element.GetScrollOffset()
		row -= step
		if row < 0 {
			row = 0
		}
		element.ScrollTo(row, column)
	}
}

func centerTreeView(element *tview.TreeView, halfPage int) {
	current := element.GetCurrentNode()
	if current == nil || halfPage < 1 {
		return
	}
	offset := element.GetScrollOffset()
	element.Move(-halfPage)
	if element.GetScrollOffset() == offset {
		element.SetCurrentNode(current)
		_, _, _, height := element.GetInnerRect()
		down := height - halfPage - 1
		if down > 0 {
			element.Move(down)
		}
	}
	// SetCurrentNode defers restoring the selection until the next draw. The
	// temporary move changes only the viewport; it never selects or opens a node.
	element.SetCurrentNode(current)
}

func halfPageSize(application *Application, primitive tview.Primitive) int {
	step := pageSize(application, primitive) / 2
	if step < 1 {
		return 1
	}
	return step
}

func pageSize(application *Application, primitive tview.Primitive) int {
	_, _, _, height := primitive.GetRect()
	if box, ok := primitive.(interface{ GetInnerRect() (int, int, int, int) }); ok {
		_, _, _, height = box.GetInnerRect()
	}
	if list, ok := primitive.(*tview.List); ok && application != nil && application.Suites != nil && list == application.Suites.Element {
		height /= 2 // Suite rows contain a title and a body preview.
	}
	if height < 1 {
		return 1
	}
	return height
}
