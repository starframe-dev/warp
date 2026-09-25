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
2. Устанавливает `Cursor` в количество рун значения
3. Вызывает `clampCursor()` для валидации диапазона

#### SetCursor

```go
func (in *Input) SetCursor(pos int)
```

Устанавливает позицию курсора в рунах.

**Параметры:**
- `pos` — позиция курсора в рунах

**Поведение:**
- Ограничивает позицию диапазоном значения
- Если позиция попадает внутрь grapheme cluster, сдвигает курсор к его правой границе

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
- `w` — доступная ширина в терминальных ячейках
- `h` — доступная высота в строках

**Поведение:**
- Если `w >= 3` и `h >= 3`: отрисовывает в боксе (с рамкой)
- Иначе: отрисовывает inline (просто текст)
- Возвращает строку с отрисованным содержимым

### View Modes

#### Boxed Mode (w >= 3, h >= 3)

Отрисовывает компонент внутри рамки:
- Верхняя граница: `╭───╮`
- Нижняя граница: `╰───╯`
- Боковые границы: `│`
- Контент центрирован вертикально

#### Inline Mode (w < 3 or h < 3)

Отрисовывает просто как строку текста без рамки.

## Геометрия Unicode

- `Cursor` и публичные методы `SetCursor`/`SetValue` измеряют позицию в рунах.
- Рендеринг, обрезка и видимость курсора измеряются в терминальных ячейках с учётом ANSI, ширины CJK/emoji и объединённых graphemes.
- При обрезке широкая графема не делится на некорректные части; если курсор попадает внутрь графемы, курсорное выделение применяется ко всей графеме.
- `KeyRunes` вставляет все руны события, а не только первый символ.

## Rendering Details

### renderLine

```go
func (in *Input) renderLine(maxW int) string
```

Строит строку с подсказкой, значением и подсветкой курсора.

**Поведение:**
1. Добавляет текст-подсказку (`Prompt`)
2. Обрезает значение, если не влезает в `maxW`
3. Подсвечивает grapheme cluster под курсором ANSI-кодом `\x1b[7m`
4. Добавляет курсор после значения, если он находится в пределах строки

### truncateInputAtCursor

```go
func truncateInputAtCursor(value string, maxCells, cursor int) (string, int)
```

Обрезает значение по терминальным ячейкам, сохраняя курсор в видимой области. Графема, которая пересекает край видимого окна, заменяется пробелами, а не разрезается.

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
  - `backspace` — удаляет grapheme cluster перед курсором
  - `delete` — удаляет grapheme cluster под курсором
  - `left`/`right` — перемещение курсора
  - `home`/`end` — перемещение к началу/концу
  - `enter` — no-op placeholder (submit пока не реализован)
  - Табы — передача родительскому компоненту
  - `tea.KeyRunes` — вставляет все руны события, включая события с несколькими рунами
- Стрелки перемещают курсор по границам grapheme clusters; курсор остаётся в допустимых границах значения

### Insert

```go
func (in *Input) insertAtCursor(s string)
```

Вставляет текст в позицию курсора.

**Поведение:**
1. Преобразует `Value` и весь вставляемый текст в slice `rune`
2. Вставляет все руны события в позицию курсора
3. Конвертирует обратно в строку
4. Перемещает курсор после вставленного текста и нормализует его к границе grapheme cluster

### Grapheme-aware Editing

`Cursor` сохраняет публичную семантику позиции в рунах, но его допустимые позиции ограничены границами grapheme clusters. Стрелки влево/вправо переходят к предыдущей/следующей границе; Backspace/Delete удаляют целый grapheme cluster. `SetCursor` и позиции внутри кластера нормализуются к его правой границе. `Home`/`End` устанавливают начало/конец строки.

### Delete Operations

#### Delete Before Cursor

```go
func (in *Input) deleteBeforeCursor()
```

Удаляет grapheme cluster перед курсором.

**Поведение:**
1. Находит предыдущую границу grapheme cluster
2. Удаляет все руны между ней и `Cursor`
3. Перемещает `Cursor` к предыдущей границе

#### Delete At Cursor

```go
func (in *Input) deleteAtCursor()
```

Удаляет grapheme cluster под курсором.

**Поведение:**
1. Находит следующую границу grapheme cluster
2. Удаляет все руны между `Cursor` и этой границей

## Cursor Clamping

```go
func (in *Input) clampCursor()
```

Ограничивает курсор валидным диапазоном.

**Поведение:**
- `Cursor < 0` → `Cursor = 0`
- `Cursor > len(runes(Value))` → `Cursor = len(runes(Value))`
- Позиция внутри grapheme cluster → правая граница этого кластера

## Type Definition

```go
type Input struct {
    Value   string   // Input value
    Cursor  int      // Cursor position in runes
    Prompt  string   // Prompt text
    Width   int      // Desired width; zero uses the View width
    focused bool     // Focus state
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

1. **Cursor всегда валиден:** `0 <= Cursor <= len(runes(Value))` и находится на границе grapheme cluster
2. **Focused изменяет стиль рамки**
3. **View всегда возвращает строку**
4. **Update игнорирует не-keyMsg и нефокусные сообщения**
5. **Ширина View ограничивает только отображение, не `Value`**

## Constraints

- Одно событие `tea.KeyRunes` может вставить несколько рун
- Длина `Value` не ограничивается шириной рендера; при переполнении отображение выбирает ячейки вокруг курсора
- Фокус хранится в поле компонента до вызова `Blur`

## Usage Pattern

```go
input := NewInput("Name:")
input.Focus()

// Handle events
func keyPress(msg tea.Msg) {
    cmd := input.Update(msg)
    if cmd != nil {
        // handle custom commands
    }
}

// Render the input
func view(w, h int) string {
    return input.View(w, h)
}

// Manage input state
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
