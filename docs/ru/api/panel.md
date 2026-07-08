---
title: Panel
description: Базовый интерфейс панели Warp.
---

# Panel

`Panel` — базовый интерфейс для всех визуальных компонентов Warp.

```go
type Panel interface {
    View(width, height int) string
    Update(msg tea.Msg) tea.Cmd
}
```

`View` рендерит панель в заданных размерах. `Update` обрабатывает сообщения Bubble Tea и возвращает команду.
