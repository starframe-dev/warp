# Архитектура проекта Warp

Warp — Go-библиотека (TUI layout engine) на базе `charmbracelet/bubbletea` и `lipgloss`.
Она не рисует UI сама, а управляет **расположением** пользовательских панелей:
собирует их в дерево `Node` (split / flex / leaf), рисует границы и рамки,
распределяет размеры и маршрутизирует события Bubbletea к панелям.

## Ключевые решения

### Warp — тонкий корневой model, а не контейнер

`Warp` реализует `tea.Model`, но **не перехватывает** ни клавиши, ни мышь:
все сообщения уходят в `root Panel` без обработки. Это позволяет
использовать Warp как поддерево внутри чужого Bubbletea-приложения.

```
Warp (Model)
  └── root Panel (по умолчанию *TabGroup)
        └── Tab
              ├── root *Node (дерево split/flex/leaf)
              ├── floats []*FloatPane
              └── focused Panel
```

### Композиция вместо наследования

Панели не имеют общего предка. Любая структура (`*Collapsible`,
`TabGroup`, чужой `Panel`) вставляется в дерево как `Panel` —
Warp лишь знает `View(w, h) string` и `Update(msg) tea.Cmd`.

### `TabGroup` — Panel, а не корневой тип

`TabGroup` сам по себе Panel и может быть вложен в split/flex
рядом с любым другим компонентом. `Warp.New()` лишь создаёт
`TabGroup` как корень — это не ограничение, а конвенция:
`SetRoot(customPanel)` заменяет корень полностью.

### `Node` — единый узел дерева

Внутри `Tab` живёт дерево `Node`:

```go
Node { Panel Panel; Split *SplitConfig; Flex *FlexConfig; Collapse *NodeCollapse }
```

- `Panel != nil` → лист.
- `Split != nil` → бинарное деление с `Fraction` и границей `│`/`─`.
- `Flex != nil` → ряд/столбец с `Grow`-весами.
- `Collapse != nil` → узел временно в режиме фиксированного размера.

Дерево мутатируется методами `Tab`: `SplitVertical`, `SplitHorizontal`,
`FlexRow`, `FlexColumn` — каждый берёт `parent Panel`, находит узел
`findNode` и превращает его в `Split`/`Flex`.

### Float'ы — overlay, а не часть дерева

Плавающие панели **не** участвуют в `Node`. `Tab` хранит их
в порядке добавления; рендеринг `overlayFloat` накладывает их
поверх строк основного дерева с учётом ANSI-последовательностей
(`StripANSI`). Клик по float'у поднимает его z-order.

### Событийная модель — «спуск» по дереву

Порядок обработки в `Tab`:

1. `FloatPane` по z-order (drag / resize / click / close).
2. Border drag — `findBorders` собирает `BorderHit`, клик по `BorderHit`
   начинает drag; движение границы пересчитывает `Fraction`.
3. Клик по leaf-панели — `Update` получает `tea.Msg`.

`TabGroup` перехватывает только `Ctrl+T / Ctrl+W / Ctrl+Tab / Ctrl+Shift+Tab`,
остальное делегируется активному `Tab`.

### Рендеринг — рекурсия + padContent

`renderNode(node, w, h)` строит `[]string` строк:

- leaf → `padContent(panel.View(w, h), w, h)` — обрезка/дополнение
  до точных размеров с `ansi.Truncate` и `ansi.ResetStyle` в конце
  каждой строки (стили не утекают в соседние панели).
- split → `renderVerticalSplit`/`renderHorizontalSplit`:
  `computeSplitSizes` считает размеры детей с учётом `MinPanelSize`
  и collapsed-состояния, между ними строка-граница.
- flex → `renderFlex`: `computeFlexSizes` сначала выделяет `Basis`
  (или 1 для collapsed), затем остаток распределяется по `Grow`.

Границы рисуются в `borderStyle`, в режиме drag — в `borderDragStyle`
(yellow). Когда одно из делений collapsed, граница опускается —
collapsed-панель стоит вплотную к раскрытой.

### Фокус — явное API, а не биндинг клавиш

Warp **не биндит** `Tab`/`Shift+Tab` сам. `Tab.FocusNext/Prev/First/Panel`
вызывает разработчик из своего `Update`. `Focusable` — интерфейс
`{Panel; Focus(); Blur(); Focused() bool}`, `RawKeyReceiver` — для PTY,
получащих все клавиши без перехвата.

### Element tree — для E2E, а не рендера

`ElementProvider.Elements(w, h)` возвращает `[]Element`
(`Role`, `Name`, `Action`, `Bounds`, `Children`). Warp экспортирует его
через HTTP `/elements` — это контракт для E2E-тестов, а не часть
отрисовки. Компоненты реализуют `ElementProvider` сами
(табы, floats, modal, popover, input, dropdown и др.).

### Состояние — локальное у каждого компонента

Ни у Warp, ни у `Tab`, ни у `TabGroup` нет глобальных подписок
или эмиттеров. Состояние живёт внутри конкретного экземпляра
`Input.Value`, `Scrollable.Offset`, `Selectable.Selection`,
`Modal.open` и т.п. Коммуникация — только через `tea.Msg`.

## Файловая карта

| Файл | Ответственность |
|-------|------------------|
| `warp.go` | Корневая `Warp`, HTTP `/elements`, `AsPanel` |
| `panel.go` | Интерфейс `Panel` |
| `split.go` | `Node`, `SplitConfig`, `FlexConfig`, `Direction`, collapse |
| `tab.go` | `Tab` (дерево, floats, focus, drag, mouse/keys) |
| `tabgroup.go` | `TabGroup` (tab bar, Ctrl+T/W/Tab) |
| `render.go` | `renderNode`, `findBorders`, `padContent` |
| `float.go` | `FloatPane`, `overlayFloat`, `StripANSI` |
| `focus.go` | `Focusable`, `collectFocusables`, `RawKeyReceiver` |
| `styles.go` + `theme.go` | Палитра, `lipgloss.Style`, `SetTheme` |
| `collapsible.go` … `popover.go` | Компоненты-обёртки |
| `wrap.go` | `WordWrap` / `SpaceWrap` |
| `drag.go` | Плейсхолдер (логика — в `tab.go`/`render.go`) |

## Ограничения по дизайну

- Nested float не поддерживается (float внутри float).
- Разрешён только один `SplitConfig`/`FlexConfig` на узел — вложенные
  layouts строятся за счёт новых узлов, а не вложенных структур.
- `MinPanelSize = 3` и `clampFraction(0.1–0.9)` — жёсткие границы
  деления; `Fraction` вне диапазона не допустим.
- Границы всегда 1 символ; padding/gap/align отсутствуют.
