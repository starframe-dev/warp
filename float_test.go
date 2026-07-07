package warp

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type recordingPanel struct {
	BasePanel
	last *tea.MouseMsg
	cmd  tea.Cmd
}

func (r *recordingPanel) Update(msg tea.Msg) tea.Cmd {
	if m, ok := msg.(tea.MouseMsg); ok {
		m2 := m
		r.last = &m2
	}
	return r.cmd
}

func TestFloatPaneRender(t *testing.T) {
	fp := &FloatPane{
		Panel:  BasePanel{},
		X:      0,
		Y:      0,
		Width:  20,
		Height: 5,
		Title:  "Test",
	}
	lines := fp.render(100, 100)
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(lines))
	}

	top := StripANSI(lines[0])
	if lipgloss.Width(top) != 20 {
		t.Errorf("expected top border width 20, got %d (%q)", lipgloss.Width(top), top)
	}
	if !strings.Contains(top, "Test") {
		t.Errorf("expected top border to contain title, got %q", top)
	}
	if !strings.Contains(top, "×") {
		t.Errorf("expected top border to contain close button, got %q", top)
	}

	content := StripANSI(lines[1])
	if !strings.Contains(content, "│") {
		t.Errorf("expected content line to contain vertical border, got %q", content)
	}

	bottom := StripANSI(lines[4])
	if !strings.Contains(bottom, "╰") || !strings.Contains(bottom, "╯") {
		t.Errorf("expected bottom border, got %q", bottom)
	}
}

func TestFloatPaneRenderTitleTruncation(t *testing.T) {
	fp := &FloatPane{
		Panel:  BasePanel{},
		Width:  20,
		Height: 3,
		Title:  strings.Repeat("a", 30),
	}
	top := StripANSI(fp.render(0, 0)[0])
	if !strings.Contains(top, "...") {
		t.Errorf("expected truncated title to contain '...', got %q", top)
	}
	if strings.Contains(top, strings.Repeat("a", 20)) {
		t.Errorf("expected title to be truncated, got %q", top)
	}
}

func TestFloatPaneHandleMouseClose(t *testing.T) {
	fp := &FloatPane{
		Panel:  BasePanel{},
		X:      5,
		Y:      5,
		Width:  20,
		Height: 10,
	}
	mx := fp.X + fp.Width - 2
	my := fp.Y
	msg := tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft, X: mx, Y: my}
	fp.handleMouse(msg, mx, my)
	if !fp.CloseRequested {
		t.Error("expected CloseRequested to be true")
	}
}

func TestFloatPaneHandleMouseDrag(t *testing.T) {
	fp := &FloatPane{
		Panel:  BasePanel{},
		X:      5,
		Y:      5,
		Width:  20,
		Height: 10,
	}

	press := tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft, X: 10, Y: 5}
	fp.handleMouse(press, 10, 5)
	if !fp.dragging {
		t.Fatal("expected dragging to be true")
	}

	motion := tea.MouseMsg{Action: tea.MouseActionMotion, Button: tea.MouseButtonLeft, X: 15, Y: 8}
	fp.handleMouse(motion, 15, 8)
	if fp.X != 10 || fp.Y != 8 {
		t.Errorf("expected X=10 Y=8, got X=%d Y=%d", fp.X, fp.Y)
	}

	release := tea.MouseMsg{Action: tea.MouseActionRelease, Button: tea.MouseButtonLeft}
	fp.handleMouse(release, 15, 8)
	if fp.dragging {
		t.Error("expected dragging to be false after release")
	}
}

func TestFloatPaneHandleMouseDragClamp(t *testing.T) {
	fp := &FloatPane{
		Panel:  BasePanel{},
		X:      2,
		Y:      2,
		Width:  20,
		Height: 10,
	}
	fp.handleMouse(tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft, X: 5, Y: 2}, 5, 2)
	fp.handleMouse(tea.MouseMsg{Action: tea.MouseActionMotion, Button: tea.MouseButtonLeft, X: -10, Y: -10}, -10, -10)
	if fp.X != 0 || fp.Y != 0 {
		t.Errorf("expected X=0 Y=0, got X=%d Y=%d", fp.X, fp.Y)
	}
}

func TestFloatPaneHandleMouseResize(t *testing.T) {
	fp := &FloatPane{
		Panel:  BasePanel{},
		X:      5,
		Y:      5,
		Width:  20,
		Height: 10,
	}
	// Bottom-right corner resize
	press := tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft, X: fp.X + fp.Width - 1, Y: fp.Y + fp.Height - 1}
	fp.handleMouse(press, fp.X+fp.Width-1, fp.Y+fp.Height-1)
	if !fp.resizing || fp.resizeEdge != "se" {
		t.Fatalf("expected resizing on se edge, got resizing=%v edge=%q", fp.resizing, fp.resizeEdge)
	}

	motion := tea.MouseMsg{Action: tea.MouseActionMotion, Button: tea.MouseButtonLeft, X: 29, Y: 19}
	fp.handleMouse(motion, 29, 19)
	if fp.Width != 25 || fp.Height != 15 {
		t.Errorf("expected Width=25 Height=15, got Width=%d Height=%d", fp.Width, fp.Height)
	}

	release := tea.MouseMsg{Action: tea.MouseActionRelease, Button: tea.MouseButtonLeft}
	fp.handleMouse(release, 29, 19)
	if fp.resizing || fp.resizeEdge != "" {
		t.Error("expected resize state to be cleared after release")
	}
}

func TestFloatPaneHandleMouseInsidePanel(t *testing.T) {
	rec := &recordingPanel{cmd: func() tea.Msg { return nil }}
	fp := &FloatPane{
		Panel:  rec,
		X:      0,
		Y:      0,
		Width:  10,
		Height: 10,
	}
	press := tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft, X: 2, Y: 2}
	cmd := fp.handleMouse(press, 2, 2)
	if rec.last == nil {
		t.Fatal("expected panel Update to be called")
	}
	if rec.last.X != 1 || rec.last.Y != 1 {
		t.Errorf("expected inner coords (1,1), got (%d,%d)", rec.last.X, rec.last.Y)
	}
	if cmd == nil {
		t.Error("expected panel command to be returned")
	}
}

func TestFloatPaneHandleMouseOutsideIgnored(t *testing.T) {
	fp := &FloatPane{
		Panel:               BasePanel{},
		X:                   0,
		Y:                   0,
		Width:               10,
		Height:              10,
		CloseOnOutsideClick: true,
	}
	press := tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft, X: 50, Y: 50}
	cmd := fp.handleMouse(press, 50, 50)
	if cmd != nil {
		t.Error("expected nil command for outside click")
	}
	if fp.CloseRequested {
		t.Error("CloseRequested should not be set by handleMouse")
	}
}

func TestFloatPaneHitEdge(t *testing.T) {
	fp := &FloatPane{Width: 10, Height: 5}
	cases := []struct {
		x, y int
		want string
	}{
		{0, 0, "nw"},
		{9, 0, "ne"},
		{0, 4, "sw"},
		{9, 4, "se"},
		{5, 0, "n"},
		{5, 4, "s"},
		{0, 2, "w"},
		{9, 2, "e"},
		{5, 2, ""},
	}
	for _, tc := range cases {
		got := fp.hitEdge(tc.x, tc.y)
		if got != tc.want {
			t.Errorf("hitEdge(%d,%d) = %q, want %q", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestFloatPaneApplyResize(t *testing.T) {
	cases := []struct {
		edge                       string
		dx, dy                     int
		wantX, wantY, wantW, wantH int
	}{
		{"n", 0, 3, 10, 13, 20, 7},
		{"s", 0, 3, 10, 10, 20, 13},
		{"w", 3, 0, 13, 10, 17, 10},
		{"e", 3, 0, 10, 10, 23, 10},
		{"nw", 3, 2, 13, 12, 17, 8},
		{"ne", 0, 2, 10, 12, 20, 8},
		{"sw", 3, 0, 13, 10, 17, 10},
		{"se", 3, 2, 10, 10, 23, 12},
	}
	for _, tc := range cases {
		fp := &FloatPane{
			X:      10,
			Y:      10,
			Width:  20,
			Height: 10,
		}
		fp.resizeEdge = tc.edge
		fp.origX, fp.origY = 10, 10
		fp.origW, fp.origH = 20, 10
		fp.applyResize(tc.dx, tc.dy)
		if fp.X != tc.wantX || fp.Y != tc.wantY || fp.Width != tc.wantW || fp.Height != tc.wantH {
			t.Errorf("edge %s dx=%d dy=%d: got X=%d Y=%d W=%d H=%d, want X=%d Y=%d W=%d H=%d",
				tc.edge, tc.dx, tc.dy, fp.X, fp.Y, fp.Width, fp.Height, tc.wantX, tc.wantY, tc.wantW, tc.wantH)
		}
	}
}

func TestFloatPaneApplyResizeClamp(t *testing.T) {
	fp := &FloatPane{
		X:      10,
		Y:      10,
		Width:  20,
		Height: 10,
	}
	fp.resizeEdge = "e"
	fp.origX, fp.origY = 10, 10
	fp.origW, fp.origH = 20, 10
	fp.applyResize(-50, 0)
	if fp.Width != floatMinWidth {
		t.Errorf("expected width clamped to %d, got %d", floatMinWidth, fp.Width)
	}

	fp = &FloatPane{X: 10, Y: 10, Width: 20, Height: 10}
	fp.resizeEdge = "w"
	fp.origX, fp.origY = 10, 10
	fp.origW, fp.origH = 20, 10
	fp.applyResize(-20, 0)
	if fp.X != 0 {
		t.Errorf("expected X clamped to 0, got %d", fp.X)
	}
	if fp.Width != 40 {
		t.Errorf("expected width 40, got %d", fp.Width)
	}
}

func TestStripANSI(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"hello", "hello"},
		{"\x1b[31mred\x1b[0m", "red"},
		{"\x1b[1m\x1b[31mbold red\x1b[0m", "bold red"},
		{"\x1b" + "[" + "mplain", "plain"}, // malformed-ish but handled
	}
	for _, tc := range cases {
		got := StripANSI(tc.in)
		if got != tc.want {
			t.Errorf("StripANSI(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestOverlayFloat(t *testing.T) {
	lines := []string{
		"0123456789",
		"1234567890",
		"2345678901",
	}
	fp := &FloatPane{
		Panel:  BasePanel{},
		X:      2,
		Y:      1,
		Width:  6,
		Height: 3,
		Title:  "T",
	}
	overlayFloat(lines, fp, 10, 3)

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	for i, line := range lines {
		if lipgloss.Width(line) != 10 {
			t.Errorf("line %d visual width = %d, want 10", i, lipgloss.Width(line))
		}
	}

	stripped := StripANSI(lines[1])
	r := []rune(stripped)
	if len(r) != 10 {
		t.Fatalf("expected 10 runes, got %d (%q)", len(r), stripped)
	}
	if r[0] != '1' || r[1] != '2' {
		t.Errorf("expected prefix '12', got %q", string(r[:2]))
	}
	if r[8] != '9' || r[9] != '0' {
		t.Errorf("expected suffix '90', got %q", string(r[8:]))
	}
	if !strings.Contains(stripped, "×") {
		t.Errorf("expected overlay to contain close button, got %q", stripped)
	}
}

func TestOverlayFloatClipped(t *testing.T) {
	lines := []string{strings.Repeat(" ", 10)}
	fp := &FloatPane{
		Panel:  BasePanel{},
		X:      6,
		Y:      0,
		Width:  8,
		Height: 3,
		Title:  "T",
	}
	overlayFloat(lines, fp, 10, 1)
	if lipgloss.Width(lines[0]) != 10 {
		t.Errorf("expected visual width 10 when clipped, got %d", lipgloss.Width(lines[0]))
	}
}

func TestOverlayFloatPreservesANSIBackground(t *testing.T) {
	lines := []string{"\x1b[34m0123456789\x1b[0m"}
	fp := &FloatPane{
		Panel:  BasePanel{},
		X:      2,
		Y:      0,
		Width:  6,
		Height: 3,
		Title:  "",
	}
	overlayFloat(lines, fp, 10, 1)
	if !strings.Contains(lines[0], "\x1b[34m") {
		t.Error("expected original ANSI prefix to be preserved")
	}
	if lipgloss.Width(lines[0]) != 10 {
		t.Errorf("expected visual width 10, got %d", lipgloss.Width(lines[0]))
	}
}

func TestOverlayFloatOffScreen(t *testing.T) {
	lines := []string{"0123456789"}
	fp := &FloatPane{
		Panel:  BasePanel{},
		X:      10,
		Y:      0,
		Width:  6,
		Height: 3,
		Title:  "T",
	}
	overlayFloat(lines, fp, 10, 1)
	if lines[0] != "0123456789" {
		t.Errorf("expected lines unchanged when float is off screen, got %q", lines[0])
	}
}
