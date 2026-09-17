# Управление состоянием

## Локальное состояние

Каждый компонент хранит своё состояние в полях:

```go
type Input struct {
    Value   string
    Cursor  int
    Prompt  string
    focused bool
}

type Scrollable struct {
    Content Panel
    Offset  int
}

type Selectable struct {
    Content   Panel
    Selection Selection // (start, end)
}
```

Внешний мир видит компонент только через `Panel`,
`Focusable`, `ElementProvider`.

## Состояние через Bubbletea

Warp использует стандартный Bubbletea-цикл:

- `Update(msg tea.Msg) tea.Cmd` — обработка сообщения,
  изменение состояния.
- `View(w, h int) string` — рендеринг текущего состояния.

Ключевые сообщения:

| Сообщение | Назначение |
|-----------|------------|
| `tea.KeyMsg` | клавиатура |
| `tea.MouseMsg` | мышь |
| `tea.WindowSizeMsg` | изменение окна |
| `ResizeMsg` | изменение размера панели |
| `ShowModalMsg` / `CloseModalMsg` | модалки |

## Drag-состояние

`Tab` хранит **активный** drag:

```go
dragging *SplitConfig // текущий split, если drag активен
flexDragging *FlexConfig // текущий flex
lastBorders []BorderHit // hit-зона после последнего рендера
```

Drag не меняет состояние компонента, а **временное**
состояние `Tab`; граница рисуется в `borderDragStyle`
пока `Dragging == true`.

## Float-состояние

Каждый `FloatPane` сам хранит:

- `dragging bool` — текущий drag
- `dragOffsetX, dragOffsetY int` — смещение курсора
- `closeOnOutsideClick bool`
- `X, Y, Width, Height`

Z-order = порядок в `Tab.floats []*FloatPane`.

## Resize через `ResizeMsg`

Когда размер панели изменился, Warp отправляет
`ResizeMsg {Width, Height}` через `tea.Msg`.
Панель может перерисовать кэшированные
строки (в `Scrollable`, `Input` и т.п. используется).

## Нет глобального состояния

Warp не использует:

- глобальных переменных состояния
- подписок
- эмиттеров / event-buses
- singleton'ов

Компоненты изолированы. Коммуникация — только
через `tea.Msg`.
