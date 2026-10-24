# Warp — аудит корректности, раунд 3

## Контекст
Baseline: `122c5eb071b78aaebfa490d1c7c3282288cd5832`. Проверялись legacy `WindowSizeMsg` для вложенного standalone `Tab` и lifecycle Warp, размещённого в дереве через повторные `AsPanel()`.

## Результаты воспроизведения до production changes
- **P1 подтверждён.** Regressions показали лишний `ResizeMsg{80,24}` после локального resize для прямого `TabGroup→Tab`, `Scrollable→Tab`, `Collapsible→Tab`, float и split/flex внутри nested Tab. Для collapsible правильный локальный размер был `80×22`, для float после viewport clamp — `78×21`; затем global WindowSize перезаписывал его на `80×24`.
- **P2 подтверждён.** После удаления первого из двух placements baseline сбрасывал `inner.ownership` в nil и устанавливал `ownsDomainRoot=true`, пока второй placement оставался достижим. `inner.Close()` размонтировал ресурс преждевременно и ломал оставшийся adapter. В nested-Warp сценарии повторился тот же сброс ownership и преждевременный Unmount.

## Исправления
1. `tab.go`: `appendLegacyWindowSizePanel` обходит `*Tab` через существующий `appendLegacyWindowSizeNode`, а затем его float panels тем же traversal. Вложенный Tab не получает global `WindowSizeMsg`; ordinary leaf-панели получают legacy message, их `tea.Cmd` сохраняются.
2. `warp.go`: `AsPanel()` сохраняет один адаптер для данного Warp под существующим mutex. Повторные вызовы сохраняют сигнатуру `AsPanel() Panel`, но возвращают одну adapter identity; несколько placements становятся ссылками на тот же Panel, которые existing ownership reachability уже дедуплицирует. Новых registry/history или ownership rewrite нет.
3. EN/RU API docs и `code-specs/warp.md` описывают стабильную identity адаптера. `ownership.go`, VERSION и зависимости не менялись.

## Regression tests
- `TestNestedTabWindowSizePreservesLocalGeometry`: direct Tab, Scrollable, Collapsible и float. Сверяет локальные ResizeMsg, legacy WindowSize delivery, возвращённые команды, rendered viewport, semantic Elements и mouse hit-testing.
- `TestNestedTabSplitAndFlexPreserveLocalWindowGeometry`: local dimensions обеих leaf-панелей, команды, View, Elements и mouse routing.
- `TestAsPanelReturnsStableAdapterIdentity`: повторные вызовы сохраняют adapter identity.
- `TestMultipleAsPanelAdaptersRetainSharedOwnershipUntilLastRemoval`: два placements в одном TabGroup, перемещение внутри ownership domain, Close при оставшемся placement, удаление последнего и повторные Close.
- `TestNestedWarpAdaptersRetainOwnershipAtEachLevel`: nested Warp и ровно-однократный Unmount.
- `BenchmarkOwnershipDetachWithRetainedAsPanel`: измеряет detach при остающемся placement.

## Локальная верификация
- `gofmt -l .`, `git diff --check`, `go vet ./...`, `go test ./...`, `go test -race ./...`, `go test -shuffle=on -count=2 ./...` — прошли; integration/E2E включены в полный suite.
- Coverage 91.7% (Round 2 baseline 91.6%).
- `golangci-lint run ./...` — 0 issues.
- `govulncheck ./...` — 0 достижимых уязвимостей; scan снова сообщил unreachable GO-2026-6809 (stdlib Go 1.27.2) и GO-2026-5024 (`x/sys` v0.24 Windows). Зависимости не менялись.
- Go 1.27.2/macOS arm64 suite и `CGO_ENABLED=0 GOTOOLCHAIN=go1.22.12 go test ./...` — прошли. Windows test packages cross-compiled.
- GitHub Actions Go CI run [40115435114](https://github.com/starframe-dev/warp/actions/runs/40115435114) завершился `success`: Linux Go 1.22.x/1.26.x/1.27.x, macOS Go 1.27, Windows Go 1.27, golangci-lint и govulncheck — все jobs успешны; 5 fuzz smoke targets прошли. Release job был `skipped` (workflow dispatch не на main push).
- VitePress build EN/RU docs прошёл.
- Ownership benchmark, Apple M4, 3 повтора: Round 3 `2.36–2.39 µs/op`, `5528 B/op`, `79 allocs/op`; baseline `2.38–2.41 µs/op`, `5616 B/op`, `82 allocs/op`. Существенной регрессии нет.

## CI matrix и приёмка
- [x] Оба дефекта воспроизведены failing regressions до production changes.
- [x] Подтверждённые дефекты исправлены минимально; regressions раундов 1–2 проходят.
- [x] Локальная геометрия, legacy leaf messages/commands и ровно-однократный Unmount сохранены.
- [x] Доступные локальные проверки прошли, benchmark не показывает существенной регрессии.
- [x] GitHub Go CI matrix прошёл на Linux/macOS/Windows; commit SHA будет указан в итоговом отчёте.

VERSION и dependencies не менялись; release job не запускался и не создавал релиз.
