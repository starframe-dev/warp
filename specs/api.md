# Публичный API проекта Warp

## Обзор

Библиотека `warp` — Go TUI layout engine на базе `charmbracelet/bubbletea`.
Управляет расположением панелей: вкладки, сплиты, flexbox, плавающие окна, модалки, popover.

## Пакет warp

### Корневая модель Warp

```go
type Warp struct {
    root   Panel
    width  int
    height int
}
```

#### Конструкторы

| Функция | Возврат | Описание |
|---------|---------|----------|
| `New()` | `*Warp` | Создаёт Warp с корневой `TabGroup` |

#### Методы

| Метод | Описание |
|-------|----------|
| `SetRoot(panel Panel)` | Заменяет корневую панель |
| `Root() Panel` | Возвращает корневую панель |
| `NewTab(name string) *Tab` | Создаёт вкладку (делегирует TabGroup) |
| `ActiveTab() *Tab` | Активная вкладка |
| `SetTabPosition(pos TabPosition)` | Позиция таб-бара |
| `NextTab()` / `PrevTab()` | Переключение вкладок |
| `AsPanel() Panel` | Адаптер для вложенных warp |
| `Run() error` | Запуск Bubbletea программы |

### Тип Panel

```go
type Panel interface {
    View(width, height int) string
    Update(msg tea.Msg) tea.Cmd
}
```

### TabGroup

```go
tg := warp.NewTabGroup(warp.TabTop)
tg.NewTab("editor")
tg.NewTab("terminal")
tab := tg.ActiveTab()
```

### Tab

```go
tab.SplitVertical(parent, 0.5, newPanel)
tab.SplitHorizontal(parent, 0.5, newPanel)
tab.FlexRow(parent, []warp.FlexItemSpec{...})
tab.FlexColumn(parent, []warp.FlexItemSpec{...})
tab.Float(panel, x, y, w, h)
tab.FocusNext()
tab.FocusPrev()
tab.FocusFirst()
tab.FocusPanel(panel)
tab.ToggleCollapsible(panel)
```

### Компоненты

| Компонент | Конструктор | Описание |
|-----------|-------------|----------|
| Collapsible | `NewCollapsible(title, panel)` | Сворачиваемая панель |
| Scrollable | `NewScrollable(panel)` | Скроллируемая панель |
| DropdownMenu | `NewDropdownMenu(label, items)` | Выпадающее меню |
| Selectable | `NewSelectable(panel)` | Выделение текста |
| Input | `NewInput(prompt)` | Текстовый ввод |
| Modal | `NewModal(title, content, buttons, onClose)` | Модальное окно |
| Popover | `&warp.Popover{Items, X, Y, OnClose}` | Контекстное меню |

### Focusable

```go
type Focusable interface {
    Panel
    Focus()
    Blur()
    Focused() bool
}
```

### ElementProvider

```go
type ElementProvider interface {
    Elements(width, height int) []Element
}
```

### Вспомогательные функции

| Функция | Описание |
|---------|----------|
| `WordWrap(text, width)` | Перенос по словам |
| `SpaceWrap(text, width)` | Перенос по пробелам |
| `StripANSI(s string)` | Удаление ANSI-кодов |
| `FindElement(elems, role, name, action)` | Поиск в дереве элементов |

## Примеры

```go
w := warp.New()
tab := w.ActiveTab()

name := warp.NewInput("Name: ")
email := warp.NewInput("Email: ")
tab.FlexRow(tab.RootPanel(), []warp.FlexItemSpec{
    {Panel: name, Grow: 1},
    {Panel: email, Grow: 1},
})

tab.Float(&demoPanel{name: "Float"}, 10, 4, 24, 6)
w.Run()
```
