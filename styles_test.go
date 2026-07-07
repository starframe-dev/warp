package warp

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestBorderStyle(t *testing.T) {
	style := BorderStyle()
	if style.GetForeground() != lipgloss.Color(borderColor) {
		t.Fatalf("expected BorderStyle foreground %v, got %v", borderColor, style.GetForeground())
	}
}

func TestBorderDragStyle(t *testing.T) {
	style := BorderDragStyle()
	if style.GetForeground() != lipgloss.Color(borderDragColor) {
		t.Fatalf("expected BorderDragStyle foreground %v, got %v", borderDragColor, style.GetForeground())
	}
}

func TestBorderHoverStyle(t *testing.T) {
	style := BorderHoverStyle()
	if style.GetForeground() != lipgloss.Color(borderHoverColor) {
		t.Fatalf("expected BorderHoverStyle foreground %v, got %v", borderHoverColor, style.GetForeground())
	}
}
