# Публичный API Warp

## Обзор

Warp — Go-библиотека TUI layout engine для Bubbletea. Она не рисует
контент сама, а собирает `Panel`-компоненты в дерево, рисует границы
и перераспределяет размеры. Контент, ввод и события остаются за
разработчиком.

Пакет: `github.com/starframe-dev/warp` (импортируется как `warp`).

## Точки входа

```go
w := warp.New()               // корень — TabGroup с одной вкладкой
w.SetRoot(myPanel)            // заменить корень полностью
tab := w.NewTab("editor")     // создать вкладку
tab2 := w.ActiveTab()         // активная вкладка
w.SetTabPosition(warp.TabLeft)
w.NextTab() / w.PrevTab()
w.Run()                      // запустить Bubbletea-программу
```

## Интерфейс `Panel`

```go
type Panel interface {
    View(width, height int) string
    Update(msg tea.Msg) tea.Cmd
}
```

Всё, что вставляется в дерево Warp — `Panel`. Это единственная
обязательная реализация у стороннего кода.

## Layout: `Tab`

`Tab` — единственная точка мутации дерева:

| Метод | Описание |
|--------|----------|
| `RootPanel() Panel` | корневая панель таба |
| `SetRootPanel(p Panel)` | заменить корневую панель |
| `SplitVertical(parent, fraction, newPanel)` | деление по вертикали |
| `SplitHorizontal(parent, fraction, newPanel)` | деление по горизонтали |
| `FlexRow(parent, items []FlexItemSpec)` | горизонтальный flex |
| `FlexColumn(parent, items []FlexItemSpec)` | вертикальный flex |
| `Float(panel, x, y, w, h)` | добавить плавающую панель |
| `CloseFloat(fp)` | удалить плавающую панель |
| `SetSplitCollapse(parent, row, onCollapse)` | символ «<» на границе |
| `ToggleSplitCollapse(parent)` | переключить collapsed |

`FlexItemSpec { Panel Panel; Grow int }` — `Grow: 0` (авто),
`Grow: 1..n` (доля остатка).

## `TabGroup` — Panel

```go
tg := warp.NewTabGroup(warp.TabLeft)
tg.NewTab("editor")
tg.NewTab("debug")
tg.ActiveTab()
tg.NextTab() / tg.PrevTab()
```

`TabGroup` сам — `Panel`, вставляется в flex/split наравне с другими
компонентами.

## Компоненты-обёртки

| Конструктор | Тип | Назначение |
|--------------|-----|------------|
| `NewCollapsible(title, panel)` | `*Collapsible` | сворачиваемая секция |
| `NewScrollable(panel)` | `*Scrollable` | прокрутка |
| `NewDropdownMenu(label, items)` | `*DropdownMenu` | выпадающее меню |
| `NewSelectable(panel)` | `*Selectable` | выделение текста |
| `NewInput(prompt)` | `*Input` | однострочный ввод |
| `NewModal(...)` / `ShowModalMsg` | `*Modal` | диалоговое окно |
| `&Popover{Items, X, Y, OnClose}` | `*Popover` | контекстное меню |

Компоненты сами реализуют `Panel`, плюс (при необходимости)
`Focusable` и `ElementProvider`.

## Фокус

```go
type Focusable interface {
    Panel
    Focus()
    Blur()
    Focused() bool
}

tab.FocusNext()      // следующая focusable
tab.FocusPrev()     // предыдущая
tab.FocusFirst()    // первая
tab.FocusPanel(p)   // конкретная панель
```

Warp не биндит `Tab`/`Shift+Tab`. Разработчик сам решает, какие
клавиши вызывают `Focus*` — и сам рисует help, если нужно.

`RawKeyReceiver` — интерфейс `{Panel; WantsRawKeys() bool}` для
PTY/терминалов, которым нужен прямой доступ ко всем клавишам.

## Element tree (E2E)

```go
type Element struct {
    Role     string
    Name     string
    Action   string
    Bounds   Bounds      // {X, Y, W, H}
    Children []Element
}

type ElementProvider interface {
    Elements(width, height int) []Element
}

FindElement(elems []Element, role, name, action string) (Element, bool)
```

HTTP: `GET /elements` — JSON-массив элементов.
`GET /healthz` — `ok`.

## Тема

```go
type ThemeColors struct {
    Background, Surface, Raised, Border, BorderMuted string
    Text, TextMuted, TextStrong string
    Accent, AccentMuted string
    Error, Success, Warning string
    SelectionBackground, SelectionForeground string
}
func SetTheme(c ThemeColors)
```

Пересчитывает все package-level `lipgloss.Style` (табы, floats,
dropdown, popover, modal, input, collapsible, split borders).
Порядок вызовов не важен: `SetTheme` перезаписывает
полный набор стилей.

## Утилиты

| Функция | Описание |
|---------|----------|
| `WordWrap(text, width)` | перенос по словам, ломает длинные слова |
| `SpaceWrap(text, width)` | перенос по пробелам, слова не ломает |
| `StripANSI(s string)` | удалить ANSI-последовательности, посчитать визуальную ширину |
| `FindElement(...)` | поиск в `[]Element` |

## Встраивание

```go
inner := warp.New()
inner.SetRoot(innerTabGroup)
outerTab.SplitVertical(parent, 0.5, inner.AsPanel())
```

`AsPanel()` возвращает адаптер, который сам вызывает
`View`/`Update` с передаваемыми размерами.

## Пример: flex-форма

```go
w := warp.New()
tab := w.ActiveTab()

name := warp.NewInput("Name: ")
email := warp.NewInput("Email: ")
preview := &statusPanel{}

tab.FlexRow(tab.RootPanel(), []warp.FlexItemSpec{
    {Panel: name,  Grow: 1},
    {Panel: email, Grow: 1},
})
tab.Float(preview, 10, 4, 24, 8)

w.Run()
```
