---
title: Фокус и ввод
description: Focusable, переключение фокуса, RawKeyReceiver и пример обработки Tab в Warp.
---

# Фокус и ввод

Warp не навязывает правила фокуса и клавиатурных биндингов. Приложение само решает, какие клавиши и сценарии управляют вводом.

## Focusable

Панель может поддерживать фокус через интерфейс `Focusable`:

```go
type Focusable interface {
	Focus()
	Blur()
	Focused() bool
}
```

Методы отвечают за состояние фокуса:

- `Focus()` активирует панель;
- `Blur()` снимает фокус;
- `Focused() bool` возвращает текущее состояние.

## Управление фокусом во вкладке

`Tab` предоставляет явные методы переключения фокуса:

```go
tab.FocusNext()
tab.FocusPrev()
tab.FocusFirst()
tab.FocusPanel(panel)
```

- `FocusNext()` переводит фокус на следующую доступную панель;
- `FocusPrev()` переводит фокус на предыдущую доступную панель;
- `FocusFirst()` фокусирует первую доступную панель;
- `FocusPanel(panel)` фокусирует конкретную панель.

## Биндинги клавиш

Warp не биндит `Tab` и `Shift+Tab` автоматически.

Это сделано намеренно: разработчик сам решает, какие клавиши переключают фокус и как они конфликтуют с вводом в конкретных компонентах.

## RawKeyReceiver

`RawKeyReceiver` предназначен для PTY и терминальных панелей.

Такая панель получает все клавиши напрямую, без обычной обработки фокуса и компонентных биндингов. Это полезно для встроенных терминалов, shell-сессий и других интерактивных TUI внутри Warp.

## Пример переключения фокуса по Tab

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.Tab.FocusNext()
			return m, nil
		case "shift+tab":
			m.Tab.FocusPrev()
			return m, nil
		}
	}

	return m, nil
}
```

В этом примере приложение само связывает `Tab` с `FocusNext()`, а `Shift+Tab` — с `FocusPrev()`.
