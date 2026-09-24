package warp

import tea "github.com/charmbracelet/bubbletea"

// Direction specifies the split orientation.
type Direction int

const (
	// Vertical splits left/right (side by side).
	Vertical Direction = iota
	// Horizontal splits top/bottom (stacked).
	Horizontal
)

// ResizeMsg is sent by the layout engine to a panel when its allocated
// rectangle changes. It carries the panel's content size in cells (without
// borders or padding).
type ResizeMsg struct {
	Width  int
	Height int
}

// MinPanelSize is the minimum size in cells when enough space is available.
const MinPanelSize = 3

// SplitConfig is an internal node in the panel tree.
// It splits its area between two children.
type SplitConfig struct {
	Direction Direction
	Fraction  float64 // Share of First child (0.0–1.0)
	First     *Node
	Second    *Node
	Dragging  bool // true during drag-and-drop

	// CollapseRow is the row (0-indexed) where the border shows "<" instead of "│".
	// When clicked, OnCollapse is called. It should return a tea.Cmd to trigger re-render.
	CollapseRow int
	OnCollapse  func() tea.Cmd // called when the "<" on the border is clicked
}

// NodeCollapse holds the collapsed state for a node.
// When Collapsed is true, the node renders with a fixed size.
type NodeCollapse struct {
	Active bool
	Width  int     // fixed width when collapsed (for vertical layouts)
	Height int     // fixed height when collapsed (for horizontal layouts)
	Saved  float64 // saved fraction to restore on expand
}

// Node is a node in the panel tree — either a terminal (Panel) or internal (Split/Flex).
type Node struct {
	Panel    Panel
	Split    *SplitConfig
	Flex     *FlexConfig
	Collapse *NodeCollapse
}

// IsLeaf returns true if this node contains only a Panel.
func (n *Node) IsLeaf() bool {
	return n != nil && !isNilPanel(n.Panel) && n.Split == nil && n.Flex == nil
}

// IsCollapsed returns true if the node is currently collapsed.
func (n *Node) IsCollapsed() bool {
	return n != nil && n.Collapse != nil && n.Collapse.Active
}

// CollapsedSize returns the collapsed size for the given direction, or 0 if not collapsed.
func (n *Node) CollapsedSize(d Direction) int {
	if n == nil || n.Collapse == nil || !n.Collapse.Active {
		return 0
	}
	if d == Vertical {
		if n.Collapse.Width > 0 {
			return n.Collapse.Width
		}
		return 1
	}
	if n.Collapse.Height > 0 {
		return n.Collapse.Height
	}
	return 1
}

// FlexItem is a single item inside a FlexConfig.
type FlexItem struct {
	Node      *Node
	Grow      int  // flex-grow weight
	Shrink    int  // flex-shrink (unused for now)
	Basis     int  // flex-basis (min size), 0 = auto
	Collapsed bool // when true, the item takes only 1 line/char
}

// FlexConfig lays out its children in a row or column with weighted sizes.
type FlexConfig struct {
	Direction Direction
	Items     []*FlexItem
	Dragging  bool
}

// findNode locates the leaf node containing the given panel.
// Returns nil if the panel is nil or not found.
func (n *Node) findNode(panel Panel) *Node {
	if n == nil || isNilPanel(panel) {
		return nil
	}
	if n.IsLeaf() && samePanel(n.Panel, panel) {
		return n
	}
	if n.Split != nil {
		if found := n.Split.First.findNode(panel); found != nil {
			return found
		}
		if found := n.Split.Second.findNode(panel); found != nil {
			return found
		}
	}
	if n.Flex != nil {
		for _, item := range n.Flex.Items {
			if item == nil {
				continue
			}
			if found := item.Node.findNode(panel); found != nil {
				return found
			}
		}
	}
	return nil
}

// findSplitParent finds the node whose Split directly contains the panel.
func (n *Node) findSplitParent(panel Panel) *Node {
	if n == nil || isNilPanel(panel) {
		return nil
	}
	if n.Split != nil {
		if n.Split.First != nil && n.Split.First.IsLeaf() && samePanel(n.Split.First.Panel, panel) {
			return n
		}
		if n.Split.Second != nil && n.Split.Second.IsLeaf() && samePanel(n.Split.Second.Panel, panel) {
			return n
		}
		if found := n.Split.First.findSplitParent(panel); found != nil {
			return found
		}
		if found := n.Split.Second.findSplitParent(panel); found != nil {
			return found
		}
	}
	if n.Flex != nil {
		for _, item := range n.Flex.Items {
			if item != nil {
				if found := item.Node.findSplitParent(panel); found != nil {
					return found
				}
			}
		}
	}
	return nil
}

// replaceNode replaces oldNode with newNode in the tree.
// Returns true if replacement was successful.
func (n *Node) replaceNode(old, new *Node) bool {
	if n == nil {
		return false
	}
	if n.Split != nil {
		if n.Split.First == old {
			n.Split.First = new
			return true
		}
		if n.Split.Second == old {
			n.Split.Second = new
			return true
		}
		return n.Split.First.replaceNode(old, new) || n.Split.Second.replaceNode(old, new)
	}
	if n.Flex != nil {
		for _, item := range n.Flex.Items {
			if item == nil {
				continue
			}
			if item.Node == old {
				item.Node = new
				return true
			}
			if item.Node.replaceNode(old, new) {
				return true
			}
		}
	}
	return false
}

// collectLeafNodes collects all leaf (Panel) nodes in order.
func (n *Node) collectLeafNodes() []*Node {
	if n == nil {
		return nil
	}
	if n.IsLeaf() {
		return []*Node{n}
	}
	var result []*Node
	if n.Split != nil {
		result = append(result, n.Split.First.collectLeafNodes()...)
		result = append(result, n.Split.Second.collectLeafNodes()...)
	}
	if n.Flex != nil {
		for _, item := range n.Flex.Items {
			if item != nil {
				result = append(result, item.Node.collectLeafNodes()...)
			}
		}
	}
	return result
}
