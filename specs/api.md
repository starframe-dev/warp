# Публичный API Warp

## Назначение

Warp — Go-библиотека для компоновки терминальных панелей Bubble Tea. Пакет `github.com/starframe-dev/warp` предоставляет модель `Warp`, дерево split/flex на уровне `Tab`, вкладки и компоненты-обёртки. Содержимое панелей и прикладная логика остаются у вызывающего кода; визуальные размеры задаются в терминальных ячейках.

## Минимальная панель и расширения

```go
type Panel interface {
    View(width, height int) string
    Update(msg tea.Msg) tea.Cmd
}
```

Дополнительные возможности необязательны и независимы:

- `Focusable` (`Panel`, `Focus()`, `Blur()`, `Focused() bool`) участвует в явном управлении фокусом.
- `RawKeyReceiver` (`Panel`, `WantsRawKeys() bool`) просит передавать клавиши без перехвата сочетаний TabGroup.
- `ElementProvider.Elements(width, height) []Element` публикует семантические элементы для инспектора и E2E.
- `ContentHeightProvider.ContentHeight(width) (height int, known bool)` сообщает intrinsic-высоту без рендеринга.
- `ViewportRenderer.ViewAt(width, height, offset int) string` и `ViewportElementProvider.ElementsAt(width, height, offset int) []Element` дают быстрый viewport-путь Scrollable при известной высоте. Bounds `ElementsAt` остаются в координатах полного содержимого.
- `Unmounter.Unmount()` освобождает ресурсы при окончательном удалении панели из домена владения.
- `SemanticStableMsg.SemanticStateUnchanged() bool` позволяет сообщению обещать неизменность semantic snapshot.

`BasePanel` — пустая базовая реализация `Panel`.

## Корень и вкладки

```go
w := warp.New()                  // TabGroup TabTop с вкладкой main
w.SetRoot(customPanel)           // заменить корневую Panel
root := w.Root()
tab := w.ActiveTab()
newTab := w.NewTab("editor")
w.SetTabPosition(warp.TabLeft)
w.NextTab(); w.PrevTab()
err := w.Run()
```

`Warp` реализует `tea.Model`. `Init` пуст; `Update` сохраняет `WindowSizeMsg`, передаёт сообщение корню и возвращает его `tea.Cmd`. `View` рендерит корневую панель; до получения размера возвращается `Loading...`. Warp не назначает собственные клавиатурные биндинги.

`NewTabGroup(position)` создаёт Panel с вкладкой `main`; `TabGroup.NewTab`, `ActiveTab`, `NextTab`, `PrevTab` управляют набором вкладок. Позиции: `TabTop`, `TabBottom`, `TabLeft`, `TabRight`, `TabNone`.

`Tab` управляет layout-деревом: `RootPanel`, `SetRootPanel`, `SplitVertical`, `SplitHorizontal`, `FlexRow`, `FlexColumn`, `Float`, `CloseFloat`, `SetSplitCollapse`, `ToggleSplitCollapse`, а также `FocusNext`, `FocusPrev`, `FocusFirst`, `FocusPanel`. Split fraction относится к первому ребёнку; `FlexItemSpec{Panel, Grow}` задаёт детей flex и веса роста. Float располагается поверх дерева.

## Компоненты

Конструкторы основных компонентов: `NewCollapsible(title, panel)`, `NewScrollable(panel)`, `NewDropdownMenu(label, items)`, `NewSelectable(panel)`, `NewInput(prompt)`, `NewModal(...)`. Контекстное меню можно создать как `&Popover{Items, X, Y, OnClose}`. Их поля и точная семантика описаны в `code-specs/*.md`; компоненты реализуют `Panel` и поддерживают соответствующие optional interfaces.

Вкладки и компоненты могут быть вложены друг в друга как `Panel`. `w.AsPanel()` возвращает адаптер Warp для такого встраивания.

## Фокус и события

Фокус управляется явно через методы `Tab.Focus*`; Warp не связывает Tab/Shift+Tab с навигацией. TabGroup резервирует `Ctrl+T` (новая вкладка), `Ctrl+W` (закрыть текущую, кроме последней), `Ctrl+Tab` и `Ctrl+Shift+Tab` (переключение), а также `Ctrl+C`. Фокусируемая raw-key панель может получать клавиши первой. Точная маршрутизация мыши описана в архитектуре и спецификациях компонентов.

## Семантические элементы

`Element` содержит `Role`, `Name`, `Action`, `Bounds` и `Children`; `Bounds` имеет координаты и размер в ячейках, `Center()` возвращает центральную ячейку. `FindElement(elems, role, name, action)` ищет первое совпадение в глубину; пустой критерий игнорируется.

## Тема и утилиты

`SetTheme(ThemeColors)` перенастраивает package-level стили компонентов. `ThemeColors` включает Background/Surface/Raised, Border/BorderMuted, Text/TextMuted/TextStrong, Accent/AccentMuted, Error/Success/Warning и цвета SelectionBackground/SelectionForeground.

Экспортируемые строковые утилиты: `WordWrap(text, width)`, `SpaceWrap(text, width)` и `StripANSI(s)`. Последняя удаляет ANSI-последовательности; визуальная ширина текста учитывает терминальные ячейки.

## Совместимость

`Panel` остаётся минимальным обязательным интерфейсом. Новые свойства добавляются через optional interfaces, поэтому пользовательские панели не обязаны реализовывать высоту, viewport, семантические элементы, фокус или lifecycle.
