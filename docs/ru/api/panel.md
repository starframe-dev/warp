---
title: Panel
description: Интерфейс, который реализует каждая панель в компоновке Warp.
---

# Panel

```go
type Panel interface {
    View(width, height int) string
    Update(msg tea.Msg) tea.Cmd
}
```

`View` получает размер панели в терминальных ячейках. `Update` обрабатывает сообщения Bubble Tea и может вернуть команду.

## BasePanel

```go
type BasePanel struct{}

func (BasePanel) View(width, height int) string
func (BasePanel) Update(msg tea.Msg) tea.Cmd
```

Встраивайте `BasePanel`, если панели достаточно пустого представления и no-op обновления. Фокус и семантическое дерево добавляются отдельными интерфейсами `Focusable` и `ElementProvider`.
