# Patterns

> Полная спецификация: [`specs/patterns.md`](https://github.com/starframe-dev/warp/blob/main/specs/patterns.md)

## TabGroup как Panel

```go
tg := warp.NewTabGroup(warp.TabLeft)
tab.FlexRow(root, []warp.FlexItemSpec{
    {Panel: tg, Grow: 2},
})
```

## Focus — явный API

Warp не биндит Tab/Shift+Tab. Разработчик сам решает.

## Float z-order

Клик по float'у поднимает его наверх. CloseOnOutsideClick закрывает popover.
