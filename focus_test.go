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
