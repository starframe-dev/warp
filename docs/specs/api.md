# API

> Полная спецификация: [`specs/api.md`](https://github.com/starframe-dev/warp/blob/main/specs/api.md)

## Warp

```go
w := warp.New()
w.SetRoot(panel)
w.Run()
```

## TabGroup

```go
tg := warp.NewTabGroup(warp.TabTop)
tg.NewTab("name")
tg.ActiveTab()
```

## Tab

```go
tab.SplitVertical(parent, 0.5, newPanel)
tab.SplitHorizontal(parent, 0.5, newPanel)
tab.FlexRow(parent, []warp.FlexItemSpec{...})
tab.Float(panel, x, y, w, h)
tab.FocusNext()
tab.FocusPrev()
```

## Panel

```go
type Panel interface {
    View(width, height int) string
    Update(msg tea.Msg) tea.Cmd
}
```

## Компоненты

- `NewCollapsible(title, panel)` — сворачиваемая панель
- `NewScrollable(panel)` — скроллируемая панель
- `NewDropdownMenu(label, items)` — выпадающее меню
- `NewSelectable(panel)` — выделение текста
- `NewInput(prompt)` — текстовый ввод
- `NewModal(title, content, buttons, onClose)` — модальное окно
- `Popover` — контекстное меню
