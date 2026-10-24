# Warp v0.1.2: корректность и стабильность

## Контекст
Статический аудит выделил возможные ошибки в геометрии и маршрутизации сообщений, lifecycle панелей и компонентах. Исправлять нужно только подтверждённые regression-тестом дефекты.

## Цель
Проверить кандидаты P1/P2 на актуальном `main`, воспроизвести подтверждённые ошибки до исправления и устранить их без нарушения публичной совместимости и композиционной архитектуры.

## Что изменено
- `tab.go`, `tabgroup.go`, `layout.go`/`render.go`, `float.go`, `warp.go`, `ownership.go`, `input.go`, `popover.go`, `selectable.go` — подтверждённые resize, drag commands, float notifications, lifecycle, Flex, terminal safety, geometry и routing defects.
- Regression tests в `correctness_regression_test.go`, `ownership_test.go`, `warp_test.go`.
- Документация API не менялась; данная спека — запись проверки.

## Критерии приёмки и результаты
- [x] Подтверждённые дефекты воспроизводятся новыми тестами до реализации: TabGroup content resize, drag command loss, float resize notifications, duplicate nested Unmount, unsafe Input.Value, Flex grow remainder, unsupported Shift keys и standalone Tab keyboard routing.
- [x] Исправлены подтверждённые дефекты, сохранён legacy `WindowSizeMsg` полного окна для leaf-панелей.
- [x] Существующий `go test ./...`, включая integration/E2E и HTTP inspector, проходит.
- [x] `go vet ./...`, `go test -race ./...`, `go test -shuffle=on -count=2 ./...`, `gofmt -l .`, `git diff --check` проходят.
- [x] Benchmark запускается; внешние lint/vulnerability tools указаны как недоступные.

## Неподтверждённые/ограниченные пункты
- P2 Popover переполнение было непосредственно видно из вычисления ширины: `contentW` допускал всю ширину viewport, затем рамка добавляла ещё две ячейки. Исправлено геометрией и проверено существующим и новым тестами.
- `Tab.Float` сохраняет void-сигнатуру; команда из синхронного resize при создании не может быть возвращена вызывающей модели без изменения API. Команды с последующих resize сохраняются. API расширение не включалось в этот совместимый bugfix.
- P3 API/theme/branch protection/Bubble Tea v2 не менялись.
