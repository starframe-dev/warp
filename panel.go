package warp

import tea "github.com/charmbracelet/bubbletea"

// Panel is the interface users implement to create content for warp panes.
// A Panel can be anything: terminal, text, graphics, form.
type Panel interface {
	// View renders the panel content at the given size.
	View(width, height int) string

	// Update handles Bubbletea messages (keys, mouse, etc).
	// The panel receives only messages that arrived while it was focused.
	Update(msg tea.Msg) tea.Cmd
}

// ContentHeightProvider is an optional interface for panels that know their
// intrinsic content height at a given width. Implementations must not render
// content to determine the height. A false known value means the height is
// unknown; consumers normalize negative known heights to zero.
type ContentHeightProvider interface {
	ContentHeight(width int) (height int, known bool)
}

// ViewportRenderer renders only the requested content rows for a known offset.
// Scrollable uses it only when the wrapped panel reports a known content height.
type ViewportRenderer interface {
	ViewAt(width, height, offset int) string
}

// Unmounter releases resources when Warp permanently removes a panel from its
// ownership domain. Implement it on pointer-backed panels so instance identity
// remains stable. Warp invokes it only after the panel is detached and no longer
// referenced by that domain.
type Unmounter interface {
	Unmount()
}

func panelContentHeight(panel Panel, width int) (int, bool) {
	if isNilPanel(panel) {
		return 0, false
	}
	provider, ok := panel.(ContentHeightProvider)
	if !ok {
		return 0, false
	}
	height, known := provider.ContentHeight(max(0, width))
	if !known {
		return 0, false
	}
	return max(0, height), true
}

// BasePanel provides a default no-op implementation of Panel.
// Embed it in your panel and override only the methods you need.
type BasePanel struct{}

func (BasePanel) View(width, height int) string {
	return ""
}

func (BasePanel) Update(msg tea.Msg) tea.Cmd {
	return nil
}
