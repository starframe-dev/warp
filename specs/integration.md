# Интеграция и внешние системы

## HTTP API

### Эндпоинты

#### `/elements`

Экспортирует дерево UI-элементов в формате JSON (через `ElementProvider`).

**Заголовки:** `Content-Type: application/json`, `Access-Control-Allow-Origin: *`

**Ответ:**
```json
[
  {"role": "button", "name": "+Folder", "bounds": {"x": 0, "y": 0, "w": 10, "h": 1}}
]
```

#### `/healthz`

```
ok
```

### Использование

```go
w := warp.New()
err := w.ServeHTTP(":8080")
// GET http://localhost:8080/elements
```

## Встраивание

Через `AsPanel()`:

```go
inner := warp.New()
outerTab.SplitVertical(parent, 0.5, inner.AsPanel())
```

## Тестирование

- Unit-тесты: `*_test.go` рядом с кодом
- Element tree: `/elements` для E2E-тестов по семантическим координатам
- Example tests: `examples_test.go` с `// Output:` проверками
