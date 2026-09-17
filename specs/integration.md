# Интеграция с внешними системами

## HTTP API

Warp запускает собственный HTTP-сервер для экспорта дерева UI
и проверки доступности. Сервер не связан с Bubbletea-циклом:
`Warp` живёт в одном goroutine, HTTP-сервер — в своём.

### Эндпоинты

**`GET /elements`** — JSON-массив элементов:

```json
[
  {
    "role": "tab",
    "name": "editor",
    "bounds": {"x": 0, "y": 0, "w": 80, "h": 1},
    "children": [
      {
        "role": "input",
        "name": "Name: ",
        "action": "focus",
        "bounds": {"x": 0, "y": 2, "w": 20, "h": 1}
      }
    ]
  }
]
```

Заголовки: `Content-Type: application/json`,
`Access-Control-Allow-Origin: *` (CORS для E2E-фреймворков).

**`GET /healthz`** — `ok`.

### Запуск

```go
w := warp.New()
_ = w.ServeHTTP(":8080")
w.HTTPAddr()          // текущий адрес ("" если не запущен)
_ = w.CloseHTTP()    // остановка перед shutdown
```

`addr == ""` → порт из `WARP_HTTP_PORT`, если не задан — `:0`
(случайный). `HTTPAddr()` возвращает фактически
присоединённый адрес.

### Контракт с E2E-тестами

- Тест-клиент запрашивает `/elements`, парсит JSON,
  ищет элемент через `FindElement(elems, role, name, action)`.
- Клик по элементу — `Bounds.Center()` → координаты мыши
  в Bubbletea-окне.
- Повторный запрос `/elements` после действия — сверка
  состояния.

`/elements` **не** рендерит UI: он опрашивает компоненты
через `ElementProvider`. Компоненты, реализующие интерфейс,
вызывают `Elements(w, h)` с текущими размерами Warp.

## Встраивание в чужое TUI

Warp — `tea.Model`, но через `AsPanel()` становится обычным
`Panel`. Пример:

```go
inner := warp.New()
inner.SetRoot(innerTabGroup)

outerTab := outerWarp.ActiveTab()
outerTab.SplitVertical(outerTab.RootPanel(), 0.5, inner.AsPanel())
```

Адаптер `warpPanel` сам вызывает `View(w, h)` с переданными
размерами и прокидывает `Update` дальше. Состояние `width`/`height`
в `Warp` синхронизируется с каждым вызовом.

## Вложенные Warp

Вложенные Warp живут как обычные `Panel`:
внешний `Warp` управляет границей, внутренний `Warp` —
своим корнем. HTTP-сервер поднимается только в том Warp,
который сам его запустил; вложенный — только если явно
`ServeHTTP`.

## Тестирование

- Unit-тесты: `*_test.go` рядом с кодом.
- E2E: HTTP `/elements` + `FindElement` в тестовом фреймворке.
- `examples_test.go` с `// Output:` — smoke-рендер.
- `go vet ./...` и `go test ./...` — стандартный приём.

## Ограничения

- HTTP-сервер не отдаёт события — только состояние.
- Нет авторизации: `/elements` — публичный.
- `HTTPAddr()` фиксирует адрес, с которым
  запущен сервер; повторный `ServeHTTP` игнорируется
  (`httpServer != nil` → `return nil`).
- `CloseHTTP` останавливает `http.Server` через
  `http.Server.Shutdown` и обнуляет `httpAddr`.
