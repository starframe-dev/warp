# Tab

Спецификация пакета `warp/tab.go` — компонент для управления вкладками и панелями в TUI-приложениях на базе Bubble Tea.

## Описание

Пакет предоставляет функционал для создания и управления вкладками (tabs) с панельной структурой. Tab поддерживает:

- Дерево панелей с поддержкой вертикальных и горизонтальных разделов (splits)
- Flex-раскладки (flex layouts)
- Плавающие панели (floating panes)
- Управление фокусом панелей
- Обработку событий мыши и клавиатуры
- Рендеринг контента и элементов

## Типы

### TabPosition

Определяет позицию панели вкладок.

```go
type TabPosition int
```

#### Константы

| Константа | Значение | Описание |
|-----------|----------|----------|
| TabTop | 0 | Вкладка сверху |
| TabBottom | 1 | Вкладка снизу |
| TabLeft | 2 | Вкладка слева |
| TabRight | 3 | Вкладка справа |
| TabNone | 4 | Нет вкладки |

### Tab

Представляет вкладку с деревом панелей, плавающими панелями и состоянием фокуса.

```go
type Tab struct {
    name    string
    root    *Node
    focused Panel
    floats  []*FloatPane
    parent  *TabGroup
    width   int
    height  int

    // Drag state
    dragging     *SplitConfig
    flexDragging *FlexConfig
    flexDragIdx  int
    lastBorders  []BorderHit
}
```

#### Поля

- `name` — имя вкладки
- `root` — корень дерева панелей
- `focused` — текущая сфокусированная панель
- `floats` — список плавающих панелей
- `parent` — родительская TabGroup (если вкладка вложена)
- `width`, `height` — размеры контента в ячейках
- `dragging` — текущий разрыв split при перетаскивании
- `flexDragging` — текущий перетаскиваемый flex элемент
- `flexDragIdx` — индекс границ для flex перетаскивания
- `lastBorders` — последняя позиция границ для обработки кликов

### TabGroup

Группа вкладок, объединяющая несколько Tab объектов.

```go
type TabGroup struct {
    tabs []*Tab
    position TabPosition
}
```

## Публичный API

### newTab

Создаёт вкладку без родительской TabGroup.

```go
func newTab(name string, parent *TabGroup) *Tab
```

**Параметры:**

- `name` — имя вкладки
- `parent` — родительская группа (может быть nil)

**Возвращает:**

- `*Tab` — новая вкладка с пустым корневым panel

### NewTab

Создаёт независимую вкладку без родительской группы. Полезно для встраивания warp layout в Panel.

```go
func NewTab(name string) *Tab
```

**Параметры:**

- `name` — имя вкладки

**Возвращает:**

- `*Tab` — независимая вкладка

### ensureRoot

Гарантирует существование корневого узла.

```go
func (t *Tab) ensureRoot()
```

### RootPanel

Возвращает корневую панель вкладки.

```go
func (t *Tab) RootPanel() Panel
```

**Возвращает:**

- `Panel` — корневая панель

### SetRootPanel

Заменяет корневую панель вкладки.

```go
func (t *Tab) SetRootPanel(panel Panel)
```

**Параметры:**

- `panel` — новая корневая панель

### SplitVertical

Вертикально делит панель (слева/справа).

```go
func (t *Tab) SplitVertical(parent Panel, fraction float64, newPanel Panel)
```

**Параметры:**

- `parent` — панель, которую нужно разделить
- `fraction` — доля левой панели (0.0–1.0)
- `newPanel` — новая панель для добавления в разделение

**Возвращает:**

- ни возвращает (побочный эффект на дереве)

### SplitHorizontal

Горизонтально делит панель (сверху/снизу).

```go
func (t *Tab) SplitHorizontal(parent Panel, fraction float64, newPanel Panel)
```

**Параметры:**

- `parent` — панель, которую нужно разделить
- `fraction` — доля верхней панели (0.0–1.0)
- `newPanel` — новая панель для добавления в разделение

**Возвращает:**

- ни возвращает (побочный эффект на дереве)

### FlexRow

Заменяет родительскую панель горизонтальной flex-раскладкой.

```go
func (t *Tab) FlexRow(parent Panel, items []FlexItemSpec)
```

**Параметры:**

- `parent` — панель для замены
- `items` — список элементов flex-раскладки

**Возвращает:**

- ни возвращает (побочный эффект на дереве)

### FlexColumn

Заменяет родительскую панель вертикальной flex-раскладкой.

```go
func (t *Tab) FlexColumn(parent Panel, items []FlexItemSpec)
```

**Параметры:**

- `parent` — панель для замены
- `items` — список элементов flex-раскладки

**Возвращает:**

- ни возвращает (побочный эффект на дереве)

### Float

Создаёт плавающую панель над layout.

```go
func (t *Tab) Float(panel Panel, x, y, width, height int)
```

**Параметры:**

- `panel` — панель для плавающего режима
- `x`, `y` — позиция в ячейках
- `width`, `height` — размеры в ячейках

**Возвращает:**

- ни возвращает

### CloseFloat

Удаляет плавающую панель.

```go
func (t *Tab) CloseFloat(fp *FloatPane)
```

**Параметры:**

- `fp` — плавающая панель для удаления

**Возвращает:**

- ни возвращает

### Focus

Возвращает текущую сфокусированную панель.

```go
func (t *Tab) Focus() Panel
```

**Возвращает:**

- `Panel` — сфокусированная панель

### SetFocus

Устанавливает сфокусированную панель.

```go
func (t *Tab) SetFocus(panel Panel) tea.Cmd
```

**Параметры:**

- `panel` — панель для фокуса

**Возвращает:**

- `tea.Cmd` — команда для обновления (обычно nil)

### FocusNext

Перемещает фокус на следующую фокусируемую панель.

```go
func (t *Tab) FocusNext()
```

**Возвращает:**

- ни возвращает

### FocusPrev

Перемещает фокус на предыдущую фокусируемую панель.

```go
func (t *Tab) FocusPrev()
```

**Возвращает:**

- ни возвращает

### FocusFirst

Перемещает фокус на первую фокусируемую панель.

```go
func (t *Tab) FocusFirst()
```

**Возвращает:**

- ни возвращает

### FocusPanel

Устанавливает фокус на конкретной панели, если она фокусируема.

```go
func (t *Tab) FocusPanel(panel Panel)
```

**Параметры:**

- `panel` — панель для фокуса

**Возвращает:**

- ни возвращает

### View

Реализует интерфейс warp.Panel для рендеринга контента.

```go
func (t *Tab) View(width, height int) string
```

**Параметры:**

- `width`, `height` — размеры в ячейках

**Возвращает:**

- `string` — строка с рендерингом

### Update

Реализует интерфейс warp.Panel для обработки сообщений.

```go
func (t *Tab) Update(msg tea.Msg) tea.Cmd
```

**Параметры:**

- `msg` — сообщение Bubble Tea

**Возвращает:**

- `tea.Cmd` — команда для отправки

### Elements

Возвращает элементы из дерева панелей с учётом разложений.

```go
func (t *Tab) Elements(w, h int) []Element
```

**Параметры:**

- `w`, `h` — размеры в ячейках

**Возвращает:**

- `[]Element` — список элементов

### HandleMouse

Обрабатывает события мыши для вкладки.

```go
func (t *Tab) HandleMouse(msg tea.MouseMsg) tea.Cmd
```

**Параметры:**

- `msg` — событие мыши

**Возвращает:**

- `tea.Cmd` — команда (если есть)

### HandleKeys

Передает клавиатурные сообщения сфокусированной панели.

```go
func (t *Tab) handleKeys(msg tea.KeyMsg) tea.Cmd
```

**Параметры:**

- `msg` — событие клавиатуры

**Возвращает:**

- `tea.Cmd` — команда (если есть)

**Примечание:** Warp не перехватывает Tab/Shift+Tab для фокуса автоматически. Используйте `FocusNext()`/`FocusPrev()` для переключения фокуса.

### SetSplitFraction

Обновляет долю разделения, содержащую указанную панель.

```go
func (t *Tab) SetSplitFraction(panel Panel, fraction float64) bool
```

**Параметры:**

- `panel` — целевая панель
- `fraction` — новая доля (0.0–1.0)

**Возвращает:**

- `bool` — true если панель найдена внутри разделения

### GetSplitFraction

Возвращает текущую долю разделения для панели.

```go
func (t *Tab) GetSplitFraction(panel Panel) (float64, bool)
```

**Параметры:**

- `panel` — целевая панель

**Возвращает:**

- `float64` — текущая доля
- `bool` — true если панель в разложении

### Collapse

Сжимает панель до фиксированного размера внутри родителя.

```go
func (t *Tab) Collapse(panel Panel, size int) bool
```

**Параметры:**

- `panel` — панель для сжатия
- `size` — желаемый размер в ячейках

**Возвращает:**

- `bool` — true если панель найдена

### Expand

Восстанавливает панель до сохранённой доли разделения.

```go
func (t *Tab) Expand(panel Panel) bool
```

**Параметры:**

- `panel` — панель для расширения

**Возвращает:**

- `bool` — true если панель найдена и была сжата

### BroadcastResize

Отправляет ResizeMsg с текущими размерами контента каждой листовой панели. Вызывается автоматически на WindowSizeMsg и после завершения перетаскивания.

```go
func (t *Tab) BroadcastResize() tea.Msg
```

**Возвращает:**

- `tea.Msg` — batch-сообщение с командами

## Внутренние функции (неэкспортируемые)

- `handleMouse` — обработка событий мыши с оффсетом для вложенных вкладок
- `panelAt` — поиск панели под курсором
- `focusStep` — перемещение фокуса на delta шагов
- `setFocusedFocusable` — установка фокуса на фокусируемую панель
- `renderContent` — рендеринг дерева панелей
- `broadcastMsg` — отправка сообщения всем панелям
- `broadcastResize` — отправка ResizeMsg всем панелям с размерами
- `collapseNode` — рекурсивное сжатие узла
- `expandNode` — рекурсивное расширение узла
- `saveCollapse` — сохранение состояния сжатия
- `restoreCollapse` — восстановление состояния сжатия
- `setSplitFractionNode` — установка доли разделения
- `getSplitFractionNode` — получение доли разделения
- `toggleCollapsibleNode` — переключение состояния collapsible
- `updateDrag` — обновление позиции при перетаскивании
- `updateSplitDrag` — обновление split drag
- `updateFlexDrag` — обновление flex drag
- `updateFlexCollapsed` — обновление состояния collapsible в flex
- `elementsNode` — сбор элементов из узла
- `elementsFlex` — сбор элементов из flex раскладки
- `broadcastNode` — отправка сообщения узлам

## Helper-функции

### clampFraction

Ограничивает долю разделения диапазоном [0.1, 0.9].

```go
func clampFraction(f float64) float64
```

**Параметры:**

- `f` — доля для ограничения

**Возвращает:**

- `float64` — ограниченная доля

### emptyPanel

Плейсхолдер для панели, когда вкладка ещё не имеет пользовательских панелей.

```go
type emptyPanel struct{}

func (emptyPanel) View(width, height int) string
func (emptyPanel) Update(msg tea.Msg) tea.Cmd
```

## События

Tab обрабатывает следующие типы сообщений Bubble Tea:

- `tea.WindowSizeMsg` — изменение размера окна
- `ResizeMsg` — изменение размера контента
- `tea.MouseMsg` — события мыши
- `tea.KeyMsg` — события клавиатуры
- `tea.PtyOutputMsg` — вывод псевдоконсоли
- `tea.PtyReadyMsg` — готовность псевдоконсоли

## События мыши

Tab обрабатывает следующие действия мыши:

- `tea.MouseActionPress` — нажатие
- `tea.MouseActionMotion` — движение
- `tea.MouseActionRelease` — отпускание

### Кнопки мыши

- `tea.MouseButtonLeft` — левая кнопка

## Примечания

- Warp не перехватывает Tab/Shift+Tab автоматически для фокуса — используйте `FocusNext()`/`FocusPrev()`
- `handleMouse` передаёт события в панели под курсором (относительные координаты)
- Плавающие панели обрабатываются первыми (z-order)
- При отпускании перетаскивания размеры панелей обновляются автоматически
