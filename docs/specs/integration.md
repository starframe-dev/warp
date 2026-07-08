# Integration

> Полная спецификация: [`specs/integration.md`](https://github.com/starframe-dev/warp/blob/main/specs/integration.md)

## HTTP API

- `GET /elements` — дерево UI-элементов (JSON)
- `GET /healthz` — проверка доступности

## Встраивание

```go
inner := warp.New()
outerTab.SplitVertical(parent, 0.5, inner.AsPanel())
```
