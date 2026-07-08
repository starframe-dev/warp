---
title: Начало работы
description: Установка Warp и первый минимальный запуск Go TUI layout engine на Bubbletea.
---

# Начало работы

Warp — это Go TUI layout engine на Bubbletea.

## Установка

Установите пакет через `go get`:

```bash
go get github.com/starframe-dev/warp
```

## Минимальный пример

```go
package main

import "github.com/starframe-dev/warp"

func main() {
	w := warp.New()
	w.Run()
}
```

`Warp` — это Bubbletea `Model`. Его можно использовать как обычную модель Bubbletea.

`New()` создаёт экземпляр `Warp` с корневой `TabGroup`.

`Run()` запускает Bubbletea-программу.

## Демо

Чтобы посмотреть рабочий пример, запустите демо:

```bash
go run ./cmd/demo/
```
