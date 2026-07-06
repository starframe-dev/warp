package warp

import "fmt"

// ExampleNewTabGroup shows how to create a TabGroup as a local Panel component.
func ExampleNewTabGroup() {
	tg := NewTabGroup(TabTop)
	tg.NewTab("editor")
	tg.NewTab("terminal")

	fmt.Println("tabs:", len(tg.tabs))
	fmt.Println("active:", tg.ActiveTab().name)

	// Output:
	// tabs: 3
	// active: terminal
}

// ExampleWordWrap demonstrates text wrapping at word boundaries.
func ExampleWordWrap() {
	lines := WordWrap("hello world this is a test", 10)
	for _, line := range lines {
		fmt.Println(line)
	}

	// Output:
	// hello
	// world this
	// is a test
}

// ExampleSpaceWrap demonstrates text wrapping only at spaces.
func ExampleSpaceWrap() {
	lines := SpaceWrap("hello world this is a test", 10)
	for _, line := range lines {
		fmt.Println(line)
	}

	// Output:
	// hello
	// world this
	// is a test
}

// ExampleNewCollapsible shows a collapsible panel wrapper.
func ExampleNewCollapsible() {
	content := &testPanel{name: "content"}
	col := NewCollapsible("Section", content)

	fmt.Println("collapsed:", col.Collapsed)
	col.Toggle()
	fmt.Println("collapsed:", col.Collapsed)

	// Output:
	// collapsed: false
	// collapsed: true
}

// ExampleNewSelectable shows text selection bounds.
func ExampleNewSelectable() {
	content := &testPanel{name: "hello world"}
	sel := NewSelectable(content)
	sel.SelectAll(20, 5)

	fmt.Println("has selection:", sel.HasSelection)
	fmt.Println("selected text:", sel.SelectedText())

	// Output:
	// has selection: true
	// selected text: hello world
}
