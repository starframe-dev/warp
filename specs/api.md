# Публичный API проекта Warp

## Обзор

Этот документ описывает публичный API проекта `warp` — библиотеку UI-компонентов для терминальных приложений на базе `charmbracelet/bubbletea`. API спроектирован для простоты использования, расширяемости и безопасности.

## Пакет warp

### Корневая модель Warp

```go
type Warp struct {
    root   Panel
    width  int
    height int
    httpServer *http.Server
    httpAddr   string
    httpMu     sync.Mutex
}
```

#### Конструкторы

| Функция | Возврат | Описание |
|---------|---------|----------|
| `New()` | `*Warp` | Создаёт новый Warp с корневой `TabGroup` |

#### Методы управления корнем

| Метод | Возврат | Описание |
|-------|---------|----------|
| `SetRoot(panel Panel)` | `void` | Заменяет корневую панель |
| `Root() Panel` | `Panel` | Возвращает текущую корневую панель |

#### Комиссионные методы (делегирование TabGroup)

| Метод | Возврат | Описание |
|-------|---------|----------|
| `NewTab(name string)` | `*Tab` | Создаёт новую вкладку |
| `ActiveTab() *Tab` | `*Tab` | Возвращает активную вкладку |
| `SetTabPosition(pos TabPosition)` | `void` | Устанавливает позицию вкладок |
| `NextTab()` | `void` | Переходит к следующей вкладке |
| `PrevTab()` | `void` | Переходит к предыдущей вкладке |

#### Bubbletea методы

| Метод | Возврат | Описание |
|-------|---------|----------|
| `Init()` | `tea.Cmd` | Инициализация |
| `Update(msg tea.Msg)` | `(tea.Model, tea.Cmd)` | Обновление |
| `View()` | `string` | Рендеринг |

#### HTTP API

| Метод | Возврат | Описание |
|-------|---------|----------|
| `ServeHTTP(addr string)` | `error` | Запуск HTTP сервера |
| `CloseHTTP()` | `error` | Остановка HTTP сервера |
| `HTTPAddr()` | `string` | Адрес HTTP сервера |

#### Вложенность

| Метод | Возврат | Описание |
|-------|---------|----------|
| `AsPanel()` | `Panel` | Адаптер для вложенных warp |

### Тип Panel

Интерфейс панели для Bubbletea:

```go
type Panel interface {
    View(width, height int) string
    Update(msg tea.Msg) tea.Cmd
}
```

### Тип Element

Семантический UI-элемент:

```go
type Element struct {
    Role     string    `json:"role"`
    Name     string    `json:"name"`
    Action   string    `json:"action,omitempty"`
    Bounds   Bounds    `json:"bounds"`
    Children []Element `json:"children,omitempty"`
}
```

#### Вспомогательные функции Element

| Функция | Возврат | Описание |
|---------|---------|----------|
| `collectElements(panel Panel, width, height int)` | `[]Element` | Агрегация элементов |
| `FindElement(elems []Element, role, name, action string)` | `(Element, bool)` | Поиск элемента |
| `FindElement` (Bounds.Center) | `(int, int)` | Центр области |

### Тип Bounds

```go
type Bounds struct {
    X int `json:"x"`
    Y int `json:"y"`
    W int `json:"w"`
    H int `json:"h"`
}
```

### Типы компонентов

#### Tab

```go
type Tab struct {
    name    string
    root    *Node
    focused Panel
    floats  []*FloatPane
    parent  *TabGroup
    width   int
    height  int
}
```

#### TabGroup

```go
type TabGroup struct {
    tabs      []*Tab
    position  TabPosition
}
```

#### Modal

```go
type Modal struct {
    Title   string
    Content string
    Buttons []ModalButton
    OnClose func()
    Width   int
    startX  int
    startY  int
    boxWidth int
    boxHeight int
}
```

#### Popover

```go
type Popover struct {
    Items   []PopoverItem
    X, Y    int
    Width   int
    OnClose func()
    selected int
}
```

#### DropdownMenu

```go
type DropdownMenu struct {
    Label   string
    Items   []DropdownItem
    Open    bool
    Hovered int
    OnSelect func(idx int)
}
```

#### DropdownItem

```go
type DropdownItem struct {
    Label   string
    Selected bool
}
```

#### Scrollable

```go
type Scrollable struct {
    Content Panel
    Offset  int
}
```

#### FloatPane

```go
type FloatPane struct {
    Panel    Panel
    X, Y     int
    Width    int
    Height   int
    Title    string
    dragging bool
    resizing bool
    CloseRequested bool
}
```

#### Collapsible

```go
type Collapsible struct {
    Title     string
    Collapsed bool
    Content   Panel
}
```

#### Selectable

```go
type Selectable struct {
    Content       Panel
    AnchorX, AnchorY int
    CursorX, CursorY int
    HasSelection  bool
    Selecting      bool
    lastW, lastH   int
    lastLines       []string
}
```

#### Focusable (интерфейс)

```go
type Focusable interface {
    Panel
    Focus()
    Blur()
    Focused() bool
}
```

#### RawKeyReceiver (интерфейс)

```go
type RawKeyReceiver interface {
    Panel
    WantsRawKeys() bool
}
```

#### ElementProvider (интерфейс)

```go
type ElementProvider interface {
    Elements(width, height int) []Element
}
```

#### ElementProviderFunc

```go
type ElementProviderFunc func(width, height int) []Element
```

## Правила использования

1. **Использовать только публичный API** — не обращайтесь напрямую к внутренним полям
2. **Соблюдать контракт Panel** — реализовать View и Update
3. **Обрабатывать ошибки** — не игнорировать ошибки
4. **Использовать ElementProvider** — для экспорта элементов UI

## Примеры использования

### Создание базовой панели

```go
import "github.com/charmbracelet/warp"

// Создание Warp
w := warp.New()

// Установка корневой панели
w.SetRoot(somePanel)

// Запуск
err := w.Run()
if err != nil {
    log.Fatal(err)
}
```

### Создание модалки

```go
modal := warp.NewModal(
    "Заголовок",
    "Контент модалки",
    []warp.ModalButton{
        {Label: "ОК", Action: func() { /* ... */ }},
        {Label: "Отмена", Action: func() { /* ... */ }},
    },
    func() { /* очистка */ },
)
```

### Создание выпадающего списка

```go
items := []warp.DropdownItem{
    {Label: "Опция 1", Selected: false},
    {Label: "Опция 2", Selected: false},
}

dropdown := warp.NewDropdownMenu("Выбор", items)
```

### Плавающая панель

```go
fp := &warp.FloatPane{
    X: 0,
    Y: 10,
    Width: 80,
    Height: 24,
    Title: "Плавающая панель",
    Panel: somePanel,
}
```

### Сворачиваемая панель

```go
content := warp.NewPanel("Контент")
collapsible := warp.NewCollapsible("Заголовок", content)
```

## Безопасность

1. **Валидация ввода** — все входные данные валидируются
2. **Защита от XSS** — содержимое экранируется
3. **Безопасность событий** — события проверяются на валидность
4. **Изоляция состояний** — состояния компонентов изолированы

## Совместимость

- Требует `github.com/charmbracelet/bubbletea`
- Требует `github.com/charmbracelet/lipgloss`
- Требует `github.com/charmbracelet/x/ansi`

## Версионирование

- API не требует явного версионирования
- Изменения бэкворд-совместимы
- Сломывающие изменения вводят новую мажорную версию
