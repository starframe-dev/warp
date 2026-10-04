---
title: Getting Started
description: Install Warp, create a minimal program, and run the demo.
---

# Getting Started

Install Warp with `go get`:

```sh
go get github.com/starframe-dev/warp
```

Create a minimal Bubbletea program with Warp:

```go
package main

import "github.com/starframe-dev/warp"

func main() {
	w := warp.New()
	w.Run()
}
```

`warp.New()` creates a Warp instance with a `TabGroup` root panel.

Warp implements the Bubbletea `Model` contract, so you can embed it in a Bubbletea app or let Warp run its own program.

`Run()` starts the Bubbletea program. `Warp` also implements `tea.Model`, so you can pass it to `tea.NewProgram` when your application needs to manage Bubbletea itself. Use `SetRoot(panel)` to install a custom root panel.

Run the project demo from the repository root:

```sh
go run ./cmd/demo/
```
