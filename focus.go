package warp

import "reflect"

// Focusable is a Panel that can receive keyboard focus.
type Focusable interface {
	Panel
	Focus()
	Blur()
	Focused() bool
}

// isFocusable reports whether a non-nil panel implements Focusable.
func isFocusable(panel Panel) (Focusable, bool) {
	if isNilPanel(panel) {
		return nil, false
	}
	f, ok := panel.(Focusable)
	return f, ok
}

func isNilPanel(panel Panel) bool {
	if panel == nil {
		return true
	}
	value := reflect.ValueOf(panel)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func samePanel(a, b Panel) bool {
	aNil := isNilPanel(a)
	bNil := isNilPanel(b)
	if aNil || bNil {
		return aNil && bNil
	}
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}
	if reflect.TypeOf(a).Comparable() {
		return a == b
	}
	return reflect.DeepEqual(a, b)
}

// collectFocusables returns all focusable panels in the tree, in visual order.
func collectFocusables(node *Node) []Focusable {
	var result []Focusable
	appendFocusables(&result, node)
	return result
}

func appendFocusables(result *[]Focusable, node *Node) {
	if node == nil {
		return
	}
	if node.IsLeaf() {
		if focusable, ok := isFocusable(node.Panel); ok {
			*result = append(*result, focusable)
		}
		return
	}
	if node.Split != nil {
		appendFocusables(result, node.Split.First)
		appendFocusables(result, node.Split.Second)
	}
	if node.Flex != nil {
		for _, item := range node.Flex.Items {
			if item != nil {
				appendFocusables(result, item.Node)
			}
		}
	}
}

// focusIndex returns the index of the currently focused panel in the list.
func focusIndex(list []Focusable, current Panel) int {
	if isNilPanel(current) {
		return -1
	}
	for i, f := range list {
		if samePanel(f, current) {
			return i
		}
	}
	return -1
}

// focusNext switches focus to the next focusable panel.
func focusNext(list []Focusable, current Panel) Focusable {
	if len(list) == 0 {
		return nil
	}
	idx := focusIndex(list, current)
	return list[(idx+1)%len(list)]
}

// focusPrev switches focus to the previous focusable panel.
func focusPrev(list []Focusable, current Panel) Focusable {
	if len(list) == 0 {
		return nil
	}
	idx := focusIndex(list, current) - 1
	if idx < 0 {
		idx = len(list) - 1
	}
	return list[idx]
}

// applyFocus moves focus from current to next, calling Blur and Focus once.
func applyFocus(current, next Focusable) {
	if samePanel(current, next) {
		return
	}
	if current != nil {
		current.Blur()
	}
	if next != nil {
		next.Focus()
	}
}

// RawKeyReceiver asks TabGroup to forward every key to the focused panel
// before handling Warp shortcuts. In particular, Ctrl+C is not treated as quit
// while a focused receiver reports WantsRawKeys() == true.
type RawKeyReceiver interface {
	Panel
	WantsRawKeys() bool
}
