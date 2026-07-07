package warp

import "testing"

type splitTestPanel struct {
	BasePanel
	id int
}

func TestDirectionConstants(t *testing.T) {
	if Vertical != 0 {
		t.Errorf("Vertical expected 0, got %d", Vertical)
	}
	if Horizontal != 1 {
		t.Errorf("Horizontal expected 1, got %d", Horizontal)
	}
}

func TestResizeMsg(t *testing.T) {
	msg := ResizeMsg{Width: 80, Height: 24}
	if msg.Width != 80 || msg.Height != 24 {
		t.Errorf("ResizeMsg fields mismatch: got %+v", msg)
	}
}

func TestMinPanelSize(t *testing.T) {
	if MinPanelSize != 3 {
		t.Errorf("MinPanelSize expected 3, got %d", MinPanelSize)
	}
}

func TestNodeIsLeaf(t *testing.T) {
	leaf := &Node{Panel: splitTestPanel{id: 1}}
	if !leaf.IsLeaf() {
		t.Error("expected leaf node to be leaf")
	}

	split := &Node{Split: &SplitConfig{}}
	if split.IsLeaf() {
		t.Error("expected split node to not be leaf")
	}

	flex := &Node{Flex: &FlexConfig{}}
	if flex.IsLeaf() {
		t.Error("expected flex node to not be leaf")
	}
}

func TestNodeIsCollapsed(t *testing.T) {
	collapsed := &Node{Collapse: &NodeCollapse{Active: true}}
	if !collapsed.IsCollapsed() {
		t.Error("expected node to be collapsed")
	}

	inactive := &Node{Collapse: &NodeCollapse{Active: false}}
	if inactive.IsCollapsed() {
		t.Error("expected inactive collapse to not be collapsed")
	}

	empty := &Node{}
	if empty.IsCollapsed() {
		t.Error("expected node without collapse to not be collapsed")
	}
}

func TestNodeCollapsedSize(t *testing.T) {
	cases := []struct {
		name      string
		node      *Node
		direction Direction
		want      int
	}{
		{"vertical explicit width", &Node{Collapse: &NodeCollapse{Active: true, Width: 5}}, Vertical, 5},
		{"vertical default width", &Node{Collapse: &NodeCollapse{Active: true}}, Vertical, 1},
		{"horizontal explicit height", &Node{Collapse: &NodeCollapse{Active: true, Height: 7}}, Horizontal, 7},
		{"horizontal default height", &Node{Collapse: &NodeCollapse{Active: true}}, Horizontal, 1},
		{"inactive collapse", &Node{Collapse: &NodeCollapse{Active: false, Width: 5}}, Vertical, 0},
		{"no collapse", &Node{}, Vertical, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.node.CollapsedSize(tc.direction); got != tc.want {
				t.Errorf("CollapsedSize(%v) = %d, want %d", tc.direction, got, tc.want)
			}
		})
	}
}

func TestNodeFindNode(t *testing.T) {
	p1 := splitTestPanel{id: 1}
	p2 := splitTestPanel{id: 2}
	p3 := splitTestPanel{id: 3}

	leaf1 := &Node{Panel: p1}
	leaf2 := &Node{Panel: p2}
	leaf3 := &Node{Panel: p3}

	root := &Node{Split: &SplitConfig{
		First: leaf1,
		Second: &Node{Split: &SplitConfig{
			First:  leaf2,
			Second: leaf3,
		}},
	}}

	if got := root.findNode(p1); got != leaf1 {
		t.Errorf("findNode(p1) = %v, want %v", got, leaf1)
	}
	if got := root.findNode(p2); got != leaf2 {
		t.Errorf("findNode(p2) = %v, want %v", got, leaf2)
	}
	if got := root.findNode(p3); got != leaf3 {
		t.Errorf("findNode(p3) = %v, want %v", got, leaf3)
	}
	if got := root.findNode(splitTestPanel{id: 99}); got != nil {
		t.Errorf("findNode(unknown) = %v, want nil", got)
	}
	if got := (*Node)(nil).findNode(p1); got != nil {
		t.Errorf("findNode on nil receiver = %v, want nil", got)
	}
}

func TestNodeReplaceNode(t *testing.T) {
	old := &Node{Panel: splitTestPanel{id: 4}}
	replacement := &Node{Panel: splitTestPanel{id: 5}}

	root := &Node{Split: &SplitConfig{
		First:  old,
		Second: &Node{Panel: splitTestPanel{id: 6}},
	}}

	if !root.replaceNode(old, replacement) {
		t.Fatal("expected replaceNode to succeed")
	}
	if root.Split.First != replacement {
		t.Errorf("expected First child to be replaced, got %v", root.Split.First)
	}

	if root.replaceNode(&Node{Panel: splitTestPanel{id: 7}}, &Node{Panel: splitTestPanel{id: 8}}) {
		t.Error("expected replaceNode to fail for missing node")
	}

	if (*Node)(nil).replaceNode(old, replacement) {
		t.Error("expected replaceNode on nil receiver to fail")
	}
}

func TestNodeCollectLeafNodes(t *testing.T) {
	p1 := splitTestPanel{id: 1}
	p2 := splitTestPanel{id: 2}
	p3 := splitTestPanel{id: 3}

	leaf1 := &Node{Panel: p1}
	leaf2 := &Node{Panel: p2}
	leaf3 := &Node{Panel: p3}

	root := &Node{Split: &SplitConfig{
		First: leaf1,
		Second: &Node{Flex: &FlexConfig{
			Items: []*FlexItem{
				{Node: leaf2},
				{Node: leaf3},
			},
		}},
	}}

	leaves := root.collectLeafNodes()
	if len(leaves) != 3 {
		t.Fatalf("expected 3 leaves, got %d", len(leaves))
	}
	if leaves[0] != leaf1 || leaves[1] != leaf2 || leaves[2] != leaf3 {
		t.Errorf("leaves order mismatch: got %v", leaves)
	}

	if got := leaf1.collectLeafNodes(); len(got) != 1 || got[0] != leaf1 {
		t.Errorf("single leaf collectLeafNodes = %v, want [%v]", got, leaf1)
	}

	if got := (*Node)(nil).collectLeafNodes(); got != nil {
		t.Errorf("nil collectLeafNodes = %v, want nil", got)
	}
}

func TestSplitConfig(t *testing.T) {
	cfg := &SplitConfig{
		Direction: Vertical,
		Fraction:  0.5,
		First:     &Node{Panel: splitTestPanel{id: 10}},
		Second:    &Node{Panel: splitTestPanel{id: 11}},
	}
	if cfg.Direction != Vertical || cfg.Fraction != 0.5 {
		t.Errorf("SplitConfig mismatch: got %+v", cfg)
	}
}

func TestFlexConfig(t *testing.T) {
	item := &FlexItem{
		Node:      &Node{Panel: splitTestPanel{id: 12}},
		Grow:      1,
		Shrink:    1,
		Basis:     10,
		Collapsed: false,
	}
	cfg := &FlexConfig{
		Direction: Horizontal,
		Items:     []*FlexItem{item},
	}
	if len(cfg.Items) != 1 || cfg.Direction != Horizontal {
		t.Errorf("FlexConfig mismatch: got %+v", cfg)
	}
}
