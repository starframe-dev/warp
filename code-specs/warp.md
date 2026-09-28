# warp.go — спецификация реализации пакета `warp`

## Обзор

Файл определяет корневую модель Bubbletea `Warp`, адаптер для встраивания `Warp` как `Panel` и HTTP-сервер для публикации snapshot дерева элементов.

## Тип `Warp`

`Warp` хранит корневую панель, ревизии корня и состояния, последние размеры окна, состояние HTTP-сервера и snapshot элементов. `mu` защищает корень, ревизии, размеры и поля HTTP-сервера; отдельный `elementsSnapshotMu` защищает опубликованный snapshot. Snapshot строится вне блокировки `mu`, а публикуется, только если ревизии корня и состояния всё ещё совпадают с использованными при построении. Дерево глубоко копируется вместе с `Children`; копирование ограничено глубиной 128 уровней и 100 000 узлами.

### Создание и корень

- `New() *Warp` создаёт `Warp` с `TabGroup` в позиции `TabTop`. Snapshot изначально пуст: nil snapshot при HTTP-кодировании даёт JSON-массив `[]`.
- `SetRoot(panel Panel)` собирает панели старого корня, обновляет и ревизует корень и состояние, затем вызывает `Unmount` лишь для экземпляров, недостижимых из нового дерева владельца. Если inspector включён, опубликованный snapshot очищается.
- `Root() Panel` возвращает текущий корень под блокировкой чтения.

### Безопасность framework labels

Текст, встроенный самим Warp в UI chrome, не является произвольным terminal stream. Перед компоновкой tab names, float/collapsible titles, dropdown labels/options, input prompts, modal titles/button labels и popover item names проходят plain-text sanitization: invalid UTF-8 нормализуется, raw layout controls нейтрализуются, ANSI/OSC sequences удаляются. Это предотвращает terminal injection и нарушение геометрии framework chrome. Пользовательский `Panel.View` и `Modal.Content` сохраняют ANSI semantics.

### Владение ресурсами

Один корневой `Warp` владеет всем деревом панелей, включая вложенные `TabGroup`, `Tab`, wrappers и float-панели. Самостоятельные `TabGroup` и `Tab` образуют отдельные домены, пока не вложены во внешний домен. Указательные панели идентифицируются типом и адресом экземпляра; `Unmount` вызывается однократно после detach и удаления последней ссылки внутри домена. Hiding, collapse, focus changes и reparenting не приводят к `Unmount`. Один экземпляр, разделённый между независимыми `Warp`/owner roots, не поддерживается и управляется вызывающим кодом; глобального реестра нет.

### Делегирование и размеры

- `tabGroup() *TabGroup` — внутренний помощник, возвращающий корень, если тот имеет тип `*TabGroup`.
- `Width() int` и `Height() int` возвращают сохранённые размеры.
- `NewTab(name string) *Tab`, `ActiveTab() *Tab`, `SetTabPosition(pos TabPosition)`, `NextTab()` и `PrevTab()` делегируют корневому `TabGroup`. Если корень не `*TabGroup`, методы, возвращающие значение, возвращают `nil`, а остальные ничего не делают.

### Bubbletea и запуск

- `Init() tea.Cmd` возвращает `nil`.
- `Update(msg tea.Msg) (tea.Model, tea.Cmd)` сохраняет размеры при `tea.WindowSizeMsg`, передаёт сообщение корню (если он не nil) и, если inspector включён, обновляет snapshot. Если сообщение реализует `SemanticStableMsg` и возвращает `true`, Update-snapshot пропускается как заведомо семантически неизменный; предыдущий immutable snapshot остаётся опубликованным до следующего `View`. Возвращает сам `Warp` и команду корневой панели.
- `View() string` возвращает пустую строку при nil-корне. При нулевой ширине или высоте возвращает `Loading...`; иначе рендерит корень. После каждого вызова обновляет snapshot, если inspector включён.
- `Run() error` запускает Bubbletea-программу с `WithAltScreen` и `WithMouseCellMotion` и возвращает ошибку `Program.Run`.
- `Close() error` останавливает HTTP inspector и, если Warp является корнем ownership domain, заменяет корень на nil, вызывая lifecycle cleanup уникальных `Unmounter` панелей. Для embedded Warp внешним lifecycle владеет родитель.

### Вложенная панель

`AsPanel() Panel` возвращает адаптер `warpPanel`.

- `warpPanel.View(width, height)` сохраняет размеры родителя во вложенном `Warp`, затем вызывает `Warp.View()`.
- `warpPanel.Update(msg)` вызывает `Warp.Update(msg)` и возвращает команду.

## HTTP-сервер

### Методы

- `ServeHTTP(addr string) error` использует безопасные defaults и делегирует `ServeHTTPWithOptions(addr, InspectorOptions{})`.
- `ServeHTTPWithOptions(addr string, options InspectorOptions) error` запускает сервер с маршрутами `/elements` и `/healthz`. При пустом адресе используется loopback `127.0.0.1` и `WARP_HTTP_PORT` либо порт 0. `InspectorOptions.AllowedOrigin` явно разрешает точный CORS origin или `"*"`; по умолчанию CORS не добавляется. `BearerToken` при непустом значении требует `Authorization: Bearer ...` для `/elements`. Повторный запуск при уже активном сервере или в процессе закрытия — no-op.
- `CloseHTTP() error` отключает inspector demand, очищает snapshot и выполняет bounded `http.Server.Shutdown` с timeout; при ошибке Shutdown выполняется `server.Close()`. Mutex Warp не удерживается во время ожидания shutdown.
- `HTTPAddr() string` возвращает сохранённый адрес или пустую строку.

### `/elements`

GET возвращает HTTP 200 с JSON-массивом immutable snapshot элементов; при отсутствии snapshot возвращается `[]`. При включённом bearer token неавторизованный запрос получает 401. OPTIONS возвращает 204 и разрешённые методы/headers; другие методы, кроме GET/OPTIONS, отклоняются. CORS-заголовок появляется только при явной настройке origin и совпадающем Origin (или настройке `*`).

При построении snapshot неизвестные размеры заменяются на 80×24. Snapshot обновляется после обычного `Update` и повторно после каждого завершённого `View`, потому что пользовательская реализация `Panel.View` может менять semantic state. Для `SemanticStableMsg` Update-snapshot можно пропустить. `SetRoot` немедленно очищает опубликованный snapshot, не вызывая `Elements` у нового root вне UI-cycle. HTTP handler читает только опубликованный snapshot и не обходит живое дерево.

### `/healthz`

Возвращает HTTP 200, заголовок `Content-Type: text/plain` и тело `ok`.

## Внутренняя функция

`parsePort(addr string) string` возвращает порт, полученный через `net.SplitHostPort`, либо пустую строку при ошибке. Эта функция в данном файле не используется.
