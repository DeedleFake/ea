Ea
==

[![Go Reference](https://pkg.go.dev/badge/deedles.dev/ea.svg)](https://pkg.go.dev/deedles.dev/ea)
[![Go Report Card](https://goreportcard.com/badge/deedles.dev/ea)](https://goreportcard.com/report/deedles.dev/ea)

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

type model struct {
	m Model
}

func (m model) Update(msg ea.Msg) (model, ea.Cmd) {
	next, cmd := m.m.Update(msg)
	m.m = next
	fmt.Println(m.m.View())
	return m, cmd
}

type Program struct {
	loop *ea.Loop
}

func New(initialModel Model) *Program {
	return &Program{
		loop: ea.New(initialModel),
	}
}

func (p *Program) view(model Model) {
	fmt.Println(model.View())
}

func (p *Program) Run() {
	p.loop.Run(context.Background())
}
```
