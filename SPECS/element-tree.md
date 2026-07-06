# HTTP element tree в warp

## Контекст
Manual E2E-тесты (`automata/e2e/manual`) и cue-tty сейчас кликают по raw-координатам ячеек. Координаты hover-иконок (`✎`, `✕`, `+`) зависят от длины имени чата/папки, поэтому тесты хрупкие и ломаются при любом рефакторинге рендера.

## Цель
Добавить в warp HTTP endpoint, который возвращает семантическое дерево UI-элементов с их экранными координатами. Тесты смогут находить элементы по `role`/`name`/`action` и кликать в центр их bbox, не вычисляя координаты вручную.

## Что изменится

1. `warp/element.go` — типы `Element`, `Bounds`, интерфейс `ElementProvider`.
2. `warp/warp.go` — HTTP-сервер с endpoint `/elements`.
3. `warp/panel.go` — optional `ElementProvider` support.
4. `automata/internal/tree/elements.go` — `Tree.Elements()` реализация.
5. `automata/main.go` — запуск HTTP-сервера, экспорт порта в env.
6. `automata/e2e/manual/main.go` — переписать на element tree.
7. Тесты: `warp/element_test.go`, `automata/e2e/element_tree_test.go`.

## Детали реализации

### Формат `/elements` JSON
```json
[
  {
    "role": "header",
    "name": "Automata",
    "bounds": {"x": 0, "y": 0, "w": 80, "h": 1},
    "children": [
      {"role": "button", "name": "+Folder", "action": "add-folder", "bounds": {...}},
      {"role": "button", "name": "+Chat", "action": "add-chat", "bounds": {...}}
    ]
  },
  {
    "role": "folder",
    "name": "Work",
    "bounds": {"x": 0, "y": 4, "w": 22, "h": 1},
    "children": [
      {"role": "action", "name": "new-folder", "action": "new-folder", "bounds": {...}},
      {"role": "action", "name": "rename", "action": "rename", "bounds": {...}},
      {"role": "action", "name": "delete", "action": "delete", "bounds": {...}}
    ]
  }
]
```

### ElementProvider interface
```go
type ElementProvider interface {
    Elements(width, height int) []Element
}

type Element struct {
    Role     string    `json:"role"`
    Name     string    `json:"name"`
    Action   string    `json:"action,omitempty"`
    Bounds   Bounds    `json:"bounds"`
    Children []Element `json:"children,omitempty"`
}

type Bounds struct {
    X int `json:"x"`
    Y int `json:"y"`
    W int `json:"w"`
    H int `json:"h"`
}
```

### HTTP сервер
- warp запускает `net/http` сервер на порту из env `WARP_HTTP_PORT` или `0` (автопорт).
- Порт пишется в `WARP_HTTP_PORT` env после запуска.
- `GET /elements` возвращает JSON от корневого panel.
- CORS: `*`.
- Сервер останавливается при закрытии warp/программы.

## Критерии приёмки
- [ ] `GET /elements` возвращает JSON с header, folders, chats, actions, modals.
- [ ] Координаты action icons корректны.
- [ ] Manual test переписан на поиск элементов.
- [ ] `go test ./...` в warp и automata проходит.
- [ ] `go run ./e2e/manual` успешно выполняет create/rename/delete/drag.
- [ ] cue-tty не изменяется.
