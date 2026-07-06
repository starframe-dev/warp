package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/starframe-dev/warp"
)

// --- appRoot wraps the TabGroup and lets the developer decide focus keys ---

type appRoot struct {
	tg *warp.TabGroup
}

func (a *appRoot) View(w, h int) string {
	return a.tg.View(w, h)
}

func (a *appRoot) Update(msg tea.Msg) tea.Cmd {
	// The developer decides how to switch focus. Here we use Tab/Shift+Tab
	// in this app, but warp itself does not impose any focus keys.
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "tab":
			if tab := a.tg.ActiveTab(); tab != nil {
				tab.FocusNext()
			}
			return nil
		case "shift+tab":
			if tab := a.tg.ActiveTab(); tab != nil {
				tab.FocusPrev()
			}
			return nil
		}
	}
	return a.tg.Update(msg)
}

// --- demoPanel — basic clickable panel ---

type demoPanel struct {
	name  string
	count int
}

func (p *demoPanel) View(w, h int) string {
	lines := make([]string, h)
	for i := range lines {
		switch {
		case i == 0:
			lines[i] = padRight(p.name, w)
		case i == h-1:
			lines[i] = padRight(fmt.Sprintf("clicks: %d", p.count), w)
		default:
			lines[i] = strings.Repeat("·", w)
		}
	}
	return strings.Join(lines, "\n")
}

func (p *demoPanel) Update(msg tea.Msg) tea.Cmd {
	if _, ok := msg.(tea.MouseMsg); ok {
		p.count++
	}
	return nil
}

func padRight(s string, w int) string {
	if len(s) >= w {
		return s[:w]
	}
	return s + strings.Repeat(" ", w-len(s))
}

// --- textPanel — scrollable wrapped text ---

type textPanel struct {
	lines []string
}

func newTextPanel(text string, wrapW int) *textPanel {
	wrapped := warp.WordWrap(text, wrapW)
	return &textPanel{lines: wrapped}
}

func (p *textPanel) View(w, h int) string {
	result := make([]string, h)
	for i := 0; i < h; i++ {
		if i < len(p.lines) {
			line := p.lines[i]
			if len(line) > w {
				line = line[:w]
			}
			result[i] = line + strings.Repeat(" ", w-len(line))
		} else {
			result[i] = strings.Repeat(" ", w)
		}
	}
	return strings.Join(result, "\n")
}

func (p *textPanel) Update(msg tea.Msg) tea.Cmd {
	return nil
}

// --- statusPanel ---

type statusPanel struct {
	msg string
}

func (p *statusPanel) View(w, h int) string {
	line := padRight(p.msg, w)
	lines := make([]string, h)
	for i := range lines {
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

func (p *statusPanel) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (p *statusPanel) Set(msg string) {
	p.msg = msg
}

func main() {
	w := warp.New()

	// ═══════════════════════════════════════════════════
	// TAB 1: Inputs + flex layout
	// ═══════════════════════════════════════════════════
	tg := warp.NewTabGroup(warp.TabTop)
	tab1 := tg.ActiveTab()

	nameInput := warp.NewInput("Name: ")
	nameInput.SetValue("Warp")
	nameInput.Focus()

	emailInput := warp.NewInput("Email: ")

	searchInput := warp.NewInput("Search: ")
	searchInput.SetValue("flex layout")

	// Flex row with 3 inputs and a text panel
	tab1.FlexRow(tab1.RootPanel(), []warp.FlexItemSpec{
		{Panel: nameInput, Grow: 1},
		{Panel: emailInput, Grow: 1},
		{Panel: searchInput, Grow: 1},
		{Panel: &demoPanel{name: "Preview"}, Grow: 2},
	})

	// ═══════════════════════════════════════════════════
	// TAB 2: Local TabGroup inside flex (tabs are local components)
	// ═══════════════════════════════════════════════════
	tab2 := tg.NewTab("local-tabs")

	localTabs := warp.NewTabGroup(warp.TabLeft)
	localTabs.NewTab("local-a")
	localTabs.NewTab("local-b")
	localTabs.ActiveTab().SplitVertical(localTabs.ActiveTab().RootPanel(), 0.5, &demoPanel{name: "Local Right"})

	longText := "This demonstrates that TabGroup is just a Panel component. " +
		"You can embed it inside splits, flex layouts, or anywhere else. " +
		"It is NOT global — each TabGroup manages its own set of tabs independently. " +
		"The root warp has its own tabs, and this local TabGroup has its own."
	textContent := newTextPanel(longText, 40)
	selectableText := warp.NewSelectable(textContent)
	scroll := warp.NewScrollable(selectableText)

	tab2.FlexRow(tab2.RootPanel(), []warp.FlexItemSpec{
		{Panel: localTabs, Grow: 2},
		{Panel: scroll, Grow: 3},
	})

	// ═══════════════════════════════════════════════════
	// TAB 3: Columns + floats
	// ═══════════════════════════════════════════════════
	tab3 := tg.NewTab("columns")
	output := warp.NewCollapsible("Build Output", &demoPanel{name: "Compiler output here"})
	bottom := &demoPanel{name: "Bottom Panel"}
	tab3.FlexColumn(tab3.RootPanel(), []warp.FlexItemSpec{
		{Panel: output, Grow: 1},
		{Panel: bottom, Grow: 1},
	})
	tab3.Float(&demoPanel{name: "Column Float"}, 5, 3, 22, 5)

	// ═══════════════════════════════════════════════════
	// TAB 4: Splits + floats
	// ═══════════════════════════════════════════════════
	tab4 := tg.NewTab("splits")
	topRight := &demoPanel{name: "Top Right"}
	tab4.SplitVertical(tab4.RootPanel(), 0.5, topRight)
	tab4.SplitHorizontal(topRight, 0.5, &demoPanel{name: "Bottom"})
	tab4.Float(&demoPanel{name: "Split Float"}, 15, 5, 20, 5)

	// Replace Warp's default root with our custom appRoot that intercepts Tab.
	w.SetRoot(&appRoot{tg: tg})

	if err := w.Run(); err != nil {
		panic(err)
	}
}
