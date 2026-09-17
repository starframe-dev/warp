---
title: Selectable
description: Выделение текста, получение выделенного текста и копирование через OSC 52.
---

# Selectable

`Selectable` оборачивает панель и хранит выделение в координатах терминальных ячеек.

## Конструктор и поля

```go
func NewSelectable(content Panel) *Selectable
```

Публичные поля выделения: `AnchorX`, `AnchorY`, `CursorX`, `CursorY`, `HasSelection` и `Selecting`.

## Методы выделения

```go
func (s *Selectable) SelectAll(width, height int)
func (s *Selectable) ClearSelection()
func (s *Selectable) SelectedText() string
func (s *Selectable) Copy() tea.Cmd
```

`Copy` возвращает команду, которая отправляет выделенный текст через OSC 52. Операционная система не вызывается напрямую.

## Управление

- выделение мышью перетаскиванием;
- `Shift` + стрелки расширяют выделение;
- `Ctrl+A` выделяет весь видимый контент;
- `Esc` очищает выделение.

## Panel

```go
func (s *Selectable) View(w, h int) string
func (s *Selectable) Update(msg tea.Msg) tea.Cmd
```
