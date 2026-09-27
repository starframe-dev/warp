# Element

Файл `element.go` определяет представление семантических UI-элементов и
вспомогательные функции для их экспозиции, сбора и поиска внутри панели.
Элемент — это узел дерева: он несёт *роль*, необязательное *имя* и
необязательное *действие*, экранные *границы* (прямоугольник в
координатах ячеек) и список дочерних элементов.

## Типы

### `Element`

``` go
type Element struct {
    Role     string    `json:"role"`
    Name     string    `json:"name"`
    Action   string    `json:"action,omitempty"`
    Bounds   Bounds    `json:"bounds"`
    Children []Element `json:"children,omitempty"`
}
```

Примечания по JSON-кодированию:

- `Role` и `Name` всегда присутствуют.
- `Action` опускается, если пусто.
- `Children` опускается, если срез пуст (используется форма указателя в
  теге `json`).

### `Bounds`

``` go
type Bounds struct {
    X int `json:"x"`
    Y int `json:"y"`
    W int `json:"w"`
    H int `json:"h"`
}
```

Определяет прямоугольную область экрана в координатах ячеек: левый
верхний угол в точке `(X, Y)`, ширина `W`, высота `H`.

### `Bounds.Center`

``` go
func (b Bounds) Center() (int, int)
```

Возвращает центральную ячейку как целочисленные координаты с помощью
целочисленного деления: `(X + W/2, Y + H/2)`.

### `ElementProvider`

``` go
type ElementProvider interface {
    Elements(width, height int) []Element
}
```

Реализуется панелями, которые могут экспонировать свои UI-элементы.
Метод принимает текущие размеры панели в ячейках и возвращает корневой
список элементов (дочерние вложены в каждый `Element`).

### `ElementProviderFunc`

``` go
type ElementProviderFunc func(width, height int) []Element

func (f ElementProviderFunc) Elements(width, height int) []Element
```

Адаптирует обычную функцию к интерфейсу `ElementProvider`. Полезно для
одноразовых провайдеров, которым не стоит завязывать структуру.

## Функции

### `collectElements`

``` go
func collectElements(panel Panel, width, height int) []Element
```

Внутренний помощник пакета. Возвращает `nil`, если `panel` nil. Иначе
приводит панель к типу `ElementProvider` и вызывает
`Elements(width, height)`; если приведение не удалось, возвращается
`nil`. Используется warp-конвейером для сбора элементов из панели.

### `FindElement`

``` go
func FindElement(elems []Element, role, name, action string) (Element, bool)
```

Рекурсивно ищет в списке элементов первый элемент, чьи роль, имя и
действие совпадают. Пустые поисковые строки трактуются как подстановки
(совпадение с чем угодно). Возвращает найденный `Element` и `true`, либо
нулевое значение и `false`.

``` go
if el, ok := FindElement(panel.Elems, "", "submit", ""); ok {
    x, y := el.Bounds.Center()
    // press at (x, y)
}
```

## Примечания о поведении

- Дерево *упорядочено*: `FindElement` возвращает *первое* совпадение при
  обходе «вглубь и по обходе (pre-order)», поэтому на вызывающем лежит
  ответственность за уточнение по роли/имени/действию.
- Семантический обход ограничен 128 уровнями и 100 000 узлами. При cloning snapshot и `FindElement` более глубокий или крупный вывод provider'а обрезается, чтобы ошибочное дерево не исчерпало стек или память.
- Перед clipping/translation Warp копирует provider-owned semantic data и не изменяет срез, который вернул `ElementProvider`.
- `collectElements` одинаково обрабатывает отсутствие provider'а и nil-панель (возвращает `nil`).
- `Bounds` задаются в терминальных **ячейках**, а не пикселях; layout и inspector Warp не выполняют преобразование в пиксели.


### `SemanticStableMsg`

```go
type SemanticStableMsg interface {
    SemanticStateUnchanged() bool
}
```

Высокочастотное сообщение может реализовать этот необязательный интерфейс и
возвращать `true`, если его обработка гарантированно не меняет semantic
element tree. При включённом HTTP inspector Warp сохраняет предыдущий
immutable snapshot после `Update` и пересобирает semantic elements на
следующем `View`.

Это предназначено для streaming/background-сообщений вроде PTY output, когда
semantic bounds, roles, names и actions остаются неизменными. Значение `true`
является явным обещанием корректности со стороны producer. Для остальных
сообщений сохраняется консервативная пересборка snapshot после `Update`.
