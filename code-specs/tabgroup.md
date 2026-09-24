# TabGroup — Спецификация

## Описание

`TabGroup` — это компонент Panel, который отображает панель вкладок (tab bar) и переключение между вкладками. Предназначен для использования внутри splits, flex layouts или как корневой panel.

## Публичный API

### Конструкторы

```go
// NewTabGroup создаёт TabGroup panel с одной дефолтной вкладкой "main".
func NewTabGroup(pos TabPosition) *TabGroup
```

```go
// NewTab создаёт новую вкладку и переключается на неё.
func (tg *TabGroup) NewTab(name string) *Tab
```

### Методы доступа

```go
// ActiveTab возвращает текущую активную вкладку.
func (tg *TabGroup) ActiveTab() *Tab
```

```go
// NextTab переключает на следующую вкладку (циклически).
func (tg *TabGroup) NextTab()
```

```go
// PrevTab переключает на предыдущую вкладку (циклически).
func (tg *TabGroup) PrevTab()
```

### Panel interface

```go
// View рендерит таббар + контент активной вкладки.
func (tg *TabGroup) View(w, h int) string

// Update обрабатывает сообщения: KeyMsg, MouseMsg, WindowSizeMsg.
func (tg *TabGroup) Update(msg tea.Msg) tea.Cmd

// Elements возвращает элементы активной вкладки со сдвигом от tab bar.
func (tg *TabGroup) Elements(w, h int) []Element
```

## Внутренние методы

```go
// closeTab закрывает вкладку по индексу. Нельзя закрыть последнюю.
func (tg *TabGroup) closeTab(idx int)

// switchTab переключает на вкладку по индексу.
func (tg *TabGroup) switchTab(idx int)

// handleKeyMsg сначала передаёт сообщение фокусированной панели, если она
// реализует RawKeyReceiver и WantsRawKeys() == true; в противном случае:
//   - Ctrl+Tab — NextTab
//   - Ctrl+Shift+Tab — PrevTab
//   - Ctrl+W — closeTab (активная вкладка)
//   - Ctrl+T — NewTab (новая вкладка)
//   - Ctrl+C — завершение программы
//   - Остальные клавиши — передаются активной вкладке
func (tg *TabGroup) handleKeyMsg(msg tea.KeyMsg) tea.Cmd

// handleMouseMsg обрабатывает клики мыши.
func (tg *TabGroup) handleMouseMsg(msg tea.MouseMsg) tea.Cmd

// isOnTabBar проверяет, находится ли клик в области таббара.
func (tg *TabGroup) isOnTabBar(x, y int) bool

// handleTabBarClick обрабатывает клики по таббару.
func (tg *TabGroup) handleTabBarClick(msg tea.MouseMsg) tea.Cmd

// renderTabBar рендерит таббар.
func (tg *TabGroup) renderTabBar(width int) string

// renderHorizontalTabBar рендерит горизонтальный таббар (TabTop, TabBottom).
func (tg *TabGroup) renderHorizontalTabBar(width int) string

// renderVerticalTabBar рендерит вертикальный таббар (TabLeft, TabRight).
func (tg *TabGroup) renderVerticalTabBar(width int) string

// contentWidth вычисляет ширину контента с учётом позиции таббара.
func (tg *TabGroup) contentWidth(totalW int) int

// contentHeight вычисляет высоту контента с учётом позиции таббара.
func (tg *TabGroup) contentHeight(totalH int) int

// contentOffset возвращает смещение X,Y для контента.
func (tg *TabGroup) contentOffset() (int, int)
```

## Типы

### TabGroup

```go
type TabGroup struct {
    tabs            []*Tab
    activeTab       int
    width           int
    height          int
    tabPosition     TabPosition
    tabRegions      []tabRegion
    newTabRegion    *tabRegion
    verticalTabWidth int
}
```

### TabPosition

```go
type TabPosition int

const (
    TabTop    TabPosition = iota // сверху
    TabBottom                    // снизу
    TabLeft                      // слева
    TabRight                     // справа
    TabNone                      // без таббара
)
```

### tabRegion

```go
type tabRegion struct {
    idx    int  // индекс вкладки, -1 если +
    startX int  // начало области X
    endX   int  // конец области X
    closeX int  // X позиция кнопки закрытия, -1 если нет
}
```

## Поведение

### Создание вкладки
1. `NewTab(name)` создаёт вкладку с заданным именем.
2. Новая вкладка становится активной.
3. `tabRegions` обновляется при следующем рендеринге.

### Закрытие вкладки
1. `closeTab(idx)` удаляет вкладку из массива `tabs`.
2. При закрытии вкладки перед активной её индекс корректируется, чтобы сохранить ту же логическую вкладку.
3. При закрытии активной выбирается следующая вкладка на её позиции; если удалена последняя, выбирается новая последняя.
4. Нельзя закрыть последнюю вкладку.

### Переключение вкладок
- `NextTab()`: `(activeTab + 1) % len(tabs)`
- `PrevTab()`: `(activeTab - 1 + len(tabs)) % len(tabs)`
- Клик на вкладке в таббаре переключает на неё.

### Рендеринг
1. `View(w, h)` вычисляет позицию таббара и смещение контента.
2. Рендерит таббар + контент активной вкладки.

### Обработка сообщений
- `tea.KeyMsg` → `handleKeyMsg` (горячие клавиши табов, остальное — в активную вкладку)
- `tea.MouseMsg` → `handleMouseMsg` (клик по таббару или в активную вкладку)
- `tea.WindowSizeMsg` → broadcast всем вкладкам

## Стили

- `inactiveTabStyle` — неактивные вкладки
- `activeTabStyle` — активная вкладка
- `newTabStyle` — кнопка +
- `tabBarStyle` — фон таббара

## Ключевые правила

1. Нельзя закрыть последнюю вкладку
2. Переключение вкладок циклическое
3. `tabRegions` пересчитывается при каждом рендеринге
4. Вертикальная ширина вычисляется по отображаемым меткам вкладок и общей геометрии; `Elements` считает её без вызова рендеринга и без изменения сохранённого состояния
5. `Elements(w, h)` не зависит от предшествующего `View` и возвращает координаты относительно исходного прямоугольника вкладки
6. Максимальная ширина названия — 20 терминальных ячеек (горизонтально) и 15 (вертикально); Unicode-текст усекается по ширине ячеек
7. Правый вертикальный таббар преобразует X мыши из экранной системы координат в координаты самого таббара
8. `newTabRegion` — отдельная область справа от таббара
9. Через Ctrl+W или кнопку × закрывается только активная вкладка
