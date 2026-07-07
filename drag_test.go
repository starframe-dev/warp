package warp

// Drag-and-drop interactions for split borders are not implemented in this
// file. They are handled in tab.go and rendered in render.go:
//
//   - tab.go implements the input handling and state management for dragging a
//     border between split panes. The relevant functions are handleMouse,
//     updateDrag, and hitBorder.
//   - render.go computes the screen positions of borders via findBorders.
