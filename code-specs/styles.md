# Styles

## Обзор

Файл `styles.go` содержит цветовые палитры и стили для UI компонентов проекта Warp, построенного на базе библиотеки [lipgloss](https://github.com/charmbracelet/lipgloss). Все стили основаны на цветовой палитре **Gruvbox Dark**.

## Цветовая палитра

### Gruvbox Dark

Используются следующие цвета из палитры Gruvbox Dark:

| Имя | Цвет | Использование |
|-----|------|---------------|
| gbDark0 | `#282828` | Основной фон (background) |
| gbDark1 | `#3c3836` | Вторичный фон, границы |
| gbDark2 | `#504945` | Фон активных элементов |
| gbDark3 | `#665c54` | Тertiary фон |
| gbDark4 | `#7c6f64` | Четвёртый уровень фона |
| gbGray | `#928374` | Серый, текст по умолчанию |
| gbLight1 | `#ebdbb2` | Светлый текст |
| gbRed | `#fb4934` | Красный (ошибки, закрытие) |
| gbGreen | `#b8bb26` | Зелёный (успех, новые элементы) |
| gbYellow | `#fabd2f` | Жёлтый (внимание, hover) |
| gbBlue | `#83a598` | Синий (информация)

## Публичный API

### Функции

| Функция | Возврат | Описание |
|---------|---------|----------|
| `BorderStyle()` | `lipgloss.Style` | Стиль для обычных split border |
| `BorderDragStyle()` | `lipgloss.Style` | Стиль при перетаскивании border |
| `BorderHoverStyle()` | `lipgloss.Style` | Стиль при наведении мыши на border |

Все публичные функции возвращают соответствующий `lipgloss.Style`, собранный из неэкспортируемых стилей (см. ниже).

## Внутренние стили и переменные

Все стили неэкспортируемые (camelCase) и создаются через `lipgloss.NewStyle()`. Ниже перечислены стили и переменные, которые действительно присутствуют в файле.

### Tab bar styles

| Стиль | Цвета |
|-------|-------|
| `tabBarStyle` | Background: gbDark0 |
| `activeTabStyle` | Background: gbDark2, Foreground: gbLight1, Bold |
| `inactiveTabStyle` | Background: gbDark0, Foreground: gbGray |
| `newTabStyle` | Foreground: gbGreen |
| `closeTabStyle` | Foreground: gbRed |

Переменные цвета для таб-бара:

| Переменная | Значение | Описание |
|------------|----------|----------|
| `tabBarBg` | gbDark0 | Фон таб-бара |
| `activeTabBg` | gbDark2 | Фон активной вкладки |
| `activeTabFg` | gbLight1 | Текст активной вкладки |
| `inactiveTabFg` | gbGray | Текст неактивной вкладки |
| `newTabFg` | gbGreen | Цвет индикатора новой вкладки |
| `closeTabFg` | gbRed | Цвет кнопки закрытия |

### Border styles

| Стиль | Цвета |
|-------|-------|
| `borderStyle` | Foreground: gbDark1 |
| `borderHoverStyle` | Foreground: gbDark3 |
| `borderDragStyle` | Foreground: gbYellow |
| `collapseStyle` | Foreground: gbDark1 |

Переменные цвета split-границ:

| Переменная | Значение | Описание |
|------------|----------|----------|
| `borderColor` | gbDark1 | Обычный цвет границы |
| `borderDragColor` | gbYellow | Цвет при drag |
| `borderHoverColor` | gbDark3 | Цвет при hover |

### Float pane styles

| Стиль | Цвета |
|-------|-------|
| `floatBorderStyle` | Foreground: gbGray |
| `floatTitleStyle` | Background: gbDark1, Foreground: gbLight1, Bold |
| `floatCloseStyle` | Foreground: gbRed, Bold |
| `floatBgStyle` | Background: gbDark0 |

Переменные цвета плавающих окон:

| Переменная | Значение | Описание |
|------------|----------|----------|
| `floatBorderColor` | gbGray | Граница плавающего окна |
| `floatTitleBg` | gbDark1 | Фон заголовка плавающего окна |
| `floatTitleFg` | gbLight1 | Текст заголовка |
| `floatBg` | gbDark0 | Фон плавающего окна |
| `floatCloseFg` | gbRed | Кнопка закрытия |

### Collapsible styles

| Стиль | Цвета |
|-------|-------|
| `collapsibleStyle` | Foreground: gbLight1, Background: gbDark1 |
| `collapsibleBorderStyle` | Foreground: gbDark4 |

### Dropdown styles

| Стиль | Цвета |
|-------|-------|
| `dropdownButtonStyle` | Background: gbDark2, Foreground: gbLight1 |
| `dropdownItemStyle` | Background: gbDark0, Foreground: gbLight1 |
| `dropdownItemHoverStyle` | Background: gbDark2, Foreground: gbYellow |
| `dropdownItemSelectedStyle` | Background: gbDark2, Foreground: gbGreen, Bold |

## Паттерны реализации

### Стилизация через lipgloss

Все стили создаются через `lipgloss.NewStyle()` с последующим вызовом методов `Background()`, `Foreground()`, `Bold()`.

```go
activeTabStyle = lipgloss.NewStyle().
        Background(activeTabBg).
        Foreground(activeTabFg).
        Bold(true)
```

### Композиция стилей

Стили строятся композицией базовых цветовых переменных. Каждая цветовая схема соответствует определённой части UI (tab bar, split borders, float panes).

## Ограничения

- Используются только цвета из палитры Gruvbox Dark
- Все стили создаются через `lipgloss`
- Нет прозрачности (opacity) в стилях
- Нет градиентов или сложной анимации

## Использование

```go
import "warp"

// Стиль для границы
style := warp.BorderStyle()

// Применение к тексту
text := lipgloss.NewStyle().Render("text")
```

## Ключевые Правила

1. **Использовать только публичный API** — вызывать `BorderStyle()`, `BorderDragStyle()` вместо прямого доступа к переменным
2. **Не изменять цветовую палитру** — цвета определены в начале файла и не должны меняться
3. **Соблюдать семантику** — каждый стиль соответствует определённой части UI
4. **Генерировать стили через конструкторы** — использовать `lipgloss.NewStyle()` с методами

### Ключевые Правила (Human Horizon)

- **Всегда проверяй ошибки** при работе с `lipgloss`
- **Маленькие интерфейсы** — каждый стиль отвечает за одну задачу
- **Явное важнее неявного** — цветовые палитры явно определены в начале файла
- **camelCase для неэкспортируемых** — все переменные стилизованы в camelCase
