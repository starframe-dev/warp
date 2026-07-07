# Интеграция и внешние системы

## Обзор

Этот документ описывает паттерны и практики интеграции проекта Warp с внешними системами и сервисами. Включает HTTP API, интеграцию с внешними UI-фреймворками, работу с состоянием и обработку событий.

## HTTP API

### Эндпоинты

#### `/elements`

Экспортирует дерево элементов UI в формате JSON.

**Заголовки:**
- `Content-Type: application/json`
- `Access-Control-Allow-Origin: *` (CORS)

**Ответ:**
```json
[
  {
    "role": "panel",
    "name": "main",
    "action": "focus",
    "bounds": {
      "x": 0,
      "y": 0,
      "w": 80,
      "h": 24
    },
    "children": [
      {
        "role": "tab-bar",
        "name": "tabs",
        "bounds": {"x": 0, "y": 0, "w": 80, "h": 1},
        "children": [...]
      }
    ]
  }
]
```

**Использование:**
```go
// Запуск HTTP сервера
w := warp.New()
err := w.ServeHTTP(":8080")
if err != nil {
    log.Fatal(err)
}

// Получение элементов
resp, err := http.Get("http://localhost:8080/elements")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

var elements []warp.Element
if err := json.NewDecoder(resp.Body).Decode(&elements); err != nil {
    log.Fatal(err)
}
```

#### `/healthz`

Эндпоинт здоровья для мониторинга.

**Ответ:**
```
ok
```

### Безопасность HTTP API

1. **CORS** — разрешены все_origin для разработки
2. **Rate limiting** — рекомендуется реализовать в продакшене
3. **Authentication** — рекомендуется добавить в продакшене
4. **CSRF** — не требуется для GET-эндпоинтов

## Интеграция с внешними UI-фреймворками

### Встраивание в Bubble Tea

Warp является нативным Bubble Tea приложением:

```go
import tea "github.com/charmbracelet/bubbletea"

// Создание программы
p := tea.NewProgram(
    warp.New(),
    tea.WithAltScreen(),
    tea.WithMouseCellMotion(),
)

_, err := p.Run()
if err != nil {
    log.Fatal(err)
}
```

### Встраивание в другие проекты

Через `AsPanel()`:

```go
w := warp.New()
panel := w.AsPanel()

// Использование panel в другом проекте
type MyPanel struct {
    warp.Panel
    // ...
}

func (m *MyPanel) View(width, height int) string {
    return m.Panel.View(width, height)
}
```

## Работа с состоянием

### Локальное состояние

Каждый компонент хранит своё состояние:

```go
type State struct {
    isOpen    bool
    value     string
    selected  []string
    focused   bool
}
```

### Глобальное состояние

Через `state.md` — управление глобальным состоянием через подписки.

### Синхронизация состояний

Состояния синхронизируются через Bubble Tea модели:

```go
// Update делегируется корневой панели
func (w *Warp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if w.root != nil {
        return w, w.root.Update(msg)
    }
    return w, nil
}
```

## Обработка событий

### Событийный цикл

1. Пользователь взаимодействует с элементом
2. Событие передаётся обработчику
3. Обработчик обновляет состояние
4. Состояние триггерит перерисовку
5. UI обновляется

### Типы событий

#### Bubble Tea события

- `tea.KeyMsg` — клавиатурные события
- `tea.MouseMsg` — мышиные события
- `tea.WindowSizeMsg` — изменение размера
- `tea.PtyOutputMsg` — вывод псевдоконсоли

#### Пользовательские события

```go
type CustomMsg struct {
    Type string
    Data interface{}
}
```

### Событийная модель

```
[Event] → [Handler] → [State Update] → [Render] → [View]
```

## Интеграция с внешними сервисами

### Загрузка данных

Асинхронная загрузка данных из внешних источников:

```go
// Пример асинхронной загрузки
type DataLoadMsg struct {
    URL     string
    Data    interface{}
    Error   error
}

func (w *Warp) LoadData(url string) tea.Cmd {
    return func() tea.Msg {
        data, err := fetchData(url)
        return DataLoadMsg{
            URL:   url,
            Data:  data,
            Error: err,
        }
    }
}
```

### Сохранение состояния

Сохранение и восстановление состояния:

```go
// Сохранение
func SaveState(component Element, prefix string) error

// Восстановление
func LoadState(component Element, prefix string) error
```

## Интеграция с тестовыми фреймворками

### Unit-тесты

```go
func TestWarp_New(t *testing.T) {
    w := warp.New()
    if w == nil {
        t.Fatal("expected Warp, got nil")
    }
}
```

### Интеграционные тесты

```go
func TestWarp_ElementsEndpoint(t *testing.T) {
    w := warp.New()
    if err := w.ServeHTTP("127.0.0.1:0"); err != nil {
        t.Fatal(err)
    }
    defer w.CloseHTTP()

    resp, err := http.Get(w.HTTPAddr() + "/elements")
    if err != nil {
        t.Fatal(err)
    }

    var elements []warp.Element
    if err := json.NewDecoder(resp.Body).Decode(&elements); err != nil {
        t.Fatal(err)
    }
}
```

## Паттерны интеграции

### Adapter Pattern

Через `AsPanel()`:

```go
func (w *Warp) AsPanel() warp.Panel {
    return &warpPanel{warp: w}
}
```

### Facade Pattern

Warp предоставляет фасад для работы с UI:

```go
func (w *Warp) NewTab(name string) *Tab
func (w *Warp) ActiveTab() *Tab
func (w *Warp) SetTabPosition(pos TabPosition)
```

### Strategy Pattern

Через `SetRoot()`:

```go
func (w *Warp) SetRoot(panel Panel) {
    w.root = panel
}
```

## Безопасность

1. **Валидация входных данных** — все внешние данные валидируются
2. **Защита от XSS** — содержимое экранируется
3. **Безопасность событий** — события проверяются
4. **Rate limiting** — рекомендуется для HTTP API
5. **Authentication** — рекомендуется для продакшена

## Производительность

### Оптимизации

1. **Ленивая загрузка** — компоненты рендерятся только при необходимости
2. **Дедупликация** — повторяющиеся операции дедуплицируются
3. **Batched updates** — обновления состояний батчуются

### Мониторинг

- `/healthz` — проверка доступности
- HTTP API — экспортирование метрик (опционально)

## Документация

### OpenAPI

Рекомендуется генерировать OpenAPI спецификацию для HTTP API.

### Примеры

- `cmd/demo/main.go` — демо-приложение
- Документация в README

## Миграция

### Сломывающие изменения

- Вводятся новой мажорной версией
- Предыдущие версии поддерживаются с депрешн-периодом

### Обратная совместимость

- API бэкворд-совместим
- Избегайте изменений существующих методов

## Поддержка

### Сообщество

- GitHub issues для отчётности о багах
- PR для вкладов в код

### Поддерживаемые платформы

- Linux
- macOS
- Windows (через WSL/Cygwin)
