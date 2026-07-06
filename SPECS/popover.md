# Popover — контекстное меню в Warp

## Контекст

Сейчас контекстное меню (⋮ → New Folder, Rename, Delete и т.д.) рендерится и обрабатывается
напрямую в `tree/render.go` (overlayMenu) и `tree/input.go` (handleMouse для меню).
Это дублирует логику, которая уже есть в `warp/modal.go`: overlay поверх lines, hit-test,
сохранение контента слева/справа.

Нужно вынести Popover (контекстное меню) в warp как переиспользуемый компонент,
по аналогии с `warp.Modal`.

## Цель

Создать `warp.Popover` — компонент для контекстного меню, который:
- Рендерится поверх existing lines (как Modal.Overlay)
- Обрабатывает mouse (клик по пункту, клик вне меню → закрыть)
- Обрабатывает keyboard (Esc → закрыть, Enter → выбрать)
- Сохраняет контент слева/справа от меню (как Modal)

## Что изменится

1. `warp/popover.go` — новый файл: Popover struct, Overlay, HandleMouse, HandleKey
2. `tree/render.go` — заменить overlayMenu на вызов warp.Popover
3. `tree/input.go` — заменить handleMouse для меню на вызов popover.HandleMouse/HandleKey
4. `tree/model.go` — убрать menuBoxW/menuBoxH/menuSelected, добавить popover *warp.Popover
5. `warp/element.go` — добавить Popover в collectElements (если нужно)

## Детали реализации

### 1. warp/popover.go

```go
package warp

type PopoverItem struct {
    Name   string
    Action func()
}

type Popover struct {
    Items    []PopoverItem
    X, Y     int       // Позиция в screen-координатах (0 = header)
    Width    int       // Ширина контента (0 = auto: 20)
    OnClose  func()    // Вызывается при закрытии

    // rendered dimensions (set by Overlay)
    boxW, boxH int

    // selection
    selected int
}
```

**Overlay(lines []string, totalW, totalH int) []string:**

1. Если Items пусто → вернуть lines как есть
2. contentW = Width (или 20 по умолчанию)
3. Для каждого item:
   - Если selected → menuSelectedStyle
   - Иначе → menuBaseStyle
   - " " + item.Name, ширина contentW
4. Склеить в menuContent через "\n"
5. Оборачиваем в border box (NormalBorder, BorderForeground=gbBg2, Background=gbBg1)
6. Получаем boxLines
7. Сохраняем boxW = lipgloss.Width(boxLines[0]), boxH = len(boxLines)
8. Позиционирование:
   - menuX = X
   - menuY = Y - 1 (Y в screen-координатах, lines — content)
   - Если menuX + boxW > totalW → menuX = totalW - boxW
   - Если menuX < 0 → menuX = 0
   - Если menuY < 0 → menuY = 0
   - Если menuY + boxH > totalH → menuY = totalH - boxH
9. Для каждой строки boxLines:
   - leftPart = ansi.Truncate(original, menuX, "")
   - rightStart = visualBytePos(original, menuX + boxW)
   - rightPart = original[rightStart:] (если не пустой ANSI)
   - Дополнить leftPart пробелами до menuX
   - lines[menuY + i] = leftPart + boxLine + rightPart

**HandleMouse(msg tea.MouseMsg) bool:**

1. Если boxW == 0 → false (не рендерили)
2. На press:
   - Если msg.X внутри [menuX, menuX+boxW) и msg.Y внутри [menuY, menuY+boxH):
     - contentIdx = msg.Y - menuY - 1 (минус top border)
     - Если contentIdx в [0, len(Items)):
       - Items[contentIdx].Action()
       - OnClose()
     - return true
   - Иначе (клик вне меню):
     - OnClose()
     - return true
3. На motion:
   - Если msg.Y внутри [menuY+1, menuY+boxH-1):
     - contentIdx = msg.Y - menuY - 1
     - Если contentIdx в [0, len(Items)): selected = contentIdx
     - return true
4. На release: return true (swallow)

**HandleKey(msg tea.KeyMsg) bool:**

1. Если boxW == 0 → false
2. KeyEsc → OnClose(), return true
3. KeyEnter → Items[selected].Action(), OnClose(), return true
4. KeyUp → selected-- (clamp to 0), return true
5. KeyDown → selected++ (clamp to len-1), return true
6. Иначе → false

### 2. tree/render.go — замена overlayMenu

```go
// В View(), вместо:
//   if t.showMenu {
//       lines = t.overlayMenu(lines, width)
//   }
// теперь:
//   if t.popover != nil {
//       lines = t.popover.Overlay(lines, width, height)
//   }
```

Удалить функцию `overlayMenu` и связанные стили (menuSelectedStyle, menuBaseStyle).

### 3. tree/input.go — замена handleMouse для меню

```go
// В handleMouse, вместо:
//   if t.showMenu {
//       ... // ручная проверка координат
//   }
// теперь:
//   if t.popover != nil {
//       if t.popover.HandleMouse(msg) {
//           t.popover = nil
//       }
//       return nil
//   }
```

Добавить keyboard handler для popover в handleKey:
```go
// В handleKey, после modalActive:
//   if t.popover != nil {
//       if t.popover.HandleKey(msg) {
//           t.popover = nil
//       }
//       return nil
//   }
```

### 4. tree/model.go — замена полей

Убрать:
- `showMenu bool`
- `menuX, menuY int`
- `menuBoxW, menuBoxH int`
- `menuItems []menuItem`
- `menuSelected int`
- `closeMenu()`
- `showContextMenu()`
- `showCreateMenu()`

Добавить:
- `popover *warp.Popover`

Изменить `showContextMenu` и `showCreateMenu` на создание `warp.Popover`:
```go
func (t *Tree) showContextMenu(x, y int) {
    sel := t.ItemAt(t.selected)
    var items []warp.PopoverItem
    if sel != nil && sel.IsFolder {
        items = []warp.PopoverItem{
            {"New Folder", func() { t.startInput("Folder name:", ...) }},
            {"New Chat", func() { t.startInput("Chat name:", ...) }},
            {"Rename", func() { t.startInput("Rename:", ...) }},
            {"Delete", func() { t.startConfirm(sel) }},
        }
    } else if ...
    t.popover = &warp.Popover{
        Items:   items,
        X:       x,
        Y:       y,
        OnClose: func() { t.popover = nil },
    }
}
```

### 5. tree/elements.go — обновить Elements()

Изменить проверку `t.showMenu` на `t.popover != nil`:
```go
if t.popover != nil {
    elems = append(elems, t.popoverElements()...)
}
```

## Критерии приёмки

- [ ] `warp/popover.go` создан, тесты проходят
- [ ] Контекстное меню открывается по клику на ⋮
- [ ] Контекстное меню открывается по правому клику
- [ ] Клик по пункту меню выполняет действие
- [ ] Клик вне меню закрывает его
- [ ] Esc закрывает меню
- [ ] Enter выбирает пункт
- [ ] ↑/↓ навигация по пунктам
- [ ] Контент слева/справа от меню сохраняется
- [ ] Все тесты tree проходят
- [ ] Все тесты warp проходят
- [ ] go vet чисто
