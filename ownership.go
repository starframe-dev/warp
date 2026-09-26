package warp

import (
	"reflect"
	"sync"
)

type panelOwnership struct {
	mu   sync.RWMutex
	root Panel
}

type panelInstanceKey struct {
	typeOf  reflect.Type
	address uintptr
	value   Panel
}

func newPanelOwnership(root Panel) *panelOwnership {
	return &panelOwnership{root: root}
}

func (o *panelOwnership) setRoot(root Panel) {
	if o == nil {
		return
	}
	o.mu.Lock()
	o.root = root
	o.mu.Unlock()
}

func (o *panelOwnership) rootPanel() Panel {
	if o == nil {
		return nil
	}
	o.mu.RLock()
	root := o.root
	o.mu.RUnlock()
	return root
}

func (o *panelOwnership) unmountRemoved(candidates []Panel) {
	if o == nil || len(candidates) == 0 {
		return
	}

	retained := make(map[panelInstanceKey]struct{})
	for _, panel := range collectPanelInstances(o.rootPanel()) {
		if key, ok := panelInstance(panel); ok {
			retained[key] = struct{}{}
		}
	}

	seen := make(map[panelInstanceKey]struct{}, len(candidates))
	for _, panel := range candidates {
		key, ok := panelInstance(panel)
		if !ok {
			continue
		}
		if _, duplicate := seen[key]; duplicate {
			continue
		}
		seen[key] = struct{}{}
		if _, stillOwned := retained[key]; stillOwned {
			continue
		}
		if unmounter, ok := panel.(Unmounter); ok {
			unmounter.Unmount()
		}
		detachPanelOwnership(panel, o)
	}
}

func detachPanelOwnership(panel Panel, ownership *panelOwnership) {
	switch current := panel.(type) {
	case *TabGroup:
		if current.ownership == ownership {
			current.ownership = nil
		}
	case *Tab:
		if current.ownership == ownership {
			current.ownership = nil
		}
	case *warpPanel:
		if current.warp == nil {
			return
		}
		current.warp.mu.Lock()
		if current.warp.ownership == ownership {
			current.warp.ownership = nil
			current.warp.ownsDomainRoot = true
		}
		current.warp.mu.Unlock()
	}
}

func panelInstance(panel Panel) (panelInstanceKey, bool) {
	if isNilPanel(panel) {
		return panelInstanceKey{}, false
	}

	typeOf := reflect.TypeOf(panel)
	value := reflect.ValueOf(panel)
	if value.Kind() == reflect.Pointer {
		return panelInstanceKey{typeOf: typeOf, address: value.Pointer()}, true
	}
	if typeOf.Comparable() {
		return panelInstanceKey{typeOf: typeOf, value: panel}, true
	}
	return panelInstanceKey{}, false
}

type panelInstanceCollector struct {
	panels  []Panel
	visited map[Panel]struct{}
}

func newPanelInstanceCollector() *panelInstanceCollector {
	return &panelInstanceCollector{visited: make(map[Panel]struct{})}
}

func collectPanelInstances(root Panel) []Panel {
	collector := newPanelInstanceCollector()
	collector.visitPanel(root)
	return collector.panels
}

func collectNodePanelInstances(root *Node) []Panel {
	collector := newPanelInstanceCollector()
	collector.visitNode(root)
	return collector.panels
}

func (collector *panelInstanceCollector) visitNode(node *Node) {
	if node == nil {
		return
	}
	collector.visitPanel(node.Panel)
	if node.Split != nil {
		collector.visitNode(node.Split.First)
		collector.visitNode(node.Split.Second)
	}
	if node.Flex != nil {
		for _, item := range node.Flex.Items {
			if item != nil {
				collector.visitNode(item.Node)
			}
		}
	}
}

func (collector *panelInstanceCollector) visitPanel(panel Panel) {
	if isNilPanel(panel) {
		return
	}
	switch current := panel.(type) {
	case *TabGroup:
		if _, ok := collector.visited[current]; ok {
			return
		}
		collector.visited[current] = struct{}{}
		collector.panels = append(collector.panels, current)
		for _, tab := range current.tabs {
			collector.visitPanel(tab)
		}
	case *Tab:
		if _, ok := collector.visited[current]; ok {
			return
		}
		collector.visited[current] = struct{}{}
		collector.panels = append(collector.panels, current)
		collector.visitNode(current.root)
		for _, float := range current.floats {
			if float != nil {
				collector.visitPanel(float.Panel)
			}
		}
	case *Scrollable:
		if _, ok := collector.visited[current]; ok {
			return
		}
		collector.visited[current] = struct{}{}
		collector.panels = append(collector.panels, current)
		collector.visitPanel(current.Content)
	case *Selectable:
		if _, ok := collector.visited[current]; ok {
			return
		}
		collector.visited[current] = struct{}{}
		collector.panels = append(collector.panels, current)
		collector.visitPanel(current.Content)
	case *Collapsible:
		if _, ok := collector.visited[current]; ok {
			return
		}
		collector.visited[current] = struct{}{}
		collector.panels = append(collector.panels, current)
		collector.visitPanel(current.Content)
	case *warpPanel:
		if _, ok := collector.visited[current]; ok {
			return
		}
		collector.visited[current] = struct{}{}
		collector.panels = append(collector.panels, current)
		if current.warp != nil {
			collector.visitPanel(current.warp.Root())
		}
	default:
		collector.panels = append(collector.panels, panel)
	}
}

func attachPanelOwnership(root Panel, ownership *panelOwnership) {
	if ownership == nil {
		return
	}
	for _, panel := range collectPanelInstances(root) {
		switch current := panel.(type) {
		case *TabGroup:
			current.ownership = ownership
			for _, tab := range current.tabs {
				if tab != nil {
					tab.parent = current
				}
			}
		case *Tab:
			current.ownership = ownership
		case *warpPanel:
			if current.warp != nil {
				current.warp.mu.Lock()
				current.warp.ownership = ownership
				current.warp.ownsDomainRoot = false
				current.warp.mu.Unlock()
			}
		}
	}
}

func collectTabPanels(tab *Tab) []Panel {
	return collectPanelInstances(tab)
}

func tabContainsPanel(tab *Tab, target Panel) bool {
	key, ok := panelInstance(target)
	if !ok {
		return false
	}
	for _, panel := range collectTabPanels(tab) {
		if candidate, ok := panelInstance(panel); ok && candidate == key {
			return true
		}
	}
	return false
}

func removeSliceAt[T any](items []T, index int) []T {
	if index < 0 || index >= len(items) {
		return items
	}
	copy(items[index:], items[index+1:])
	var zero T
	items[len(items)-1] = zero
	return items[:len(items)-1]
}

func (tab *Tab) ensureOwnership() *panelOwnership {
	if tab == nil {
		return nil
	}
	if tab.ownership == nil {
		if tab.parent != nil {
			tab.parent.ensureOwnership()
			tab.ownership = tab.parent.ownership
		} else {
			tab.ownership = newPanelOwnership(tab)
		}
		attachPanelOwnership(tab, tab.ownership)
	}
	return tab.ownership
}

func (group *TabGroup) ensureOwnership() *panelOwnership {
	if group == nil {
		return nil
	}
	if group.ownership == nil {
		group.ownership = newPanelOwnership(group)
		attachPanelOwnership(group, group.ownership)
	}
	return group.ownership
}
