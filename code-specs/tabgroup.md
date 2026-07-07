# TabGroup — Спецификация

## Описание

`TabGroup` — это компонент Panel, который отображает панель вкладок (tab bar) и переключение между вкладками. Предназначен для использования внутри splits, flex layouts или как корневой panel.

## Публичный API

### Конструкторы

```go
// NewTabGroup создаёт TabGroup panel с одной дефолтной вкладкой "main".
//
// @param pos — позиция панели вкладок (TabTop, TabBottom, TabLeft, TabRight, TabNone)
// @returns *TabGroup — новый TabGroup
func NewTabGroup(pos TabPosition) *TabGroup
```

```go
// NewTab создаёт новую вкладку и переключается на неё.
//
// @param name — имя вкладки для отображения в tab bar
// @returns *Tab — ссылка на созданную вкладку
func (tg *TabGroup) NewTab(name string) *Tab
```

### Методы доступа

```go
// ActiveTab возвращает текущую активную вкладку.
//
// @returns *Tab — активная вкладка, или nil если вкладок нет
func (tg *TabGroup) ActiveTab() *Tab
```

```go
// CloseTab закрывает вкладку по индексу.
//
// @param idx — индекс вкладки для закрытия
func (tg *TabGroup) CloseTab(idx int)
```

```go
// SwitchTab переключает на вкладку по индексу.
//
// @param idx — индекс вкладки для переключения
func (tg *TabGroup) SwitchTab(idx int)
```

```go
// NextTab переключает на следующую вкладку.
//
// Работает циклически (последняя → первая).
func (tg *TabGroup) NextTab()
```

```go
// PrevTab переключает на предыдущую вкладку.
//
// Работает циклически (первая → последняя).
func (tg *TabGroup) PrevTab()
```

```go
// NewTab создаёт новую вкладку с именем "tab" и переключается на неё.
//
// @returns tea.Cmd — команда (nil)
func (tg *TabGroup) NewTab() tea.Cmd
```

### Обработка ввода

```go
// NextTab переключает на следующую вкладку по Ctrl+Tab.
// @sideeffect net — не используется
func (tg *TabGroup) NextTab() tea.Cmd
```

```go
// PrevTab переключает на предыдущую вкладку по Ctrl+Shift+Tab.
// @sideeffect net — не используется
func (tg *TabGroup) PrevTab() tea.Cmd
```

```go
// CloseTab закрывает активную вкладку по Ctrl+W.
//
// Нельзя закрыть последнюю вкладку.
// @sideeffect mutation — удаление из списка вкладок
func (tg *TabGroup) CloseTab() tea.Cmd
```

```go
// NewTab создаёт новую вкладку по Ctrl+T.
// @sideeffect mutation — добавление новой вкладки
func (tg *TabGroup) NewTab() tea.Cmd
```

### Рендеринг

```go
// View рендерит таббар + контент активной вкладки.
//
// @param w — ширина компонента
// @param h — высота компонента
// @returns string — отрендеренный текст
func (tg *TabGroup) View(w, h int) string
```

```go
// Elements возвращает элементы активной вкладки со сдвигом от tab bar.
//
// @param w — ширина компонента
// @param h — высота компонента
// @returns []Element — элементы с учётом позиции таббара
func (tg *TabGroup) Elements(w, h int) []Element
```

## Типы

### TabGroup

```go
// TabGroup представляет панель вкладок.
//
// Может использоваться как компонент внутри splits, flex layouts или как корневой panel.
// Поддерживает несколько позиций: TabTop, TabBottom, TabLeft, TabRight, TabNone.
//
// @sideeffect render — рендеринг через View()
// @sideeffect mutation — создание/удаление вкладок через NewTab(), CloseTab()
type TabGroup struct {
    tabs      []*Tab           // список всех вкладок
    activeTab int              // индекс активной вкладки
    width     int              // текущая ширина
    height    int              // текущая высота
    tabPosition      TabPosition    // позиция таббара
    tabRegions       []tabRegion    // области клика по вкладкам
    newTabRegion     *tabRegion     // область для создания новой вкладки
    verticalTabWidth int            // ширина вертикального таббара
}
```

### Tab

```go
// Tab представляет отдельную вкладку.
// Содержит контент и метаданные для отображения в tab bar.
//
// @sideeffect render — рендеринг контента через renderContent()
type Tab struct {
    name      string           // имя вкладки
    content   Panel            // панель контента
    tabRegion tabRegion        // область клика в tab bar
}
```

### tabRegion

```go
// tabRegion описывает кликабельную область в таббаре.
//
// @sideeffect — внутренний тип, не используется напрямую
type tabRegion struct {
    idx    int          // индекс вкладки, -1 если это область создания новой вкладки
    startX int          // начало области X
    endX   int          // конец области X
    closeX int          // X позиция кнопки закрытия, -1 если нет
}
```

### TabPosition

```go
// TabPosition определяет позицию таббара относительно контента.
type TabPosition int

// Константы:
const (
    TabTop    TabPosition = iota // сверху
    TabBottom                    // снизу
    TabLeft                      // слева
    TabRight                     // справа
    TabNone                      // без таббара (только контент)
)
```

## Внутренние типы

### padRight

```go
// padRight добавляет пробелы справа до указанной ширины.
// Используется для выравнивания вертикальных вкладок.
//
// @param s — строка для выравнивания
// @param w — целевая ширина
// @returns string — отцентрированная строка
func padRight(s string, w int) string
```

## Реализация

### Обновление состояния (Update)

```go
// Update обрабатывает сообщения: KeyMsg, MouseMsg, WindowSizeMsg, ResizeMsg.
// Передаёт сообщения в панели вкладок для broadcast.
//
// @param msg tea.Msg — входящее сообщение
// @returns tea.Cmd — команда (batch для множественных)
func (tg *TabGroup) Update(msg tea.Msg) tea.Cmd
```

### Обработка клавиатуры (handleKeyMsg)

```go
// handleKeyMsg обрабатывает клавиатурные события:
//   - Ctrl+C — завершение работы (tea.Quit)
//   - Ctrl+Tab — NextTab
//   - Ctrl+Shift+Tab — PrevTab
//   - Ctrl+W — CloseTab (активная вкладка)
//   - Ctrl+T — NewTab (новая вкладка)
//   - Передаёт остальные клавиши активной вкладке
//
// @param msg tea.KeyMsg — клавиатурное сообщение
// @returns tea.Cmd — команда (обычно nil)
func (tg *TabGroup) handleKeyMsg(msg tea.KeyMsg) tea.Cmd
```

### Обработка мыши (handleMouseMsg)

```go
// handleMouseMsg обрабатывает клики:
//   - Если клик в таббаре (isOnTabBar) — handleTabBarClick
//   - Иначе передаёт клику активной вкладке
//
// @param msg tea.MouseMsg — мышиное сообщение
// @returns tea.Cmd — команда (обычно nil)
func (tg *TabGroup) handleMouseMsg(msg tea.MouseMsg) tea.Cmd
```

### Проверка положения таббара (isOnTabBar)

```go
// isOnTabBar проверяет, находится ли клик в области таббара.
//
// @param x — координата X
// @param y — координата Y
// @returns bool — true если клик в таббаре
func (tg *TabGroup) isOnTabBar(x, y int) bool
```

### Обработка клика по таббару (handleTabBarClick)

```go
// handleTabBarClick обрабатывает клики по таббару:
//   - Клик на вкладке — переключение (switchTab)
//   - Клик на + — создание новой вкладки (NewTab)
//   - Клик на × — закрытие вкладки (closeTab)
//
// @param msg tea.MouseMsg — мышиное сообщение с координатами
// @returns tea.Cmd — команда (обычно nil)
func (tg *TabGroup) handleTabBarClick(msg tea.MouseMsg) tea.Cmd
```

### Рендеринг таббара (renderTabBar)

```go
// renderTabBar рендерит горизонтальный или вертикальный таббар.
//
// @param width — ширина для рендеринга
// @returns string — отрендеренный текст таббара
func (tg *TabGroup) renderTabBar(width int) string
```

#### renderHorizontalTabBar

```go
// renderHorizontalTabBar рендерит горизонтальный таббар (TabTop, TabBottom).
//
// Форматирует названия вкладок:
//   - Активная вкладка: "▎ NAME ×"
//   - Неактивная вкладка: " NAME "
//   - Максимум 20 символов в названии
//
// @param width — полная ширина таббара
// @returns string — горизонтальный таббар
func (tg *TabGroup) renderHorizontalTabBar(width int) string
```

#### renderVerticalTabBar

```go
// renderVerticalTabBar рендерит вертикальный таббар (TabLeft, TabRight).
//
// Вычисляет максимальную ширину названий, выравнивает по правому краю.
//
// @returns string — вертикальный таббар (строка за строкой)
func (tg *TabGroup) renderVerticalTabBar(width int) string
```

### Калькуляция размеров (contentWidth, contentHeight, contentOffset)

```go
// contentWidth вычисляет ширину контента с учётом позиции таббара.
func (tg *TabGroup) contentWidth(totalW int) int

// contentHeight вычисляет высоту контента с учётом позиции таббара.
func (tg *TabGroup) contentHeight(totalH int) int

// contentOffset возвращает смещение X,Y для контента.
func (tg *TabGroup) contentOffset() (int, int)
```

### Стилизация (Styles)

Используются следующие стили (константы в коде):

- `inactiveTabStyle` — стиль неактивных вкладок
- `activeTabStyle` — стиль активной вкладки
- `newTabStyle` — стиль кнопки создания новой вкладки (+)
- `tabBarStyle` — общий стиль таббара

## Поведение

### Создание вкладки

1. Вызов `NewTab(name)` создаёт новую вкладку с заданным именем.
2. Новая вкладка становится активной.
3. Обновляется `tabRegions` для отрисовки таббара.

### Закрытие вкладки

1. Вызов `closeTab(idx)` удаляет вкладку из массива `tabs`.
2. Если активная вкладка была удалена, выбирается последняя оставшаяся.
3. Нельзя закрыть последнюю оставшуюся вкладку.

### Переключение вкладок

- `NextTab()`: `(activeTab + 1) % len(tabs)`
- `PrevTab()`: `(activeTab - 1 + len(tabs)) % len(tabs)`
- Клик на вкладке в таббаре переключает на неё.

### Рендеринг

1. Вызывается `View(w, h)` при каждом рендеринге.
2. Вычисляется позиция таббара и смещение контента.
3. Рендерится таббар + контент активной вкладки через `lipgloss.Join`.

### Обработка сообщений

- `tea.KeyMsg` — клавиатурные события
- `tea.MouseMsg` — мышиные события
- `tea.WindowSizeMsg` — изменение размера окна
- `ResizeMsg` — изменение размера сессии
- Остальные сообщения (PtyReadyMsg, PtyOutputMsg) — передаются всем вкладкам через `broadcastMsg`

## Тесты

Рекомендуемые тесты:

```go
func TestTabGroup_NewTab(t *testing.T) {
    tg := NewTabGroup(TabTop)
    tab := tg.NewTab("test")
    assert.Equal(t, 1, len(tg.tabs))
    assert.Equal(t, 0, tg.activeTab)
}

func TestTabGroup_CloseTab(t *testing.T) {
    tg := NewTabGroup(TabTop)
    tg.NewTab("a")
    tg.NewTab("b")
    tg.closeTab(0)
    assert.Equal(t, 1, len(tg.tabs))
    assert.Equal(t, "b", tg.ActiveTab().name)
}

func TestTabGroup_NextTab(t *testing.T) {
    tg := NewTabGroup(TabTop)
    tg.NewTab("a")
    tg.NewTab("b")
    tg.NewTab("c")
    tg.NextTab()
    assert.Equal(t, 1, tg.activeTab) // b активна
}

func TestTabGroup_PrevTab(t *testing.T) {
    tg := NewTabGroup(TabTop)
    tg.NewTab("a")
    tg.NewTab("b")
    tg.NewTab("c")
    tg.PrevTab()
    assert.Equal(t, 2, tg.activeTab) // c активна
}
```

## Ключевые Правила

1. Нельзя закрыть последнюю вкладку
1. Переключение вкладок циклическое
1. Активная вкладка всегда индекс 0-н-1 в массиве `tabs`
1. `tabRegions` пересчитывается при каждом рендеринге горизонтального таббара
1. Вертикальный таббар использует `padRight` для выравнивания
1. Максимум 20 символов в названии вкладки (горизонтальный)
1. Максимум 15 символов в названии вкладки (вертикальный)
1. `verticalTabWidth` устанавливается из максимального названия в вертикальном таббаре
1. `newTabRegion` — отдельная область справа от таббара для создания новых вкладок
1. Клик на + всегда создаёт новую вкладку, независимо от позиции
1. Закрыть можно только активную вкладку через Ctrl+W или клик на ×

## Совместимость

- Требует Bubble Tea (`github.com/charmbracelet/bubbletea`)
- Требует Lip Gloss (`github.com/charmbracelet/lipgloss`)
- Использует `unicode/utf8` для подсчёта символов
- Использует `fmt` и `strings` для форматирования
