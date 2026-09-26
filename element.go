package warp

import (
	"encoding/json"
	"sync"
)

// Element describes a semantic UI element with its screen bounds.
type Element struct {
	Role     string    `json:"role"`
	Name     string    `json:"name"`
	Action   string    `json:"action,omitempty"`
	Bounds   Bounds    `json:"bounds"`
	Children []Element `json:"children,omitempty"`
}

// Bounds defines a rectangular screen region in cell coordinates.
type Bounds struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type elementSnapshot struct {
	elements []Element
	jsonOnce sync.Once
	json     []byte
	jsonErr  error
}

func newElementSnapshot(elements []Element) *elementSnapshot {
	if elements == nil {
		elements = []Element{}
	}
	return &elementSnapshot{elements: elements}
}

func (snapshot *elementSnapshot) jsonBytes() ([]byte, error) {
	if snapshot == nil {
		return []byte("[]\n"), nil
	}
	snapshot.jsonOnce.Do(func() {
		encoded, err := json.Marshal(snapshot.elements)
		if err != nil {
			snapshot.jsonErr = err
			return
		}
		snapshot.json = append(encoded, '\n')
	})
	return snapshot.json, snapshot.jsonErr
}

// Center returns the center cell of the bounds.
func (b Bounds) Center() (int, int) {
	return b.X + b.W/2, b.Y + b.H/2
}

// ElementProvider is implemented by panels that can expose their UI elements.
type ElementProvider interface {
	Elements(width, height int) []Element
}

// ViewportElementProvider returns semantic elements intersecting the requested
// content viewport. Returned bounds remain relative to the full content origin.
type ViewportElementProvider interface {
	ElementsAt(width, height, offset int) []Element
}

// collectElements returns elements from a panel if it implements ElementProvider.
func collectElements(panel Panel, width, height int) []Element {
	if isNilPanel(panel) {
		return nil
	}
	if ep, ok := panel.(ElementProvider); ok {
		return ep.Elements(width, height)
	}
	return nil
}

func clipElements(elements []Element, viewport Bounds) []Element {
	if len(elements) == 0 || viewport.W <= 0 || viewport.H <= 0 {
		return nil
	}

	var clipped []Element
	for _, element := range elements {
		bounds, visible := intersectBounds(element.Bounds, viewport)
		if !visible {
			continue
		}
		element.Bounds = bounds
		element.Children = clipElements(element.Children, viewport)
		clipped = append(clipped, element)
	}
	return clipped
}

func intersectBounds(bounds, viewport Bounds) (Bounds, bool) {
	if bounds.W <= 0 || bounds.H <= 0 || viewport.W <= 0 || viewport.H <= 0 {
		return Bounds{}, false
	}

	left := max(bounds.X, viewport.X)
	top := max(bounds.Y, viewport.Y)
	right := min(elementBoundsEnd(bounds.X, bounds.W), elementBoundsEnd(viewport.X, viewport.W))
	bottom := min(elementBoundsEnd(bounds.Y, bounds.H), elementBoundsEnd(viewport.Y, viewport.H))
	if right <= left || bottom <= top {
		return Bounds{}, false
	}
	return Bounds{X: left, Y: top, W: right - left, H: bottom - top}, true
}

func elementBoundsEnd(start, size int) int {
	if size <= 0 {
		return start
	}
	maxInt := int(^uint(0) >> 1)
	if start > maxInt-size {
		return maxInt
	}
	return start + size
}

// ElementProviderFunc adapts a plain function to the ElementProvider interface.
type ElementProviderFunc func(width, height int) []Element

// Elements implements ElementProvider.
func (f ElementProviderFunc) Elements(width, height int) []Element {
	return f(width, height)
}

// FindElement recursively searches elements for the first matching role/name/action.
func FindElement(elems []Element, role, name, action string) (Element, bool) {
	for _, el := range elems {
		if (role == "" || el.Role == role) &&
			(name == "" || el.Name == name) &&
			(action == "" || el.Action == action) {
			return el, true
		}
		if found, ok := FindElement(el.Children, role, name, action); ok {
			return found, true
		}
	}
	return Element{}, false
}
