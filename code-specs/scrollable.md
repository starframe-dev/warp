# scrollable.md

## Обзор

Scrollable — это обёртка для компонента Panel, добавляющая поддержку прокрутки содержимого при превышении высоты отведённого пространства. Позволяет пользователям прокручивать содержимое через колёсико мыши или клавиатуру.

## Публичный API

### Тип `Scrollable`

```go
// Scrollable обёртка для Panel с поддержкой прокрутки.
// Когда содержимое превышает отведённую высоту, пользователь может
// прокручивать через колёсико мыши.
type Scrollable struct {
    Content Panel
    Offset  int // смещение прокрутки в строках
}
```

| Поле | Тип | Описание |
|------|-----|----------|
| `Content` | Panel | Содержимое, которое может быть прокручено |
| `Offset` | int | Текущее смещение прокрутки (в строках), 0-значение — верх viewport |

### Функции

#### `NewScrollable`

```go
// NewScrollable создаёт новую обёртку scrollable.
//
// @param content — Panel-компонент, который будет обёрнут
// @returns *Scrollable — новый экземпляр Scrollable
// @example
// ```go
// s := NewScrollable(content)
// ```
func NewScrollable(content Panel) *Scrollable
```

#### `View`

```go
// View рендерит видимую область viewport содержимого.
//
// @param w — ширина viewport в символах
// @param h — высота viewport в строках
// @returns string — рендеренный вид
//
// Поведение:
// 1. Если Content nil — возвращает пустые строки высотой h
// 2. Рендерит полное содержимое с неограниченной высотой
// 3. Кламит Offset в диапазоне [0, maxOffset]
// 4. Возвращает видимый срез с padding по ширине
//
// @sideeffect io — рендеринг в строку
func (s *Scrollable) View(w, h int) string
```

#### `Update`

```go
// Update обрабатывает сообщения прокрутки (колёсико мыши, клавиши).
//
// @param msg tea.Msg — сообщение от tea.Msg
// @returns tea.Cmd — команда (возвращает nil)
//
// Обработчики:
// - tea.MouseMsg.MouseButtonWheelUp: смещение вверх на 3 строки
// - tea.MouseMsg.MouseButtonWheelDown: смещение вниз на 3 строки
// - tea.KeyMsg: "up" — вверх на 1, "down" — вниз на 1,
//   "pgup" — вверх на 10, "pgdown" — вниз на 10
//
// @sideeffect mutation — изменение поля Offset
// @pure
func (s *Scrollable) Update(msg tea.Msg) tea.Cmd
```

#### `padLine` (внутренняя)

```go
// padLine выравнивает строку по ширине w.
//
// @param line — исходная строка
// @param w — целевая ширина в символах
// @returns string — отцентрированная/обрезанная строка
//
// Логика:
// 1. Если ширина строки >= w — обрезает до w-1 символа
// 2. Иначе добавляет пробелы справа до ширины w
//
// @pure
func padLine(line string, w int) string
```

## Детали реализации

### Структура `Scrollable`

```go
type Scrollable struct {
    Content Panel
    Offset  int
}
```

| Поле | Тип | Инициализация | Описание |
|------|-----|---------------|----------|
| `Content` | Panel | По умолчанию | Panel-компонент, содержащий контент |
| `Offset` | int | По умолчанию (0) | Смещение прокрутки в строках |

### Конструктор

```go
func NewScrollable(content Panel) *Scrollable {
    return &Scrollable{Content: content}
}
```

**Поведение:**
- Создаёт новый экземпляр Scrollable
- Не инициализирует Offset (получает zero value — 0)

### Метод `View`

```go
func (s *Scrollable) View(w, h int) string
```

**Логика рендеринга:**

1. **Проверка nil:** Если `Content == nil` — возвращает `h` пустых строк
2. **Полный рендер:** Вызывает `s.Content.View(w, 9999)` для получения всего контента
3. **Разбиение на строки:** `strings.Split(fullContent, "\n")`
4. **Кламминг Offset:**
   - `maxOffset = len(lines) - h`
   - Если `maxOffset < 0` → `maxOffset = 0`
   - Если `s.Offset < 0` → `s.Offset = 0`
   - Если `s.Offset > maxOffset` → `s.Offset = maxOffset`
5. **Срез видимого:** Берёт `h` строк начиная с `s.Offset`
6. **Padding:** Выравнивает каждую строку по ширине `w` через `padLine`
7. **Объединение:** `strings.Join(visible, "\n")`

### Метод `Update`

```go
func (s *Scrollable) Update(msg tea.Msg) tea.Cmd
```

**Обработка сообщений:**

#### `tea.MouseMsg`

| Button | Действие |
|--------|----------|
| `tea.MouseButtonWheelUp` | `Offset -= 3`, кламп до 0 |
| `tea.MouseButtonWheelDown` | `Offset += 3` |

#### `tea.KeyMsg`

| Key | Действие |
|-----|----------|
| `"up"` | `Offset -= 1`, кламп до 0 |
| `"down"` | `Offset += 1` |
| `"pgup"` | `Offset -= 10`, кламп до 0 |
| `"down"` | `Offset += 10` |

**Побочные эффекты:**
- Изменение поля `Offset` (mutation)
- Вызов `s.Content.Update(msg)` если Content не nil

**Возвращаемое:**
- Всегда `nil` (нет команд)

### Внутренняя функция `padLine`

```go
func padLine(line string, w int) string
```

**Логика:**
1. Вычисляет ширину строки через `lipgloss.Width(line)`
2. Если `lw >= w`:
   - Ищет байтовую границу, где ширина становится > w
   - Возвращает `line[:i-1]` (обрезает на 1 символ меньше)
   - Если нет такой границы — возвращает исходную строку
3. Иначе добавляет пробелы справа: `line + strings.Repeat(" ", w-lw)`

**Примечание:** Используется `lipgloss.Width` для корректного учёта ширины символов (CJK и т.п.)

## Тесты

### TestNewScrollable

```go
func TestNewScrollable(t *testing.T) {
    t.Run("creates scrollable with nil offset", func(t *testing.T) {
        s := NewScrollable(nil)
        if s.Offset != 0 {
            t.Errorf("Offset = %d, want 0", s.Offset)
        }
        if s.Content != nil {
            t.Errorf("Content = %v, want nil", s.Content)
        }
    })
}
```

### TestView_EmptyContent

```go
func TestView_EmptyContent(t *testing.T) {
    s := NewScrollable(nil)
    result := s.View(80, 24)
    if len(strings.Split(result, "\n")) != 24 {
        t.Errorf("got %d lines, want 24", len(strings.Split(result, "\n")))
    }
}
```

### TestView_Scrolling

```go
func TestView_Scrolling(t *testing.T) {
    content := Panel{
        View: func(w, h int) string {
            return strings.Repeat("A\n", 100)
        },
    }
    s := NewScrollable(content)

    // Top view
    s.View(80, 10)
    if s.Offset != 0 {
        t.Errorf("Offset = %d, want 0", s.Offset)
    }

    // Scroll down
    s.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown})
    if s.Offset != 3 {
        t.Errorf("Offset after wheel down = %d, want 3", s.Offset)
    }

    // Clamp test
    s.Offset = 1000
    s.View(80, 10)
    if s.Offset != 99 {
        t.Errorf("Offset after clamp = %d, want 99", s.Offset)
    }
}
```

### TestUpdate_Keyboard

```go
func TestUpdate_Keyboard(t *testing.T) {
    s := NewScrollable(nil)

    s.Update(tea.KeyMsg{String: "up"})
    if s.Offset != -1 {
        t.Errorf("Offset after up = %d, want -1", s.Offset)
    }

    s.Update(tea.KeyMsg{String: "down"})
    if s.Offset != 0 {
        t.Errorf("Offset after down = %d, want 0", s.Offset)
    }

    s.Update(tea.KeyMsg{String: "pgup"})
    if s.Offset != -10 {
        t.Errorf("Offset after pgup = %d, want -10", s.Offset)
    }
}
```

### TestPadLine

```go
func TestPadLine(t *testing.T) {
    t.Run("short line gets padded", func(t *testing.T) {
        got := padLine("hello", 10)
        expected := "hello     "
        if got != expected {
            t.Errorf("padLine(%q, 10) = %q, want %q", "hello", got, expected)
        }
    })

    t.Run("long line gets truncated", func(t *testing.T) {
        got := padLine("this is a very long line that should be truncated", 20)
        if len(got) != 20 {
            t.Errorf("len(got) = %d, want 20", len(got))
        }
    })
}
```

## Чеклист

- [ ] Публичные функции имеют JSDoc/doc comments
- [ ] Есть примеры использования
- [ ] Тесты Red-Green-Refactor
- [ ] README актуален

## Ключевые правила

1. **Offset всегда клампится** в диапазоне `[0, maxOffset]`
2. **Mouse wheel** двигает на 3 строки
3. **Keys** двигают на 1 или 10 строк (pgup/pgdown)
4. **Padding** использует `lipgloss.Width` для корректного учёта ширины символов
5. **Nil Content** возвращается пустым viewportом высотой h
