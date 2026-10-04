# Panel ownership

`ownership.go` implements internal ownership tracking for panels in the `warp` package. It does not declare exported types or functions, so it adds no direct public API. Its helpers coordinate panel-tree attachment, detect panels removed from a tree, and ensure removed panels are unmounted at most once per removal pass.

## Ownership state

### `panelOwnership`

`panelOwnership` stores a root `Panel` and protects access to that root with an `sync.RWMutex`.

- `newPanelOwnership(root)` creates an ownership record with the supplied root.
- `setRoot(root)` replaces the root under a write lock. A nil receiver is ignored.
- `rootPanel()` returns the root under a read lock; a nil receiver returns nil.

The mutex protects the root reference itself. It does not make the panels or their tree contents generally thread-safe.

### `panelInstanceKey`

`panelInstanceKey` identifies a panel for retention and duplicate detection. Pointer panels are keyed by their dynamic `reflect.Type` and pointer address. Comparable non-pointer panels are keyed by type and value. Nil-like panels and non-comparable non-pointer values cannot produce a key.

`panelInstance` uses `isNilPanel` to reject nil panels (including typed nils as recognized by that helper). Type is part of the key, preventing values of different dynamic types from colliding solely because of an address or equal value.

## Tree traversal

`collectPanelInstances` walks a panel and the panel structures nested within it. `collectNodePanelInstances` starts the traversal from a layout `Node`. The collector recognizes these built-in panel relationships:

| Panel | Nested panels visited |
| --- | --- |
| `*TabGroup` | Its tabs |
| `*Tab` | Panels in its root node and floating panels |
| `*Scrollable`, `*Selectable`, `*Collapsible` | `Content` |
| `*warpPanel` | The root returned by its associated warp, if present |

Node traversal visits its panel, both children of a split, and each non-nil flex item's node. Nil nodes, panels, and flex items are skipped. A visited set prevents repeated traversal of the built-in pointer panel types, including cycles. Other panel implementations are appended as leaves; their internal contents are not introspected.

## Attaching and ensuring ownership

`attachPanelOwnership(root, ownership)` traverses the supplied panel tree and associates ownership where supported:

- Tab groups and tabs receive the ownership pointer.
- Tabs found in a group have their `parent` set to that group.
- A `warpPanel` with a non-nil warp receives the ownership under the warp's mutex and has `ownsDomainRoot` set to `false`.

A nil ownership pointer makes attachment a no-op.

`(*Tab).ensureOwnership` and `(*TabGroup).ensureOwnership` lazily establish ownership and return it. A tab inherits ownership from its parent group when possible; otherwise it creates ownership rooted at itself. A group creates ownership rooted at itself. Both attach ownership through their current panel tree after initialization. Nil receivers return nil.

`collectTabPanels` traverses a tab's contents, and `tabContainsPanel` compares instance keys to determine whether a target occurs in that tree. `removeSliceAt` is a generic slice helper: it returns the original slice for an out-of-range index; otherwise it shifts later elements left, clears the old final slot to the element type's zero value, and returns the shortened slice.

## Removing panels and unmounting

`(*panelOwnership).unmountRemoved(candidates)` compares candidate panels with the panels still reachable from the ownership root at call time:

1. It collects current root panels and records each panel with a valid instance key as retained.
2. It skips candidates that are nil-like, have no usable key, or repeat a key already considered in this call.
3. It leaves candidates that remain in the root tree untouched.
4. For each unique candidate no longer retained, it calls `Unmount()` if the panel implements the public `Unmounter` interface, then detaches this ownership from supported built-in panels.

A nil receiver or an empty candidate list is a no-op. Detachment only clears an ownership pointer when it still equals the ownership being removed, so it does not erase a newer association. For a detached `warpPanel`, the warp mutex guards the update; clearing the matching ownership also sets `ownsDomainRoot` to `true`. Tab and tab-group ownership pointers are cleared on a matching association.

Unmounting and detachment are performed once per unique candidate key in a call, not globally: later calls can process the same panel again if it is supplied as a removed candidate. Candidates without an instance key cannot be tracked and therefore are not unmounted or detached by this method.

## Public API note

All declarations in this file are unexported implementation details. The only public contract it relies on directly is `Panel` and the optional `Unmounter` interface; ownership setup and cleanup are driven by other package operations that manipulate panel trees.
