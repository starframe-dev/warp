package integration_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	warp "github.com/starframe-dev/warp"
)

type lifecyclePanel struct {
	unmounts int
}

func (*lifecyclePanel) View(int, int) string   { return "" }
func (*lifecyclePanel) Update(tea.Msg) tea.Cmd { return nil }
func (p *lifecyclePanel) Unmount()             { p.unmounts++ }

func TestRootReplacementWaitsForLastNestedReference(t *testing.T) {
	shared := &lifecyclePanel{}
	removed := &lifecyclePanel{}

	inner := warp.NewTabGroup(warp.TabNone)
	inner.ActiveTab().SetRootPanel(warp.NewScrollable(shared))
	inner.NewTab("also shared").SetRootPanel(shared)
	outer := warp.NewTabGroup(warp.TabNone)
	outer.ActiveTab().SetRootPanel(inner)
	outer.NewTab("other").SetRootPanel(removed)

	w := warp.New()
	w.SetRoot(outer)

	// Keep one occurrence in the replacement tree while removing the nested
	// group and its duplicate references from the ownership domain.
	replacement := warp.NewTabGroup(warp.TabNone)
	replacement.ActiveTab().SetRootPanel(warp.NewCollapsible("retained", shared))
	w.SetRoot(replacement)

	if shared.unmounts != 0 {
		t.Fatalf("shared panel Unmounts after replacement retained a reference = %d, want 0", shared.unmounts)
	}
	if removed.unmounts != 1 {
		t.Fatalf("removed panel Unmounts after root replacement = %d, want 1", removed.unmounts)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	if shared.unmounts != 1 {
		t.Fatalf("shared panel Unmounts after final reference was removed = %d, want 1", shared.unmounts)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("second Close failed: %v", err)
	}
	if shared.unmounts != 1 {
		t.Fatalf("shared panel Unmounts after repeated Close = %d, want 1", shared.unmounts)
	}
}

func TestEmbeddedWarpCloseDefersUnmountToOuterOwnershipDomain(t *testing.T) {
	shared := &lifecyclePanel{}
	embeddedRoot := warp.NewTabGroup(warp.TabNone)
	embeddedRoot.ActiveTab().SetRootPanel(warp.NewScrollable(shared))
	embeddedRoot.NewTab("another reference").SetRootPanel(shared)

	embedded := warp.New()
	embedded.SetRoot(embeddedRoot)
	outerRoot := warp.NewTabGroup(warp.TabNone)
	outerRoot.ActiveTab().SetRootPanel(warp.NewCollapsible("embedded", embedded.AsPanel()))

	outer := warp.New()
	outer.SetRoot(outerRoot)

	if err := embedded.Close(); err != nil {
		t.Fatalf("embedded Close failed: %v", err)
	}
	if shared.unmounts != 0 {
		t.Fatalf("embedded Close unmounted a panel still owned by the outer Warp: %d, want 0", shared.unmounts)
	}

	// The nested tab references the panel more than once. Removing the outer
	// root removes the entire ownership domain and must unmount it just once.
	if err := outer.Close(); err != nil {
		t.Fatalf("outer Close failed: %v", err)
	}
	if shared.unmounts != 1 {
		t.Fatalf("outer Close Unmount calls = %d, want 1", shared.unmounts)
	}
}
