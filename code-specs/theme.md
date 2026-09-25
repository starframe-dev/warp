# theme.go

## Назначение

Файл `theme.go` предоставляет публичный API для перенастройки цветов и связанных `lipgloss.Style` пакета `warp`. `SetTheme` записывает новые значения в package-level переменные, объявленные в других файлах пакета, и пересоздаёт используемые пакетные стили.

## Публичный API

### `ThemeColors`

Структура содержит строки для семантических цветов: `Background`, `Surface`, `Raised`, `Border`, `BorderMuted`, `Text`, `TextMuted`, `TextStrong`, `Accent`, `AccentMuted`, `Error`, `Success`, `Warning`, `SelectionBackground` и `SelectionForeground`. Строки передаются в `lipgloss.Color`; структура не проверяет, что они имеют формат hex-кода.

### `SetTheme(colors ThemeColors)`

Функция не возвращает ошибку и выполняет следующие назначения:

- `Background` → `gbDark0`; `Surface` → `gbDark1`; `Raised` → `gbDark2`; `BorderMuted` → `gbDark3`; `Border` → `gbDark4`.
- `TextMuted` → `gbGray`; `TextStrong` → `gbLight1`; `Error` → `gbRed`; `Success` → `gbGreen`; `Warning` → `gbYellow`; `Accent` → `gbBlue`.
- `Text` и `AccentMuted` в текущей реализации не используются.
- `tabBarBg` получает `gbDark0`; `activeTabBg` и `activeTabFg` получают соответственно `SelectionBackground` и `SelectionForeground`; `inactiveTabFg`, `newTabFg` и `closeTabFg` получают `gbGray`, `gbGreen` и `gbRed`.
- `borderColor`, `borderDragColor`, `borderHoverColor` получают `gbDark1`, `gbYellow`, `gbDark3`. `floatBorderColor`, `floatTitleBg`, `floatTitleFg`, `floatBg`, `floatCloseFg` получают `gbGray`, `gbDark1`, `gbLight1`, `gbDark0`, `gbRed`.

Затем пересоздаются стили табов, границ, плавающих панелей, collapsible-компонентов, dropdown, popover, modal, dim и input: `tabBarStyle`, `activeTabStyle`, `inactiveTabStyle`, `newTabStyle`, `closeTabStyle`, `borderStyle`, `borderHoverStyle`, `borderDragStyle`, `collapseStyle`, `floatBorderStyle`, `floatTitleStyle`, `floatCloseStyle`, `floatBgStyle`, `collapsibleStyle`, `collapsibleBorderStyle`, `dropdownButtonStyle`, `dropdownItemStyle`, `dropdownItemHoverStyle`, `dropdownItemSelectedStyle`, `popoverBaseStyle`, `popoverSelectedStyle`, `modalBorderStyle`, `dimStyle`, `inputStyle`, `inputBorderStyle`, `inputFocusBorderStyle`. Их атрибуты (фон, передний план, граница, отступы и жирность) задаются непосредственно в `SetTheme`; в частности, `modalBorderStyle` получает скруглённую границу и `Padding(1, 2)`.

`SetTheme` меняет общие для пакета переменные и стили, а не тему отдельного экземпляра `Warp`. Вызовы не синхронизированы с рендерингом.

## Связанные объявления

Палитровые и производные цветовые переменные и большинство перечисленных стилей объявлены в `styles.go`; стили popover, modal, dim и input также объявлены в соответствующих файлах пакета. `SetTheme` пересваивает их значения. Обновлённые стили применяются последующими обращениями к этим package-level переменным.

## Тест

`TestSetTheme` вызывает `SetTheme` с одинаковым цветом `#123456`, проверяет обновление палитровых переменных, нескольких производных цветов и ряда атрибутов стилей. Он сохраняет и восстанавливает палитровые переменные. Тест не проверяет каждый стиль и каждый его атрибут, перечисленные выше.
