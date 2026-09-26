# Warp — Handoff

[2026-09-26] Проблема: общий `lastBorders = collectLayoutBorders(layout)` повторялся в render и mouse paths, из-за чего точечная замена была неоднозначной → Решение: привязывать замену к уникальному контексту функции и сначала перечитывать оба места.
[2026-09-26] Проблема: lifecycle-патч TabGroup API docs не совпал с предположенным текстом абзаца → Решение: искать и перечитывать точный фрагмент через grep/read перед правкой локализованного документа.
[2026-09-25] Проблема: legacy regression жёстко ожидала запрос `Offset + viewport`, несовместимый с новым безопасным probe → Решение: обновить ожидаемую высоту до `Offset + viewport + 1` и явно назвать probe-строку в assertion.
[2026-09-25] Проблема: новая View regression посчитала отступ для строки длиной 4 при ширине 8 как 3 ячейки → Решение: в ожиданиях рассчитывать padding как `width - ansi.StringWidth(line)`.
[2026-09-25] Проблема: при поиске Node-кода обращение к `node.go` завершилось ENOENT, поскольку тип определён в `split.go` → Решение: находить фактический файл через `grep`/`find`, прежде чем читать ожидаемый путь.
[2026-09-25] Проблема: корневой `package.json` отсутствует, поэтому его чтение завершилось ENOENT → Решение: проверять фактические каталоги; команды документации объявлены в `docs/package.json`.
[2026-09-25] Проблема: ожидался отдельный `http_test.go`, но HTTP lifecycle-тесты находятся в `http_concurrency_test.go` → Решение: искать фактическое объявление теста через `find`/`grep` перед чтением.
[2026-09-25] Проблема: массовая точечная правка `tab.go` не прошла, потому что общий блок `if len(items) == 0` встречается в FlexRow и FlexColumn → Решение: включать имя функции/соседний контекст в точечную замену и повторять правку только после перечитывания файла.
[2026-09-25] Проблема: пакетная правка `warp.go` не совпала с полями после gofmt и текущего diff → Решение: перечитывать актуальные блоки непосредственно перед редактированием и не полагаться на прежнее форматирование.
[2026-09-25] Проблема: создание JSON cache записало некорректный rune literal `\\n`, и gofmt остановился с `illegal rune literal` → Решение: перечитывать экранированные литералы в целевом Go-файле и проверять их через gofmt сразу после добавления.
[2026-09-25] Проблема: `/elements` regression тесты ожидали снимок, созданный до старта demand-gated инспектора → Решение: после `ServeHTTP` выполнить UI `View`/`Update` перед проверкой semantic snapshot.
[2026-09-25] Проблема: общий блок запуска HTTP сервера встретился в нескольких element tests и массовая правка остановилась → Решение: ограничивать замену именем конкретного теста и редактировать повторяющиеся сценарии отдельно.
[2026-09-25] Проблема: RU Scrollable Markdown patch не совпал из-за лишнего escape-символа в буквальной строке → Решение: перечитать точный фрагмент документа, включая обратные слеши, до пакетной замены.
[2026-09-25] Проблема: точечное добавление troubleshooting-правила в `CONTEXT.md` ссылалось на отсутствующий заголовок `## Recent lessons` → Решение: читать фактическое начало файла и вставлять правило после существующего заголовка.
[2026-09-25] Проблема: тестовые helper-блоки с одинаковым выражением не давали уникальную адресную замену → Решение: добавлять сигнатуру метода в контекст патча и раздельно менять одинаковые блоки.
[2026-09-25] Проблема: поиск HTTP-инспекторных тестов обращался к отсутствующему `http_test.go`, а замена одинаковых тестовых таблиц не была уникальной → Решение: сначала находить фактические тестовые файлы через `grep`/`find`, а точечные замены повторяющихся блоков связывать с именем тестовой функции.
[2026-09-25] Проблема: точечная замена русского Selectable API-документа не совпала из-за неточно восстановленного текста → Решение: перечитать точный абзац файла, включая переносы строк и написание `escape`, перед применением замены.
[2026-09-25] Проблема: прежний mouse-selection regression ожидал исключение конечной ячейки после перехода на end-exclusive границы → Решение: переводить cell-coordinate мыши `x` в границу `x+1` и обновлять ожидание до точного включения последней ячейки.
[2026-09-25] Проблема: точечная правка `SetSplitCollapse` не совпала с фактическим уровнем табуляции метода → Решение: перечитать точный блок и строить `oldText` с буквальным отступом из файла.
[2026-09-25] Проблема: Markdown-патчи ссылались на отсутствующие/неидентичные фрагменты, а параллельные замены перекрывались; `.lore/weft/.gitignore` отсутствует → Решение: перечитывать точные блоки каждого файла, объединять пересекающиеся edits и изменять существующий `.lore/.gitignore`.
[2026-09-25] Проблема: `functions.job` safety блокировал длинную команду с `&&` и subshell, не запуская проверки → Решение: выполнять проверки упорядоченными отдельными job или последовательными строками с `set -e` и прямым `cd`.
[2026-09-24] Проблема: цикл zsh использовал зарезервированный массив `path` и перезаписал `PATH`, из-за чего команды в цикле не находились → Решение: не использовать `path` как имя переменной; применять `file_path` и проверять shell-лог перед чтением результата.
[2026-09-24] Проблема: поиск корневого `package.json` завершился ENOENT, хотя VitePress manifest находится в `docs/package.json` → Решение: сначала находить build manifest через `find`, затем читать найденный путь.
[2026-09-24] Проблема: на macOS вызвана Windows-only команда PowerShell для Git-status → Решение: сразу выбирать shell job/`functions.job` для команд Git на этой машине.
[2026-09-24] Проблема: существующий `TestFloatOffScreen` ожидал float за правой границей viewport после рендера → Решение: обновить ожидание на новую нормализацию float в доступные bounds.
[2026-09-24] Проблема: тест shutdown собрал URL из `HTTPAddr()` уже после очистки адреса → Решение: завершать проверочный запрос до закрытия сервера; конкурентный HTTP-vs-close lifecycle проверять отдельным тестом с зафиксированным адресом.
[2026-09-24] Проблема: первая целевая сборка выявила неиспользуемый `strconv` в HTTP concurrency regression → Решение: удалить неиспользуемый импорт и повторно проверить целевые тесты.
[2026-09-24] Проблема: zsh не разбил переносимый через newline список Go-файлов в переменную аргументов `gofmt`, передав один слишком длинный путь → Решение: строить проверку форматирования через `find -print0 | xargs -0 gofmt -l`.
[2026-09-24] Проблема: PowerShell-команда недоступна на macOS, чтение каталога как файла и неоднозначные md-замены завершились ошибками инструмента → Решение: использовать shell job для Git, `ls` для каталогов, перечитывать Markdown и применять уникальные точечные фрагменты.
[2026-09-24] Проблема: HTTP-инспектор по умолчанию слушал все интерфейсы, а `TabGroup.Elements` менял состояние рендеринга → Решение: использовать loopback с портом `WARP_HTTP_PORT`/0, синхронизировать Warp state и вычислять `Elements` без рендеринговых побочных эффектов.
[2026-09-24] Проблема: shutdown regression test разрешал обработчику читать HTTPAddr до того, как goroutine CloseHTTP очистит поле, и ошибочно требовал пустой адрес → Решение: синхронизировать тест по запуску закрытия и проверять отсутствие дедлока, а очистку адреса проверять после CloseHTTP.
[2026-09-24] Проблема: zero-viewport Selectable test использовал статический Panel, который игнорирует размеры View, и ошибочно ожидал пустой ответ → Решение: использовать размеро-зависимую панель для проверки контракта viewport, не навязывая обёртке клиппинг произвольного Content.
[2026-09-24] Проблема: Unicode tests суммировали ширины нескольких строк как одну и `Hardwrap` ломал слова вопреки публичному поведению → Решение: проверять строки отдельно и использовать ANSI-aware `ansi.Wrap` с теми же word-break правилами.
[2026-09-24] Проблема: Input добавлял placeholder-ячейку курсора даже когда CJK/emoji уже занимали всю ширину → Решение: добавлять placeholder только при наличии свободной terminal cell.
[2026-09-24] Проблема: modal resize regression ожидал ширину 24, забыв минимальную ширину авто-модалки 30 → Решение: сверять тестовые ожидания с действующим clamp 30–50 до ограничения шириной viewport.
[2026-09-24] Проблема: существующий modal test закреплял устаревший кэш размеров и провалился при реальном resize → Решение: заменить ожидание «не пересчитывать» на регрессию нового viewport и проверять координаты после повторного `EnsureDimensions`.
[2026-09-24] Проблема: старый west-resize тест требовал растянуть панель за фиксированный правый край, а zero-size тест принимал невалидный float → Решение: сохранять противоположный край (ширина 30 при X=0) и явно проверять отказ нулевых размеров.
[2026-09-24] Проблема: после замены FloatPane на ANSI-ячейковую реализацию остался неиспользуемый импорт `lipgloss`, из-за чего сборка остановилась → Решение: удалять импорты старой реализации вместе с её кодом и запускать тесты сразу после компиляции.
[2026-09-24] Проблема: правка спецификации `TabGroup` ссылалась на список правил как на нумерованный список рендеринга, хотя он лежит в разделе «Ключевые правила» → Решение: перечитать заголовки/точные блоки и обновлять каждый раздел отдельно.
[2026-09-24] Проблема: новый Node regression test запросил четыре результата у helper, возвращающего три панели, и не компилировался → Решение: создавать четвёртую тестовую панель явно и проверять число возвращаемых значений helper.
[2026-09-24] Проблема: повторно не были различены одинаковые блоки `FlexRow`/`FlexColumn`, поэтому второй пакетный патч тоже не применился → Решение: привязывать замену к уникальной окружающей функции и не включать no-op правки.
[2026-09-24] Проблема: в `tab.go` повторяющиеся nil-normalization блоки помешали пакетной точечной правке → Решение: включать имя функции в контекст каждого изменения и повторно читать участок после неоднозначности.
[2026-09-24] Проблема: точечная правка tab bar была неоднозначной из-за двух одинаковых `lipgloss.Width(label)` блоков → Решение: менять горизонтальный и вертикальный рендерер отдельными правками с уникальным контекстом.
[2026-09-24] Проблема: geometry regression tests измеряли X через UTF-8 byte index, поэтому многобайтовая граница сдвигала ожидаемую координату → Решение: измерять ширину префикса через `ansi.StringWidth`.
[2026-09-24] Проблема: пакетная правка `focus.go` не совпала с фактическим блоком `applyFocus` и не применилась → Решение: перечитать точный участок и переписать небольшой файл целиком, проверив все методы.
[2026-09-24] Проблема: точечная правка `tab.go` не совпала с исходным форматированием и не применилась → Решение: сначала привести файл к `gofmt`, затем повторить адресные изменения.
[2026-09-24] Проблема: точечная правка `tab.go` не совпала с исходным форматированием и не применилась → Решение: сначала привести файл к `gofmt`, затем повторить адресные изменения.
[2026-09-24] Проблема: существующие ожидания закрепляли переполнение flex-размеров, короткий вывод свёрнутой колонки и смещённую границу → Решение: проверять размеры flex как `[2, 3]` в доступных пяти ячейках, сохранять точную высоту контейнера (7 строк) и ожидать границу на Y=2.
[2026-09-24] Проблема: тестовая функция `blankLines` конфликтовала с новым helper рендера → Решение: переименовать внутренний helper раскладки, повторно запустить тесты.
[2026-09-24] Проблема: общий layout-слой был создан до расширения `BorderHit`, поэтому сборка не находила поля `Bounds` и `FlexIndex` → Решение: добавить в `BorderHit` прямоугольник раскладки и локальный индекс flex-границы, затем повторить проверки.
[2026-09-24] Проблема: существующие ожидания закрепляли переполнение flex-размеров, короткий вывод свёрнутой колонки и смещённую границу → Решение: проверять размеры flex как `[2, 3]` в доступных пяти ячейках, сохранять точную высоту контейнера (7 строк) и ожидать границу на Y=2.
[2026-09-24] Проблема: тестовая функция `blankLines` конфликтовала с новым helper рендера → Решение: переименовать внутренний helper раскладки, повторно запустить тесты.
[2026-09-24] Проблема: общий layout-слой был создан до расширения `BorderHit`, поэтому сборка не находила поля `Bounds` и `FlexIndex` → Решение: добавить в `BorderHit` прямоугольник раскладки и локальный индекс flex-границы, затем повторить проверки.
[2026-09-24] Проблема: тестовая функция `blankLines` конфликтовала с новым helper рендера → Решение: переименовать внутренний helper раскладки, повторно запустить тесты.
[2026-09-24] Проблема: общий layout-слой был создан до расширения `BorderHit`, поэтому сборка не находила поля `Bounds` и `FlexIndex` → Решение: добавить в `BorderHit` прямоугольник раскладки и локальный индекс flex-границы, затем повторить проверки.
[2026-09-17] Problem: code-check generated TabGroup tests assumed a tab-bar width for every position and used row zero for a vertical active tab → Solution: align the tests with the current vertical-only width calculation, active-tab row, and visual padding behavior.
[2026-09-17] Problem: generated code-check pages were exposed through a separate reference page and led to 404s → Solution: remove that page and place the unmodified EN/RU output in the API section; retain docs/en and docs/ru as code-check source paths.

## Что это

Warp — Go-библиотека (Bubbletea layout engine) для создания TUI с гибким управлением
пространством: вкладки, сплиты, плавающие панели, flexbox, модальные окна, popover.
Пользователь реализует интерфейс `Panel`, а warp управляет их расположением.

## Состояние проекта

**Текущее состояние** — Input, explicit focus API, Modal, Popover, Element tree и runtime-тема через `SetTheme`. Тестовый набор расширен code-check.

## Рабочая версия: lifecycle и performance (2026-09-26)

- Реализованы локальные lifecycle-домены: корневой Warp владеет полной вложенной иерархией; standalone TabGroup/Tab — отдельными доменами. `Unmounter` выполняется после detach и только если экземпляр больше не достижим в домене; ссылки дедуплицируются по identity. Использование одного экземпляра между независимыми корнями остаётся caller-managed и документировано как неподдерживаемое.
- Scrollable получил необязательные `ViewportRenderer`/`ViewportElementProvider`, используемые только при известной intrinsic-высоте; legacy Panel путь сохранён. На Apple M1 Pro offset 19 990: View 395 µs / 410 KB против 0.82 µs / 1.3 KB, Elements 973 µs / 4.16 MB против 1.33 µs / 5.7 KB в текущем бенчмарке.
- Добавлены demand-gated inspector snapshots, JSON cache, bounded HTTP shutdown, accumulator collectors, reusable layout/border state, flex drag layout reuse и cache grapheme layout для Input.
- Добавлены `performance_test.go`, lifecycle/viewport regressions и EN/RU API docs/code specs. Полный ручной benchmark: `go test -run='^$' -bench='.' -benchmem -benchtime=50ms -count=1`.
- Проверки 2026-09-26: `gofmt -w .`, `go vet ./...`, `go test ./...`, `go test -race ./...`, `pnpm install --frozen-lockfile` в `docs/`, `pnpm exec vitepress build .` — пройдены.
- Не коммитить; пользовательская Scrollable extent работа сохранена в том же worktree.

## Новое в v0.7

- **Input component** — `NewInput(prompt)` с курсором, backspace, delete, стрелками, home/end
- **Focus API** — `Focusable` interface, `FocusNext()`, `FocusPrev()`, `FocusFirst()`, `FocusPanel()`
  Warp **не биндит** Tab/Shift+Tab автоматически — разработчик сам решает
- **RawKeyReceiver** — интерфейс для PTY/терминалов, которым нужны все клавиши без перехвата
- **Modal** — `ShowModalMsg`/`CloseModalMsg`, overlay поверх lines, drag, close, buttons
- **Popover** — контекстное меню с Overlay, HandleMouse, HandleKey
- **Element tree** — `ElementProvider` интерфейс для семантического UI-дерева (HTTP endpoint)
- **ContextMenu удалён** — заменён на Popover
- **Тесты** расширены проверками edge cases и Theme API
- **Theme API** — `ThemeColors` и `SetTheme` для runtime-переопределения палитры и стилей

## Демо (`cmd/demo/main.go`)

```bash
go run ./cmd/demo/
```

**Tab 1 — «main»**: FlexRow с 3 input-полями (Name, Email, Search) + Preview panel.
Tab/Shift+Tab переключает фокус между input'ами (через custom appRoot).

**Tab 2 — «local-tabs»**: Local TabGroup(TabLeft) внутри FlexRow + Scrollable с Selectable.

**Tab 3 — «columns»**: FlexColumn с Collapsible + float.

**Tab 4 — «splits»**: Split-панели + float.

**Горячие клавиши** (определяет разработчик, не warp):
- `Tab` / `Shift+Tab` — переключение фокуса (в demo)
- `Ctrl+T` — новый таб
- `Ctrl+W` — закрыть таб
- `Ctrl+Tab` / `Ctrl+Shift+Tab` — следующий/предыдущий таб
- `q` / `Ctrl+C` — выход

## Модуль и зависимости

- **Модуль:** `github.com/starframe-dev/warp`
- **Директория:** `/Users/a/Space/Projects/Starframe/warp`
- **Go:** 1.22+
- **Зависимости:** `bubbletea v1.1.0`, `lipgloss v0.13.0`, `charmbracelet/x/ansi`, `rivo/uniseg`
- **Проверки:** точное состояние тестов и `go vet` фиксировать после запуска обязательной финальной верификации.
- **Git:** `https://github.com/starframe-dev/warp.git`, ветка `main`

## Архитектура

```
warp.go         — tea.Model, тонкая обёртка вокруг root Panel
tabgroup.go     — TabGroup: Panel с таб-баром, переключением табов, keyboard/mouse
tab.go          — Tab: дерево splits/flex, float-панели, фокус, mouse handling, рендеринг
panel.go        — интерфейс Panel{View(w,h) string; Update(Msg) Cmd}
split.go        — Node, SplitConfig, FlexConfig, Direction, MinPanelSize=3
layout.go       — общая геометрия для render, Elements, hit-testing, resize и drag
render.go       — renderNode (рекурсивный), findBorders, padContent, computeFlexSizes
float.go        — FloatPane: рамка, drag, resize, overlayFloat, StripANSI, CloseOnOutsideClick
styles.go       — lipgloss-стили (Gruvbox Dark)
theme.go        — ThemeColors и SetTheme для runtime-переопределения темы
collapsible.go  — Collapsible Panel с заголовком и toggle
scrollable.go   — Scrollable Panel с viewport и mouse wheel
dropdown.go     — DropdownMenu Panel с кнопкой и раскрывающимся списком
selectable.go   — Selectable Panel с text selection (mouse drag, Shift+arrows, Ctrl+A, OSC 52)
wrap.go         — WordWrap, SpaceWrap утилиты для переноса текста
modal.go        — Modal dialog с overlay, drag, close, buttons
popover.go      — Popover контекстное меню с Overlay, HandleMouse, HandleKey
element.go      — Element, Bounds, ElementProvider для семантического UI-дерева
focus.go        — Focusable interface, collectFocusables, RawKeyReceiver
input.go        — Input Panel с курсором, backspace, delete, стрелками
drag.go         — заглушка (drag-логика в tab.go)
```

### Дерево панелей

```go
Node { Panel Panel | Split *SplitConfig | Flex *FlexConfig }
SplitConfig { Direction, Fraction, First *Node, Second *Node, Dragging bool }
FlexConfig  { Direction, Items []FlexItem }
```

### TabGroup как Panel

```go
// Табы как корень (как раньше)
w := warp.New()  // root = TabGroup с 1 табом

// Табы как компонент внутри flex
tg := warp.NewTabGroup(warp.TabLeft)
tg.NewTab("code")
tg.NewTab("debug")
tab.FlexRow(root, []warp.FlexItemSpec{
    {Panel: tg, Grow: 2},  // ← TabGroup внутри flex!
})

// Warp вообще без табов
w := warp.New()
w.SetRoot(myCustomPanel)
```

### Focus API

```go
// Разработчик сам решает, какие клавиши биндить:
tab.FocusNext()      // следующая focusable панель
tab.FocusPrev()      // предыдущая
tab.FocusFirst()     // первая
tab.FocusPanel(p)    // конкретная панель

// Focusable interface
type Focusable interface {
    Panel
    Focus()
    Blur()
    Focused() bool
}
```

### Modal / Popover

```go
// Modal — через сообщения
warp.ShowModalMsg{Title: "Confirm", Content: "Delete?", Buttons: [...]}

// Popover — контекстное меню
pop := &warp.Popover{
    Items:   []warp.PopoverItem{{Name: "Copy", Action: ...}},
    X:       x, Y: y,
    OnClose: func() { ... },
}
lines = pop.Overlay(lines, totalW, totalH)
```

### Element tree

```go
type ElementProvider interface {
    Elements(width, height int) []Element
}
// HTTP endpoint /elements возвращает JSON-дерево UI-элементов
```

## API

```go
w := warp.New()
tab := w.NewTab("name")
w.SetTabPosition(warp.TabBottom)

// Layouts
tab.SplitVertical(parent, 0.5, newPanel)
tab.SplitHorizontal(parent, 0.5, newPanel)
tab.FlexRow(parent, []warp.FlexItemSpec{{Panel: p1, Grow: 1}, ...})
tab.Float(panel, x, y, w, h)

// Collapsible
col := warp.NewCollapsible("Title", panel)
tab.ToggleCollapsible(col)

// Scrollable
scroll := warp.NewScrollable(panel)

// Dropdown
dd := warp.NewDropdownMenu("Menu", []warp.DropdownItem{...})
dd.OnSelect = func(idx int) { ... }

// Input
in := warp.NewInput("Name: ")
in.Focus()
in.SetValue("hello")

// Selectable
sel := warp.NewSelectable(panel)
sel.SelectedText()
sel.Copy()  // OSC 52 clipboard

// Focus
tab.FocusNext()
tab.FocusPrev()
tab.FocusFirst()
tab.FocusPanel(panel)

// Word wrap
lines := warp.WordWrap(text, 40)
lines := warp.SpaceWrap(text, 40)

// Nested warps
inner := warp.New()
tab.SplitVertical(tab.RootPanel(), 0.5, inner.AsPanel())

w.Run()
```

## Что не доделано / Ideas

- **Стилизация** — `SetTheme` меняет семантическую палитру; отдельного API для настройки каждого внутреннего style нет
- **Анимации** — нет (drag без анимации, переключение табов мгновенное)
- **Nested float** — float внутри float не поддерживается
- **List / Table** — нет компонентов для списков и таблиц
- **Textarea** — нет многострочного ввода
- **Subscriptions** — нет таймеров, Spinner, Progress
- **Layout constraints** — нет padding, gap, align, justify как в CSS
- **Help overlay** — нет встроенного help с key bindings
- **HTTP element tree** — `/elements` отдаёт дерево `ElementProvider`; события через HTTP не передаются

## Changelog

### v0.7
- **Input** — `NewInput(prompt)` с курсором, backspace, delete, стрелками
- **Focus API** — `Focusable`, `FocusNext()`, `FocusPrev()`, `FocusFirst()`, `FocusPanel()`
- **RawKeyReceiver** — для PTY/терминалов
- **Modal** — `ShowModalMsg`/`CloseModalMsg`, overlay, drag, close, buttons
- **Popover** — контекстное меню с Overlay, HandleMouse, HandleKey
- **Element tree** — `ElementProvider`, `Element`, `Bounds`
- **ContextMenu удалён** — заменён на Popover
- **Float close-on-outside-click** — `CloseOnOutsideClick`
- **OSC 52 clipboard** — `Selectable.Copy()`
- **35 тестов**

### v0.6
- **TabGroup как Panel** — табы внутри splits/flex, Warp — thin wrapper
- **Backward-compatible API** — `w.NewTab()`, `w.ActiveTab()` делегируют root TabGroup

### v0.5
- Табы не прыгают, Scrollable, DropdownMenu, ContextMenu, WordWrap, Collapsible, Flexbox
- Gruvbox Dark, вложенные табы, WindowSizeMsg broadcast
- Float overlay fix, ANSI isolation, float close button
- 24→27 тестов

## Правила кода

- Не использовать `os.Exit()` — только `tea.Quit`
- Не добавлять свои `signal.Notify` — Bubbletea сам обрабатывает SIGINT
- Минимальный размер панели: `MinPanelSize = 3`
- Fraction всегда через `clampFraction(0.1–0.9)`
- Файлы завершаются переносом строки
- Комментарии только на английском
- Warp не биндит Tab/Shift+Tab — фокус управляется разработчиком

## Как запустить тесты

```bash
cd /Users/a/Space/Projects/Starframe/warp
go vet ./...
go test ./... -v
```
