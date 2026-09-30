package integration_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	warp "github.com/starframe-dev/warp"
)

func modalMouse(x, y int, action tea.MouseAction) tea.MouseMsg {
	return tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: action}
}

func TestModalOverlayGeometrySafeLabelsAndHitTesting(t *testing.T) {
	const width, height = 80, 20
	background := make([]string, height)
	for i := range background {
		background[i] = "background " + strings.Repeat("x", width)
	}
	background[0] = "\x1b[31mANSI styled background\x1b[0m"

	buttonCalls, closeCalls := 0, 0
	modal := warp.NewModal("Title\x1b[31mBAD", "\x1b[32mcontent\x1b[0m", []warp.ModalButton{
		{Label: "OK\x1b[2J", Action: func() { buttonCalls++ }},
	}, func() { closeCalls++ })
	modal.Width = 40
	lines := modal.Overlay(background, width, height)

	if modal.BoxWidth() != 40 || modal.BoxHeight() != 7 {
		t.Fatalf("modal geometry = %dx%d, want 40x7", modal.BoxWidth(), modal.BoxHeight())
	}
	if modal.StartX() != (width-40)/2 || modal.StartY() != (height-7)/2 {
		t.Fatalf("modal origin = (%d,%d), want centered (%d,%d)", modal.StartX(), modal.StartY(), (width-40)/2, (height-7)/2)
	}
	if len(lines) != height {
		t.Fatalf("overlay has %d lines, want %d", len(lines), height)
	}
	plain := warp.StripANSI(strings.Join(lines, "\n"))
	if !strings.Contains(plain, "TitleBAD") || !strings.Contains(plain, "[OK]") || !strings.Contains(plain, "content") {
		t.Fatalf("overlay missing safe title, button, or ANSI content: %q", plain)
	}
	if strings.Contains(plain, "\x1b") || strings.Contains(plain, "2J") {
		t.Fatalf("framework labels retained injected terminal controls: %q", plain)
	}
	if got := ansi.StringWidth(warp.StripANSI(lines[modal.StartY()])); got != width {
		t.Fatalf("overlay row visual width = %d, want %d", got, width)
	}
	if !strings.Contains(warp.StripANSI(lines[0]), "ANSI styled background") {
		t.Fatalf("background content was lost: %q", warp.StripANSI(lines[0]))
	}

	// The visible label's text cell is clickable; unrelated cells are not.
	if modal.HandleMouse(modalMouse(modal.StartX()+4, modal.StartY()+4, tea.MouseActionPress)) != true || buttonCalls != 1 {
		t.Fatalf("button click was not consumed or invoked exactly once: calls=%d", buttonCalls)
	}
	if modal.HandleMouse(modalMouse(modal.StartX()+4, modal.StartY()+3, tea.MouseActionPress)) {
		t.Fatal("click outside button hit area was consumed")
	}
	if modal.HandleMouse(modalMouse(modal.StartX()+modal.BoxWidth()-4, modal.StartY()+2, tea.MouseActionPress)) != true || closeCalls != 1 {
		t.Fatalf("close click was not consumed or invoked: calls=%d", closeCalls)
	}
}

func TestModalOverlayDragAndClamp(t *testing.T) {
	const width, height = 70, 16
	modal := warp.NewModal("Move", "body", nil, nil)
	modal.Width = 36
	modal.Overlay(modalBlankLines(width, height), width, height)
	initialX, initialY := modal.StartX(), modal.StartY()

	// Only the top padding strip initiates a drag.
	if !modal.HandleMouse(modalMouse(initialX+2, initialY+1, tea.MouseActionPress)) {
		t.Fatal("press on draggable padding was not consumed")
	}
	if !modal.HandleMouse(modalMouse(initialX+7, initialY+3, tea.MouseActionMotion)) {
		t.Fatal("motion during drag was not consumed")
	}
	if !modal.HandleMouse(modalMouse(initialX+7, initialY+3, tea.MouseActionRelease)) {
		t.Fatal("release after drag was not consumed")
	}
	if modal.StartX() != initialX+5 || modal.StartY() != initialY+2 {
		t.Fatalf("dragged origin = (%d,%d), want (%d,%d)", modal.StartX(), modal.StartY(), initialX+5, initialY+2)
	}

	// A further drag beyond the lower-right edge is clamped to the viewport.
	x, y := modal.StartX()+2, modal.StartY()+1
	modal.HandleMouse(modalMouse(x, y, tea.MouseActionPress))
	modal.HandleMouse(modalMouse(width+100, height+100, tea.MouseActionMotion))
	modal.HandleMouse(modalMouse(width+100, height+100, tea.MouseActionRelease))
	if modal.StartX() != width-modal.BoxWidth() || modal.StartY() != height-modal.BoxHeight() {
		t.Fatalf("drag was not clamped: origin=(%d,%d) size=%dx%d viewport=%dx%d", modal.StartX(), modal.StartY(), modal.BoxWidth(), modal.BoxHeight(), width, height)
	}
}

func modalBlankLines(width, height int) []string {
	lines := make([]string, height)
	for i := range lines {
		lines[i] = strings.Repeat(" ", width)
	}
	return lines
}
