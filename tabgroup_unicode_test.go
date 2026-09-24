package warp

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestTabGroupTruncatesNamesByTerminalCells(t *testing.T) {
	name := strings.Repeat("界", 15) + "👩‍💻e\u0301"

	t.Run("horizontal", func(t *testing.T) {
		tg := NewTabGroup(TabTop)
		tg.NewTab(name)
		tg.View(80, 5)
		region := tg.tabRegions[tg.activeTab]
		truncated := ansi.Truncate(name, 20, "...")
		wantRegionWidth := ansi.StringWidth(fmt.Sprintf("▎ %s ×", truncated))
		if got := region.endX - region.startX; got != wantRegionWidth {
			t.Fatalf("horizontal tab region width=%d, want %d", got, wantRegionWidth)
		}
		if ansi.StringWidth(truncated) > 20 {
			t.Fatalf("horizontal tab name uses %d cells, want at most 20", ansi.StringWidth(truncated))
		}
	})

	t.Run("vertical right", func(t *testing.T) {
		tg := NewTabGroup(TabRight)
		tg.NewTab(name)
		tg.View(50, 6)
		region := tg.tabRegions[tg.activeTab]
		truncated := ansi.Truncate(name, 15, "...")
		wantRegionWidth := ansi.StringWidth(fmt.Sprintf("▎ %s ×", truncated))
		if got := region.endX - region.startX; got != wantRegionWidth {
			t.Fatalf("vertical tab region width=%d, want %d", got, wantRegionWidth)
		}
		if ansi.StringWidth(truncated) > 15 {
			t.Fatalf("vertical tab name uses %d cells, want at most 15", ansi.StringWidth(truncated))
		}
		if tg.verticalTabWidth < region.endX {
			t.Fatalf("vertical bar width=%d does not contain region ending at %d", tg.verticalTabWidth, region.endX)
		}
	})
}
