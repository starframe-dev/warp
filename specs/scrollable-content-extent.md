# Intrinsic content extent для Scrollable

## Контекст

`Scrollable` выводит максимальный `Offset` из числа строк `Content.View`, хотя валидная Panel может дополнять результат до запрошенной высоты. Такое дополнение не раскрывает intrinsic content height и позволяет бесконечно прокручивать пустоту. Отдельный расчёт в `Elements` также может не совпадать с `View`.

## Цель

Добавить необязательный width-aware `ContentHeightProvider`, использовать один effective offset для `View`, `Elements` и нормализации после `Update`, а для старых Panel применять только консервативное обнаружение конца. Не менять обязательный интерфейс `Panel` и маршрутизацию сообщений.

## Что изменится

1. `panel.go` — публичный необязательный `ContentHeightProvider`.
2. `scrollable.go` — общая effective-offset геометрия, консервативный fallback, локальный viewport из `View`/`ResizeMsg`/`WindowSizeMsg`, clipping семантических элементов.
3. `selectable.go`, `collapsible.go` — передача или вычисление intrinsic extent; вложенный `Scrollable` явно не обещает intrinsic height.
4. `scrollable_test.go` и новая узкая регрессия — провайдер, fallback, snapshot и границы.
5. `docs/api/{panel,scrollable}.md`, `docs/ru/api/{panel,scrollable}.md`, `code-specs/{panel,scrollable}.md` — контракт и фактический forwarding без других component docs.

## Детали реализации

1. `ContentHeightProvider.ContentHeight(width) (height, known)` остаётся optional; потребитель нормализует отрицательную известную высоту к нулю. Метод не рендерит содержимое.
2. Общая функция effective offset сначала использует известную explicit высоту; иначе пробует `View(width, saturatingAdd(offset, viewportHeight, 1))`. Только результат с меньшим количеством строк подтверждает конец. Если строк ровно столько, сколько запрошено, extent остаётся неизвестным, и запрошенный неотрицательный offset не ограничивается.
3. Не кэшировать fallback extent между вызовами: ширина и изменяемое содержимое могут поменяться без надёжного сигнала инвалидирования. Для fallback `View` использует строки того же probe, которым рассчитан effective offset.
4. `View`, `Elements` и `Update` используют одну effective-offset функцию. `View` и `Update` могут нормализовать `Offset`; `Elements` не меняет его. `Elements` запрашивает дочернее дерево до effective offset + viewport, клипает по viewport и переводит Y в локальные координаты.
5. `Update` сохраняет локальные размеры из `ResizeMsg`; `WindowSizeMsg` служит provisional-размером, пока не получен локальный resize. Scroll-сообщения продолжают передаваться дочерней панели.
6. `Selectable.ContentHeight` прозрачно проксирует известную высоту. `Collapsible` сообщает одну строку в свёрнутом состоянии; раскрытая высота известна только при известной высоте содержимого. `Scrollable` не представляет вложенный viewport как intrinsic высоту.
7. Все суммы высот насыщаются на `MaxInt`; известный max offset равен `max(0, contentHeight-viewportHeight)`.

## Критерии приёмки

- [x] Panel API совместим; провайдер optional, negative height нормализуется потребителем.
- [x] Явный extent ограничивает overscroll в рендеринге и семантических координатах.
- [x] Update → `/elements` до `View` и последующий `View` используют одинаковый effective offset.
- [x] Изменение высоты и ширины немедленно отражается без stale extent.
- [x] Natural-height fallback обнаруживает конец только по числу строк меньше probe height.
- [x] Panel, дополняющая View до запрошенной высоты, остаётся с unknown/unbounded offset.
- [x] Resize/update размеры корректно поступают Scrollable; события прокрутки продолжают forwarding.
- [x] Selectable и Collapsible сообщают extent согласно wrapper-семантике; nested Scrollable остаётся unknown.
- [x] Сложение около `MaxInt` не переполняет размер запроса и не даёт отрицательной высоты.
- [x] API и code-spec docs объясняют optional extent, неизвестный fallback и forwarding сообщений.
- [x] Пройдены Go formatting/vet/test/race и frozen pnpm/VitePress проверки.
