---
title: Warp
description: Корневая Bubble Tea-модель, вкладки, вложенность и HTTP-инспекция.
---

# Warp

`Warp` — корневая Bubble Tea-модель. Она владеет корневой `Panel` и передаёт ей сообщения.

## Конструктор

```go
func New() *Warp
```

Создаёт `Warp` с корнем `TabGroup` и одной вкладкой `main`.

## Корень и делегирование вкладок

```go
func (w *Warp) SetRoot(panel Panel)
func (w *Warp) Root() Panel

func (w *Warp) NewTab(name string) *Tab
func (w *Warp) ActiveTab() *Tab
func (w *Warp) SetTabPosition(pos TabPosition)
func (w *Warp) NextTab()
func (w *Warp) PrevTab()
```

Методы вкладок работают только если корень — `*TabGroup`. После замены корня через `SetRoot` они ничего не делают, а `ActiveTab` и `NewTab` возвращают `nil`.

## Модель Bubble Tea

```go
func (w *Warp) Init() tea.Cmd
func (w *Warp) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (w *Warp) View() string
```

До первого ненулевого `WindowSizeMsg` метод `View` возвращает `Loading...`.

## Вложенность

```go
func (w *Warp) AsPanel() Panel
```

Возвращает адаптер, который передаёт вложенному Warp выделенный размер и сообщения.

## HTTP API дерева элементов

```go
func (w *Warp) ServeHTTP(addr string) error
func (w *Warp) CloseHTTP() error
func (w *Warp) HTTPAddr() string
```

`ServeHTTP` предоставляет:

- `GET /elements` — JSON-дерево элементов; до первого resize используются размеры `80 × 24`;
- `GET /healthz` — текстовый ответ `ok`.

Пустой адрес использует `WARP_HTTP_PORT`; если переменная не задана, ОС выбирает свободный порт. Повторный вызов `ServeHTTP` во время работы ничего не меняет.

## Размеры и запуск

```go
func (w *Warp) Width() int
func (w *Warp) Height() int
func (w *Warp) Run() error
```

`Width` и `Height` возвращают последние полученные размеры. `Run` запускает Bubble Tea с alternate screen и отслеживанием движения мыши по ячейкам.
