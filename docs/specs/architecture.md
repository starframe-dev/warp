# Architecture

> Полная спецификация: [`specs/arhitektura.md`](https://github.com/starframe-dev/warp/blob/main/specs/arhitektura.md)

## Дерево панелей

```
Node { Panel | Split | Flex }
SplitConfig { Direction, Fraction, First, Second }
FlexConfig  { Direction, Items[] }
```

## Рендеринг

1. `renderNode` — рекурсивно рендерит дерево
2. `overlayFloat` — ANSI-aware наложение float поверх
3. `padContent` — обрезка/дополнение до точных размеров

## Фокус

- `Focusable` — Panel с Focus/Blur/Focused
- `Tab.FocusNext()/FocusPrev()` — явное переключение
- Warp не биндит Tab/Shift+Tab
