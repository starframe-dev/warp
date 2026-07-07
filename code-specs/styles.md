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
| gbBlue | `#83a598` | Синий (информация) |

## Публичный API

### Функции

| Функция | Возврат | Описание |
|---------|---------|----------|
| `BorderStyle()` | `lipgloss.Style` | Стиль для обычных split border |
| `BorderDragStyle()` | `lipgloss.Style` | Стиль при перетаскивании border |
| `BorderHoverStyle()` | `lipgloss.Style` | Стиль при наведении мыши на border |

### Типы и стили

#### Tab bar styles

| Стиль | Описание | Цвета |
|-------|----------|-------|
| `tabBarStyle` | Фон таб-барa | Background: gbDark0 |
| `activeTabStyle` | Активная вкладка | Background: gbDark2, Foreground: gbLight1, Bold |
| `inactiveTabStyle` | Неактивная вкладка | Background: gbDark0, Foreground: gbGray |
| `newTabStyle` | Новая вкладка (индикатор) | Foreground: gbGreen |
| `closeTabStyle` | Кнопка закрытия | Foreground: gbRed |

#### Border styles

| Стиль | Описание | Цвета |
|-------|----------|-------|
| `borderStyle` | Обычная граница | Foreground: gbDark1 |
| `borderHoverStyle` | Hover граница | Foreground: gbDark3 |
| `borderDragStyle` | Drag граница | Foreground: gbYellow |

#### Float pane styles

| Стиль | Описание | Цвета |
|-------|----------|-------|
| `floatBorderStyle` | Граница плавающего окна | Foreground: gbGray |
| `floatTitleStyle` | Заголовок плавающего окна | Background: gbDark1, Foreground: gbLight1, Bold |
| `floatCloseStyle` | Кнопка закрытия плавающего окна | Foreground: gbRed, Bold |
| `floatBgStyle` | Фон плавающего окна | Background: gbDark0 |

#### Collapsible styles

| Стиль | Описание | Цвета |
|-------|----------|-------|
| `collapsibleStyle` | Сворачиваемый элемент | Foreground: gbLight1, Background: gbDark1 |
| `collapsibleBorderStyle` | Граница сворачиваемого | Foreground: gbDark4 |

#### Dropdown styles

| Стиль | Описание | Цвета |
|-------|----------|-------|
| `dropdownButtonStyle` | Кнопка dropdown | Background: gbDark2, Foreground: gbLight1 |
| `dropdownItemStyle` | Элемент списка dropdown | Background: gbDark0, Foreground: gbLight1 |
| `dropdownItemHoverStyle` | Hover элемент dropdown | Background: gbDark2, Foreground: gbYellow |
| `dropdownItemSelectedStyle` | Выбранный элемент dropdown | Background: gbDark2, Foreground: gbGreen, Bold |

## Внутренние переменные

### Tab bar colors

| Переменная | Значение | Описание |
|------------|----------|----------|
| `tabBarBg` | gbDark0 | Фон таб-бара |
| `activeTabBg` | gbDark2 | Фон активной вкладки |
| `activeTabFg` | gbLight1 | Текст активной вкладки |
| `inactiveTabFg` | gbGray | Текст неактивной вкладки |
| `newTabFg` | gbGreen | Цвет индикатора новой вкладки |
| `closeTabFg` | gbRed | Цвет кнопки закрытия |

### Split border colors

| Переменная | Значение | Описание |
|------------|----------|----------|
| `borderColor` | gbDark1 | Обычный цвет границы |
| `borderDragColor` | gbYellow | Цвет при drag |
| `borderHoverColor` | gbDark3 | Цвет при hover |

### Float pane colors

| Переменная | Значение | Описание |
|------------|----------|----------|
| `floatBorderColor` | gbGray | Граница плавающего окна |
| `floatTitleBg` | gbDark1 | Фон заголовка плавающего окна |
| `floatTitleFg` | gbLight1 | Текст заголовка |
| `floatBg` | gbDark0 | Фон плавающего окна |
| `floatCloseFg` | gbRed | Кнопка закрытия |

### Collapsible colors

| Переменная | Значение | Описание |
|------------|----------|----------|
| `collapsibleStyle` | gbLight1 / gbDark1 | Стиль сворачиваемого |
| `collapsibleBorderStyle` | gbDark4 | Граница |

### Dropdown colors

| Переменная | Значение | Описание |
|------------|----------|----------|
| `dropdownButtonStyle` | gbDark2 / gbLight1 | Кнопка |
| `dropdownItemStyle` | gbDark0 / gbLight1 | Элемент списка |
| `dropdownItemHoverStyle` | gbDark2 / gbYellow | Hover |
| `dropdownItemSelectedStyle` | gbDark2 / gbGreen, Bold | Выбрано |

## Паттерны реализации

### Стилизация через липгloss

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
import "github.com/charmbracelet/warp/styles"

// Использование стилей
style := styles.BorderStyle()
text := lipgloss.NewStyle().Foreground(styles.gbLight1).Render("text")
```

## Ключевые Правила

1. **Использовать только публичный API** — вызывать `BorderStyle()`, `BorderDragStyle()` вместо прямого доступа к переменным
2. **Не изменять цветовую палитру** — цвета определены в начале файла и не должны меняться
3. **Соблюдать семантику** — каждый стиль соответствует определённой части UI
4. **Генерировать стили через конструкторы** — использовать `lipgloss.NewStyle()` с методами

## Ключевые Правила (Human Horizon)

- **Всегда проверяй ошибки** при работе с `lipgloss`
- **Маленькие интерфейсы** — каждый стиль отвечает за одну задачу
- **Явное важнее неявного** — цветовые палитры явно определены в начале файла
- **Английские комментарии** — все публичные функции должны иметь English doc comments
- **camelCase для неэкспортируемых** — все переменные стилизованы в camelCase
