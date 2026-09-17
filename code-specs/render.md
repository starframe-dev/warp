# Render Package Specification

## Описание

Файл `render.go` отвечает за рендеринг дерева узлов (node tree) в строковые представления с поддержкой различных макетных паттернов: splits (вертикальные/горизонтальные), flex-контейнеры и leaf-панели.

## Публичный API

### Типы

#### BorderHit

```go
// BorderHit описывает позиционируемую границу (drag handle) для drag-and-drop операций.
type BorderHit struct {
    Split     *SplitConfig
    Flex      *FlexConfig
    Direction Direction
    X, Y      int   // Стартовая позиция границы (в ячейках)
    Length    int   // Длина границы в ячейках
}
```

### Внутренние функции

Публичные API-элементы (`Node`, `SplitConfig`, `FlexConfig`, `FlexItem`, `Direction` и т.п.) описаны в других файлах пакета. В `render.go` используются поля:

- `Node.IsLeaf()`, `Node.Panel.View(w, h)`, `Node.IsCollapsed()`, `Node.CollapsedSize(direction)`
- `SplitConfig.Direction`, `SplitConfig.First`, `SplitConfig.Second`, `SplitConfig.Fraction`, `SplitConfig.Dragging`, `SplitConfig.OnCollapse`, `SplitConfig.CollapseRow`
- `FlexConfig.Direction`, `FlexConfig.Items`, `FlexConfig.Dragging`
- `FlexItem.Node`, `FlexItem.Basis`, `FlexItem.Grow`, `FlexItem.Collapsed`

## Функции

### renderNode

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

### renderVerticalSplit

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

### renderHorizontalSplit

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

### renderFlex

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

### renderFlexRow

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

### renderFlexColumn

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

### computeFlexSizes

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

### computeSplitSizes

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

### padContent

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

### makeEmptyLines

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

### findBorders

```go
// findBorders рекурсивно собирает все позиции границ для drag-and-drop.
//
// @param node - Укореняющий узел
// @param x, y - Стартовая позиция (верхний левый угол)
// @param w, h - Размеры области
// @returns []BorderHit — Список собранных границ
//
// @sideeffect none — Функция не имеет побочных эффектов (чистая)
// @pure
func findBorders(node *Node, x, y, w, h int) []BorderHit
```

### findFlexBorders

```go
// findFlexBorders собирает границы внутри flex-контейнера.
//
// @param flex - Конфигурация flex-распределения
// @param x, y - Стартовая позиция
// @param w, h - Размеры области
// @returns []BorderHit — Список собранных границ
//
// @sideeffect none — Функция не имеет побочных эффектов (чистая)
// @pure
func findFlexBorders(flex *FlexConfig, x, y, w, h int) []BorderHit
```

## Поведение

### Split-рендеринг

- **Vertical split** рендерит две колонки, разделённые вертикальной границей `│`.
- **Horizontal split** рендерит две группы строк, разделённые горизонтальной границей `─`.
- Когда один из узлов свернут (`IsCollapsed()`), граница между узлами не рендерится — узлы встают вплотную.
- При `Dragging = true` граница рендерится через `borderDragStyle` вместо `borderStyle`.
- Границы обёрнуты `ansi.ResetStyle` с обеих сторон, чтобы стили панелей не утекали через границу.
- Когда `OnCollapse != nil` и `CollapseRow >= 0`, и граница рендерится (не collapsed), символ границы на строке `CollapseRow` заменяется на collapse-символ `collapseStyle.Render("<")` (изолированный ANSI-стилями).
- Размеры первого/второго вычисляются через `computeSplitSizes` с учётом `Fraction` и collapsed-состояний.

### Flex-рендеринг

- **Horizontal flex** — элементы в строке, разделённые вертикальными границами `│`.
- **Vertical flex** — элементы в колонке, разделённые горизонтальными границами `─`.
- Граница между двумя элементами рендерится только если оба соседних элемента не collapsed.
- Границы обёрнуты `ansi.ResetStyle`, как и в split-рендеринге.
- Размеры элементов вычисляются через `computeFlexSizes`.

### Вычисление размеров flex

`computeFlexSizes`:

1. Для каждого элемента берётся базовый размер: `basis = Basis` (если не collapsed и `Basis <= 0`, то `basis = MinPanelSize`); для collapsed элементов `basis = 1`.
2. `remaining = avail - totalBasis`. Если `remaining <= 0`, возвращаются `sizes` как есть.
3. Если нет grow-весов (`totalGrow == 0`), `remaining` распределяется поровну между всеми не-collapsed элементами.
4. Иначе `remaining` распределяется пропорционально `Grow` для каждого не-collapsed элемента, а остаток (остаток от целочисленного деления) дописывается в последний не-collapsed элемент.
5. Collapsed элементы получают `basis` (обычно 1) и не участвуют в распределении.

### Вычисление размеров split

`computeSplitSizes`:

1. Если `firstCollapsed`: `first = firstSize` (минимум 1), `second = avail - first`; если `second < MinPanelSize`, то `second = MinPanelSize`, `first = avail - second`.
2. Если `secondCollapsed`: симметрично.
3. Иначе `first = int(avail * fraction)` (не меньше `MinPanelSize`), `second = avail - first` (не меньше `MinPanelSize`, иначе коррекция).

### Leaf-рендеринг

- Leaf-узел рендерится через `node.Panel.View(w, h)`, результат прогоняется через `padContent`.
- `padContent`:
  1. Обрезает каждую строку по визуальной ширине `w` через `ansi.Truncate(line, w, "")`.
  2. Дописывает пробелы до визуальной ширины `w` через `ansi.StringWidth`.
  3. Добавляет `ansi.ResetStyle` в конец каждой строки, чтобы стили не утекали.
  4. Возвращает ровно `h` строк.
- Если `w <= 0` или `h <= 0`, возвращает `makeEmptyLines(w, h)`.

### makeEmptyLines

- Если `w <= 0` или `h <= 0`, возвращает `nil`.
- Иначе возвращает `h` строк по `w` пробелов.

### findBorders / findFlexBorders

- `findBorders` рекурсивно собирает позиции границ:
  - Для split: позиция границы вычисляется через `computeSplitSizes`; граница добавляется только если оба узла не collapsed.
  - Для flex: позиция границ вычисляется через `computeFlexSizes`; граница добавляется только если оба соседних элемента не collapsed.
- `findFlexBorders` — внутренняя версия для flex-контейнера, без split-ветки.
- Границы внутри flex-контейнера обходятся через `findBorders(item.Node, ...)`, позиции `cx`/`cy` накапливаются по мере обхода.

## Side Effects Contract

- Все функции файла `render.go` помечены как `@pure` / `@sideeffect none`.
- Функции не производят I/O, не модифицируют глобальное состояние, не вызывают внешние API.
- Рендеринг полностью детерминирован при одинаковых входных параметрах.
- Допустимая внешняя зависимость — пакет `github.com/charmbracelet/x/ansi` (публичные константы/функции `ansi.ResetStyle`, `ansi.StringWidth`, `ansi.Truncate`).

## Ключевые Правила

1. **Всегда проверяй размеры** — `w <= 0` или `h <= 0` возвращает `nil` или пустые линии.
2. **Учитывай collapsed** — границы не рендерятся между collapsed элементами, collapsed flex-элементы не участвуют в распределении, а collapsed split-узлы используют `CollapsedSize(direction)`.
3. **Изолируй ANSI стили** — каждая строка и граница обёрнута в `ansi.ResetStyle`.
4. **Минимальный размер** — `MinPanelSize` используется как fallback при недостатке пространства.
5. **Border isolation** — границы изолированы через `ansi.ResetStyle` для предотвращения утечки стилей.

## Чеклист

- [x] Рендер вернёт точно `w × h` строк (для leaf через `padContent`; для split/flex через компоновку из `renderNode`)
- [x] Границы рендерятся только между не-collapsed элементами
- [x] ANSI стили изолированы на каждой строке и границе
- [x] Collapsed split-узлы используют `CollapsedSize(direction)`
- [x] Все функции помечены как `@pure` / `@sideeffect none`