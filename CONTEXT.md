# Warp — Handoff

## Что это

Warp — Go-библиотека (Bubbletea layout engine) для создания TUI с гибким управлением
пространством: вкладки, сплиты, плавающие панели, flexbox, модальные окна, popover.
Пользователь реализует интерфейс `Panel`, а warp управляет их расположением.

## Состояние проекта

**v0.7** — Input component, explicit focus API, Modal, Popover, Element tree, 35 тестов.

## Новое в v0.7

- **Input component** — `NewInput(prompt)` с курсором, backspace, delete, стрелками, home/end
- **Focus API** — `Focusable` interface, `FocusNext()`, `FocusPrev()`, `FocusFirst()`, `FocusPanel()`
  Warp **не биндит** Tab/Shift+Tab автоматически — разработчик сам решает
- **RawKeyReceiver** — интерфейс для PTY/терминалов, которым нужны все клавиши без перехвата
- **Modal** — `ShowModalMsg`/`CloseModalMsg`, overlay поверх lines, drag, close, buttons
- **Popover** — контекстное меню с Overlay, HandleMouse, HandleKey
- **Element tree** — `ElementProvider` интерфейс для семантического UI-дерева (HTTP endpoint)
- **ContextMenu удалён** — заменён на Popover
- **35 тестов** (было 27)

## Демо (`cmd/demo/main.go`)

```bash
go run ./cmd/demo/
```

**Tab 1 — «main»**: FlexRow с 3 input-полями (Name, Email, Search) + Preview panel.
Tab/Shift+Tab переключает фокус между input'ами (через custom appRoot).

**Tab 2 — «local-tabs»**: Local TabGroup(TabLeft) внутри FlexRow + Scrollable с Selectable.

**Tab 3 — «columns»**: FlexColumn с Collapsible + float.

**Tab 4 — «splits»**: Split-панели + float.

**Горячие клавиши** (определяет разработчик, не warp):
- `Tab` / `Shift+Tab` — переключение фокуса (в demo)
- `Ctrl+T` — новый таб
- `Ctrl+W` — закрыть таб
- `Ctrl+Tab` / `Ctrl+Shift+Tab` — следующий/предыдущий таб
- `q` / `Ctrl+C` — выход

## Модуль и зависимости

- **Модуль:** `github.com/starframe-dev/warp`
- **Директория:** `/Users/a/Space/Projects/Starframe/warp`
- **Go:** 1.22+
- **Зависимости:** `bubbletea v1.1.0`, `lipgloss v0.13.0`, `charmbracelet/x/ansi`, `rivo/uniseg`
- **Тесты:** 35 тестов проходят, `go vet` чист.
- **Git:** `https://github.com/starframe-dev/warp.git`, ветка `main`

## Архитектура

```
warp.go         — tea.Model, тонкая обёртка вокруг root Panel
tabgroup.go     — TabGroup: Panel с таб-баром, переключением табов, keyboard/mouse
tab.go          — Tab: дерево splits/flex, float-панели, фокус, mouse handling, рендеринг
panel.go        — интерфейс Panel{View(w,h) string; Update(Msg) Cmd}
split.go        — Node, SplitConfig, FlexConfig, Direction, MinPanelSize=3
render.go       — renderNode (рекурсивный), findBorders, padContent, computeFlexSizes
float.go        — FloatPane: рамка, drag, resize, overlayFloat, StripANSI, CloseOnOutsideClick
styles.go       — lipgloss-стили (Gruvbox Dark)
collapsible.go  — Collapsible Panel с заголовком и toggle
scrollable.go   — Scrollable Panel с viewport и mouse wheel
dropdown.go     — DropdownMenu Panel с кнопкой и раскрывающимся списком
selectable.go   — Selectable Panel с text selection (mouse drag, Shift+arrows, Ctrl+A, OSC 52)
wrap.go         — WordWrap, SpaceWrap утилиты для переноса текста
modal.go        — Modal dialog с overlay, drag, close, buttons
popover.go      — Popover контекстное меню с Overlay, HandleMouse, HandleKey
element.go      — Element, Bounds, ElementProvider для семантического UI-дерева
focus.go        — Focusable interface, collectFocusables, RawKeyReceiver
input.go        — Input Panel с курсором, backspace, delete, стрелками
drag.go         — заглушка (drag-логика в tab.go)
```

### Дерево панелей

```go
Node { Panel Panel | Split *SplitConfig | Flex *FlexConfig }
SplitConfig { Direction, Fraction, First *Node, Second *Node, Dragging bool }
FlexConfig  { Direction, Items []FlexItem }
```

### TabGroup как Panel

```go
// Табы как корень (как раньше)
w := warp.New()  // root = TabGroup с 1 табом

// Табы как компонент внутри flex
tg := warp.NewTabGroup(warp.TabLeft)
tg.NewTab("code")
tg.NewTab("debug")
tab.FlexRow(root, []warp.FlexItemSpec{
    {Panel: tg, Grow: 2},  // ← TabGroup внутри flex!
})

// Warp вообще без табов
w := warp.New()
w.SetRoot(myCustomPanel)
```

### Focus API

```go
// Разработчик сам решает, какие клавиши биндить:
tab.FocusNext()      // следующая focusable панель
tab.FocusPrev()      // предыдущая
tab.FocusFirst()     // первая
tab.FocusPanel(p)    // конкретная панель

// Focusable interface
type Focusable interface {
    Panel
    Focus()
    Blur()
    Focused() bool
}
```

### Modal / Popover

```go
// Modal — через сообщения
warp.ShowModalMsg{Title: "Confirm", Content: "Delete?", Buttons: [...]}

// Popover — контекстное меню
pop := &warp.Popover{
    Items:   []warp.PopoverItem{{Name: "Copy", Action: ...}},
    X:       x, Y: y,
    OnClose: func() { ... },
}
lines = pop.Overlay(lines, totalW, totalH)
```

### Element tree

```go
type ElementProvider interface {
    Elements(width, height int) []Element
}
// HTTP endpoint /elements возвращает JSON-дерево UI-элементов
```

## API

```go
w := warp.New()
tab := w.NewTab("name")
w.SetTabPosition(warp.TabBottom)

// Layouts
tab.SplitVertical(parent, 0.5, newPanel)
tab.SplitHorizontal(parent, 0.5, newPanel)
tab.FlexRow(parent, []warp.FlexItemSpec{{Panel: p1, Grow: 1}, ...})
tab.Float(panel, x, y, w, h)

// Collapsible
col := warp.NewCollapsible("Title", panel)
tab.ToggleCollapsible(col)

// Scrollable
scroll := warp.NewScrollable(panel)

// Dropdown
dd := warp.NewDropdownMenu("Menu", []warp.DropdownItem{...})
dd.OnSelect = func(idx int) { ... }

// Input
in := warp.NewInput("Name: ")
in.Focus()
in.SetValue("hello")

// Selectable
sel := warp.NewSelectable(panel)
sel.SelectedText()
sel.Copy()  // OSC 52 clipboard

// Focus
tab.FocusNext()
tab.FocusPrev()
tab.FocusFirst()
tab.FocusPanel(panel)

// Word wrap
lines := warp.WordWrap(text, 40)
lines := warp.SpaceWrap(text, 40)

// Nested warps
inner := warp.New()
tab.SplitVertical(tab.RootPanel(), 0.5, inner.AsPanel())

w.Run()
```

## Что не доделано / Ideas

- **Стилизация** — цвета захардкожены в styles.go, нет публичного API для кастомизации
- **Анимации** — нет (drag без анимации, переключение табов мгновенное)
- **Nested float** — float внутри float не поддерживается
- **List / Table** — нет компонентов для списков и таблиц
- **Textarea** — нет многострочного ввода
- **Subscriptions** — нет таймеров, Spinner, Progress
- **Layout constraints** — нет padding, gap, align, justify как в CSS
- **Help overlay** — нет встроенного help с key bindings
- **HTTP element tree** — ElementProvider есть, но HTTP сервер не реализован в warp

## Changelog

### v0.7
- **Input** — `NewInput(prompt)` с курсором, backspace, delete, стрелками
- **Focus API** — `Focusable`, `FocusNext()`, `FocusPrev()`, `FocusFirst()`, `FocusPanel()`
- **RawKeyReceiver** — для PTY/терминалов
- **Modal** — `ShowModalMsg`/`CloseModalMsg`, overlay, drag, close, buttons
- **Popover** — контекстное меню с Overlay, HandleMouse, HandleKey
- **Element tree** — `ElementProvider`, `Element`, `Bounds`
- **ContextMenu удалён** — заменён на Popover
- **Float close-on-outside-click** — `CloseOnOutsideClick`
- **OSC 52 clipboard** — `Selectable.Copy()`
- **35 тестов**

### v0.6
- **TabGroup как Panel** — табы внутри splits/flex, Warp — thin wrapper
- **Backward-compatible API** — `w.NewTab()`, `w.ActiveTab()` делегируют root TabGroup

### v0.5
- Табы не прыгают, Scrollable, DropdownMenu, ContextMenu, WordWrap, Collapsible, Flexbox
- Gruvbox Dark, вложенные табы, WindowSizeMsg broadcast
- Float overlay fix, ANSI isolation, float close button
- 24→27 тестов

## Правила кода

- Не использовать `os.Exit()` — только `tea.Quit`
- Не добавлять свои `signal.Notify` — Bubbletea сам обрабатывает SIGINT
- Минимальный размер панели: `MinPanelSize = 3`
- Fraction всегда через `clampFraction(0.1–0.9)`
- Файлы завершаются переносом строки
- Комментарии только на английском
- Warp не биндит Tab/Shift+Tab — фокус управляется разработчиком

## Как запустить тесты

```bash
cd /Users/a/Space/Projects/Starframe/warp
go vet ./...
go test ./... -v
```
