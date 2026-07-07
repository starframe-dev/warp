# Modal

## Роль

Диалоговое окно, отображаемое поверх контента панели (panel). Предоставляет пользователю модальное взаимодействие с ограниченной областью внимания.

## API

### Типы

#### ShowModalMsg

Сообщение для отображения модалки.

| Поле | Тип | Описание |
| :---- | :---- | :---- |
| Title | string | Заголовок модалки (отображается в верхней области границы) |
| Content | string | Рендеримый контент (ANSI-строка) |
| Buttons | []ModalButton | Кнопки в нижней части модалки |
| OnClose | func() | Вызывается при закрытии модалки (✕ или Esc) |
| Width | int | Ширина рамки (0 = авто: 3/5 экрана, clamp 30-50) |

#### CloseModalMsg

Сообщение для закрытия текущей модалки.

| Поле | Тип | Описание |
| :---- | :---- | :---- |
| — | — | — |

#### Modal

Структура модалого окна.

| Поле | Тип | Описание |
| :---- | :---- | :---- |
| Title | string | Заголовок (топ-бorders) |
| Content | string | Рендеримый ANSI-контент |
| Buttons | []ModalButton | Кнопки внизу |
| OnClose | func() | Callback при закрытии |
| Width | int | Ширина рамки |
| startX | int | X-позиция рамки (после Overlay) |
| startY | int | Y-позиция рамки (после Overlay) |
| boxWidth | int | Ширина рамки |
| boxHeight | int | Высота рамки |
| dragging | bool | Флаг перетаскивания |
| dragX | int | X-позиция при drag |
| dragY | int | Y-позиция при drag |
| offsetX | int | Смещение X |
| offsetY | int | Смещение Y |
| dimsSet | bool | Дименшены вычислены |
| totalW | int | Общая ширина экрана |
| totalH | int | Общая высота экрана |

#### ModalButton

Описание кнопки в модалке.

| Поле | Тип | Описание |
| :---- | :---- | :---- |
| Label | string | Текст кнопки |
| Action | func() | Action при клике |

### Функции

#### NewModal

Создаёт новый Modal из его частей.

```go
func NewModal(title, content string, buttons []ModalButton, onClose func()) *Modal
```

**Параметры:**

- `title` — заголовок модалки
- `content` — рендеримый ANSI-контент
- `buttons` — массив кнопок
- `onClose` — callback при закрытии

**Возвращает:**

- `*Modal` — новый экземпляр модалки

#### EnsureDimensions

Вычисляет и хранит дименшены рамки, если они ещё не установлены.

```go
func (m *Modal) EnsureDimensions(totalW, totalH int)
```

**Параметры:**

- `totalW` — общая ширина экрана
- `totalH` — общая высота экрана

**Поведение:**

- Вызывается перед HandleMouse, если Overlay ещё не был вызван
- Вычисляет boxWidth: 3/5 от totalW, clamp 30-50
- boxHeight фиксирован: 7 линий (RoundedBorder + Padding(1,2) + 3 content lines)
- Центрирует модалку с учётом offsetX/offsetY
- Устанавливает dimsSet = true после вычисления

#### Overlay

Рендерит модалку поверх существующих строк контента.

```go
func (m *Modal) Overlay(lines []string, totalW, totalH int) []string
```

**Параметры:**

- `lines` — исходные строки контента
- `totalW` — общая ширина экрана
- `totalH` — общая высота экрана

**Возвращает:**

- `[]string` — строки с модалкой наложенной поверх

**Поведение:**

- Если totalW <= 0 или lines пуста — возвращает lines как есть
- Вызывает EnsureDimensions перед рендерингом
- Затемняет фон (lines[i] = dimStyle.Background(gbDark0).Render(stripANSI(lines[i])))
- Рендерит рамку с RoundedBorder и Padding(1,2)
- Внутренняя ширина = boxWidth - 6 (2 бордеры + 4 padding)
- Строит три линии контента:
  - Заголовок + ✕ в конце
  - Контент (с обрезкой если длиннее innerWidth)
  - Кнопки в формате [Label] [Label] ...
- Накладывает рамку на затемнённые строки, сохраняя контент слева/справа

#### HandleMouse

Обработка мыши для модалки.

```go
func (m *Modal) HandleMouse(msg tea.MouseMsg) bool
```

**Параметры:**

- `msg` — событие мыши (tea.MouseMsg)

**Возвращает:**

- `bool` — true если событие было обработано (consumed)

**Поведение:**

- Если boxHeight == 0 — возвращает false
- Поддерживает:
  - Клик по ✕ (закрытие)
  - Клик по кнопкам (вызов Action)
  - Перетаскивание за верхней полоской (startY+1)
- ✕ находится на позиции `startX + boxWidth - 4` в строке заголовка
- Кнопки находятся в строке `startY + 4` (и `startY + 5` если wrap)
- Drag работает только на полоске `startY + 1`, не на заголовке
- При motion обновляет offsetX/offsetY/startX/startY
- При release сбрасывает dragging = false

#### StartX

Возвращает X-позицию модалки.

```go
func (m *Modal) StartX() int
```

**Возвращает:**

- `int` — startX

#### StartY

Возвращает Y-позицию модалки.

```go
func (m *Modal) StartY() int
```

**Возвращает:**

- `int` — startY

#### BoxWidth

Возвращает ширину модалки.

```go
func (m *Modal) BoxWidth() int
```

**Возвращает:**

- `int` — boxWidth

#### BoxHeight

Возвращает высоту модалки.

```go
func (m *Modal) BoxHeight() int
```

**Возвращает:**

- `int` — boxHeight

## Вспомогательные функции

### buildButtonLine

Создаёт строку с кнопками.

```go
func (m *Modal) buildButtonLine() string
```

**Поведение:**

- Форматирует кнопки как `[Label]`
- Объединяет их с разделителем "  "

### findBracketPair

Находит пару скобок в строке.

```go
func findBracketPair(s string, pos int) (int, int)
```

**Параметры:**

- `s` — строка для поиска
- `pos` — позиция для начала поиска

**Возвращает:**

- `start` — позиция открытия скобки
- `end` — позиция закрытия скобки

**Поведение:**

- Ищет "[" начиная с pos
- Ищет "]" после "["
- Возвращает (-1, -1) если пар нет

### visualBytePos

Возвращает byte-позицию в строке, где визуальная ширина достигает targetW.

```go
func visualBytePos(s string, targetW int) int
```

**Параметры:**

- `s` — строка
- `targetW` — целевая визуальная ширина

**Возвращает:**

- `int` — byte-позицию

**Поведение:**

- Использует ansi/parser для обработки ANSI-секвенций
- Использует uniseg для обработки графем
- Если визуальная ширина никогда не достигает targetW — возвращает len(s)

### stripANSI

Удаляет ANSI-секвенции из строки.

```go
func stripANSI(s string) string
```

**Поведение:**

- Удаляет все escape-секвенции через regexp

### max

Возвращает максимум из двух значений.

```go
func max(a, b int) int
```

## Стили

### modalBorderStyle

Стиль рамки модалки.

| Свойство | Значение |
| :---- | :---- |
| Background | gbDark1 |
| Foreground | gbLight1 |
| BorderStyle | RoundedBorder() |
| BorderForeground | gbBlue |
| Padding | (1, 2) |

### dimStyle

Стиль затемнения фона.

| Свойство | Значение |
| :---- | :---- |
| Foreground | gbGray |
| Background | gbDark0 |

## Поведение

### Dragging

- Dragging работает только на полоске заголовка (`startY + 1`)
- Не работает на строке заголовка (`startY + 2`)
- При движении обновляется offsetX/offsetY
- При выходе за границы экрана clamp позиция

### Button Layout

```
box[0]: top border
box[1]: top padding ← draggable strip
box[2]: title + ✕
box[3]: content
box[4]: buttons ← first row
box[5]: bottom pad
box[6]: bottom border
```

### Close Button Position

- ✕ находится на позиции `startX + boxWidth - 4` в строке заголовка
- Это последняя визуальная колонка внутреннего контента

## Ограничения

- `Width <= 0` — автоширина: 3/5 экрана, clamp 30-50
- `Width > totalW` — clamp до totalW
- `totalW <= 0` — Overlay возвращает lines как есть
- `boxHeight` фиксирован: 7 линий
- Dragging clamp не позволяет выйти за границы экрана

## Ключевые Правила

1. **EnsureDimensions** должен быть вызван перед первым `HandleMouse`
1. **Overlay** должен быть вызван перед рендерингом контента
1. **Width** может быть авто (0) или явно заданным (30-50)
1. **OnClose** вызывается только один раз при закрытии
1. **Button Action** вызывается при первом клике в пределах кнопки
1. **Dragging** работает только на верхней полоске, не на заголовке
