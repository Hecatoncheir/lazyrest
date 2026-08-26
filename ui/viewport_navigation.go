package ui

import (
	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var viewportActions = []keymap.Action{
	keymap.HalfPageDown,
	keymap.HalfPageUp,
	keymap.PageDown,
	keymap.PageUp,
	keymap.GoToTop,
	keymap.GoToBottom,
	keymap.AlignTop,
	keymap.CenterView,
	keymap.AlignBottom,
}

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
	if application.pendingViewFocus != focused {
		application.resetViewportSequence()
	}
	sequence := append(append([]*tcell.EventKey(nil), application.pendingViewKeys...), event)
	action, match := matchViewportAction(bindings, sequence)
	switch match {
	case keymap.SequenceFull:
		application.resetViewportSequence()
		applyViewportAction(application, focused, action)
		return true
	case keymap.SequencePrefix:
		application.pendingViewKeys = sequence
		application.pendingViewFocus = focused
		return true
	}

	if len(application.pendingViewKeys) > 0 {
		application.resetViewportSequence()
		action, match = matchViewportAction(bindings, []*tcell.EventKey{event})
		switch match {
		case keymap.SequenceFull:
			applyViewportAction(application, focused, action)
			return true
		case keymap.SequencePrefix:
			application.pendingViewKeys = []*tcell.EventKey{event}
			application.pendingViewFocus = focused
			return true
		}
	}
	return false
}

func matchViewportAction(bindings *keymap.Bindings, events []*tcell.EventKey) (keymap.Action, keymap.SequenceMatch) {
	matchedPrefix := false
	for _, action := range viewportActions {
		switch bindings.MatchesSequence(action, events) {
		case keymap.SequenceFull:
			return action, keymap.SequenceFull
		case keymap.SequencePrefix:
			matchedPrefix = true
		}
	}
	if matchedPrefix {
		return "", keymap.SequencePrefix
	}
	return "", keymap.SequenceNoMatch
}

func applyViewportAction(application *Application, primitive tview.Primitive, action keymap.Action) {
	switch action {
	case keymap.HalfPageDown:
		moveViewport(application, primitive, 1, halfPageSize(application, primitive))
	case keymap.HalfPageUp:
		moveViewport(application, primitive, -1, halfPageSize(application, primitive))
	case keymap.PageDown:
		moveViewport(application, primitive, 1, pageSize(application, primitive))
	case keymap.PageUp:
		moveViewport(application, primitive, -1, pageSize(application, primitive))
	case keymap.GoToTop:
		moveToBoundary(application, primitive, false)
	case keymap.GoToBottom:
		moveToBoundary(application, primitive, true)
	case keymap.AlignTop:
		alignViewport(application, primitive, alignTop)
	case keymap.CenterView:
		alignViewport(application, primitive, alignCenter)
	case keymap.AlignBottom:
		alignViewport(application, primitive, alignBottom)
	}
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

func moveViewport(application *Application, primitive tview.Primitive, direction, step int) {
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

func moveToBoundary(application *Application, primitive tview.Primitive, bottom bool) {
	switch element := primitive.(type) {
	case *tview.TreeView:
		step := -element.GetRowCount()
		if bottom {
			step = element.GetRowCount()
		}
		element.Move(step)
	case *tview.List:
		_, horizontal := element.GetOffset()
		if !bottom {
			element.SetCurrentItem(0).SetOffset(0, horizontal)
			return
		}
		last := element.GetItemCount() - 1
		offset := element.GetItemCount() - pageSize(application, primitive)
		if offset < 0 {
			offset = 0
		}
		element.SetCurrentItem(last).SetOffset(offset, horizontal)
	case *tview.TextView:
		if bottom {
			element.ScrollToEnd()
		} else {
			element.ScrollToBeginning()
		}
	}
}

type viewportAlignment uint8

const (
	alignTop viewportAlignment = iota
	alignCenter
	alignBottom
)

func alignViewport(application *Application, primitive tview.Primitive, alignment viewportAlignment) {
	shift := alignmentShift(application, primitive, alignment)
	switch element := primitive.(type) {
	case *tview.TreeView:
		alignTreeView(element, alignment)
	case *tview.List:
		_, horizontal := element.GetOffset()
		offset := element.GetCurrentItem() - shift
		if offset < 0 {
			offset = 0
		}
		element.SetOffset(offset, horizontal)
	case *tview.TextView:
		row, column := element.GetScrollOffset()
		row -= shift
		if row < 0 {
			row = 0
		}
		element.ScrollTo(row, column)
	}
}

func alignTreeView(element *tview.TreeView, alignment viewportAlignment) {
	current := element.GetCurrentNode()
	if current == nil {
		return
	}
	_, _, _, height := element.GetInnerRect()
	if height < 2 {
		return
	}
	switch alignment {
	case alignTop:
		element.Move(height - 1)
	case alignCenter:
		centerTreeView(element, current, height/2)
	case alignBottom:
		element.Move(-(height - 1))
	}
	// SetCurrentNode defers restoring the selection until the next draw. The
	// temporary move changes only the viewport; it never opens a node.
	element.SetCurrentNode(current)
}

func centerTreeView(element *tview.TreeView, current *tview.TreeNode, halfPage int) {
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
}

func alignmentShift(application *Application, primitive tview.Primitive, alignment viewportAlignment) int {
	switch alignment {
	case alignCenter:
		return halfPageSize(application, primitive)
	case alignBottom:
		return pageSize(application, primitive) - 1
	default:
		return 0
	}
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
