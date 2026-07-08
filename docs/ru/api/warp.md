---
title: Warp
description: Корневой объект приложения Warp и управление вкладками, панелями, HTTP и размером экрана.
---

# Warp

`Warp` — основной объект приложения. Он хранит корневую панель, управляет вкладками и запускает TUI.

## Создание

```go
New() *Warp
```

Создаёт новый экземпляр `Warp`.

## Корневая панель

```go
SetRoot(panel Panel)
Root() Panel
```

`SetRoot` задаёт корневую панель. `Root` возвращает текущую корневую панель.

## Вкладки

```go
NewTab(name string) *Tab
ActiveTab() *Tab
SetTabPosition(pos TabPosition)
NextTab()
PrevTab()
```

`NewTab` создаёт вкладку. `ActiveTab` возвращает активную вкладку. `SetTabPosition` меняет расположение вкладок. `NextTab` и `PrevTab` переключают активную вкладку.

## Panel-совместимость

```go
AsPanel() Panel
```

Возвращает `Warp` как `Panel` для вложенного использования.

## Запуск

```go
Run() error
```

Запускает Bubble Tea приложение.

## HTTP

```go
ServeHTTP(addr string) error
CloseHTTP() error
HTTPAddr() string
```

`ServeHTTP` запускает HTTP-сервер. `CloseHTTP` останавливает его. `HTTPAddr` возвращает текущий адрес HTTP-сервера.

## Размер

```go
Width() int
Height() int
```

Возвращают текущую ширину и высоту терминала.
