# panel.go

The `panel.go` file defines the `Panel` abstraction used by the Warp TUI
framework. A **Panel** is the fundamental content unit for warp panes —
it can represent anything from plain text to interactive graphics and
forms. The file also provides `BasePanel`, a no-op base type for
convenient embedding.

## Public API

### Panel (interface)

` type Panel interface { View(width, height int) string Update(msg tea.Msg) tea.Cmd } `

`Panel` is the core interface that users implement to render content
inside warp panes. It exposes two methods:

| Method | Description |
|----|----|
| `View(width, height int) string` | Renders the panel content for the given dimensions. |
| `Update(msg tea.Msg) tea.Cmd` | Handles incoming Bubbletea messages (keys, mouse, etc.) while the panel is focused. |

### BasePanel (struct)

` type BasePanel struct{} func (BasePanel) View(width, height int) string { return "" } func (BasePanel) Update(msg tea.Msg) tea.Cmd { return nil } `

`BasePanel` provides a default no-op implementation of `Panel`. Embed it
in your concrete panel type and override only the methods you need —
this avoids boilerplate while keeping the interface contract satisfied.

## Usage

Implement `Panel` directly, or embed `BasePanel` for partial overrides.
The panel receives only messages that arrived while it was focused.

``` go

type MyPanel struct {
    BasePanel // embed for no-op defaults
    title string
}

func (p *MyPanel) View(width, height int) string {
    return fmt.Sprintf("[%s]", p.title)
}

func (p *MyPanel) Update(msg tea.Msg) tea.Cmd {
    // handle keys while focused
    return nil
}
```

## Implementation Notes

- **Message routing** — only focused panels receive `Update` calls;
  unfocused panels are not disturbed by incoming messages.
- **Dimension-aware rendering** — `View` receives explicit `width` and
  `height`, so panels can adapt their layout to the available space
  without inspecting global state.
- **Embedding pattern** — Go's struct embedding gives a clean
  composition path: embed `BasePanel` to inherit no-op behavior and
  override only what you need.
- **Bubbletea integration** — the interface mirrors the `Model` contract
  from `github.com/charmbracelet/bubbletea` (specifically the
  `View`/`Update` pair), making panels plug-and-play with the
  framework's update loop.
