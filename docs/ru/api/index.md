---
title: API Warp
description: Публичные типы, компоненты, интерфейсы и функции Warp.
---

# API

Warp — Go TUI layout engine для Bubble Tea. API строится вокруг корневого `Warp`, составных `Panel` и дерева компоновки `Tab`.

## Основные типы

| Тип | Назначение |
|-----|-----------|
| [`Warp`](./warp) | Корневая Bubble Tea-модель и HTTP-сервер дерева элементов |
| [`Panel`](./panel) | Интерфейс компонента |
| [`TabGroup`](./tabgroup) | Таб-бар и управление активной вкладкой |
| [`Tab`](./tab) | Операции split, flex, float, фокус и сворачивание |
| [`Node`](./split) | Узел дерева компоновки |
| [`SplitConfig`](./split) | Конфигурация разделения на два дочерних узла |
| [`FlexConfig`](./split) | Горизонтальная или вертикальная flex-компоновка |
| [`FlexItem`](./split) / `FlexItemSpec` | Элемент flex и его вес |
| [`FloatPane`](./float) | Состояние плавающей панели |
| [`Element`](./element) / [`Bounds`](./element) | Значения семантического дерева UI |
| [`ThemeColors`](./theme) | Семантическая палитра темы |

## Компоненты

| Компонент | Конструктор | Назначение |
|-----------|-------------|-----------|
| [`Collapsible`](./collapsible) | `NewCollapsible` | Сворачиваемая секция |
| [`Scrollable`](./scrollable) | `NewScrollable` | Прокручиваемая область |
| [`DropdownMenu`](./dropdown) | `NewDropdownMenu` | Выпадающий список |
| [`Selectable`](./selectable) | `NewSelectable` | Выделение и копирование текста |
| [`Input`](./input) | `NewInput` | Однострочный ввод |
| [`Modal`](./modal) | `NewModal` / `ShowModalMsg` | Модальное окно |
| [`Popover`](./popover) | `&Popover{...}` | Контекстное меню |

## Интерфейсы

- [`Focusable`](./focus) — явный фокус панели
- [`RawKeyReceiver`](./focus) — намерение получать необработанные клавиши
- [`ElementProvider`](./element) — семантическое дерево элементов

## Вспомогательные функции

- [`WordWrap`](./wrap) и `SpaceWrap` — перенос текста
- [`StripANSI`](./float) — удаление ANSI-последовательностей
- [`FindElement`](./element) — рекурсивный поиск элемента
- `WrapToString` — перенос текста с объединением строк

## Дополнительные типы

`Direction`, `TabPosition`, `ResizeMsg`, `NodeCollapse`, `ModalButton` и `DropdownItem` описаны на страницах компонентов, которые их используют.

## Сгенерированный справочник файлов

code-check также создаёт подробную HTML-документацию для каждого Go-файла. Она остаётся в `docs/en/` и `docs/ru/`, потому что эти пути являются контрактом code-check. Для работы с API рекомендуются страницы VitePress выше. Ссылки находятся в [сгенерированном справочнике](../guide/generated-reference).
