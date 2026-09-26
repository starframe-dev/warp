# Scrollable — спецификация

## Назначение

`Scrollable` оборачивает `Panel` и показывает viewport высотой `h` строк с вертикальным смещением. Intrinsic-высота необязательна и передаётся через `ContentHeightProvider` без изменения интерфейса `Panel`.

## Публичный API

```go
func NewScrollable(content Panel) *Scrollable
func (s *Scrollable) View(w, h int) string
func (s *Scrollable) Elements(w, h int) []Element
func (s *Scrollable) ContentHeight(width int) (height int, known bool)
func (s *Scrollable) Update(msg tea.Msg) tea.Cmd
```

`Scrollable.ContentHeight` всегда сообщает `known == false`: видимая область вложенного `Scrollable` не является его intrinsic-высотой. `Offset` — запрошенная строковая позиция; отрицательное значение нормализуется к нулю.

## Определение effective offset

Все `View`, `Elements` и `Update` используют общую функцию effective-offset.

1. Если вложенный `Panel` сообщает известную высоту через `ContentHeightProvider`, отрицательная высота нормализуется к нулю, а offset ограничивается `max(0, contentHeight-viewportHeight)`.
2. Иначе запрашивается probe высотой `saturatingAdd(max(0, Offset), max(0, h), 1)`. Конец считается найденным только когда `strings.Split(View, "\n")` содержит меньше строк, чем probe. Если число строк совпало с запросом, intrinsic-высота остаётся неизвестной и положительный offset не ограничивается.
3. Fallback extent не кэшируется между вызовами: он повторно определяется при текущей ширине/состоянии содержимого. Результат probe используется `View`, чтобы не рендерить fallback содержимое повторно в одном вызове.
4. Суммы saturate на `MaxInt`; известная высота и размеры viewport нормализуются к неотрицательным значениям.

Панели, дополняющие каждый `View` до запрошенной высоты, остаются с неизвестным extent. Для точной ограниченной прокрутки следует реализовать `ContentHeightProvider`.

## View и Elements

- `View(w, h)` нормализует отрицательные размеры, сохраняет последний viewport и рендерит ровно `h` строк. При известной высоте и наличии `ViewportRenderer` вызывает `ViewAt(w,h,offset)`; иначе при известном extent запрашивает `effectiveOffset+h`, а после fallback использует строки probe. Недостающие строки дополняются пробелами, `Offset` сохраняет effective значение.
- `Elements(w, h)` использует тот же effective offset. При известном extent и наличии `ViewportElementProvider` запрашивает только viewport; иначе получает элементы до `effectiveOffset+h`. Затем клонирует дерево, обрезает его по `[0,w) × [effectiveOffset,effectiveOffset+h)` и сдвигает Y на `-effectiveOffset`. `ElementsAt` возвращает bounds в координатах полного содержимого. Запрос не изменяет `Offset`.
- При nil-панели `View` возвращает ровно `h` пустых строк; `Elements` возвращает nil.

## Update и размеры

- Колесо мыши вверх/вниз меняет смещение на 3 строки; `up`/`down` — на 1; `pgup`/`pgdown` — на 10. Смещение не уходит ниже нуля и положительное сложение насыщается на `MaxInt`.
- `ResizeMsg` задаёт локальные размеры viewport. `WindowSizeMsg` используется предварительно до получения локального resize; размеры, известные из `View`, также используются для нормализации после прокрутки.
- После прокрутки/изменения viewport `Update` нормализует сохранённый offset, если доступно его разрешение.
- Все сообщения, включая те, что обработаны как прокрутка, передаются в `Content.Update(msg)`.

## Optional viewport fast path

```go
type ViewportRenderer interface {
    ViewAt(width, height, offset int) string
}

type ViewportElementProvider interface {
    ElementsAt(width, height, offset int) []Element
}
```

Эти интерфейсы не меняют `Panel` и используются только вместе с известным `ContentHeightProvider`. Реализация должна возвращать данные ровно запрошенного диапазона, не строя строки или semantic elements перед `offset`. Без интерфейсов остаётся обычный путь, включая fallback по probe.

## Wrapper extent

- `Selectable.ContentHeight` прозрачно возвращает известную высоту вложенного контента; unknown остаётся unknown.
- `Collapsible.ContentHeight` возвращает 1 в свёрнутом состоянии. В раскрытом состоянии добавляет строку заголовка только к известной высоте содержимого.

## Проверки

Регрессии проверяют provider и padded `View`, overscroll, `Update` → `/elements` до `View`, изменение ширины и высоты контента, natural-height fallback, неизвестный padded fallback, forwarding, wrapper propagation и overflow около `MaxInt`.
