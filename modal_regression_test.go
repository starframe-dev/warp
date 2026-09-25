package warp

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestModalResizePreservesOffsetAndClampsToViewport(t *testing.T) {
	modal := NewModal("界面", "Resize 👩‍💻", []ModalButton{{Label: "続行"}}, nil)
	modal.EnsureDimensions(80, 24)
	modal.offsetX = 100
	modal.offsetY = 100
	modal.EnsureDimensions(40, 10)
	if modal.BoxWidth() != 30 || modal.StartX() != 10 || modal.StartY() != 3 {
		t.Fatalf("resized modal geometry=(%d,%d %dx%d), want (10,3 30x7)", modal.StartX(), modal.StartY(), modal.BoxWidth(), modal.BoxHeight())
	}

	lines := make([]string, 10)
	for i := range lines {
		lines[i] = strings.Repeat(" ", 40)
	}
	lines = modal.Overlay(lines, 40, 10)
	for i, line := range lines {
		if got := ansi.StringWidth(line); got != 40 {
			t.Errorf("resized overlay line %d width=%d, want 40", i, got)
		}
	}
	if modal.StartX()+modal.BoxWidth() > 40 || modal.StartY()+modal.BoxHeight() > 10 {
		t.Fatalf("modal exceeds resized viewport: start=(%d,%d) box=%dx%d", modal.StartX(), modal.StartY(), modal.BoxWidth(), modal.BoxHeight())
	}
}

func TestModalOverlayStripsOSCAndStyledBackgroundSequences(t *testing.T) {
	const width, height = 80, 20
	styled := "\x1b]8;;https://example.test\x1b\\\x1b[31m界é\x1b[0m\x1b]8;;\x1b\\"
	lines := make([]string, height)
	for i := range lines {
		lines[i] = strings.Repeat(" ", width)
	}
	lines[0] = padVisualLine(styled, width)

	modal := NewModal("Title", "Content", nil, nil)
	got := modal.Overlay(lines, width, height)
	if strings.Contains(got[0], "\x1b]8;") {
		t.Fatal("background retained an OSC hyperlink sequence")
	}
	if strings.Contains(got[0], "\x1b[31m") {
		t.Fatal("background retained the original foreground color sequence")
	}
	want := "界é" + strings.Repeat(" ", width-ansi.StringWidth("界é"))
	if stripped := ansi.Strip(got[0]); stripped != want {
		t.Fatalf("dimmed background text=%q, want %q", stripped, want)
	}
}

func TestModalOverlayTinyUnicodeViewports(t *testing.T) {
	for width := 1; width <= 8; width++ {
		for height := 1; height <= 8; height++ {
			modal := NewModal("界面 👩‍💻", "内容 e\u0301", []ModalButton{{Label: "続行"}}, nil)
			lines := make([]string, height)
			for i := range lines {
				lines[i] = padVisualLine("背景界🙂", width)
			}
			got := modal.Overlay(lines, width, height)
			if len(got) != height {
				t.Fatalf("viewport %dx%d returned %d lines", width, height, len(got))
			}
			for y, line := range got {
				if gotWidth := ansi.StringWidth(line); gotWidth != width {
					t.Errorf("viewport %dx%d line %d width=%d", width, height, y, gotWidth)
				}
			}
		}
	}
}
