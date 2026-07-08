# State

> Полная спецификация: [`specs/state.md`](https://github.com/starframe-dev/warp/blob/main/specs/state.md)

## Локальное состояние

Каждый компонент хранит своё состояние внутри экземпляра.

## Bubbletea модель

- `Update(msg tea.Msg) tea.Cmd` — обработка сообщений
- `View(w, h int) string` — рендеринг

## Нет глобального состояния

Warp не использует глобальное состояние, подписки или эмиттеры.
