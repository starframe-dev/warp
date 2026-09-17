# split.go

Модуль `split` (пакет `warp`) определяет структурные типы для
иерархического дерева раскладки панелей. Узлы являются либо листовыми
панелями, либо внутренними контейнерами split/flex. Модуль содержит
модель данных раскладки, но не содержит логики её рендеринга или
обработки нажатий.

## Публичный API

### `Direction`

``` go
type Direction int

const (
    Vertical   Direction = iota // side by side (left/right)
    Horizontal                  // stacked (top/bottom)
)
```

Целочисленный тип (`Direction`), определяющий ориентацию разделения в
`SplitConfig` или `FlexConfig`. `Vertical` располагает дочерние элементы
рядом, а `Horizontal` — один под другим сверху вниз. Значение хранится
непосредственно как `Direction` в структурах конфигурации, а не как
указатель.

### `ResizeMsg`

``` go
type ResizeMsg struct {
    Width  int
    Height int
}
```

Тип сообщения, который движок раскладки отправляет панели, когда её
выделенный прямоугольник изменяется. `Width` и `Height` измеряются в
клетках, без границ и отступов. Это команда bubbletea (`tea.Cmd`),
поэтому она участвует в цикле обработки сообщений в каждом кадре
работающего приложения.

### `SplitConfig`

``` go
type SplitConfig struct {
    Direction   Direction
    Fraction    float64   // share of First (0.0..1.0)
    First       *Node
    Second      *Node
    Dragging   bool       // true during drag-and-drop
    CollapseRow  int     // border row showing "<"
    OnCollapse   func() tea.Cmd
}
```

Внутренний узел, распределяющий свою область между двумя дочерними
элементами. `Fraction` — доля первого *первого* дочернего элемента,
ограниченная диапазоном \[0.0, 1.0\]. `CollapseRow` — номер строки (с
нуля), в которой граница рендерит символ сворачивания (“\|” или “\<”);
клик по нему вызывает `OnCollapse`, который возвращает `tea.Cmd` для
повторного рендеринга.

### `NodeCollapse`

``` go
type NodeCollapse struct {
    Active  bool
    Width   int  // fixed width when collapsed (vertical layouts)
    Height  int  // fixed height when collapsed (horizontal layouts)
    Saved   float64 // fraction to restore on expand
}
```

Хранит состояние сворачивания узла. Когда `Active` равно true, узел
рендерится с фиксированными `Width`/`Height`; `Saved` сохраняет долю до
сворачивания для последующего восстановления.

### `Node`

``` go
type Node struct {
    Panel    Panel
    Split    *SplitConfig
    Flex     *FlexConfig
    Collapse *NodeCollapse
}
```

Универсальный узел дерева панелей. Ровно один из `Panel`, `Split`,
`Flex` установлен: у листа `Panel` не nil и внутренняя конфигурация nil;
у внутреннего узела один из `Split`/`Flex` и `Panel` nil. `Collapse`
присутствует в любом узле, который поддерживает
сворачивание/разворачивание.

### Методы `Node`

| Подпись | Описание |
|----|----|
| `func (n *Node) IsLeaf() bool` | Возвращает `true`, если `n.Panel` не nil (листовой терминальный узел). |
| `func (n *Node) IsCollapsed() bool` | Возвращает `true`, если узел в данный момент находится в свёрнутом состоянии (`Collapse.Active == true`). |
| `func (n *Node) CollapsedSize(d Direction) int` | Возвращает свёрнутый размер вдоль направления `d`, или 0, если узел не свёрнут. |

Оба `IsLeaf` и `IsCollapsed` — дешёвые проверки полей; они являются
основными точками входа для обхода дерева. `CollapsedSize` возвращает
сохранённое измерение (или 1 как минимум), чтобы движок раскладки мог
запросить единообразный размер, независимо от того, свёрнут узел в
данный момент или нет.

### `FlexItem` / `FlexConfig`

``` go
type FlexItem struct {
    Node      *Node
    Grow      int  // flex-grow weight
    Shrink    int  // flex-shrink (currently unused)
    Basis     int  // flex-basis (min size); 0 = auto
    Collapsed bool
}

type FlexConfig struct {
    Direction Direction
    Items     []*FlexItem
    Dragging  bool
}
```

Взвешенная раскладка по строкам/столбцам. Каждый элемент оборачивает
`Node` и хранит его веса grow/shrink/basis, а также локальный флаг
`Collapsed`. `FlexConfig` располагает свои дочерние элементы в одну
строку или столбец с пропорциональным распределением весов. `Shrink`
объявлен, но намеренно не используется (сжатие не реализовано в текущей
версии раскладки).

## Поведение

Дерево является *бинарным* или *мульти-дочерним* деревом в зависимости
от того, какой внутренний узел используется: у `SplitConfig` всегда
ровно два дочерних элемента, у `FlexConfig` — упорядоченный список
`FlexItem` (каждый оборачивает `Node`). `Node` является либо листом
(`Panel != nil`), либо внутренним контейнером (`Split != nil` или
`Flex != nil`), никогда не обоими сразу.

Размеры зависят от направления: `SplitConfig` и `FlexConfig` используют
одно `Direction` для выбора оси разделения; свёрнутые размеры
запрашиваются через `CollapsedSize(d)`, так что вызывающая сторона может
запросить любую ось единообразно.

Состояние перетаскивания отслеживается на уровне конфигурации через
`Dragging bool`; установка в `true` приостанавливает пересчёт раскладки
до прихода следующего сообщения изменения размера.

## Детали реализации

- `findNode`, `findSplitParent`, `replaceNode` и `collectLeafNodes` —
  рекурсивные обходчики, определённые как методы на `*Node`; они
  обрабатывают дочерние элементы как `SplitConfig`, так и `FlexConfig`.
- `findNode` возвращает листовой узел, соответствующий заданному
  `Panel`; `findSplitParent` возвращает родительский `Split` узел, чьи
  два дочерних элемента являются непосредственными родителями целевой
  панели; `replaceNode` переписывает указатель на месте и возвращает
  флаг успеха; `collectLeafNodes` возвращает все листовые `*Node` в
  порядке обхода (split слева направо / flex сверху вниз).
- Все обходчики защищают от nil-приёмника, поэтому их можно безопасно
  вызывать на nil `*Node`.
