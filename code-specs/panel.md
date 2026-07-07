# Specification: panel.go

## Overview

`panel.go` определяет интерфейс и базовую реализацию для панелей в Warp приложении. Панели являются основным строительным блоком для создания контента в терминальных интерфейсах Warp.

## Purpose

Панели позволяют создавать разнообразный контент для терминальных панелей: терминал, текст, графики, формы и другие компоненты. Интерфейс следует паттерну Model/Update/View из библиотеки Bubble Tea.

## Types

### Panel Interface

```go
type Panel interface {
    View(width, height int) string
    Update(msg tea.Msg) tea.Cmd
}
```

**Поля интерфейса:**

| Метод | Описание |
|-------|----------|
| `View(width, height int) string` | Рендерит содержимое панели на заданных размерах. Возвращает строковое представление. |
| `Update(msg tea.Msg) tea.Cmd` | Обрабатывает сообщения Bubble Tea (клавиши, мышь и т.д.). Возвращает команду для асинхронного выполнения. |

**Особенности:**

- Панель получает только сообщения, которые пришли в момент, когда панель была сфокусирована.
- `View` принимает размеры экрана и должен рендерить контент соответствующего размера.
- `Update` обрабатывает события ввода и состояния.

### BasePanel Struct

```go
type BasePanel struct{}

func (BasePanel) View(width, height int) string {
    return ""
}

func (BasePanel) Update(msg tea.Msg) tea.Cmd {
    return nil
}
```

**Назначение:**

- Предоставляет тривиальную реализацию `Panel` интерфейса.
- Встраивайте `BasePanel` в свои панели и переопределяйте только нужные методы.
- Это позволяет избежать дублирования методов, которые вы не используете.

**Пример использования:**

```go
type MyPanel struct {
    BasePanel // Embed base implementation
    state     MyState
}

func (m *MyPanel) View(width, height int) string {
    // Custom view implementation
    return m.BasePanel.View(width, height)
}

func (m *MyPanel) Update(msg tea.Msg) tea.Cmd {
    // Custom update implementation
    return m.BasePanel.Update(msg)
}
```

## Usage Patterns

### 1. Простая текстовая панель

```go
type SimpleTextPanel struct {
    BasePanel
    text string
}

func (p *SimpleTextPanel) View(width, height int) string {
    return p.text
}

func (p *SimpleTextPanel) Update(msg tea.Msg) tea.Cmd {
    // No custom handling
    return nil
}
```

### 2. Панель с состоянием

```go
type FormPanel struct {
    BasePanel
    fields []FormField
}

func (f *FormPanel) View(width, height int) string {
    // Render form fields
    var output strings.Builder
    for _, field := range f.fields {
        output.WriteString(field.Render())
    }
    return output.String()
}

func (f *FormPanel) Update(msg tea.Msg) tea.Cmd {
    // Handle form submissions
    return nil
}
```

### 3. Панель с дочерними панелями

```go
type CompositePanel struct {
    BasePanel
    children []Panel
}

func (c *CompositePanel) View(width, height int) string {
    var output strings.Builder
    for _, child := range c.children {
        output.WriteString(child.View(width, height))
    }
    return output.String()
}

func (c *CompositePanel) Update(msg tea.Msg) tea.Cmd {
    // Propagate message to children
    var cmds []tea.Cmd
    for _, child := range c.children {
        cmd := child.Update(msg)
        cmds = append(cmds, cmd)
    }
    return tea.Batch(cmds...)
}
```

## Guidelines

### Interface Design

- **Маленькие интерфейсы**: `Panel` содержит только два метода — минимальная необходимая поверхность.
- **Принимайте интерфейсы, возвращайте структуры**: Конкретные типы панелей лучше возвращать, чем абстрактные интерфейсы.

### Error Handling

- Всегда проверяйте ошибки при создании панелей.
- Оберните ошибки с контекстом через `fmt.Errorf("...: %w", err)`.
- Используйте sentinel-ошибки для особых случаев (`ErrNotFound`, `ErrInvalidInput`).

### Zero-initialization

Используйте нулевые значения для полей, не требующих инициализации:

```go
type MyPanel struct {
    BasePanel
    state     State
    focused   bool    // zero value is false
    callbacks []func() // zero value is nil
}
```

### Documentation

- Используйте английский язык в комментариях и документации.
- Документируйте публичные методы (экспортируемые).
- Добавляйте примеры использования в doc comments.

### Formatting

- Используйте `gofmt` или `gofumpt` перед каждым коммитом.
- Один табуляция для отступов.
- Максимум 120 символов в строке.
- Сортируйте импорты: std → external → internal.

### Testing

- Используйте table-driven tests для unit-тестов.
- Используйте subtests (`t.Run`) для группировки тестовых случаев.
- Используйте test helpers для повторяющихся операций.

## API Reference

### Panel Interface Methods

#### View(width, height int) string

Рендерит панель на заданных размерах.

**Параметры:**

- `width` — ширина отображения в символах
- `height` — высота отображения в строках

**Возвращаемое значение:**

- `string` — строковое представление панели, готовое к отрисовке

**Пример:**

```go
func (p *MyPanel) View(width, height int) string {
    if width == 0 || height == 0 {
        return ""
    }
    // render content
    return "content"
}
```

#### Update(msg tea.Msg) tea.Cmd

Обрабатывает сообщения Bubble Tea.

**Параметры:**

- `msg tea.Msg` — сообщение Bubble Tea (клавиши, мышь, таймеры и т.д.)

**Возвращаемое значение:**

- `tea.Cmd` — команда для асинхронного выполнения (или `nil`)

**Пример:**

```go
func (p *MyPanel) Update(msg tea.Msg) tea.Cmd {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c":
            return tea.Quit
        case "q":
            return tea.Quit
        }
    }
    return nil
}
```

### BasePanel Methods

#### View(width, height int) string

Возвращает пустую строку. Переопределите этот метод для кастомного рендеринга.

#### Update(msg tea.Msg) tea.Cmd

Возвращает `nil`, игнорируя все сообщения. Переопределите для обработки событий.

## Implementation Details

### Bubble Tea Integration

Панели используют библиотеку Bubble Tea для управления состоянием и обработкой событий.

**Импорт:**

```go
import tea "github.com/charmbracelet/bubbletea"
```

**Поток сообщений:**

1. Пользователь вводит данные → событие отправляется в модель панели
2. `Update` обрабатывает событие → возвращает команду
3. Команда выполняется асинхронно
4. Результат команды может обновить состояние
5. `View` рендерит новое состояние

### Command Execution

Возвращаемые `Update` команды выполняются в отдельной goroutine. Используйте `tea.Batch` для группировки нескольких команд:

```go
func (p *MyPanel) Update(msg tea.Msg) tea.Cmd {
    // Handle message
    cmds := []tea.Cmd{
        tea.Println("Hello"),
        p.saveData(),
    }
    return tea.Batch(cmds...)
}
```

## Code Examples

### Complete Panel Example

```go
package warp

import (
    "fmt"
    "strings"

    tea "github.com/charmbracelet/bubbletea"
)

type StatusPanel struct {
    BasePanel
    status string
    log    []string
}

type LogMsg string

func (p *StatusPanel) Init() tea.Cmd {
    return tea.Batch(
        tea.Println("Welcome to Warp"),
        p.logEntry("Panel initialized"),
    )
}

func (p *StatusPanel) View(width, height int) string {
    var output strings.Builder
    output.WriteString(fmt.Sprintf("Status: %s\n", p.status))
    output.WriteString("Log:\n")
    for _, entry := range p.log {
        output.WriteString(fmt.Sprintf("  - %s\n", entry))
    }
    return strings.Repeat("-", width) + "\n" + output.String() + strings.Repeat("-", width)
}

func (p *StatusPanel) Update(msg tea.Msg) tea.Cmd {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c":
            return tea.Quit
        case "q":
            return tea.Quit
        }
    case LogMsg:
        p.log = append(p.log, string(msg))
    }
    return nil
}

func (p *StatusPanel) logEntry(msg string) tea.Cmd {
    return func() {
        p.log = append(p.log, msg)
    }
}
```

## Rules

1. **Интерфейс Panel** содержит только два метода: `View` и `Update`.
2. **BasePanel** предоставляет тривиальную реализацию, которую можно встраивать.
3. **Панели должны обрабатывать ошибки** — не игнорируйте ошибки.
4. **Zero-value инициализация** — используйте zero values для полей.
5. **Английская документация** — все комментарии на английском.
6. **Маленькие интерфейсы** — не расширяйте `Panel` без необходимости.
7. **Принимайте интерфейсы** — не принимайте конкретные реализации.

## See Also

- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [UI TUI Specification](/Users/a/Space/Projects/HumanHorizon/code-specs/allowed/Библиотеки/UI TUI.md)
- [Go Style Guide](/Users/a/Space/Projects/HumanHorizon/code-specs/guides/common/Стиль_Кода_Go.md)
