---
title: Warp — Go TUI Layout Engine
description: Вкладки, сплиты, flex, floats, модалки, popover
---

# Warp

**Go TUI Layout Engine** — вкладки, сплиты, flexbox, плавающие панели, модалки, popover.

На базе [Bubbletea](https://github.com/charmbracelet/bubbletea) и [Lipgloss](https://github.com/charmbracelet/lipgloss).

```go
w := warp.New()
tab := w.ActiveTab()
tab.SplitVertical(tab.RootPanel(), 0.5, leftPanel)
tab.SplitVertical(tab.RootPanel(), 0.5, rightPanel)
w.Run()
```

## Возможности

- **Вкладки** — TabGroup как Panel, вложенные куда угодно
- **Сплиты** — вертикальные/горизонтальные с перетаскиваемыми границами
- **Flexbox** — row/column с grow весами
- **Floats** — перетаскиваемые, изменяемые, закрываемые панели
- **Collapsible** — сворачиваемые секции
- **Scrollable** — скролл с колесом мыши
- **Dropdown** — меню с ховером и выбором
- **Selectable** — выделение текста мышью и клавиатурой
- **Input** — текстовый ввод с курсором
- **Modal** — диалоговые окна с затемнением
- **Popover** — контекстные меню
- **Focus API** — явное переключение фокуса
- **Element tree** — семантическое дерево для E2E тестов
- **Тема** Gruvbox Dark

## Быстрый старт

```bash
go get github.com/starframe-dev/warp
```

```go
package main

import (
    "github.com/starframe-dev/warp"
)

func main() {
    w := warp.New()
    tab := w.ActiveTab()
    tab.Float(&myPanel{}, 10, 5, 20, 10)
    w.Run()
}
```

## Документация

- [Руководство](/ru/guide/getting-started)
- [API](/ru/api/)
- [English docs](/guide/getting-started)
