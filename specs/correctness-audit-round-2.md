# Warp — аудит корректности, раунд 2

## Контекст
Проверка выполнена на baseline `abac1a4d8b06f8d497ec99155f513a705d52bef8`. Все пять заявленных дефектов воспроизведены до исправления и получили регрессионные тесты.

## Исправления
1. **Вертикальная геометрия TabGroup.** Бар пересчитывает свою ширину для актуального viewport и подписей вкладок. Resize, offsets, View, Elements и mouse hit-testing используют согласованную геометрию.
2. **Глобальный WindowSizeMsg.** Legacy сообщение проходит через встроенные wrappers, nested TabGroup/Warp и float-панели до обычных пользовательских листьев. Вложенные layout-контейнеры не получают полный размер вместо локального `ResizeMsg`; legacy доставка leaf-панелям сохранена.
3. **Начальная команда Float.** Добавлен совместимый `Tab.FloatWithCmd(...) tea.Cmd`; `Tab.Float(...)` сохраняет прежнюю void-сигнатуру. Вызывающая Bubble Tea модель может вернуть команду из `Update`.
4. **Lifecycle вложенного Warp.** При detach ownership и ledger уже размонтированных `Unmounter`-ресурсов переходят в локальный ownership domain вложенного Warp. Повторная замена root не вызывает второй `Unmount`.
5. **Keyboard routing.** Standalone `Tab` без фокуса больше не рассылает `tea.KeyMsg` всем leaf-панелям. Не-клавиатурные пользовательские сообщения продолжают broadcast.

## Регрессионные тесты
`round2_regression_test.go` проверяет resize 80→10→80 для TabLeft/TabRight (ResizeMsg, View, Elements и hit-testing), wrapper/split/flex/Warp/float resize routing, legacy WindowSize delivery обычным float-leaf, `FloatWithCmd`, nested Warp detach, а также unfocused keyboard routing. Каждый из пяти исходных failing regression test был добавлен до production fix.

## Проверки
- `gofmt -l .`, `go vet ./...`, `go test ./...`, `go test -race ./...`, `go test -shuffle=on -count=2 ./...` — прошли.
- Coverage: 91.6% на baseline `91.5%` (baseline точное значение 91.5086%, текущий результат округляется до 91.6%).
- `golangci-lint run ./...` — 0 issues.
- `govulncheck ./...` — 0 достижимых уязвимостей. Отдельно обнаружены advisory вне reachable call graph: GO-2026-6809 для stdlib `crypto/x509` в Go 1.27.2 (исправлено в Go 1.27.3) и GO-2026-5024 для `golang.org/x/sys/windows` v0.24.0 (исправлено в v0.44.0); зависимости и версии намеренно не менялись.
- Все пять CI fuzz smoke targets прошли.
- Go 1.22.12: `CGO_ENABLED=0 go test ./...` прошёл. На локальной macOS попытка с cgo завершилась проблемой host linker (`dyld: missing LC_UUID`), не тестовым падением.
- Go 1.27.2/macOS arm64: полный test suite прошёл. Windows test packages cross-compiled; тестовый код на Windows здесь не исполнялся. Загрузка toolchain Go 1.26.2 зависла и была остановлена; Linux и полноценный Windows CI этим локальным запуском не заменены.
- VitePress docs build прошёл.
- Isolated `BenchmarkTabGroup/broadcast-resize-8-tabs`: текущая версия ~1.70–1.72 µs/op, 5392 B/op, 125 allocs/op; baseline ~1.70–1.73 µs/op, те же 5392 B/op и 125 allocs/op.

## Приёмка
- [x] Все пять гипотез проверены; подтверждённые дефекты воспроизведены до изменения production-кода.
- [x] Совместимость публичного API и lifecycle сохранена; новый Float API аддитивный.
- [x] Доступные обязательные проверки и ограничения платформ отражены выше.
- [x] Результат финального прогона зафиксирован; локальный commit SHA будет включён в итоговый отчёт.

## Ограничения
GitHub Actions не запускался из этой сессии; фактическое исполнение на Windows/Linux требует CI. Новый release/tag не создавался, `VERSION` и зависимости не менялись; изменения не публиковались.
