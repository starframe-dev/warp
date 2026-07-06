package warp

// Focusable is a Panel that can receive keyboard focus.
type Focusable interface {
	Panel
	Focus()
	Blur()
	Focused() bool
}

// isFocusable reports whether a panel implements Focusable.
func isFocusable(panel Panel) (Focusable, bool) {
	if panel == nil {
		return nil, false
	}
	f, ok := panel.(Focusable)
	return f, ok
}

// collectFocusables returns all focusable panels in the tree, in visual order.
func collectFocusables(node *Node) []Focusable {
	if node == nil {
		return nil
	}
	if node.IsLeaf() {
		if f, ok := isFocusable(node.Panel); ok {
			return []Focusable{f}
		}
		return nil
	}
	var result []Focusable
	if node.Split != nil {
		result = append(result, collectFocusables(node.Split.First)...)
		result = append(result, collectFocusables(node.Split.Second)...)
	}
	if node.Flex != nil {
		for _, item := range node.Flex.Items {
			result = append(result, collectFocusables(item.Node)...)
		}
	}
	return result
}

// focusIndex returns the index of the currently focused panel in the list.
func focusIndex(list []Focusable, current Panel) int {
	if current == nil {
		return -1
	}
	for i, f := range list {
		if f == current {
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
	next := (idx + 1) % len(list)
	return list[next]
}

// focusPrev switches focus to the previous focusable panel.
func focusPrev(list []Focusable, current Panel) Focusable {
	if len(list) == 0 {
		return nil
	}
	idx := focusIndex(list, current)
	prev := idx - 1
	if prev < 0 {
		prev = len(list) - 1
	}
	return list[prev]
}

// applyFocus moves focus from current to next, calling Blur/Focus.
func applyFocus(current, next Focusable) {
	if current != nil && current != next {
		current.Blur()
	}
	if next != nil {
		next.Focus()
	}
}

// RawKeyReceiver is a panel that wants to receive all keyboard input
// without interception (e.g., a terminal emulator running inside a PTY).
type RawKeyReceiver interface {
	Panel
	WantsRawKeys() bool
}
