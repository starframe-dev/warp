# Стили (`styles.go`)

## Обзор

Файл пакета `warp` определяет палитру Gruvbox Dark, внутренние цветовые переменные и стили компонентов интерфейса на базе `github.com/charmbracelet/lipgloss`. Также файл предоставляет публичные функции для получения трёх стилей границ.

## Цветовая палитра

| Переменная | Значение | Использование в файле |
|---|---|---|
| `gbDark0` | `#282828` | Фон таб-бара, элементов dropdown и плавающей панели |
| `gbDark1` | `#3c3836` | Граница split, фон заголовка плавающей панели и collapsible |
| `gbDark2` | `#504945` | Фон активной вкладки и кнопки/состояний dropdown |
| `gbDark3` | `#665c54` | Цвет hover-границы split |
| `gbDark4` | `#7c6f64` | Цвет границы collapsible |
| `gbGray` | `#928374` | Текст неактивной вкладки и граница плавающей панели |
| `gbLight1` | `#ebdbb2` | Текст активной вкладки, заголовка плавающей панели, collapsible и dropdown |
| `gbRed` | `#fb4934` | Цвет закрытия вкладки и плавающей панели |
| `gbGreen` | `#b8bb26` | Цвет индикатора новой вкладки и выбранного элемента dropdown |
| `gbYellow` | `#fabd2f` | Цвет drag-границы и hover-элемента dropdown |
| `gbBlue` | `#83a598` | Определён в палитре, в этом файле не используется |

## Публичный API

| Функция | Возвращаемое значение | Назначение |
|---|---|---|
| `BorderStyle()` | `lipgloss.Style` | Стиль обычной границы split |
| `BorderDragStyle()` | `lipgloss.Style` | Стиль границы split во время перетаскивания |
| `BorderHoverStyle()` | `lipgloss.Style` | Стиль границы split при наведении указателя |

Функции возвращают соответствующие внутренние переменные `borderStyle`, `borderDragStyle` и `borderHoverStyle`.

## Внутренние переменные и стили

Все перечисленные ниже переменные неэкспортируемые. Стили создаются через `lipgloss.NewStyle()`.

### Таб-бар

Цветовые переменные: `tabBarBg = gbDark0`, `activeTabBg = gbDark2`, `activeTabFg = gbLight1`, `inactiveTabFg = gbGray`, `newTabFg = gbGreen`, `closeTabFg = gbRed`.

| Стиль | Настройки |
|---|---|
| `tabBarStyle` | Background: `tabBarBg` |
| `activeTabStyle` | Background: `activeTabBg`, Foreground: `activeTabFg`, Bold |
| `inactiveTabStyle` | Background: `tabBarBg`, Foreground: `inactiveTabFg` |
| `newTabStyle` | Foreground: `newTabFg` |
| `closeTabStyle` | Foreground: `closeTabFg` |

### Границы split

Цветовые переменные: `borderColor = gbDark1`, `borderDragColor = gbYellow`, `borderHoverColor = gbDark3`.

| Стиль | Настройки |
|---|---|
| `borderStyle` | Foreground: `borderColor` |
| `borderHoverStyle` | Foreground: `borderHoverColor` |
| `borderDragStyle` | Foreground: `borderDragColor` |
| `collapseStyle` | Foreground: `borderColor` |

### Плавающие панели

Цветовые переменные: `floatBorderColor = gbGray`, `floatTitleBg = gbDark1`, `floatTitleFg = gbLight1`, `floatBg = gbDark0`, `floatCloseFg = gbRed`.

| Стиль | Настройки |
|---|---|
| `floatBorderStyle` | Foreground: `floatBorderColor` |
| `floatTitleStyle` | Background: `floatTitleBg`, Foreground: `floatTitleFg`, Bold |
| `floatCloseStyle` | Foreground: `floatCloseFg`, Bold |
| `floatBgStyle` | Background: `floatBg` |

### Collapsible

| Стиль | Настройки |
|---|---|
| `collapsibleStyle` | Foreground: `gbLight1`, Background: `gbDark1` |
| `collapsibleBorderStyle` | Foreground: `gbDark4` |

### Dropdown

| Стиль | Настройки |
|---|---|
| `dropdownButtonStyle` | Background: `gbDark2`, Foreground: `gbLight1` |
| `dropdownItemStyle` | Background: `gbDark0`, Foreground: `gbLight1` |
| `dropdownItemHoverStyle` | Background: `gbDark2`, Foreground: `gbYellow` |
| `dropdownItemSelectedStyle` | Background: `gbDark2`, Foreground: `gbGreen`, Bold |

## Дополнительные сведения

Стили используют заданные в файле цвета Gruvbox Dark и настройки `Background`, `Foreground` и `Bold`. В файле не задаются прозрачность, градиенты или анимация. Комментарий в исходном файле указывает, что стили popover находятся в `popover.go`; они здесь не определены.
