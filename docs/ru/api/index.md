---
title: Справочник API Warp
description: Публичные типы, компоненты, интерфейсы и функции Warp.
---

# Справочник API

Warp — движок компоновки TUI на Go для Bubbletea. Создайте экран из корневого <a href="./panel.html"><code>Panel</code></a>, затем объединяйте вкладки, сплиты, flex-компоновки, плавающие панели и интерактивные компоненты.

## Основной API

| Тип | Назначение |
|-----|-----------|
| <a href="./warp.html"><code>Warp</code></a> | Модель Bubbletea, корневая панель и необязательный HTTP-метод дерева элементов |
| <a href="./panel.html"><code>Panel</code></a> | Интерфейс каждой панели в дереве |
| <a href="./tabgroup.html"><code>TabGroup</code></a> | Панель вкладок и управление активной вкладкой |
| <a href="./tab.html"><code>Tab</code></a> | Дерево компоновки, floats, фокус, изменение размеров и сворачивание |
| <a href="./split.html"><code>Node</code></a> | Узел дерева компоновки |
| <a href="./split.html"><code>SplitConfig</code></a> / <a href="./split.html"><code>FlexConfig</code></a> | Конфигурация split и flex |
| <a href="./float.html"><code>FloatPane</code></a> | Состояние плавающей панели |
| <a href="./element.html"><code>Element</code></a> / <a href="./element.html"><code>Bounds</code></a> | Семантическое дерево UI и координаты в ячейках |
| <a href="./theme.html"><code>ThemeColors</code></a> | Палитра runtime-темы |

## Компоненты

| Компонент | Назначение |
|-----------|-----------|
| <a href="./collapsible.html"><code>Collapsible</code></a> | Сворачиваемая секция панели |
| <a href="./scrollable.html"><code>Scrollable</code></a> | Область просмотра со скроллом мышью и клавиатурой |
| <a href="./dropdown.html"><code>DropdownMenu</code></a> | Кнопка с раскрывающимся списком |
| <a href="./selectable.html"><code>Selectable</code></a> | Выделение текста мышью и клавиатурой |
| <a href="./input.html"><code>Input</code></a> | Однострочный редактируемый ввод |
| <a href="./modal.html"><code>Modal</code></a> | Перетаскиваемый диалоговый overlay |
| <a href="./popover.html"><code>Popover</code></a> | Overlay контекстного меню |

## Компоновка и утилиты

- <a href="./split.html"><code>Split</code></a> — направления, сообщения resize, состояние collapse и типы flex
- <a href="./focus.html"><code>Focus</code></a> — интерфейсы явного фокуса
- <a href="./styles.html"><code>Styles</code></a> — публичные стили границ и палитра по умолчанию
- <a href="./wrap.html"><code>WordWrap</code></a> и `SpaceWrap` — перенос текста по ширине терминала
- <a href="./float.html"><code>StripANSI</code></a> — удаление CSI-последовательностей из строки
- <a href="./element.html"><code>FindElement</code></a> — рекурсивный поиск семантического элемента
- `WrapToString` — перенос текста с объединением строк

## Контракты

- Размеры панелей измеряются в ячейках терминала, а не в пикселях или байтах.
- `ResizeMsg` передаёт листовым панелям выделенный размер содержимого.
- Фокус переключается явно: приложение само выбирает клавиши и вызывает `FocusNext`, `FocusPrev` или `FocusPanel`.
- Warp не резервирует `Tab` и `Shift+Tab` для перехода фокуса.
- По умолчанию используется Gruvbox Dark; <a href="./theme.html"><code>SetTheme</code></a> меняет семантическую палитру во время работы.
