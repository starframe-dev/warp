package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/starframe-dev/warp"
)

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

func (p *statusPanel) Update(msg tea.Msg) tea.Cmd { return nil }

func main() {
	tg := warp.NewTabGroup(warp.TabTop)
	tab := tg.NewTab("demo")

	tab.SplitVertical(tab.RootPanel(), 0.5, &demoPanel{name: "Left"})
	tab.SplitHorizontal(tab.RootPanel(), 0.5, &demoPanel{name: "Bottom at left"})
	tab.SetRootPanel(&demoPanel{name: "Right"})

	w := warp.New()
	w.SetRoot(tg)

	if err := w.Run(); err != nil {
		panic(err)
	}
}