package warp

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type lifecyclePanel struct {
	name      string
	unmounts  int
	onUnmount func()
}

func (*lifecyclePanel) View(_, _ int) string   { return "" }
func (*lifecyclePanel) Update(tea.Msg) tea.Cmd { return nil }

func (p *lifecyclePanel) Unmount() {
	p.unmounts++
	if p.onUnmount != nil {
		p.onUnmount()
	}
}

func TestUnmountWaitsForLastReferenceInTab(t *testing.T) {
	shared := &lifecyclePanel{name: "shared"}
	tab := NewTab("shared")
	tab.FlexRow(tab.RootPanel(), []FlexItemSpec{{Panel: shared}, {Panel: shared}})

	replacement := &lifecyclePanel{name: "replacement"}
	tab.FlexRow(shared, []FlexItemSpec{{Panel: replacement}})
	if shared.unmounts != 0 {
		t.Fatalf("Unmount after removing one of two references = %d, want 0", shared.unmounts)
	}

	tab.SetRootPanel(&lifecyclePanel{name: "final"})
	if shared.unmounts != 1 {
		t.Fatalf("Unmount after removing the last reference = %d, want 1", shared.unmounts)
	}
	if replacement.unmounts != 1 {
		t.Fatalf("replacement Unmount after root replacement = %d, want 1", replacement.unmounts)
	}
}

func TestClosingTabKeepsPanelReferencedByAnotherTab(t *testing.T) {
	shared := &lifecyclePanel{name: "shared"}
	group := NewTabGroup(TabTop)
	group.ActiveTab().SetRootPanel(shared)
	group.NewTab("second").SetRootPanel(shared)

	group.closeTab(0)
	if shared.unmounts != 0 {
		t.Fatalf("Unmount while another tab retains the panel = %d, want 0", shared.unmounts)
	}

	group.ActiveTab().SetRootPanel(&lifecyclePanel{name: "last"})
	if shared.unmounts != 1 {
		t.Fatalf("Unmount after final TabGroup reference = %d, want 1", shared.unmounts)
	}
}

func TestCloseFloatKeepsPanelReferencedByRoot(t *testing.T) {
	shared := &lifecyclePanel{name: "shared"}
	tab := NewTab("root-float")
	tab.SetRootPanel(shared)
	tab.Float(shared, 1, 1, 20, 8)
	float := tab.floats[0]

	tab.CloseFloat(float)
	if shared.unmounts != 0 {
		t.Fatalf("Unmount after closing float while root retains panel = %d, want 0", shared.unmounts)
	}
	if float.Panel != nil {
		t.Fatal("closed FloatPane retains its Panel pointer")
	}

	tab.SetRootPanel(&lifecyclePanel{name: "last"})
	if shared.unmounts != 1 {
		t.Fatalf("Unmount after removing root reference = %d, want 1", shared.unmounts)
	}
}

func TestReparentingDoesNotUnmountPanel(t *testing.T) {
	shared := &lifecyclePanel{name: "shared"}
	tab := NewTab("reparent")
	tab.SetRootPanel(shared)
	tab.SplitVertical(shared, 0.5, &lifecyclePanel{name: "sibling"})
	if shared.unmounts != 0 {
		t.Fatalf("Unmount during Split reparenting = %d, want 0", shared.unmounts)
	}

	tab.FlexRow(shared, []FlexItemSpec{{Panel: shared}, {Panel: &lifecyclePanel{name: "flex-sibling"}}})
	if shared.unmounts != 0 {
		t.Fatalf("Unmount while Flex retains the panel = %d, want 0", shared.unmounts)
	}
}

func TestClosingWarpDomainUnmountsEachUniquePanelOnce(t *testing.T) {
	shared := &lifecyclePanel{name: "shared"}
	other := &lifecyclePanel{name: "other"}
	group := NewTabGroup(TabTop)
	group.ActiveTab().SetRootPanel(shared)
	second := group.NewTab("second")
	second.SetRootPanel(shared)
	second.Float(shared, 1, 1, 20, 8)
	group.NewTab("third").SetRootPanel(other)

	warp := New()
	warp.SetRoot(&Scrollable{Content: group})
	warp.SetRoot(nil)

	if shared.unmounts != 1 {
		t.Fatalf("shared panel Unmounts after closing Warp domain = %d, want 1", shared.unmounts)
	}
	if other.unmounts != 1 {
		t.Fatalf("other panel Unmounts after closing Warp domain = %d, want 1", other.unmounts)
	}
}

func TestUnmountUsesPanelInstanceIdentity(t *testing.T) {
	first := &lifecyclePanel{name: "same"}
	second := &lifecyclePanel{name: "same"}
	group := NewTabGroup(TabTop)
	group.ActiveTab().SetRootPanel(first)
	group.NewTab("second").SetRootPanel(second)

	group.closeTab(0)
	if first.unmounts != 1 {
		t.Fatalf("removed pointer instance Unmounts = %d, want 1", first.unmounts)
	}
	if second.unmounts != 0 {
		t.Fatalf("distinct but deeply equal panel Unmounts = %d, want 0", second.unmounts)
	}
}

func TestIndependentWarpRootsDoNotShareLifecycleTracking(t *testing.T) {
	shared := &lifecyclePanel{name: "unsupported-shared"}
	first := New()
	second := New()
	first.SetRoot(shared)
	second.SetRoot(shared)

	first.SetRoot(nil)
	if shared.unmounts != 1 {
		t.Fatalf("first ownership domain Unmounts = %d, want 1 without global tracking", shared.unmounts)
	}
	if second.Root() != shared {
		t.Fatal("independent Warp unexpectedly coordinated shared ownership")
	}
}

func TestRemovedPointerSlicesAndLayoutReferencesAreCleared(t *testing.T) {
	group := NewTabGroup(TabTop)
	group.NewTab("second")
	group.NewTab("third")
	tabCapacity := cap(group.tabs)
	group.closeTab(1)
	if group.tabs[:tabCapacity][len(group.tabs)] != nil {
		t.Fatal("removed Tab remains in the backing array")
	}

	tab := NewTab("float-tail")
	for i := 0; i < 3; i++ {
		tab.Float(&lifecyclePanel{name: "float"}, i, 0, 20, 8)
	}
	floatCapacity := cap(tab.floats)
	removedFloat := tab.floats[1]
	tab.CloseFloat(removedFloat)
	if tab.floats[:floatCapacity][len(tab.floats)] != nil {
		t.Fatal("removed FloatPane remains in the backing array")
	}

	root := tab.RootPanel()
	tab.SplitVertical(root, 0.5, &lifecyclePanel{name: "split"})
	_ = tab.View(80, 24)
	oldBorders := tab.lastBorders
	if len(oldBorders) == 0 || oldBorders[0].Split == nil {
		t.Fatal("test setup did not create a stored split border")
	}
	tab.SetRootPanel(&lifecyclePanel{name: "new-root"})
	if oldBorders[0].Split != nil {
		t.Fatal("removed layout config remains in the old border slice")
	}
}

func TestUnmountRunsAfterPanelIsDetached(t *testing.T) {
	tab := NewTab("detach-before-unmount")
	panel := &lifecyclePanel{name: "old"}
	panel.onUnmount = func() {
		if tabContainsPanel(tab, panel) {
			t.Error("Unmount ran while the panel was still reachable from its domain")
		}
	}
	tab.SetRootPanel(panel)

	tab.SetRootPanel(&lifecyclePanel{name: "new"})
	if panel.unmounts != 1 {
		t.Fatalf("Unmount calls = %d, want 1", panel.unmounts)
	}
}
