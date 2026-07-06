package warp

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

// Center returns the center cell of the bounds.
func (b Bounds) Center() (int, int) {
	return b.X + b.W/2, b.Y + b.H/2
}

// ElementProvider is implemented by panels that can expose their UI elements.
type ElementProvider interface {
	Elements(width, height int) []Element
}

// collectElements returns elements from a panel if it implements ElementProvider.
func collectElements(panel Panel, width, height int) []Element {
	if panel == nil {
		return nil
	}
	if ep, ok := panel.(ElementProvider); ok {
		return ep.Elements(width, height)
	}
	return nil
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
