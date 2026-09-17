---
title: Дерево элементов
description: Семантические элементы и границы UI для тестов и HTTP-инспекции.
---

# Дерево элементов

Панели могут предоставлять семантическое представление UI. Warp использует его для E2E-тестов и HTTP-эндпоинта `/elements`.

## Element

```go
type Element struct {
    Role     string    `json:"role"`
    Name     string    `json:"name"`
    Action   string    `json:"action,omitempty"`
    Bounds   Bounds    `json:"bounds"`
    Children []Element `json:"children,omitempty"`
}
```

`Role`, `Name` и `Action` описывают элемент. `Children` позволяет строить вложенные контролы.

## Bounds

```go
type Bounds struct {
    X int `json:"x"`
    Y int `json:"y"`
    W int `json:"w"`
    H int `json:"h"`
}

func (b Bounds) Center() (int, int)
```

Координаты указаны в терминальных ячейках. `Center` возвращает центральную ячейку прямоугольника.

## ElementProvider

```go
type ElementProvider interface {
    Elements(width, height int) []Element
}

type ElementProviderFunc func(width, height int) []Element
```

`ElementProviderFunc` адаптирует функцию к интерфейсу без отдельного типа.

## FindElement

```go
func FindElement(elems []Element, role, name, action string) (Element, bool)
```

Пустое значение фильтра совпадает с любым значением. Функция рекурсивно ищет первый подходящий элемент, включая дочерние.

## HTTP

`Warp.ServeHTTP` отдаёт текущее дерево как JSON через `GET /elements`. `GET /healthz` возвращает `ok`.
