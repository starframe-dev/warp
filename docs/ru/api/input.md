---
title: Input
description: Однострочный текстовый ввод с курсором и фокусом.
---

# Input

`Input` — однострочная панель с prompt, редактируемым значением и курсором по рунам.

## Конструктор

```go
func NewInput(prompt string) *Input
```

## Публичные поля

```go
type Input struct {
    Value  string
    Cursor int
    Prompt string
    Width  int
}
```

`Cursor` хранит индекс руны. При `Width == 0` `View` использует доступную ширину.

## Методы

```go
func (in *Input) SetValue(value string)
func (in *Input) SetCursor(position int)
func (in *Input) Focus()
func (in *Input) Blur()
func (in *Input) Focused() bool
func (in *Input) View(width, height int) string
func (in *Input) Update(msg tea.Msg) tea.Cmd
```

`Input` реализует `Panel` и `Focusable`.

## Клавиши редактирования

- Текст вставляется на позиции курсора.
- `Backspace` удаляет руну перед курсором.
- `Delete` удаляет руну после курсора.
- Стрелки перемещают курсор.
- `Home` и `End` переходят в начало и конец строки.
