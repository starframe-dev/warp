# Modal Windows in Warp

## Архитектура

Модальные окна рендерятся в **warp**, а не в tree. Tree отправляет `ShowModalMsg`/`CloseModalMsg` через `pendingCmd`.

```
Tree.startInput() → pendingCmd = ShowModalMsg{...}
  → Warp.Update(ShowModalMsg) → w.modal = NewModal(msg)
  → Warp.View() → root.View() + modal.overlay(lines, w, h)

Tree.handleKey() → sendInputUpdate() → pendingCmd = ShowModalMsg{...}
  → Warp.Update(ShowModalMsg) → w.modal обновляется

Tree.cancelInput() → pendingCmd = CloseModalMsg{}
  → Warp.Update(CloseModalMsg) → w.modal = nil
```

## Сообщения

- `ShowModalMsg` — открыть/обновить модалку
  - `Title`, `Content`, `Buttons []ModalButton`, `OnClose func()`, `Width int`
- `CloseModalMsg` — закрыть модалку

## Правила

1. **Warp перехватывает mouse/key** когда модалка активна
2. **Key events пробрасываются в root** для ввода текста (кроме Esc)
3. **Mouse events НЕ пробрасываются** — всё обрабатывает `Modal.handleMouse()`
4. **✕ кнопка** в правом верхнем углу (content line)
5. **Drag** за верхнюю полоску (padding line под border'ом)
6. **Кнопки**: клик по кнопке вызывает `Button.Action()`, модалка закрывается

## Детали реализации

### 1. Modal.overlay(lines, totalW, totalH) []string

Алгоритм:

1. Затемнить все строки через `dimStyle`
2. Вычислить `startX, startY` для центрирования
3. Построить box через lipgloss с `RoundedBorder() + Padding(1, 2)`
4. Для каждой строки box'а:
   - `leftPart = ansi.Truncate(original, startX, "")`
   - `rightStart = visualBytePos(original, startX+boxWidth)` — находит byte-позицию в оригинале
   - `rightPart = original[rightStart:]`
   - Если `rightPart` только ANSI-коды → пропустить
   - Дополнить `leftPart` пробелами до `startX` визуальных колонок
   - `lines[startY+i] = leftPart + boxLine + rightPart`

**Важно**: `visualBytePos` используется вместо `len(leftAndModal)`, потому что `ansi.Truncate` **не возвращает byte-prefix** оригинальной строки — он добавляет `\x1b[0m` в конце результата, из-за чего `original[len(leftAndModal):]` даёт неправильный срез.

### 2. Modal.handleMouse(msg tea.MouseMsg)

- ✕ кнопка: `x == startX + boxWidth - 4 && y == startY + 3` (box[3] — input line)
- Drag: `y == startY + 1` (box[1] — padding line), `x >= startX && x < startX + boxWidth`
- Кнопки: сканировать `btnLine` через `findBracketPair`, проверять `y == startY + 5` (box[5])

### 3. Warp.Update изменения

```go
case ShowModalMsg:
    w.modal = NewModal(msg)
case CloseModalMsg:
    w.modal = nil
case tea.KeyMsg:
    if w.modal != nil {
        if msg.Type == tea.KeyEsc { ... }
        // forward key to root for text input
    }
case tea.MouseMsg:
    if w.modal != nil {
        w.modal.handleMouse(msg)
    }
```

### 4. visualBytePos

`visualBytePos(s string, targetW int) int` — находит byte-позицию в строке `s`, где визуальная ширина достигает `targetW`. Использует тот же парсер ANSI-последовательностей, что и `ansi.Truncate`, но возвращает позицию в оригинальной строке, а не новую строку.

```go
func visualBytePos(s string, targetW int) int {
    curWidth := 0
    pstate := parser.GroundState
    b := []byte(s)
    i := 0
    for i < len(b) {
        state, action := parser.Table.Transition(pstate, b[i])
        if state == parser.Utf8State {
            cluster, _, width, _ := uniseg.FirstGraphemeCluster(b[i:], -1)
            i += len(cluster)
            if curWidth+width > targetW {
                return i - len(cluster)
            }
            curWidth += width
            pstate = parser.GroundState
            continue
        }
        switch action {
        case parser.PrintAction:
            if curWidth >= targetW { return i }
            curWidth++
            fallthrough
        default:
            i++
        }
        pstate = state
        if curWidth > targetW { return i - 1 }
    }
    return len(s)
}
```

## Координаты box'а

Стиль `RoundedBorder() + Padding(1, 2)` даёт box из 7 строк:

```
box[0]: ╭──────────────────────╮   top border
box[1]: │                      │   top padding (drag zone)
box[2]: │  Title            ✕  │   title + close button
box[3]: │  content             │   content line
box[4]: │  [Del]  [Esc]        │   button line
box[5]: │                      │   bottom padding
box[6]: ╰──────────────────────╯   bottom border
```

- `titleY = startY + 1` — drag zone (padding line)
- `xBtnY = startY + 3` — ✕ кнопка (content line)
- `btnY = startY + 5` — кнопки (button line)
- `xBtnX = startX + boxWidth - 4` — ✕ по X
