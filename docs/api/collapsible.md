# Collapsible

`Collapsible` wraps a `Panel` with a persistent title row and collapse/expand behavior.

## Public API

```go
type Collapsible struct {
    Title     string
    Collapsed bool
    Content   Panel
}

func NewCollapsible(title string, content Panel) *Collapsible
func (c *Collapsible) View(w, h int) string
func (c *Collapsible) Elements(w, h int) []Element
func (c *Collapsible) Update(msg tea.Msg) tea.Cmd
func (c *Collapsible) Toggle()
```

`NewCollapsible` starts expanded. `Toggle` flips `Collapsed`; a nil `Content` does not prevent state changes.

## Rendering

The title row is visible in both states: `▼` when expanded and `▶` when collapsed. In expanded mode, the inner panel is rendered below the title with height `max(0, h-1)`, clipped and padded to the available viewport. In collapsed mode, only the title row is rendered. A non-positive height produces an empty view; with non-positive width, the title is blank and expanded mode still preserves exactly `h` rows.

A nil `Content` still renders the title when `h > 0`; the remaining expanded rows are blank.

## Semantic Elements

`Elements(w, h)` returns no child elements while collapsed. When expanded,
it requests content elements at height `max(0, h-1)`, clips them to the
content viewport, and adds one to each visible Y coordinate. The title row
is not reported as an element.

## Messages and mouse coordinates

`Update` forwards ordinary messages to `Content`. For `ResizeMsg`, expanded content receives height `max(0, h-1)` and collapsed content receives height `0`. Mouse events on the title row and events targeting hidden collapsed content are not forwarded. For visible content, `Y` is reduced by one so the first content row receives `Y == 0`.

When used in a `Tab`, a left click on the visible title row toggles the component. `Tab.Collapse` and `Tab.Expand` explicitly synchronize the component state with its flex item and are idempotent when repeated.

## Styling

The title's left portion uses `collapsibleStyle`; its horizontal rule and right corner use `collapsibleBorderStyle`. The title is truncated using terminal-cell width.

## Example

```go
section := NewCollapsible("Settings", settingsPanel)
view := section.View(40, 8) // title plus up to seven content rows
section.Toggle()
```
