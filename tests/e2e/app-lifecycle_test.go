package e2e_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	warp "github.com/starframe-dev/warp"
)

type lifecycleRenderPanel struct {
	renders chan [2]int
}

func (p *lifecycleRenderPanel) View(width, height int) string {
	select {
	case p.renders <- [2]int{width, height}:
	default:
	}
	return strings.Repeat("panel\n", height)
}

func (*lifecycleRenderPanel) Update(tea.Msg) tea.Cmd { return nil }

func TestAppLifecycleWindowSizeAndCleanShutdown(t *testing.T) {
	const width, height = 48, 16

	panel := &lifecycleRenderPanel{renders: make(chan [2]int, 16)}
	app := warp.New()
	app.ActiveTab().SetRootPanel(panel)

	var output bytes.Buffer
	program := tea.NewProgram(app,
		tea.WithInput(strings.NewReader("")),
		tea.WithOutput(&output),
	)
	done := make(chan error, 1)
	go func() {
		_, err := program.Run()
		done <- err
	}()

	program.Send(tea.WindowSizeMsg{Width: width, Height: height})

	// The tab bar consumes one row; the active panel receives the remaining
	// content area while the Warp itself retains the requested terminal size.
	deadline := time.After(3 * time.Second)
	for {
		select {
		case size := <-panel.renders:
			if app.Width() != width || app.Height() != height {
				t.Fatalf("app size=(%d,%d), want (%d,%d)", app.Width(), app.Height(), width, height)
			}
			if size != [2]int{width, height - 1} {
				t.Fatalf("active panel rendered at %dx%d, want %dx%d", size[0], size[1], width, height-1)
			}
			goto rendered
		case <-deadline:
			t.Fatal("active panel was not rendered after the window-size event")
		}
	}

rendered:
	program.Quit()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("app exited with error: %v", err)
		}
	case <-time.After(3 * time.Second):
		program.Kill()
		t.Fatal("app did not shut down cleanly")
	}

	if !strings.Contains(output.String(), "main") {
		t.Fatalf("initial tab label was not rendered; output=%q", output.String())
	}
	if !strings.Contains(output.String(), "panel") {
		t.Fatalf("initial tab panel content was not rendered; output=%q", output.String())
	}
	if err := app.Close(); err != nil {
		t.Fatalf("close Warp app: %v", err)
	}
	if err := app.Close(); err != nil {
		t.Fatalf("repeated Warp close: %v", err)
	}
}
