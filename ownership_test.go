package warp

import "testing"

type ownershipTestPanel struct {
	BasePanel
	unmounts int
}

func (p *ownershipTestPanel) Unmount() { p.unmounts++ }

func TestPanelOwnershipRootAndNilReceivers(t *testing.T) {
	var nilOwnership *panelOwnership
	nilOwnership.setRoot(nil)
	if nilOwnership.rootPanel() != nil {
		t.Fatal("nil ownership should have no root")
	}
	root := &ownershipTestPanel{}
	o := newPanelOwnership(root)
	if o.rootPanel() != root {
		t.Fatal("rootPanel did not return configured root")
	}
	o.setRoot(nil)
	if o.rootPanel() != nil {
		t.Fatal("setRoot did not clear root")
	}
	o.setRoot(root)
	if o.rootPanel() != root {
		t.Fatal("setRoot did not update root")
	}
	o.unmountRemoved([]Panel{root})
	o.unmountRemoved(nil)
	o.unmountRemoved([]Panel{root})
	if root.unmounts != 0 {
		t.Fatal("retained root was unmounted")
	}
}

func TestPanelInstanceIdentityAndRemoval(t *testing.T) {
	p := &ownershipTestPanel{}
	key, ok := panelInstance(p)
	if !ok || key.address == 0 {
		t.Fatal("pointer-backed panel should have an instance key")
	}
	if _, ok := panelInstance(nil); ok {
		t.Fatal("nil panel should not have an instance key")
	}
	var typedNil *ownershipTestPanel
	if _, ok := panelInstance(typedNil); ok {
		t.Fatal("typed nil panel should not have an instance key")
	}
	valueKey, ok := panelInstance(BasePanel{})
	if !ok || valueKey.typeOf == nil || valueKey.value != (BasePanel{}) {
		t.Fatal("comparable value panel should have an instance key")
	}

	o := newPanelOwnership(&ownershipTestPanel{})
	o.unmountRemoved([]Panel{p, p, nil, typedNil})
	if p.unmounts != 1 {
		t.Fatalf("duplicate candidates unmounted %d times, want 1", p.unmounts)
	}
	o.unmountRemoved([]Panel{p})
	if p.unmounts != 2 {
		t.Fatalf("repeated removal should unmount again, got %d", p.unmounts)
	}
	var nilOwnership *panelOwnership
	nilOwnership.unmountRemoved([]Panel{p})
	if p.unmounts != 2 {
		t.Fatal("nil ownership should not unmount candidates")
	}
}

func TestOwnershipDetachAndAttach(t *testing.T) {
	o := newPanelOwnership(nil)
	wp := &warpPanel{warp: &Warp{}}
	tab := &Tab{root: &Node{Panel: wp}}
	group := &TabGroup{}
	root := &TabGroup{tabs: []*Tab{tab}}
	other := newPanelOwnership(nil)

	tab.ownership = other
	group.ownership = other
	wp.warp.ownership = other
	wp.warp.ownsDomainRoot = false
	detachPanelOwnership(tab, o)
	detachPanelOwnership(group, o)
	detachPanelOwnership(wp, o)
	if tab.ownership != other || group.ownership != other || wp.warp.ownership != other {
		t.Fatal("detach changed ownership belonging to another domain")
	}

	attachPanelOwnership(root, o)
	if root.ownership != o || tab.ownership != o || tab.parent != root {
		t.Fatal("ownership was not attached throughout tab group")
	}
	if wp.warp.ownsDomainRoot {
		t.Fatal("attached nested warp should not own its domain root")
	}
	detachPanelOwnership(tab, o)
	if tab.ownership != nil {
		t.Fatal("tab ownership was not detached")
	}
	detachPanelOwnership(root, o)
	if root.ownership != nil {
		t.Fatal("group ownership was not detached")
	}
	detachPanelOwnership(wp, o)
	if wp.warp.ownership != nil || !wp.warp.ownsDomainRoot {
		t.Fatal("detached nested warp should regain root ownership")
	}
	detachPanelOwnership(&warpPanel{}, o)
	attachPanelOwnership(root, nil)
}

func TestPanelInstanceCollectionAndTabContainment(t *testing.T) {
	leaf := &ownershipTestPanel{}
	tab := &Tab{root: &Node{Panel: leaf}}
	group := &TabGroup{tabs: []*Tab{tab}}
	scroll := &Scrollable{Content: leaf}
	selectable := &Selectable{Content: scroll}
	collapsible := &Collapsible{Content: selectable}
	warpRoot := &ownershipTestPanel{}
	wp := &warpPanel{warp: &Warp{root: warpRoot}}

	node := &Node{Split: &SplitConfig{First: &Node{Panel: collapsible}, Second: &Node{Panel: wp}}, Flex: &FlexConfig{Items: []*FlexItem{nil, {Node: &Node{Panel: leaf}}}}}
	panels := collectNodePanelInstances(node)
	if len(panels) != 7 {
		t.Fatalf("collected %d unique panel instances, want 7", len(panels))
	}
	if got := collectPanelInstances(group); len(got) != 3 {
		t.Fatalf("collected %d group instances, want group, tab and leaf", len(got))
	}
	if !tabContainsPanel(tab, leaf) || tabContainsPanel(tab, &ownershipTestPanel{}) || tabContainsPanel(nil, leaf) || tabContainsPanel(tab, nil) {
		t.Fatal("tab containment returned an unexpected result")
	}
	if len(collectPanelInstances(nil)) != 0 || len(collectNodePanelInstances(nil)) != 0 {
		t.Fatal("nil roots should produce empty collections")
	}
}

func TestEnsureOwnershipAndSliceRemoval(t *testing.T) {
	var tab *Tab
	var group *TabGroup
	if tab.ensureOwnership() != nil || group.ensureOwnership() != nil {
		t.Fatal("nil receivers should return nil ownership")
	}
	tab = &Tab{}
	o := tab.ensureOwnership()
	if o == nil || tab.ensureOwnership() != o || o.rootPanel() != tab {
		t.Fatal("standalone tab ownership not initialized consistently")
	}
	group = &TabGroup{}
	go1, go2 := group.ensureOwnership(), group.ensureOwnership()
	if go1 == nil || go1 != go2 || go1.rootPanel() != group {
		t.Fatal("group ownership not initialized consistently")
	}
	child := &Tab{parent: group}
	if child.ensureOwnership() != go1 {
		t.Fatal("child tab did not inherit its parent group's ownership")
	}

	items := []int{1, 2, 3}
	if got := removeSliceAt(items, -1); len(got) != 3 {
		t.Fatal("negative index should leave slice unchanged")
	}
	if got := removeSliceAt(items, 3); len(got) != 3 {
		t.Fatal("out-of-range index should leave slice unchanged")
	}
	got := removeSliceAt(items, 1)
	if len(got) != 2 || got[0] != 1 || got[1] != 3 || items[2] != 0 {
		t.Fatalf("removeSliceAt produced %v (backing slice %v)", got, items)
	}
}
