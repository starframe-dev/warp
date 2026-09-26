package warp

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type performancePanel struct {
	content  string
	elements []Element
	height   int
	commands bool
}

var performanceNoopCommand tea.Cmd = func() tea.Msg { return nil }

func (p *performancePanel) View(_, height int) string {
	if p.height > 0 {
		return strings.TrimSuffix(strings.Repeat("row\n", max(0, height)), "\n")
	}
	return p.content
}

func (p *performancePanel) Update(tea.Msg) tea.Cmd {
	if p.commands {
		return performanceNoopCommand
	}
	return nil
}

func (p *performancePanel) Elements(_, _ int) []Element { return p.elements }

func (p *performancePanel) ContentHeight(int) (int, bool) {
	return p.height, p.height > 0
}

type performanceMessage struct{}

type performanceRowsPanel struct {
	rows int
}

func (p *performanceRowsPanel) View(_, height int) string {
	return strings.TrimSuffix(strings.Repeat("row\n", min(p.rows, max(0, height))), "\n")
}

func (*performanceRowsPanel) Update(tea.Msg) tea.Cmd { return nil }

func (p *performanceRowsPanel) Elements(_, height int) []Element {
	count := min(p.rows, max(0, height))
	elements := make([]Element, count)
	for i := range elements {
		elements[i] = Element{Role: "row", Name: "row", Bounds: Bounds{Y: i, W: 3, H: 1}}
	}
	return elements
}

func (p *performanceRowsPanel) ContentHeight(int) (int, bool) {
	return p.rows, true
}

type performanceViewportRowsPanel struct {
	performanceRowsPanel
}

func (p *performanceViewportRowsPanel) ViewAt(_, height, offset int) string {
	start := max(0, offset)
	end := min(p.rows, saturatingAddNonNegative(start, max(0, height)))
	return strings.TrimSuffix(strings.Repeat("row\n", max(0, end-start)), "\n")
}

func (p *performanceViewportRowsPanel) ElementsAt(_, height, offset int) []Element {
	start := min(p.rows, max(0, offset))
	end := min(p.rows, saturatingAddNonNegative(start, max(0, height)))
	elements := make([]Element, max(0, end-start))
	for i := range elements {
		elements[i] = Element{Role: "row", Name: "row", Bounds: Bounds{Y: start + i, W: 3, H: 1}}
	}
	return elements
}

func benchmarkElements(count int) []Element {
	elements := make([]Element, count)
	for i := range elements {
		elements[i] = Element{
			Role:   "item",
			Name:   fmt.Sprintf("item-%d", i),
			Bounds: Bounds{X: i % 80, Y: i % 24, W: 1, H: 1},
		}
	}
	return []Element{{Role: "root", Name: "root", Bounds: Bounds{W: 80, H: 24}, Children: elements}}
}

func benchmarkSplitTab(leaves int) *Tab {
	tab := NewTab("bench")
	parent := tab.RootPanel()
	for i := 1; i < leaves; i++ {
		panel := &performancePanel{content: "content"}
		tab.SplitVertical(parent, 0.5, panel)
		parent = panel
	}
	return tab
}

func benchmarkFlexTab(items int) *Tab {
	tab := NewTab("bench")
	specs := make([]FlexItemSpec, items)
	for i := range specs {
		specs[i] = FlexItemSpec{Panel: &performancePanel{content: "content"}, Grow: 1}
	}
	tab.FlexRow(tab.RootPanel(), specs)
	return tab
}

func benchmarkTabGroup(tabCount, panelsPerTab int) *TabGroup {
	group := NewTabGroup(TabTop)
	for i := 0; i < tabCount; i++ {
		tab := group.ActiveTab()
		if i > 0 {
			tab = group.NewTab(fmt.Sprintf("tab-%d", i))
		}
		tab.SetRootPanel(&performancePanel{content: "content", commands: true})
		for j := 1; j < panelsPerTab; j++ {
			parent := tab.RootPanel()
			tab.SplitVertical(parent, 0.5, &performancePanel{content: "content", commands: true})
		}
	}
	return group
}

func benchmarkOverlayLines() []string {
	lines := make([]string, 24)
	for i := range lines {
		lines[i] = strings.Repeat("x", 80)
	}
	return lines
}

func BenchmarkRenderLayouts(b *testing.B) {
	for _, size := range []int{4, 16} {
		b.Run(fmt.Sprintf("nested-splits-%d", size), func(b *testing.B) {
			tab := benchmarkSplitTab(size)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = tab.View(120, 40)
			}
		})
		b.Run(fmt.Sprintf("flex-%d", size), func(b *testing.B) {
			tab := benchmarkFlexTab(size)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = tab.View(120, 40)
			}
		})
	}
}

func BenchmarkLayoutAndCollectors(b *testing.B) {
	tab := benchmarkSplitTab(32)
	bounds := layoutRect{w: 160, h: 48}
	b.Run("layout", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = newLayout(tab.root, bounds)
		}
	})
	layout := newLayout(tab.root, bounds)
	b.Run("borders", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = collectLayoutBorders(layout)
		}
	})
	b.Run("elements", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = elementsFromLayout(layout)
		}
	})
	b.Run("hit-test", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = findLayoutPanel(layout, 80, 20)
		}
	})
}

func BenchmarkMouseDrag(b *testing.B) {
	b.Run("split", func(b *testing.B) {
		tab := NewTab("drag")
		tab.SplitVertical(tab.RootPanel(), 0.5, &performancePanel{content: "right"})
		tab.View(120, 40)
		tab.dragging = tab.root.Split
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tab.updateDrag(55+i%2, 20, 120, 40)
		}
	})
	b.Run("flex", func(b *testing.B) {
		tab := benchmarkFlexTab(16)
		tab.View(160, 40)
		tab.flexDragging = tab.root.Flex
		tab.flexDragIdx = 0
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tab.updateDrag(55+i%2, 20, 160, 40)
		}
	})
	b.Run("handle-flex", func(b *testing.B) {
		tab := benchmarkFlexTab(16)
		tab.View(160, 40)
		tab.flexDragging = tab.root.Flex
		tab.flexDragIdx = 0
		msg := tea.MouseMsg{X: 55, Y: 20, Action: tea.MouseActionMotion, Button: tea.MouseButtonLeft}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			msg.X = 55 + i%2
			tab.handleMouse(msg, 0, 0, 160, 40)
		}
	})
}

func BenchmarkMouseHitTest(b *testing.B) {
	tab := benchmarkSplitTab(32)
	const width, height = 160, 48
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tab.handleMouse(tea.MouseMsg{
			X:      120,
			Y:      24,
			Action: tea.MouseActionMotion,
			Button: tea.MouseButtonLeft,
		}, 0, 0, width, height)
	}
}

func BenchmarkTabGroup(b *testing.B) {
	for _, count := range []int{8, 32} {
		b.Run(fmt.Sprintf("tab-bar-%d", count), func(b *testing.B) {
			group := benchmarkTabGroup(count, 1)
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = group.renderTabBar(160)
			}
		})
		b.Run(fmt.Sprintf("tab-bar-vertical-%d", count), func(b *testing.B) {
			group := benchmarkTabGroup(count, 1)
			group.tabPosition = TabLeft
			group.height = 48
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = group.renderTabBar(160)
			}
		})
		b.Run(fmt.Sprintf("broadcast-%d-tabs", count), func(b *testing.B) {
			group := benchmarkTabGroup(count, 4)
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = group.Update(performanceMessage{})
			}
		})
		b.Run(fmt.Sprintf("broadcast-resize-%d-tabs", count), func(b *testing.B) {
			group := benchmarkTabGroup(count, 4)
			msg := tea.WindowSizeMsg{Width: 160, Height: 48}
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = group.Update(msg)
			}
		})
	}
}

func BenchmarkFloatRendering(b *testing.B) {
	for _, count := range []int{0, 8, 32} {
		b.Run(fmt.Sprintf("floats-%d", count), func(b *testing.B) {
			tab := NewTab("floats")
			tab.SetRootPanel(&performancePanel{content: "background"})
			for i := 0; i < count; i++ {
				tab.Float(&performancePanel{content: "float"}, i, i, 24, 8)
			}
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = tab.View(120, 40)
			}
		})
	}
}

func BenchmarkInspector(b *testing.B) {
	b.Run("update-without-http", func(b *testing.B) {
		w := New()
		w.SetRoot(&performancePanel{elements: benchmarkElements(128)})
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = w.Update(performanceMessage{})
		}
	})
	b.Run("update-with-http", func(b *testing.B) {
		w := New()
		w.SetRoot(&performancePanel{elements: benchmarkElements(128)})
		if err := w.ServeHTTP("127.0.0.1:0"); err != nil {
			b.Fatal(err)
		}
		defer w.CloseHTTP()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = w.Update(performanceMessage{})
		}
	})
	b.Run("elements-json", func(b *testing.B) {
		w := New()
		w.SetRoot(&performancePanel{elements: benchmarkElements(128)})
		if err := w.ServeHTTP("127.0.0.1:0"); err != nil {
			b.Fatal(err)
		}
		defer w.CloseHTTP()
		_, _ = w.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		request := httptest.NewRequest("GET", "/elements", nil)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			response := httptest.NewRecorder()
			w.handleElements(response, request)
		}
	})
}

func BenchmarkScrollableOffset(b *testing.B) {
	for _, offset := range []int{0, 100, 1000, 10000, 19990} {
		b.Run(fmt.Sprintf("view-%d", offset), func(b *testing.B) {
			scrollable := NewScrollable(&performanceRowsPanel{rows: 20000})
			scrollable.Offset = offset
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = scrollable.View(40, 10)
			}
		})
		b.Run(fmt.Sprintf("elements-%d", offset), func(b *testing.B) {
			scrollable := NewScrollable(&performanceRowsPanel{rows: 20000})
			scrollable.Offset = offset
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = scrollable.Elements(40, 10)
			}
		})
		b.Run(fmt.Sprintf("view-viewport-%d", offset), func(b *testing.B) {
			scrollable := NewScrollable(&performanceViewportRowsPanel{performanceRowsPanel{rows: 20000}})
			scrollable.Offset = offset
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = scrollable.View(40, 10)
			}
		})
		b.Run(fmt.Sprintf("elements-viewport-%d", offset), func(b *testing.B) {
			scrollable := NewScrollable(&performanceViewportRowsPanel{performanceRowsPanel{rows: 20000}})
			scrollable.Offset = offset
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = scrollable.Elements(40, 10)
			}
		})
	}
}

func BenchmarkInput(b *testing.B) {
	value := strings.Repeat("á漢🙂", 64)
	b.Run("render-graphemes", func(b *testing.B) {
		input := NewInput("> ")
		input.SetValue(value)
		input.SetCursor(len([]rune(value)) / 2)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = input.View(32, 1)
		}
	})
	b.Run("insert", func(b *testing.B) {
		input := NewInput("> ")
		input.Focus()
		key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			input.SetValue(value)
			input.SetCursor(len([]rune(value)) / 2)
			input.Update(key)
		}
	})
}

func BenchmarkSelectableView(b *testing.B) {
	content := strings.Repeat("selectable text line\n", 23) + "selectable text line"
	selectable := NewSelectable(&performancePanel{content: content})
	selectable.SelectAll(80, 24)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = selectable.View(80, 24)
	}
}

func BenchmarkOverlays(b *testing.B) {
	b.Run("modal", func(b *testing.B) {
		modal := NewModal("Title", "Modal content", []ModalButton{{Label: "OK"}, {Label: "Cancel"}}, nil)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			lines := benchmarkOverlayLines()
			_ = modal.Overlay(lines, 80, 24)
		}
	})
	b.Run("popover", func(b *testing.B) {
		popover := &Popover{X: 12, Y: 5, Items: []PopoverItem{{Name: "Open"}, {Name: "Rename"}, {Name: "Close"}}}
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			lines := benchmarkOverlayLines()
			_ = popover.Overlay(lines, 80, 24)
		}
	})
}
