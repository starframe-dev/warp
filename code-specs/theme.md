# theme.go

## Наззначение

Файл `theme.go` реализует механизм переопределения палитры и стилей TUI-пакета `warp`. Он предоставляет публичный API `SetTheme`, который по данным `ThemeColors` пересчитывает все package-level цветовые константы (`gbDark0`, `gbDark1`, ...) и производные `lipgloss.Style`, используемые в контролах и оверлеях (табы, фолоты, дропдауны, поповеры, модалки, инпуты, коллапсабилы).

## Публичный API

### `ThemeColors` (struct)

```go
type ThemeColors struct {
    Background          string  // цвет фона
    Surface             string  // цвет поверхности
    Raised             string  // приподнятый уровень
    Border             string  // цвет границы
    BorderMuted      string  // приглушённая граница
    Text                string  // основной текст
    TextMuted         string  // приглушённый текст
    TextStrong        string  // сильный/яркий текст
    Accent            string  // акцентный цвет
    AccentMuted       string  // приглушённый акцент
    Error             string  // цвет ошибки
    Success           string  // цвет успеха
    Warning           string  // цвет предупреждения
    SelectionBackground  string  // фон выделения
    SelectionForeground  string  // передний план выделения
}
```

Все поля — строки hex-кода (например, `"#123456"`), интерпретируемые `lipgloss.Color(...)`.

### `SetTheme(colors ThemeColors)`

Глобальная функция, пересчитывающая все package-level палитровые константы и производные стили.

Поведение:
1. Преобразует каждое поле `ThemeColors` в `lipgloss.Color` и присваивает их пакетным переменным-константам (`gbDark0` = Background, `gbDark1` = Surface, `gbDark2` = Raised, `gbDark3` = BorderMuted, `gbDark4` = Border, `gbGray` = TextMuted, `gbLight1` = TextStrong, `gbRed` = Error, `gbGreen` = Success, `gbYellow` = Warning, `gbBlue` = Accent).
2. Пересчитает производные цвета (`tabBarBg`, `activeTabBg`, `activeTabFg`, `inactiveTabFg`, `newTabFg`, `closeTabFg`, `borderColor`, `borderDragColor`, `borderHoverColor`, `floatBorderColor`, `floatTitleBg`, `floatTitleFg`, `floatBg`, `floatCloseFg`).
3. Пересоздаёт все производные `lipgloss.Style` (`tabBarStyle`, `activeTabStyle`, `inactiveTabStyle`, `newTabStyle`, `closeTabStyle`, `borderStyle`, `borderHoverStyle`, `borderDragStyle`, `collapseStyle`, `floatBorderStyle`, `floatTitleStyle`, `floatCloseStyle`, `floatBgStyle`, `collapsibleStyle`, `collapsibleBorderStyle`, `dropdownButtonStyle`, `dropdownItemStyle`, `dropdownItemHoverStyle`, `dropdownItemSelectedStyle`, `popoverBaseStyle`, `popoverSelectedStyle`, `modalBorderStyle`, `dimStyle`, `inputStyle`, `inputBorderStyle`, `inputFocusBorderStyle`).

Важные детали:
- `Text` и `AccentMuted` из `ThemeColors` не используются в `SetTheme` — они записаны в структуру, но не имеют соответствующих пакетных констант.
- `SelectionBackground` и `SelectionForeground` используются только для `activeTabBg`/`activeTabFg`.
- Функция не возвращает ошибок: `lipgloss.Color` принимает произвольные строки; невалидные hex-значения не вызывают ошибку (поведение зависит от `lipgloss`).

## Взаимосвязь с остальным кодом

- Палитровые константы (`gbDark0..gbDark4`, `gbGray`, `gbLight1`, `gbRed`, `gbGreen`, `gbYellow`, `gbBlue`) и производные цвета (`tabBarBg`, `activeTabBg`, ...) объявлены в `styles.go`.
- `SetTheme` перезаписывает значения, инициализированные в `styles.go` при первом вызове.
- Производные стили используются в `tabgroup.go` (таб-бар), `float.go` (фолот-окна), `dropdown.go` (выпадающие списки), `modal.go` (модальные окна), `input.go` (поле ввода), `collapsible.go` (коллапс-секции).

## Тестирование

- `TestSetTheme` в `theme_test.go` сохраняет оригинальные значения палитровых констант, вызывает `SetTheme` с единым hex-кодом `"#123456"` и проверяет, что все палитровые константы и производные стили принимают заданный цвет.
- Тест восстанавливает исходные значения по завершении.

## Ограничения

- API не поддерживает именованную палитру уровней (например, `gbDark2` напрямую) — только семантические роли из `ThemeColors`.
- `SetTheme` меняет package-level переменные и стили всего процесса, а не отдельного `Warp`.
- `SetTheme` не синхронизирован с рендерингом; вызывайте его до запуска Bubble Tea и не меняйте тему одновременно с `View`.
- Поля `Text` и `AccentMuted` пока не используются реализацией `SetTheme`.
- После вызова последующие рендеры использующих эти стили контролов используют новые цвета без перезапуска.
