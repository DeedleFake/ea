Ea
==

Ea is a [Bubble Tea](https://github.com/charmbracelet/bubbletea)-inspired Go implementation of the Elm Architecture. It is designed primarily for use by other libraries that want to expose an API based on the Elm Architecture.

Example
-------

```go
package boba

import (
	"context"
	"fmt"

	"deedles.dev/ea"
)

type Model interface {
	ea.Model[Model]
	View() string
}

type Program struct {
	loop *ea.Loop
}

func New(initialModel Model) *Program {
	p := &Program{
		loop: ea.New(initialModel),
	}
	p.loop.PostUpdate = p.view

	return p
}

func (p *Program) view(model Model) {
	fmt.Println(model.View())
}

func (p *Program) Run() {
	p.loop.Run(context.Background())
}
```
