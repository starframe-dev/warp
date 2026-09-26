# Производительность и lifecycle ресурсов Warp

## Контекст

Warp сейчас собирает semantic snapshot независимо от использования HTTP-инспектора, может обходить дерево повторно за один Bubble Tea цикл, а некоторые удаления оставляют pointers за длиной слайса. У resource-owning панелей нет hook завершения владения; HTTP shutdown не ограничен временем. Текущие hot paths не имеют воспроизводимой benchmark-базы.

## Цель

Снизить измеренные лишние обходы/аллокации без изменения визуального поведения, освободить ресурсы и ссылки при окончательном удалении панелей, ограничить HTTP lifecycle и добавить ручной benchmark suite с `ReportAllocs`.

## Что изменится

1. `warp.go`, `element.go`, HTTP tests — demand-gated inspector snapshots, дедупликация snapshot в Update/View цикле, неизменяемый snapshot с кэшированным JSON, bounded shutdown и server timeouts.
2. `panel.go`, новый `ownership.go`, `warp.go`, `tab.go`, `tabgroup.go`, layout tree helpers — необязательный `Unmounter`, локальные ownership domains, cleanup постоянных удалений, очистка хвоста pointer slices и stale layout refs.
3. `tab.go`, `tabgroup.go`, `layout.go`, `render.go`, `focus.go`, `split.go`, `selectable.go`, `input.go`, `scrollable.go` — сначала benchmark baseline; изменения только при доказанной экономии или устранении O(offset) cliff.
4. `performance_test.go`, `resource_lifecycle_test.go` и узкие существующие тесты — benchmarks, cleanup/ownership regressions, snapshot/server race tests.
5. API docs и code specs для Panel/Warp/Tab/TabGroup — новые lifecycle/ownership contracts; без redesign рендерера и unrelated APIs.

## Baseline

Краткий baseline `go test -run='^$' -bench='^BenchmarkScrollableOffset$' -benchmem -benchtime=50ms -count=1` на Apple M1 Pro показал линейную стоимость legacy пути:

| Сценарий | 10 000 строк | offset у конца (`19 990`) |
|---|---:|---:|
| `Scrollable.View` | 121 795 ns/op, 205 856 B/op | 248 091 ns/op, 410 656 B/op |
| `Scrollable.Elements` | 538 620 ns/op, 2 100 585 B/op | 1 149 364 ns/op, 4 164 978 B/op |

Числа служат направляющим локальным baseline; сравнение выполнять на одной машине сериями benchmark.

### Результаты реализации

Показатели ниже — направляющее сравнение одного запуска на Apple M1 Pro, а не статистический отчёт `benchstat`:

| Сценарий | До | После |
|---|---:|---:|
| Scrollable.View, offset 19 990 (legacy → viewport) | 395 494 ns/op, 410 662 B/op | 815 ns/op, 1 264 B/op |
| Scrollable.Elements, offset 19 990 (legacy → viewport) | 972 820 ns/op, 4 164 983 B/op | 1 330 ns/op, 5 728 B/op |
| Mouse hit-test | 9 300 ns/op, 18 736 B/op, 221 allocs/op | 3 652 ns/op, 7 408 B/op, 104 allocs/op |
| Flex drag через `handleMouse` | 6 499 ns/op, 16 648 B/op, 103 allocs/op | 4 437 ns/op, 10 256 B/op, 73 allocs/op |
| Inspector Update без HTTP | 2 056 ns/op, 13 680 B/op, 2 allocs/op | 20 ns/op, 0 B/op, 0 allocs/op |
| `/elements` JSON | 35 957 ns/op, 10 511 B/op, 12 allocs/op | 2 974 ns/op, 10 450 B/op, 10 allocs/op |
| Border collection | 1 292 ns/op, 4 016 B/op, 15 allocs/op | 602 ns/op, 1 392 B/op, 4 allocs/op |

`Selectable.View` повторно разбивает строки только один раз: один запуск дал 3 008 → 2 624 B/op и 28 → 27 allocs/op; время 14 756 → 15 517 ns/op шумное и не считается доказанным ускорением.

## Детали реализации

1. Сначала добавить и запустить воспроизводимые benchmarks для render/layout/drag, inspector/HTTP JSON, broadcasts, Scrollable offsets, Input и overlays. Зафиксировать команды и baseline локально; generated output не сохранять.
2. Inspector snapshot строится только при активном HTTP inspector. `Update` и следующий `View` не собирают один и тот же snapshot дважды; публикация проверяет revision и demand. Immutable snapshot cache сериализует JSON один раз на snapshot.
3. `CloseHTTP` отключает demand и очищает публикацию, не удерживает Warp mutex во время ожидания, ограничивает graceful shutdown 5 секундами и принудительно закрывает сервер после timeout. HTTP server получает консервативные header/idle/write timeouts.
4. Добавить optional `Unmounter { Unmount() }`. Один top-level `Warp` образует один lifecycle domain всего своего дерева; standalone `TabGroup` или `Tab` образует отдельный domain, пока не вложен в другой. Переключение, collapse и потеря focus не вызывают hook.
5. При постоянном удалении сначала собрать candidate панели удаляемого subtree, затем изменить дерево, затем обойти оставшееся дерево всего domain и сравнить конкретные identities. Только candidates, которых больше нет в domain, дедуплицировать и вызвать `Unmount()` один раз. Root replacement, `Tab.SetRootPanel`, Flex replacement, закрытие tab/float используют этот общий путь. Reparenting в том же domain и оставшиеся references не прерывают lifecycle. Global registry/refcount не создавать.
6. Identity pointer-backed Panel — dynamic type + pointer address. Не использовать `reflect.DeepEqual`; одинаковые по содержимому, но разные pointer instances должны оставаться разными. Для non-pointer non-comparable `Unmounter`-значений Go не предоставляет надёжной instance identity; в документации определить pointer-backed `Unmounter` как поддерживаемый resource-owning pattern, а сравнимые value types сравнивать штатным Go equality.
7. Изменения slices обнуляют удалённый tail. Layout replacement сбрасывает ссылки на устаревшие split/flex drag/border state. Mouse hit-testing повторно использует layout текущего события. Рекурсивные collectors переписываются на accumulator, остальные аллокационные правки — только при benchmark-подтверждении. `sync.Pool` не добавлять.
8. Baseline подтвердил O(offset) и для Scrollable.View, и для Scrollable.Elements. Добавить optional `ViewportRenderer { ViewAt(width, height, offset int) string }` и `ViewportElementProvider { ElementsAt(width, height, offset int) []Element }`. Scrollable вызывает их только при известном intrinsic extent; `ViewAt` возвращает строки только запрошенного viewport, `ElementsAt` — видимые элементы с bounds в координатах исходного content. Остальные панели сохраняют прежний fallback. Renderer/canvas rewrite вне scope.
9. Benchmarks используют `b.ReportAllocs()`; recommended comparison: `go test -run='^$' -bench=. -benchmem -count=10 ./...` + `benchstat`. Benchmark execution не добавлять в push CI.

## Критерии приёмки

- [x] Inspector-disabled Warp не вызывает ElementProvider и не строит snapshot.
- [x] Один Update/View цикл не выполняет одинаковый semantic traversal дважды.
- [x] JSON snapshot сериализуется один раз на immutable revision; concurrent HTTP читатели race-safe.
- [x] Удалённые Tab/Float/root/subtree не оставляют pointer references за `len` backing slice.
- [x] Один top-level Warp сканируется как единый domain; standalone TabGroup/Tab образуют отдельные domains. Между независимыми domains одна instance не разделяется lifecycle-системой, без global registry.
- [x] Все восемь пользовательских regressions проходят: повторная ссылка внутри Tab/TabGroup; последний reference; закрытие Tab при ссылке в другом Tab; последнее удаление в TabGroup; root+float; reparent без промежуточного Unmount; закрытие domain с dedup; документированный unsupported shared instance между независимыми Warp.
- [x] Candidate panels собираются до detach; проверка оставшихся references и вызов `Unmount` происходят только после логического удаления. Identity — pointer instance, не `reflect.DeepEqual`.
- [x] `Unmounter` вызывается ровно один раз при окончательной потере последней ссылки в domain; switch/collapse/focus вызывают ноль раз.
- [x] HTTP shutdown bounded, принудительно закрывает зависшие соединения по timeout и не ждёт под Warp mutex.
- [x] HTTP server настроен с безопасными, но не чрезмерно короткими timeouts.
- [x] Hit-test reuse и accumulator collectors сохраняют результаты; hot-path аллокационные изменения измерены.
- [x] `Selectable.View` не разбивает те же строки повторно; Input/overlays изменяются только при доказанной пользе.
- [x] Benchmarks покрывают все категории handoff, печатают ns/op, B/op, allocs/op и запускаются вручную.
- [x] Scrollable large-offset baseline измерен; при известном extent viewport interfaces обрабатывают только нужные строки/elements, legacy Panel сохраняет прежний fallback.
- [x] Ресурсные, GC/backing-slice, concurrent inspector и server lifecycle tests стабильны; без goroutine-count sleeps.
- [x] Пройдены `gofmt -w .`, `go vet ./...`, `go test ./...`, `go test -race ./...` и docs frozen install/build.
