# Управление состоянием

## Локальное состояние

Каждый компонент хранит своё состояние внутри экземпляра:

```go
type Input struct {
    Value   string
    Cursor  int
    Prompt  string
    focused bool
}
```

## Состояние через Bubbletea

Warp использует стандартную Bubbletea модель:
- `Update(msg tea.Msg) tea.Cmd` — обработка сообщений, изменение состояния
- `View(w, h int) string` — рендеринг текущего состояния

Сообщения:
- `tea.KeyMsg` — клавиатура
- `tea.MouseMsg` — мышь
- `tea.WindowSizeMsg` — изменение размера
- `ShowModalMsg` / `CloseModalMsg` — модальные окна

## Нет глобального состояния

Warp не использует глобальное состояние, подписки или эмиттеры.
Каждый компонент изолирован. Состояние передаётся через сообщения Bubbletea.
