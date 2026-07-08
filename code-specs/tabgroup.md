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

// handleKeyMsg обрабатывает клавиатурные события:
//   - Ctrl+Tab — NextTab
//   - Ctrl+Shift+Tab — PrevTab
//   - Ctrl+W — closeTab (активная вкладка)
//   - Ctrl+T — NewTab (новая вкладка)
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
2. Если активная вкладка удалена, выбирается последняя оставшаяся.
3. Нельзя закрыть последнюю вкладку.

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
4. Максимум 20 символов в названии вкладки (горизонтальный)
5. Максимум 15 символов (вертикальный)
6. `newTabRegion` — отдельная область справа от таббара
7. Закрыть можно только активную вкладку через Ctrl+W или клик на ×
