---
title: Tab
description: API вкладки для компоновки, float-панелей, фокуса, размеров split и сворачивания.
---

# Tab

`Tab` владеет одним деревом панелей, его float-панелями и текущим фокусом.

## Создание и корень

```go
func NewTab(name string) *Tab
func (t *Tab) RootPanel() Panel
func (t *Tab) SetRootPanel(panel Panel)
```

`NewTab` создаёт самостоятельную вкладку. `TabGroup.NewTab` создаёт вкладку, связанную с группой.

## Компоновка

```go
func (t *Tab) SplitVertical(parent Panel, fraction float64, newPanel Panel)
func (t *Tab) SplitHorizontal(parent Panel, fraction float64, newPanel Panel)
func (t *Tab) FlexRow(parent Panel, items []FlexItemSpec)
func (t *Tab) FlexColumn(parent Panel, items []FlexItemSpec)
```

Split заменяет `parent` узлом с двумя дочерними узлами. Доля ограничивается диапазоном `0.1..0.9`. Flex-методы игнорируют пустой список.

## Размер и сворачивание split

```go
func (t *Tab) SetSplitFraction(panel Panel, fraction float64) bool
func (t *Tab) GetSplitFraction(panel Panel) (float64, bool)
func (t *Tab) Collapse(panel Panel, size int) bool
func (t *Tab) Expand(panel Panel) bool
func (t *Tab) SetSplitCollapse(parent Panel, collapseRow int, onCollapse func() tea.Cmd)
func (t *Tab) ToggleSplitCollapse(parent Panel)
```

Методы возвращают `false`, если соответствующий узел не найден. `SetSplitCollapse` задаёт callback для маркера сворачивания.

## Float-панели

```go
func (t *Tab) Float(panel Panel, x, y, width, height int)
func (t *Tab) CloseFloat(fp *FloatPane)
```

`Float` создаёт панель поверх основной компоновки. Float можно перемещать, изменять его размер и закрывать мышью.

## Фокус

```go
func (t *Tab) Focus() Panel
func (t *Tab) SetFocus(panel Panel) tea.Cmd
func (t *Tab) FocusNext()
func (t *Tab) FocusPrev()
func (t *Tab) FocusFirst()
func (t *Tab) FocusPanel(panel Panel)
```

Warp не назначает `Tab` и `Shift+Tab` автоматически. Биндинги задаёт приложение.

## Жизненный цикл Panel

```go
func (t *Tab) Update(msg tea.Msg) tea.Cmd
func (t *Tab) View(width, height int) string
func (t *Tab) HandleMouse(msg tea.MouseMsg) tea.Cmd
func (t *Tab) Elements(width, height int) []Element
func (t *Tab) BroadcastResize() tea.Msg
```

`Elements` отдаёт семантическое дерево UI вкладки. `BroadcastResize` создаёт сообщение о размере текущей компоновки.

Контекстное меню создаётся отдельно через `Popover`; у `Tab` нет фабрики контекстного меню.
