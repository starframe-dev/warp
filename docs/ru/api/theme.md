---
title: Тема
description: Семантическая палитра цветов Warp для контролов и overlay.
---

# Тема

Warp использует палитру Gruvbox Dark. `SetTheme` заменяет цвета уровня пакета и пересоздаёт стили табов, границ, float-панелей, dropdown, popover, модалок, input и collapsible.

## ThemeColors

```go
type ThemeColors struct {
    Background          string
    Surface             string
    Raised              string
    Border              string
    BorderMuted         string
    Text                string
    TextMuted           string
    TextStrong          string
    Accent              string
    AccentMuted         string
    Error               string
    Success             string
    Warning             string
    SelectionBackground string
    SelectionForeground string
}
```

Поля содержат строки цветов, которые принимает `lipgloss.Color`, обычно hex-значения вроде `"#282828"`.

`Text` и `AccentMuted` зарезервированы как семантические поля и пока не связаны с цветом уровня пакета. Остальные поля соответствуют внутренней палитре Warp. `SelectionBackground` и `SelectionForeground` задают цвета активной вкладки и выбранного popover.

## SetTheme

```go
func SetTheme(colors ThemeColors)
```

Вызовите `SetTheme` до отрисовки, чтобы применить палитру. Функция не проверяет строки цветов и не возвращает ошибку. Повторный вызов заменяет предыдущую палитру.

## Пример

```go
warp.SetTheme(warp.ThemeColors{
    Background:          "#1e1e1e",
    Surface:             "#252525",
    Raised:              "#333333",
    Border:              "#555555",
    BorderMuted:         "#666666",
    Text:                "#cccccc",
    TextMuted:           "#999999",
    TextStrong:          "#eeeeee",
    Accent:              "#4f8aff",
    AccentMuted:         "#7aa0d4",
    Error:               "#e55",
    Success:             "#4c4",
    Warning:             "#ee8811",
    SelectionBackground: "#336699",
    SelectionForeground: "#ffffff",
})
```

Состояние темы действует только внутри процесса; используется последняя установленная палитра.
