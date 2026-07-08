# Паттерны проектирования

## Композиция вместо наследования

```go
// Компоненты — это Panel, вложенные в дерево Node
type Collapsible struct {
    Title     string
    Collapsed bool
    Content   Panel  // вложенная панель
}
```

## TabGroup как Panel

`TabGroup` реализует `Panel` — может быть вложен в splits/flex как обычный компонент:

```go
tg := warp.NewTabGroup(warp.TabLeft)
tab.FlexRow(root, []warp.FlexItemSpec{
    {Panel: tg, Grow: 2},  // TabGroup внутри flex
})
```

## Focus — явный API

Warp не биндит Tab/Shift+Tab. Разработчик сам решает:

```go
// В appRoot.Update:
case "tab":
    tab.FocusNext()
case "shift+tab":
    tab.FocusPrev()
```

## Float z-order

Float'ы хранятся в порядке создания. Клик по float'у поднимает его наверх.
CloseOnOutsideClick закрывает popover/context menu при клике вне.

## ANSI-aware overlay

Float/Modal/Popover рендерятся поверх существующих строк с сохранением ANSI-кодов.
`StripANSI` для подсчёта визуальной ширины, CSI-aware позиционирование.
