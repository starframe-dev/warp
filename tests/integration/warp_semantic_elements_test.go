package integration_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	warp "github.com/starframe-dev/warp"
)

type semanticFixturePanel struct {
	name   string
	bounds warp.Bounds
	height int
}

func (p *semanticFixturePanel) View(_, h int) string {
	lines := make([]string, h)
	for i := range lines {
		lines[i] = p.name
	}
	return stringsJoinLines(lines)
}

func (p *semanticFixturePanel) Update(tea.Msg) tea.Cmd { return nil }

func (p *semanticFixturePanel) Elements(_, _ int) []warp.Element {
	return []warp.Element{{Role: "status", Name: p.name, Bounds: p.bounds}}
}

func (p *semanticFixturePanel) ContentHeight(int) (int, bool) { return p.height, true }

func (p *semanticFixturePanel) ElementsAt(_, _, _ int) []warp.Element {
	return p.Elements(0, 0)
}

func stringsJoinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	result := lines[0]
	for _, line := range lines[1:] {
		result += "\n" + line
	}
	return result
}

func requireSemanticElement(t *testing.T, elements []warp.Element, role, name, action string) warp.Element {
	t.Helper()
	element, ok := warp.FindElement(elements, role, name, action)
	if !ok {
		t.Fatalf("element role=%q name=%q action=%q not found in %+v", role, name, action, elements)
	}
	return element
}

func TestSemanticElementsFlowThroughTabsLayoutWrappersFloatsAndScrollable(t *testing.T) {
	outer := warp.NewTabGroup(warp.TabTop)
	tab := outer.ActiveTab()

	scrolledContent := &semanticFixturePanel{
		name: "scrolled content", height: 12,
		bounds: warp.Bounds{X: 2, Y: 4, W: 5, H: 1},
	}
	scrollable := warp.NewScrollable(scrolledContent)
	scrollable.Offset = 4

	nested := warp.NewTabGroup(warp.TabTop)
	nested.ActiveTab().SetRootPanel(&semanticFixturePanel{
		name: "nested content", bounds: warp.Bounds{X: 2, Y: 3, W: 4, H: 1},
	})
	tab.SetRootPanel(scrollable)
	tab.SplitVertical(scrollable, .5, nested)
	floatContent := &semanticFixturePanel{
		name: "float content", bounds: warp.Bounds{X: 0, Y: 0, W: 3, H: 1},
	}
	tab.Float(floatContent, 30, 7, 12, 5)

	const width, height = 40, 12
	elements := outer.Elements(width, height)
	// The known content extent clamps the requested offset to one row;
	// visible bounds are shifted to the viewport origin and below the tab bar.
	scrolled := requireSemanticElement(t, elements, "status", "scrolled content", "")
	if want := (warp.Bounds{X: 2, Y: 4, W: 5, H: 1}); scrolled.Bounds != want {
		t.Fatalf("scrolled bounds=%+v, want %+v", scrolled.Bounds, want)
	}
	// The split starts its second child at x=20. The nested top tab adds one
	// content row to the child's local semantic coordinates.
	nestedElement := requireSemanticElement(t, elements, "status", "nested content", "")
	if want := (warp.Bounds{X: 22, Y: 5, W: 4, H: 1}); nestedElement.Bounds != want {
		t.Fatalf("nested bounds=%+v, want %+v", nestedElement.Bounds, want)
	}
	floatElement := requireSemanticElement(t, elements, "status", "float content", "")
	if want := (warp.Bounds{X: 29, Y: 8, W: 3, H: 1}); floatElement.Bounds != want {
		t.Fatalf("float bounds=%+v, want %+v", floatElement.Bounds, want)
	}
	closeFloat := requireSemanticElement(t, elements, "button", "Close Float", "close-float")
	if want := (warp.Bounds{X: 38, Y: 7, W: 1, H: 1}); closeFloat.Bounds != want {
		t.Fatalf("float close bounds=%+v, want %+v", closeFloat.Bounds, want)
	}
	activeTab := requireSemanticElement(t, elements, "tab", "main", "activate-tab")
	if activeTab.Bounds != (warp.Bounds{X: 0, Y: 0, W: 8, H: 1}) {
		t.Fatalf("outer tab bounds=%+v", activeTab.Bounds)
	}
	if _, ok := warp.FindElement(elements, "button", "Close main", "close-tab"); !ok {
		t.Fatalf("active tab close action missing: %+v", elements)
	}

	// A repeated semantic collection must not mutate the Scrollable's requested
	// offset or the provider's full-content coordinates.
	if scrollable.Offset != 4 || scrolledContent.bounds.Y != 4 {
		t.Fatalf("semantic collection changed source state: offset=%d bounds=%+v", scrollable.Offset, scrolledContent.bounds)
	}
}

type semanticSnapshotPanel struct {
	name string
}

func (p *semanticSnapshotPanel) View(_, _ int) string { return p.name }
func (p *semanticSnapshotPanel) Elements(_, _ int) []warp.Element {
	return []warp.Element{{Role: "status", Name: p.name, Bounds: warp.Bounds{X: 1, Y: 2, W: 3, H: 1}}}
}
func (p *semanticSnapshotPanel) Update(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case semanticStableTestMsg:
		p.name = "changed by stable message"
	case semanticOrdinaryTestMsg:
		p.name = "changed by ordinary message"
	}
	return nil
}

type semanticStableTestMsg struct{}

func (semanticStableTestMsg) SemanticStateUnchanged() bool { return true }

type semanticOrdinaryTestMsg struct{}

func semanticSnapshotRequest(t *testing.T, address string) []warp.Element {
	t.Helper()
	response, err := http.Get("http://" + address + "/elements")
	if err != nil {
		t.Fatalf("GET /elements: %v", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read /elements: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("GET /elements status=%d body=%q", response.StatusCode, body)
	}
	var elements []warp.Element
	if err := json.Unmarshal(body, &elements); err != nil {
		t.Fatalf("decode semantic snapshot %q: %v", body, err)
	}
	return elements
}

func TestSemanticSnapshotStableMessageDefersRebuildUntilView(t *testing.T) {
	w := warp.New()
	panel := &semanticSnapshotPanel{name: "initial"}
	w.SetRoot(panel)
	if err := w.ServeHTTP("127.0.0.1:0"); err != nil {
		t.Fatalf("start inspector: %v", err)
	}
	defer func() { _ = w.Close() }()
	w.Update(tea.WindowSizeMsg{Width: 30, Height: 8})
	w.View()
	if got := requireSemanticElement(t, semanticSnapshotRequest(t, w.HTTPAddr()), "status", "initial", ""); got.Bounds != (warp.Bounds{X: 1, Y: 2, W: 3, H: 1}) {
		t.Fatalf("initial snapshot bounds=%+v", got.Bounds)
	}

	w.Update(semanticStableTestMsg{})
	if got := semanticSnapshotRequest(t, w.HTTPAddr()); requireSemanticElement(t, got, "status", "initial", "").Name != "initial" {
		t.Fatalf("stable update replaced published snapshot: %+v", got)
	}
	w.View()
	if got := semanticSnapshotRequest(t, w.HTTPAddr()); requireSemanticElement(t, got, "status", "changed by stable message", "").Name != "changed by stable message" {
		t.Fatalf("View did not rebuild deferred snapshot: %+v", got)
	}

	w.Update(semanticOrdinaryTestMsg{})
	if got := semanticSnapshotRequest(t, w.HTTPAddr()); requireSemanticElement(t, got, "status", "changed by ordinary message", "").Name != "changed by ordinary message" {
		t.Fatalf("ordinary update did not publish a fresh snapshot: %+v", got)
	}
}
