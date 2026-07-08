---
title: Input
description: Текстовый ввод Warp с фокусом и редактированием строки.
---

# Input

`Input` — однострочный текстовый ввод.

## Создание

```go
NewInput(prompt string) *Input
```

Создаёт поле ввода с приглашением.

## Значение

```go
SetValue(v string)
Value() string
```

`SetValue` задаёт текст. `Value` возвращает текущий текст.

## Фокус

```go
Focus()
Blur()
Focused() bool
```

`Focus` включает фокус. `Blur` снимает фокус. `Focused` возвращает состояние фокуса.

## Интерфейсы

`Input` реализует `Focusable` и `Panel`.

## Управление

Поддерживаются курсор, `backspace`, `delete`, стрелки, `home` и `end`.
