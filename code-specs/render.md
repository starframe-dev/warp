# Render Package Specification

## Описание

Пакет `render` отвечает за рендеринг дерева узлов (node tree) в строковые представления с поддержкой различных макетных паттернов: splits (вертикальные/горизонтальные), flex-контейнеры и leaf-панели.

## Публичный API

### Типы

#### BorderHit

```go
// BorderHit описывает позиционируемую границю (drag handle) для drag-and-drop операций.
type BorderHit struct {
    Split     *SplitConfig
    Flex      *FlexConfig
    Direction Direction
    X, Y      int   // Стартовая позиция границы (в ячейках)
    Length    int   // Длина границы в ячейках
}
```

#### Node

```go
// Node представляет узел дерева интерфейса.
type Node struct {
    Panel *Panel
    Split *SplitConfig
    Flex  *FlexConfig
    // ... внутренние поля
}

// IsLeaf возвращает true, если узел является листом (leaf node с Panel).
func (n *Node) IsLeaf() bool
// CollapsedSize возвращает размер при свернутом состоянии.
func (n *Node) CollapsedSize(direction Direction) int
// IsCollapsed проверяет состояние свёрнутости.
func (n *Node) IsCollapsed() bool
```

#### SplitConfig

```go
// SplitConfig конфигурирует split-распределение между двумя узлами.
type SplitConfig struct {
    Direction  Direction // Vertical | Horizontal
    First      *Node     // Первый (верхний/левый) узел
    Second     *Node     // Второй (нижний/правый) узел
    Fraction   float64   // Доля первого узла (0.0-1.0)
}

// Dragging указывает, что лицевая граница перетаскивается.
// IsCollapsed проверяет свернутость любого из узлов.
// CollapsedSize возвращает фиксированный размер при свернутом состоянии.
```

#### FlexConfig

```go
// FlexConfig конфигурирует flex-распределение между несколькими узлами.
type FlexConfig struct {
    Direction Direction // Horizontal | Vertical
    Items     []*FlexItem
}

// Dragging указывает, что лицевая граница перетаскивается.
```

#### FlexItem

```go
// FlexItem описывает элемент внутри flex-контейнера.
type FlexItem struct {
    Node      *Node
    Basis     int    // Базовый размер (0 = авто)
    Grow      int    // Коэффициент роста
    Collapsed bool   // Свёрнуто ли данный элемент
}
```

### Функции

#### renderNode

```go
// renderNode рендерит дерево узлов в строковые линии заданных размеров.
//
// @param node - Укореняющий узел для рендеринга
// @param w - Ширина в ячейках
// @param h - Высота в строках
// @returns []string — Массив строк с контентом (точно w × h)
//
// @sideeffect none — Функция не имеет побочных эффектов (чистая)
// @pure
func renderNode(node *Node, w, h int) []string
```

#### renderVerticalSplit

```go
// renderVerticalSplit рендерит вертикальный split (левый/правый колонки).
//
// @param split - Конфигурация split-распределения
// @param w - Ширина доступной области
// @param h - Высота в строках
// @returns []string — Массив строк
//
// @sideeffect none — Функция не имеет побочных эффектов (чистая)
// @pure
func renderVerticalSplit(split *SplitConfig, w, h int) []string
```

#### renderHorizontalSplit

```go
// renderHorizontalSplit рендерит горизонтальный split (верхний/нижний строки).
//
// @param split - Конфигурация split-распределения
// @param w - Ширина в ячейках
// @param h - Высота доступной области
// @returns []string — Массив строк
//
// @sideeffect none — Функция не имеет побочных эффектов (чистая)
// @pure
func renderHorizontalSplit(split *SplitConfig, w, h int) []string
```

#### renderFlex

```go
// renderFlex рендерит flex-контейнер с распределением по весам.
//
// @param flex - Конфигурация flex-распределения
// @param w - Ширина в ячейках
// @param h - Высота в строках
// @returns []string — Массив строк
//
// @sideeffect none — Функция не имеет побочных эффектов (чистая)
// @pure
func renderFlex(flex *FlexConfig, w, h int) []string
```

#### renderFlexRow

```go
// renderFlexRow рендерит flex-строку (горизонтальное направление).
//
// @param flex - Конфигурация flex-распределения
// @param w - Ширина в ячейках
// @param h - Высота в строках
// @param sizes - Вычисленные размеры каждого элемента
// @returns []string — Массив строк
//
// @sideeffect none — Функция не имеет побочных эффектов (чистая)
// @pure
func renderFlexRow(flex *FlexConfig, w, h int, sizes []int) []string
```

#### renderFlexColumn

```go
// renderFlexColumn рендерит flex-колонку (вертикальное направление).
//
// @param flex - Конфигурация flex-распределения
// @param w - Ширина в ячейках
// @param h - Высота в строках
// @param sizes - Вычисленные размеры каждого элемента
// @returns []string — Массив строк
//
// @sideeffect none — Функция не имеет побочных эффектов (чистая)
// @pure
func renderFlexColumn(flex *FlexConfig, w, h int, sizes []int) []string
```

#### computeFlexSizes

```go
// computeFlexSizes вычисляет размеры для flex-элементов по базовым и grow-весам.
//
// @param avail — Доступное пространство
// @param items — Массив flex-элементов
// @returns []int — Вычисленные размеры для каждого элемента
//
// @sideeffect none — Функция не имеет побочных эффектов (чистая)
// @pure
func computeFlexSizes(avail int, items []*FlexItem) []int
```

#### computeSplitSizes

```go
// computeSplitSizes вычисляет размеры для split-распределения.
// Учитывает состояние свёрнутости (collapsed) элементов.
//
// @param avail - Доступное пространство
// @param fraction - Доля первого элемента
// @param firstCollapsed - Свёрнут первый элемент?
// @param secondCollapsed - Свёрнут второй элемент?
// @param firstSize - Размер свернутого первого элемента
// @param secondSize - Размер свернутого второго элемента
// @returns first, second — Вычисленные размеры первого и второго элемента
//
// @sideeffect none — Функция не имеет побочных эффектов (чистая)
// @pure
func computeSplitSizes(avail int, fraction float64, firstCollapsed, secondCollapsed bool, firstSize, secondSize int) (first, second int)
```

#### padContent

```go
// padContent обеспечивает точные размеры w × h для контента.
// Обрезает по визуальной ширине (не байтам) для корректной работы с UTF-8 и ANSI.
// Каждая строка заканчивается ANSI reset, чтобы стили не утекали в соседей.
//
// @param content - Исходный контент (строка)
// @param w - Ширина в ячейках
// @param h - Высота в строках
// @returns []string — Массив строк с отформатированным контентом
//
// @sideeffect none — Функция не имеет побочных эффектов (чистая)
// @pure
func padContent(content string, w, h int) []string
```

#### makeEmptyLines

```go
// makeEmptyLines создаёт массив пустых строк заданных размеров.
//
// @param w - Ширина в ячейках
// @param h - Высота в строках
// @returns []string — Массив из h строк, каждая длиной w
//
// @sideeffect none — Функция не имеет побочных эффектов (чистая)
// @pure
func makeEmptyLines(w, h int) []string
```

#### findBorders

```go
// findBorders рекурсивно собирает все позиции границ для drag-and-drop.
//
// @param node - Укореняющий узел
// @param x, y - Стартовая позиция (верхний левый угол)
// @param w, h - Размеры области
// @returns []BorderHit — Сословие границ
//
// @sideeffect none — Функция не имеет побочных эффектов (чистая)
// @pure
func findBorders(node *Node, x, y, w, h int) []BorderHit
```

#### findFlexBorders

```go
// findFlexBorders собирает границы внутри flex-контейнера.
//
// @param flex - Конфигурация flex-распределения
// @param x, y - Стартовая позиция
// @param w, h - Размеры области
// @returns []BorderHit — Сословие границ
//
// @sideeffect none — Функция не имеет побочных эффектов (чистая)
// @pure
func findFlexBorders(flex *FlexConfig, x, y, w, h int) []BorderHit
```

## Поведение

### Splits

- **Vertical split** рендерит две колонки с вертикальной границей `│`
- **Horizontal split** рендерит две строки с горизонтальной границей `─`
- Когда один из элементов свернут (collapsed), граница не рендерится — элементы встают вплотную
- При перетаскивании границы (dragging) используется другой стиль (`borderDragStyle`)
- Размеры вычисляются с учётом `Fraction` и `CollapsedSize`

### Flex

- **Horizontal flex** — элементы в строке, разделённые вертикальными границами
- **Vertical flex** — элементы в колонке, разделённые горизонтальными границами
- Размеры вычисляются по формуле: `basis + (remaining / grow_weights)`
- Свёрнутые элементы получают размер 1 и не участвуют в распределении оставшегося пространства

### Collapsed Panels

- Свернутые панели получают фиксированный размер (`CollapsedSize`)
- Границы не рендерятся между свернутыми и соседними панелями
- Размер не может быть меньше `MinPanelSize`

### Border Rendering

- Границы имеют толщину 1 ячейка
- Вертикальная граница: `│`
- Горизонтальная граница: `─`
- При перетаскивании: `borderDragStyle.Render("│")` или `borderDragStyle.Render("─")`
- ANSI стили изолированы через `ansi.ResetStyle`

## Side Effects Contract

- Все функции пакета `render` помечены как `@pure` или `@sideeffect none`
- Функции не производят I/O, не модифицируют глобальное состояние, не вызывают внешние API
- Рендеринг полностью детерминирован при одинаковых входных параметрах

## Примеры

### Вертикальный split

```go
result := renderVerticalSplit(&SplitConfig{
    Direction: Vertical,
    First:     firstNode,
    Second:    secondNode,
    Fraction:  0.5,
}, 80, 24)
// result: []string — 24 строки по 80 символов
```

### Горизонтальный flex

```go
result := renderFlex(&FlexConfig{
    Direction: Horizontal,
    Items: []*FlexItem{
        {Node: node1, Basis: 10, Grow: 1},
        {Node: node2, Basis: 10, Grow: 2},
    },
}, 80, 24, sizes)
// sizes вычисляется через computeFlexSizes
```

### Pad content

```go
content := "Hello\nWorld"
result := padContent(content, 40, 2)
// result: []string — 2 строки по 40 символов каждая
```

## Ключевые Правила

1. **Всегда проверяй размеры** — `w <= 0` или `h <= 0` возвращает `nil` или пустые линии
2. **Учитывай collapsed** — свернутые элементы не получают границы и не участвуют в распределении
3. **Изолируй ANSI стили** — каждая строка заканчивается `ansi.ResetStyle`
4. **Минимальный размер** — `MinPanelSize` используется как fallback при недостатке пространства
5. **Border isolation** — границы изолированы через `ansi.ResetStyle` для предотвращения утечки стилей

## Чеклист

- [ ] Рендер вернёт точно `w × h` строк
- [ ] Границы рендерятся только между не свернутыми элементами
- [ ] ANSI стили изолированы на каждой строке
- [ ] collapsed элементы используют фиксированный размер
- [ ] Функции помечены с `@sideeffect` или `@pure`