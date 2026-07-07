package warp

import (
	"testing"
)

type testFocusable struct {
	testPanel
	focused bool
}

func (p *testFocusable) Focus() {
	p.focused = true
}

func (p *testFocusable) Blur() {
	p.focused = false
}

func (p *testFocusable) Focused() bool {
	return p.focused
}

type testRawKeyReceiver struct {
	testPanel
}

func (p *testRawKeyReceiver) WantsRawKeys() bool {
	return true
}

func TestFocusNext(t *testing.T) {
	w := New()
	tab := w.ActiveTab()

	p1 := &testFocusable{testPanel: testPanel{name: "p1"}}
	p2 := &testFocusable{testPanel: testPanel{name: "p2"}}
	p3 := &testFocusable{testPanel: testPanel{name: "p3"}}

	tab.FlexRow(tab.RootPanel(), []FlexItemSpec{
		{Panel: p1, Grow: 1},
		{Panel: p2, Grow: 1},
		{Panel: p3, Grow: 1},
	})

	tab.FocusFirst()
	if tab.Focus() != p1 {
		t.Fatalf("expected p1 focused, got %v", tab.Focus())
	}
	if !p1.focused {
		t.Error("p1.Focused() should be true")
	}

	tab.FocusNext()
	if tab.Focus() != p2 {
		t.Errorf("expected p2 focused, got %v", tab.Focus())
	}
	if p1.focused {
		t.Error("p1 should be blurred after FocusNext")
	}
	if !p2.focused {
		t.Error("p2 should be focused")
	}

	tab.FocusNext()
	if tab.Focus() != p3 {
		t.Errorf("expected p3 focused, got %v", tab.Focus())
	}

	tab.FocusNext()
	if tab.Focus() != p1 {
		t.Errorf("expected wrap back to p1, got %v", tab.Focus())
	}
}

func TestFocusPrev(t *testing.T) {
	w := New()
	tab := w.ActiveTab()

	p1 := &testFocusable{testPanel: testPanel{name: "p1"}}
	p2 := &testFocusable{testPanel: testPanel{name: "p2"}}

	tab.FlexRow(tab.RootPanel(), []FlexItemSpec{
		{Panel: p1, Grow: 1},
		{Panel: p2, Grow: 1},
	})

	tab.FocusPanel(p2)
	if tab.Focus() != p2 {
		t.Fatalf("expected p2 focused")
	}

	tab.FocusPrev()
	if tab.Focus() != p1 {
		t.Errorf("expected p1 focused, got %v", tab.Focus())
	}

	tab.FocusPrev()
	if tab.Focus() != p2 {
		t.Errorf("expected wrap back to p2, got %v", tab.Focus())
	}
}

func TestFocusPanel(t *testing.T) {
	w := New()
	tab := w.ActiveTab()

	p1 := &testFocusable{testPanel: testPanel{name: "p1"}}
	p2 := &testFocusable{testPanel: testPanel{name: "p2"}}

	tab.SplitVertical(tab.RootPanel(), 0.5, p1)
	tab.SplitVertical(tab.RootPanel(), 0.5, p2)

	tab.FocusPanel(p1)
	if tab.Focus() != p1 {
		t.Fatalf("expected p1 focused")
	}
	if !p1.focused {
		t.Error("p1.Focused() should be true")
	}

	tab.FocusPanel(p2)
	if tab.Focus() != p2 {
		t.Fatalf("expected p2 focused")
	}
	if p1.focused {
		t.Error("p1 should be blurred")
	}
	if !p2.focused {
		t.Error("p2 should be focused")
	}
}

func TestIsFocusable(t *testing.T) {
	if _, ok := isFocusable(nil); ok {
		t.Error("isFocusable(nil) should be false")
	}

	plain := &testPanel{name: "plain"}
	if _, ok := isFocusable(plain); ok {
		t.Error("isFocusable(non-focusable) should be false")
	}

	f := &testFocusable{testPanel: testPanel{name: "focusable"}}
	got, ok := isFocusable(f)
	if !ok {
		t.Fatal("isFocusable(focusable) should be true")
	}
	if got != f {
		t.Error("isFocusable should return the same panel")
	}
}

func TestCollectFocusables(t *testing.T) {
	if got := collectFocusables(nil); got != nil {
		t.Errorf("collectFocusables(nil) should be nil, got %v", got)
	}

	f := &testFocusable{testPanel: testPanel{name: "leaf"}}
	leaf := &Node{Panel: f}
	if got := collectFocusables(leaf); len(got) != 1 || got[0] != f {
		t.Errorf("expected one focusable, got %v", got)
	}

	plain := &testPanel{name: "plain"}
	if got := collectFocusables(&Node{Panel: plain}); len(got) != 0 {
		t.Errorf("expected no focusables for plain panel, got %v", got)
	}

	f1 := &testFocusable{testPanel: testPanel{name: "f1"}}
	f2 := &testFocusable{testPanel: testPanel{name: "f2"}}
	split := &Node{
		Split: &SplitConfig{
			First:  &Node{Panel: f1},
			Second: &Node{Panel: f2},
		},
	}
	if got := collectFocusables(split); len(got) != 2 {
		t.Errorf("expected two focusables from split, got %v", got)
	}

	f3 := &testFocusable{testPanel: testPanel{name: "f3"}}
	flex := &Node{
		Flex: &FlexConfig{
			Items: []*FlexItem{
				{Node: &Node{Panel: f3}},
			},
		},
	}
	if got := collectFocusables(flex); len(got) != 1 || got[0] != f3 {
		t.Errorf("expected one focusable from flex, got %v", got)
	}
}

func TestFocusIndex(t *testing.T) {
	f1 := &testFocusable{testPanel: testPanel{name: "f1"}}
	f2 := &testFocusable{testPanel: testPanel{name: "f2"}}
	list := []Focusable{f1, f2}

	if got := focusIndex(list, nil); got != -1 {
		t.Errorf("focusIndex(nil current) should be -1, got %d", got)
	}

	if got := focusIndex(list, f1); got != 0 {
		t.Errorf("focusIndex(f1) should be 0, got %d", got)
	}

	if got := focusIndex(list, f2); got != 1 {
		t.Errorf("focusIndex(f2) should be 1, got %d", got)
	}

	other := &testFocusable{testPanel: testPanel{name: "other"}}
	if got := focusIndex(list, other); got != -1 {
		t.Errorf("focusIndex(missing) should be -1, got %d", got)
	}
}

func TestFocusNextFunc(t *testing.T) {
	if got := focusNext(nil, nil); got != nil {
		t.Errorf("focusNext(nil) should be nil, got %v", got)
	}

	f1 := &testFocusable{testPanel: testPanel{name: "f1"}}
	if got := focusNext([]Focusable{f1}, nil); got != f1 {
		t.Errorf("focusNext single item with nil current should return f1, got %v", got)
	}

	f2 := &testFocusable{testPanel: testPanel{name: "f2"}}
	list := []Focusable{f1, f2}
	if got := focusNext(list, f1); got != f2 {
		t.Errorf("focusNext should advance to f2, got %v", got)
	}
	if got := focusNext(list, f2); got != f1 {
		t.Errorf("focusNext should wrap to f1, got %v", got)
	}
}

func TestFocusPrevFunc(t *testing.T) {
	if got := focusPrev(nil, nil); got != nil {
		t.Errorf("focusPrev(nil) should be nil, got %v", got)
	}

	f1 := &testFocusable{testPanel: testPanel{name: "f1"}}
	if got := focusPrev([]Focusable{f1}, nil); got != f1 {
		t.Errorf("focusPrev single item with nil current should return f1, got %v", got)
	}

	f2 := &testFocusable{testPanel: testPanel{name: "f2"}}
	list := []Focusable{f1, f2}
	if got := focusPrev(list, f1); got != f2 {
		t.Errorf("focusPrev should wrap to f2, got %v", got)
	}
	if got := focusPrev(list, f2); got != f1 {
		t.Errorf("focusPrev should move back to f1, got %v", got)
	}
}

func TestApplyFocus(t *testing.T) {
	f1 := &testFocusable{testPanel: testPanel{name: "f1"}}
	f2 := &testFocusable{testPanel: testPanel{name: "f2"}}

	f1.Focus()
	applyFocus(f1, f2)
	if f1.focused {
		t.Error("previous focus should be blurred")
	}
	if !f2.focused {
		t.Error("new focus should be focused")
	}

	f1.focused = true
	applyFocus(f1, f1)
	if !f1.focused {
		t.Error("same focus should not be blurred")
	}

	f2.focused = false
	applyFocus(nil, f2)
	if !f2.focused {
		t.Error("nil current should still focus next")
	}

	applyFocus(f1, nil)
	if f1.focused {
		t.Error("nil next should blur current")
	}
}

func TestRawKeyReceiverInterface(t *testing.T) {
	r := &testRawKeyReceiver{testPanel: testPanel{name: "raw"}}
	var receiver RawKeyReceiver = r
	if !receiver.WantsRawKeys() {
		t.Error("RawKeyReceiver.WantsRawKeys() should return true")
	}

	plain := &testPanel{name: "plain"}
	if _, ok := Panel(plain).(RawKeyReceiver); ok {
		t.Error("non-raw panel should not implement RawKeyReceiver")
	}
}
