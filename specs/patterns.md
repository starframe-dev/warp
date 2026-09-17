# Паттерны проектирования

## Композиция вместо наследования

Все компоненты — `Panel`. Вложенные компоненты не знают
о своих родителях: `Node` хранит `*Node`/`Panel` без типа.
Новый компонент = реализация `Panel` (+ опционально
`Focusable`, `ElementProvider`, `RawKeyReceiver`).

```go
type Collapsible struct {
    Title     string
    Collapsed bool
    Content   Panel
}
```

## TabGroup — Panel, а не корневой тип

`TabGroup` сам по себе `Panel` и вставляется в flex/split
рядом с любым другим компонентом:

```go
tg := warp.NewTabGroup(warp.TabLeft)
tab.FlexRow(root, []warp.FlexItemSpec{
    {Panel: tg, Grow: 2},
})
```

Warp-конструктор лишь создаёт `TabGroup` как корень по
умолчанию; `SetRoot` заменяет его полностью.

## Warp — тонкий model, а не контейнер

`Warp` не биндит клавиши, не обрабатывает мышь, не хранит
глобальное состояние. `Update` смотрит только
`WindowSizeMsg`; остальное уходит в `root.Panel`.
Это позволяет использовать Warp как поддерево
в чужом Bubbletea-приложении.

## Focus — явное API

Warp **не** биндит `Tab`/`Shift+Tab`. Разработчик сам
решает, какие клавиши вызывают `FocusNext/Prev/First`:

```go
case "tab":       tab.FocusNext()
case "shift+tab": tab.FocusPrev()
```

Для PTY/терминалов есть `RawKeyReceiver` — `WantsRawKeys()`
возвращает `true`, и Warp не перехватывает ни одного
клавиша.

## Float z-order

`FloatPane` — не часть дерева, а отдельный срез `[]*FloatPane`
у `Tab`. Порядок в срезе = z-order. Клик по float'у
перемещает его в конец (z-подъём). `overlayFloat` нарисует
поверх строк основного дерева с учётом ANSI-кодов.

## CloseOnOutsideClick

`FloatPane.CloseOnOutsideClick` — если клик пришёлся
не по float'у, float закрывается. Аналогично
`Popover` (у `Popover` есть `OnClose` — вызывается
при клике вне, а также по `Esc`).

## Element tree — контракт для E2E

`ElementProvider` — интерфейс для семантического
дерева UI: `{Elements(w, h) []Element}`. Компоненты
реализуют сами; Warp сам не реализует
(кроме `Warp.AsPanel` через `collectElements`).

HTTP `/elements` — экспорт этого дерева для E2E-тестов.

## ANSI-aware рендеринг

- `StripANSI` считает **визуальную** ширину строки
  (не bytes), чтобы `padContent` не резала ANSI-сиквенции.
- `padContent` заканчивает каждую строку `ansi.ResetStyle` —
  стили панелей не утекают в границы между ними.
- Границы `│`/`─` рисуются в `borderStyle`
  с обёрткой `ResetStyle` по краям.

## Resize через `ResizeMsg`

Когда размер панели изменился, Warp отправляет
`ResizeMsg { Width, Height }` через `tea.Msg`.
Панели могут использовать его для перерисовки
кэшированных строк. `Tab` сам отслеживает размеры
через `View(w, h)`.

## Состояние — локальное у каждого компонента

Никаких глобалок, подписок, эмиттеров. Каждый
компонент хранит своё состояние в полях:

```go
type Input struct {
    Value   string
    Cursor  int
    Prompt  string
    focused bool
}
```

Коммуникация — только через `tea.Msg`.

## Theme — единая точка настройки

`SetTheme(ThemeColors)` — пересчитывает **все** package-level
`lipgloss.Style` (`tabBarStyle`, `borderStyle`, `floatTitleStyle`,
`modalBorderStyle`, `inputStyle`, …) из одной структуры.
До `SetTheme` — палитра Gruvbox Dark из `styles.go`.
