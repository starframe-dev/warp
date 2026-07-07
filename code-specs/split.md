# Split

## Роль

Внутренний узел дерева панелей для разделения пространства между двумя дочерними узлами (Split/Flex). Используется для создания гибких сеточных макетов в Warp UI.

## API

### Типы

#### Direction

```go
type Direction int
```

Указывает направление разделения:

| Константа | Значение | Описание |
|-----------|----------|----------|
| Vertical  | 0        | Вертикальное деление (слева/справа) |
| Horizontal| 1        | Горизонтальное деление (сверху/снизу) |

#### SplitConfig

```go
type SplitConfig struct {
    Direction Direction
    Fraction  float64 // Доля первого дочернего узла (0.0–1.0)
    First     *Node
    Second    *Node
    Dragging  bool
}
```

**Поля:**

| Поле | Тип | Описание |
|------|-----|----------|
| Direction | Direction | Направление разделения |
| Fraction | float64 | Доля пространства для первого узла |
| First | *Node | Первый дочерний узел |
| Second | *Node | Второй дочерний узел |
| Dragging | bool | true во время drag-and-drop |

#### Node

```go
type Node struct {
    Panel    Panel
    Split    *SplitConfig
    Flex     *FlexConfig
    Collapse *NodeCollapse
}
```

**Поля:**

| Поле | Тип | Описание |
|------|-----|----------|
| Panel | Panel | Лист дерева (терминальный узел) |
| Split | *SplitConfig | Разделение пространства |
| Flex | *FlexConfig | Гибкая верстка |
| Collapse | *NodeCollapse | Свёртывание узла |

#### NodeCollapse

```go
type NodeCollapse struct {
    Active bool
    Width  int
    Height int
    Saved  float64
}
```

**Поля:**

| Поле | Тип | Описание |
|------|-----|----------|
| Active | bool | true — узел свернут |
| Width | int | Фиксированная ширина при свёртке (для Vertical) |
| Height | int | Фиксированная высота при свёртке (для Horizontal) |
| Saved | float64 | Сохранённая доля для восстановления при развёртке |

#### ResizeMsg

```go
type ResizeMsg struct {
    Width  int
    Height int
}
```

**Поле:**

| Поле | Тип | Описание |
|------|-----|----------|
| Width | int | Ширина панели в ячейках |
| Height | int | Высота панели в ячейках |

#### FlexItem

```go
type FlexItem struct {
    Node      *Node
    Grow      int
    Shrink    int
    Basis     int
    Collapsed bool
}
```

**Поля:**

| Поле | Тип | Описание |
|------|-----|----------|
| Node | *Node | Узел элемента |
| Grow | int | Вес flex-grow |
| Shrink | int | Вес flex-shrink (не используется) |
| Basis | int | Flex-basis, 0 = auto |
| Collapsed | bool | true — элемент занимает 1 линию/символ |

## Поведение

### findNode

```go
func (n *Node) findNode(panel Panel) *Node
```

Локализирует узел, содержащий указанный Panel в дереве.

**Возвращает:**
- `*Node` — найденный узел, если Panel найден
- `nil` — если Panel не найден

**Алгоритм:**
1. Проверяет текущий узел на равенство Panel
2. Если это Split — рекурсивно ищет в First и Second
3. Если это Flex — ищет во всех Items
4. Возвращает первый найденный узел или nil

### replaceNode

```go
func (n *Node) replaceNode(old, new *Node) bool
```

Заменяет oldNode на newNode в дереве.

**Возвращает:**
- `true` — замена успешна
- `false` — oldNode не найден

**Алгоритм:**
1. Проверяет Split.First и Split.Second
2. Рекурсивно ищет в поддеревах Split
3. Ищет в Items Flex
4. Возвращает результат первой успешной замены

### collectLeafNodes

```go
func (n *Node) collectLeafNodes() []*Node
```

Собирает все листы (Panel) в порядке глубинного обхода.

**Возвращает:**
- `[]*Node` — срез указателей на листы

**Алгоритм:**
1. Если узел — лист, возвращает срез из одного элемента
2. Если Split — собирает из First, затем из Second
3. Если Flex — собирает из всех Items

### IsLeaf

```go
func (n *Node) IsLeaf() bool
```

Проверяет, является ли узел листом (содержит Panel).

**Возвращает:**
- `true` — узел содержит Panel
- `false` — узел внутренний (Split/Flex)

### IsCollapsed

```go
func (n *Node) IsCollapsed() bool
```

Проверяет, свернут ли узел.

**Возвращает:**
- `true` — Collapse активен
- `false` — Collapse не активен или отсутствует

### CollapsedSize

```go
func (n *Node) CollapsedSize(d Direction) int
```

Возвращает размер узла в свернутом состоянии для заданного направления.

**Параметры:**
- `d` — Direction (Vertical/Horizontal)

**Возвращает:**
- `int` — размер (Width для Vertical, Height для Horizontal), 0 если не свернут

## Реализация

### Структура данных

Дерево панелей представляет собой иерархическую структуру:
- **Node** — корневой или промежуточный узел
- **Split** — делит пространство на две части
- **Flex** — гибкая верстка с весами
- **Panel** — терминальный узел с контентом

### Иерархия

```
Node (root)
├─ Split
│  ├─ First (Node)
│  └─ Second (Node)
└─ Flex
   └─ Items []*FlexItem
      └─ Node (First/Second)
```

### Константы

| Константа | Значение | Описание |
|-----------|----------|----------|
| MinPanelSize | 3 | Минимальный размер панели в символах |

## Примечания

- `Dragging` в SplitConfig используется только во время drag-and-drop операций
- `Fraction` не может выходить за границы 0.0–1.0
- `Node` может содержать только один тип внутреннего узла (Split, Flex, Collapse, Panel — не более одного одновременно)
