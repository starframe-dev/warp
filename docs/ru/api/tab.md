# Tab

Тип `Tab` в `tab.go` — это самодостаточная компоновка warp: она владеет
корневым деревом панелей (`Node`), необязательным набором плавающих
панелей и фокусом клавиатуры. Реализует интерфейс `Panel`
(`View`/`Update`), поэтому сам вклад можно встраивать как панель в более
крупную компоновку (например, внутрь `Container`).

Вклад создаётся по standalone через `NewTab` (без родительской группы)
или внутренне через `newTab(name, parent)`, когда он принадлежит
`TabGroup`. В обоих случаях корень стартует с пустой
плейсхолдер-панелью, чтобы вклад всегда был рендерим.

## Публичный API

| Метод | Назначение |
|----|----|
| `NewTab(name string) *Tab` | Создаёт standalone-вклад с пустым корнем. |
| `RootPanel() Panel` | Возвращает корневую панель (можно безопасно передавать как первого родительского сплита). |
| `SetRootPanel(panel Panel)` | Заменяет корневую панель. |
| `SplitVertical(parent Panel, fraction float64, newPanel Panel)` | Разбивает панель на левую/правую (fraction — доля левой). |
| `SplitHorizontal(parent Panel, fraction float64, newPanel Panel)` | Разбивает панель на верхнюю/нижнюю (fraction — доля верхней). |
| `FlexRow(parent Panel, items []FlexItemSpec)` | Заменяет панель горизонтальным flex-рядом. |
| `FlexColumn(parent Panel, items []FlexItemSpec)` | Заменяет панель вертикальным flex-столбцом. |
| `Float(panel Panel, x, y, width, height int)` | Добавляет плавающую панель поверх компоновки. |
| `CloseFloat(fp *FloatPane)` | Удаляет плавающую панель. |
| `Focus() Panel` | Возвращает текущую сфокусированную панель. |
| `SetFocus(panel Panel) tea.Cmd` | Устанавливает сфокусированную панель. |
| `FocusNext() / FocusPrev()` | Перемещает фокус на следующую/предыдущую фокусируемую панель (по кругу). |
| `FocusFirst()` | Перемещает фокус на первую фокусируемую панель. |
| `FocusPanel(panel Panel)` | Фокусирует заданную панель, если она `Focusable`. |
| `SetSplitFraction(panel Panel, fraction float64) bool` | Устанавливает долю сплита, содержащего `panel`. |
| `GetSplitFraction(panel Panel) (float64, bool)` | Читает долю сплита, содержащего `panel`. |
| `Collapse(panel Panel, size int) bool` | Сжимает `panel` до фиксированного размера, сохраняя её долю сплита для восстановления. |
| `Expand(panel Panel) bool` | Восстанавливает схлопнутую панель до сохранённой доли. |
| `ToggleSplitCollapse(parent Panel)` | Переключает состояние схлопывания сплита, содержащего `parent`. |
| `SetSplitCollapse(parent, collapseRow, onCollapse)` | Привязывает символ схлопывания + колбэк к границе сплита. |
| `ToggleCollapsible(panel Panel)` | Переключает состояние схлопывания сворачиваемой панели. |
| `BroadcastResize() tea.Msg` | Рассылает `ResizeMsg` каждому листу с вычисленным размером. |
| `HandleMouse(msg tea.MouseMsg) tea.Cmd` | Публичная точка входа для мыши во встроенных вкладах. |
| `View(width, height int) string` | Рендерит содержимое вклада (реализует `Panel`). |
| `Update(msg tea.Msg) tea.Cmd` | Диспетчит сообщения по дереву панелей (реализует `Panel`). |
| `Elements(w, h int) []Element` | Собирает интерактивные элементы из дерева. |

## Типы

### TabPosition

Целочисленный тип, определяющий, где рисуется панель вкладов: `TabTop`,
`TabBottom`, `TabLeft`, `TabRight` или `TabNone`.

### FlexItemSpec

Описывает один элемент для добавления в flex-компоновку: `Panel`
(панель) и `Grow` (её вес flex-grow, ограниченный значением ≥ 0).

### emptyPanel

Внутренняя плейсхолдер-панель, используемая как начальный корень, чтобы
вклад всегда мог опрашиваться как провайдер элементов до того, как к
нему привязаны реальные панели.

## Поведение и детали реализации

- **Разбиение:** `SplitVertical`/`SplitHorizontal` оборачивают
  существующую и новую панель в `SplitConfig` с ограниченных долями
  (`clampFraction` держит значения в \[0.1, 0.9\]).
- **Flex-компоновки:** `FlexRow`/`FlexColumn` заменяют содержимое узла
  на `FlexConfig`. Отрицательные веса роста обрезаются до 0; пустой срез
  элементов игнорируется.
- **Поток сообщений:** `Update` обрабатывает `tea.WindowSizeMsg`
  (записывает размер, рассылает `ResizeMsg` плюс исходное
  window-сообщение), `ResizeMsg` (только рассылка resize), а остальные
  сообщения (например, `PtyOutputMsg`) пересылает каждому листу.
- **Рендер:** `View` рендерит дерево узлов, обновляет данные границ
  (`findBorders`), затем накладывает плавающие панели.
- **Мышь:** `handleMouse` (экспонируется как `HandleMouse`) сначала
  обрабатывает плавающие панели (z-порядок, выведение наверх, закрытие
  по клику снаружи), затем перетаскивание границ (сплит или flex),
  символы схлопывания и в конце пересылает события панели под курсором в
  относительных координатах. Перетаскивание live-рассылает resize, так
  что панели обновляются в процессе перетаскивания.
- **Клавиатура:** `handleKeys` пересылает только сфокусированной панели;
  warp никогда не перехватывает Tab/Shift+Tab автоматически —
  используйте явные методы `FocusNext`/`FocusPrev`.
- **Схлопывание/Разворачивание:** `Collapse` сохраняет долю сплита и
  сжимает панель до фиксированного размера; `Expand` восстанавливает
  сохранённую долю. Flex-элементы получают флаг `Collapsed`, а панели
  `Collapsible` переключают собственное состояние.
- **Сбор элементов:** `Elements` рекурсивно обходит сплиты и
  flex-компоновки, вычисляя размеры по каждой ветке (с учётом 1-cell
  границ) и сдвигая границы элементов, чтобы координаты были
  относительны к области содержимого вклада.

## Использование

``` go
tab := warp.NewTab("main")
tab.SplitVertical(tab.RootPanel(), 0.6, rightPanel)
tab.SplitHorizontal(leftPanel, 0.7, bottomPanel)
tab.FlexRow(centerPanel, []warp.FlexItemSpec{{Panel: a, Grow: 1}, {Panel: b, Grow: 2}})
tab.Float(overlay, 0, 0, 40, 10)

// focus traversal
tab.FocusNext()
tab.FocusPrev()
tab.FocusFirst()

// split geometry
tab.SetSplitFraction(leftPanel, 0.7)
frac, ok := tab.GetSplitFraction(leftPanel)

// collapse / expand
tab.Collapse(leftPanel, 2)
tab.Expand(leftPanel)

// broadcast size to leaves
cmd := tab.BroadcastResize()
```
