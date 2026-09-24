package warp

import "testing"

func TestNodeMutationsKeepPanelAndContainerVariantsExclusive(t *testing.T) {
	a, b, c := geometryPanels()
	d := &geometryTestPanel{name: "D"}
	tab := NewTab("node-invariants")
	tab.SetRootPanel(a)
	root := tab.root
	if !root.IsLeaf() || root.Split != nil || root.Flex != nil {
		t.Fatalf("root leaf state = %+v", root)
	}

	tab.SplitVertical(a, 0.5, b)
	if root.Panel != nil || root.Split == nil || root.Flex != nil {
		t.Fatalf("split root has mixed variants: %+v", root)
	}
	flexNode := root.Split.Second
	tab.FlexRow(b, []FlexItemSpec{{Panel: b, Grow: 1}, {Panel: c, Grow: 1}})
	if flexNode.Panel != nil || flexNode.Split != nil || flexNode.Flex == nil {
		t.Fatalf("flex node has mixed variants: %+v", flexNode)
	}

	nested := flexNode.Flex.Items[1].Node
	tab.FlexColumn(c, []FlexItemSpec{{Panel: c, Grow: 1}, {Panel: d, Grow: 1}})
	if nested.Panel != nil || nested.Split != nil || nested.Flex == nil {
		t.Fatalf("nested flex node has mixed variants: %+v", nested)
	}
	for _, leaf := range root.collectLeafNodes() {
		if !leaf.IsLeaf() || leaf.Panel == nil || leaf.Split != nil || leaf.Flex != nil {
			t.Errorf("invalid leaf after repeated mutations: %+v", leaf)
		}
	}

	oldRoot := root
	tab.SetRootPanel(a)
	if tab.root == oldRoot || tab.root.Panel != a || tab.root.Split != nil || tab.root.Flex != nil {
		t.Fatalf("SetRootPanel did not restore a clean root leaf: %+v", tab.root)
	}
}

func TestNilPanelLookupsAndMutationsAreNoOps(t *testing.T) {
	a, b, c := geometryPanels()
	tab := NewTab("nil-panels")
	tab.SetRootPanel(a)
	tab.SplitVertical(a, 0.5, b)
	root := tab.root
	split := root.Split

	if root.findNode(nil) != nil || root.findSplitParent(nil) != nil {
		t.Fatal("nil panel unexpectedly matched a tree node")
	}
	tab.SplitVertical(nil, 0.2, c)
	tab.SplitHorizontal(nil, 0.2, c)
	tab.FlexRow(nil, []FlexItemSpec{{Panel: c, Grow: 1}})
	tab.FlexColumn(nil, []FlexItemSpec{{Panel: c, Grow: 1}})
	tab.SetSplitCollapse(nil, 0, nil)
	tab.ToggleSplitCollapse(nil)
	if tab.root != root || tab.root.Split != split || len(tab.root.collectLeafNodes()) != 2 {
		t.Fatalf("nil-panel mutation changed layout: %+v", tab.root)
	}
}

func TestTypedNilPanelNormalizesToEmptyLeaf(t *testing.T) {
	var panel *geometryTestPanel
	tab := NewTab("typed-nil")
	tab.SetRootPanel(panel)
	if tab.root.Panel == nil || !tab.root.IsLeaf() || tab.root.Split != nil || tab.root.Flex != nil {
		t.Fatalf("typed nil panel was not normalized: %+v", tab.root)
	}
}
