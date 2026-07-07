# Specification: Input Component

## Overview

`Input` — это компонент однострочного текстового ввода для TUI-приложений на базе Bubble Tea. Компонент предоставляет интерактивное поле ввода с поддержкой курсора, фокуса и обработки клавиатуры.

## Purpose

Предоставляет удобную основу для создания полей ввода в TUI-интерфейсах с полной поддержкой:
- Отображения текста с подсветкой курсора
- Обработки ввода с клавиатуры
- Поддержки фокуса и размытия
- Отрисовки в боксе или inline-режиме

## Public API

### Constructor

```go
func NewInput(prompt string) *Input
```

Создаёт новый экземпляр `Input` с заданным текстом-подсказкой.

**Параметры:**
- `prompt` — текст-подсказка, отображаемый перед значением

**Поведение:**
- Создаётся пустой экземпляр с пустым значением
- Курсор установлен в начале (позиция 0)
- `Prompt` установлен на переданный параметр

### Mutator Methods

#### SetValue

```go
func (in *Input) SetValue(v string)
```

Замениает значение ввода и перемещает курсор в конец.

**Параметры:**
- `v` — новое значение

**Поведение:**
1. Устанавливает `Value` на переданный параметр
2. Устанавливает `Cursor` в длину строки в символах (runes)
3. Вызывает `clampCursor()` для валидации диапазона

#### SetCursor

```go
func (in *Input) SetCursor(pos int)
```

Устанавливает позицию курсора в символах.

**Параметры:**
- `pos` — позиция курсора в символах

**Поведение:**
- Устанавливает `Cursor` на переданный параметр
- Вызывает `clampCursor()` для валидации диапазона

### State Queries

#### Focused

```go
func (in *Input) Focused() bool
```

Возвращает состояние фокуса компонента.

**Возвращает:**
- `true` — если компонент имеет фокус
- `false` — иначе

### Focus Management

#### Focus

```go
func (in *Input) Focus()
```

Давает фокус компоненту.

**Поведение:**
- Устанавливает `focused` на `true`

#### Blur

```go
func (in *Input) Blur()
```

Убирает фокус с компонента.

**Поведение:**
- Устанавливает `focused` на `false`

## Rendering API

### View

```go
func (in *Input) View(w, h int) string
```

Отрисовывает компонент.

**Параметры:**
- `w` — доступная ширина в символах
- `h` — доступная высота в строках

**Поведение:**
- Если `h >= 3`: отрисовывает в боксе (с рамкой)
- Иначе: отрисовывает inline (просто текст)
- Возвращает строку с отрисованным содержимым

### View Modes

#### Boxed Mode (h >= 3)

Отрисовывает компонент внутри рамки:
- Верхняя граница: `╭───╮`
- Нижняя граница: `╰───╯`
- Боковые границы: `│`
- Контент центрирован вертикально

#### Inline Mode (h < 3)

Отрисовывает просто как строку текста без рамки.

## Rendering Details

### renderLine

```go
func (in *Input) renderLine(maxW int) string
```

Строит строку с подсказкой, значением и подсветкой курсора.

**Поведение:**
1. Добавляет текст-подсказку (`Prompt`)
2. Обрезает значение, если не влезает в `maxW`
3. Подсвечивает символ под курсором ANSI-кодом `\x1b[7m`
4. Добавляет курсор после значения, если он находится в пределах строки

### truncateTailToWidth

```go
func truncateTailToWidth(s string, maxW, cursor int) string
```

Обрезает строку так, чтобы курсор оставался видимым.

**Стратегия:**
- Если `cursor < maxW`: берёт первые `maxW` символов
- Иначе: центрирует содержимое вокруг курсора

## Update API

### Update

```go
func (in *Input) Update(msg tea.Msg) tea.Cmd
```

Обработчик ввода с клавиатуры (Bubble Tea).

**Параметры:**
- `msg` — сообщение типа `tea.KeyMsg`

**Возвращает:**
- `tea.Cmd` — nil (команды не возвращаются)

**Поведение:**
- Игнорирует сообщения, если компонент не в фокусе
- Обрабатывает:
  - `backspace` — удаляет символ перед курсором
  - `delete` — удаляет символ под курсором
  - `left`/`right` — перемещение курсора
  - `home`/`end` — перемещение к началу/концу
  - `enter` — отправка (submit)
  - Табы — передача родительскому компоненту
  - Любые одиночные символы — вставка
- Клавиши управления курсором ограничены границами значения

### Insert

```go
func (in *Input) insertAtCursor(s string)
```

Вставляет текст в позицию курсора.

**Поведение:**
1. Преобразует `Value` в slice `rune`
2. Вставляет символы после курсора
3. Конвертирует обратно в строку
4. Сдвигает курсор на длину вставленного текста

### Delete Operations

#### Delete Before Cursor

```go
func (in *Input) deleteBeforeCursor()
```

Удаляет символ перед курсором.

**Поведение:**
1. Преобразует `Value` в slice `rune`
2. Удаляет символ на позиции `Cursor-1`
3. Сдвигает `Cursor` на `-1`

#### Delete At Cursor

```go
func (in *Input) deleteAtCursor()
```

Удаляет символ под курсором.

**Поведение:**
1. Преобразует `Value` в slice `rune`
2. Удаляет символ на позиции `Cursor`

## Cursor Clamping

```go
func (in *Input) clampCursor()
```

Ограничивает курсор валидным диапазоном.

**Поведение:**
- `Cursor < 0` → `Cursor = 0`
- `Cursor > len(runes(Value))` → `Cursor = len(runes(Value))`

## Type Definition

```go
type Input struct {
    Value   string   // текст значения
    Cursor  int      // позиция курсора в символах
    Prompt  string   // текст-подсказка
    Width   int      // желаемая ширина (0 = авто)
    focused bool     // состояние фокуса
}
```

## Private Methods

### viewBoxed

```go
func (in *Input) viewBoxed(w, h int) string
```

Отрисовывает компонент в режиме бокса.

**Использует:**
- `inputBorderStyle` для нефокусированного состояния
- `inputFocusBorderStyle` для фокусированного состояния
- `inputStyle` для контента

### viewInline

```go
func (in *Input) viewInline(w, h int) string
```

Отрисовывает компонент в inline-режиме.

## Styles

```go
var inputStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color(gbLight1))
var inputBorderStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(gbDark4))
var inputFocusBorderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(gbBlue))
```

- `inputStyle` — стиль контента
- `inputBorderStyle` — стиль рамки (нефокус)
- `inputFocusBorderStyle` — стиль рамки (фокус)

## Invariants

1. **Cursor всегда валиден:** `0 <= Cursor <= len(runes(Value))`
2. **Focused изменяет стиль рамки**
3. **View всегда возвращает строку**
4. **Update игнорирует не-keyMsg и нефокусные сообщения**
5. **Вставка ограничена доступной шириной**

## Constraints

- Не поддерживается мультисимвольный ввод за один раз (только по одному символу)
- Обрезание при переполнении: приоритет у курсора
- Фокус не сохраняется между рендерами (внешний контроль)

## Usage Pattern

```go
input := NewInput("Name:")
input.Focus()

// Обработка событий
func keyPress(msg tea.Msg) {
    cmd := input.Update(msg)
    if cmd != nil {
        // handle custom commands
    }
}

// Рендеринг
func view(w, h int) string {
    return input.View(w, h)
}

// Управление состоянием
input.SetValue("Alice")
input.SetCursor(5)
```

## TDD Considerations

Тестируемые аспекты:
- Обработка разных клавиш
- Ограничение курсора
- Отрисовка в разных режимах
- Обрезание при переполнении

Требуются моки для:
- Bubble Tea messaging
- Липгloss стилей (можно тестировать логику без отрисовки)

## References

- Bubble Tea: https://github.com/charmbracelet/bubbletea
- Lip Gloss: https://github.com/charmbracelet/lipgloss
- Human Horizon Go Code Style: `guides/common/Стиль_Кода_Go.md`
- Human Horizon Code Documentation: `guides/common/Документация_Кода.md`
