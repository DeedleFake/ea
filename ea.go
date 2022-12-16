package ea

import (
	"context"
	"sync"

	"deedles.dev/xsync/cq"
)

// Model is a state that is capable of producing a new state from
// itself and a Msg.
type Model[N any] interface {
	Update(Msg) (N, Cmd)
}

// A Cmd is a function that is run concurrently. The Msg returned is
// passed to Model.Update to produce a new Model.
type Cmd func() Msg

// A Msg is an arbitrary piece of data passed to a Model's Update
// method to produce a new Model.
type Msg any

// quitMsg is an internal Msg type used to signal to the main loop
// that it's time to exit. It is returned by Quit.
type quitMsg struct{}

// Quit is a command that causes the main loop to exit.
func Quit() Msg {
	return quitMsg{}
}

// batchMsg is an internal Msg type used to run multiple commands from
// a single loop.
type batchMsg []Cmd

// Batch returns a Cmd that runs multiple other Cmds.
func Batch(cmds ...Cmd) Cmd {
	b := make(batchMsg, 0, len(cmds))
	for _, cmd := range cmds {
		if cmd != nil {
			b = append(b, cmd)
		}
	}
	if len(b) == 0 {
		return nil
	}

	return func() Msg {
		return b
	}
}

// Loop runs a update loop.
type Loop[M Model[M]] struct {
	done  chan struct{}
	close sync.Once
	msgs  *cq.BulkQueue[Msg, []Msg]
	m     M
}

// New returns a Loop with an initial Model.
func New[M Model[M]](model M) *Loop[M] {
	loop := &Loop[M]{
		m:    model,
		done: make(chan struct{}),
		msgs: cq.Simple[Msg](),
	}

	return loop
}

// do runs a Cmd, sending the result back into the Loop.
func (loop *Loop[M]) do(ctx context.Context, cmd Cmd) {
	loop.Enqueue(cmd())
}

// Run runs the Loop with an optional initial command. It blocks until
// the loop exits, returning the final Model.
//
// Behavior is undefined if Run is called more than once.
func (loop *Loop[M]) Run(ctx context.Context, cmd Cmd) M {
	defer loop.close.Do(func() { close(loop.done) })

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if cmd != nil {
		go loop.do(ctx, cmd)
	}

	for {
		select {
		case <-ctx.Done():
			return loop.m

		case msgs, ok := <-loop.msgs.Get():
			if !ok {
				return loop.m
			}

			for _, msg := range msgs {
				switch msg := msg.(type) {
				case quitMsg:
					return loop.m

				case batchMsg:
					for _, cmd := range msg {
						go loop.do(ctx, cmd)
					}

				default:
					m, cmd := loop.m.Update(msg)
					loop.m = m
					if cmd != nil {
						go loop.do(ctx, cmd)
					}
				}
			}
		}
	}
}

// Enqueue adds msg to the Loop's internal queue of messages to be
// handled.
func (loop *Loop[M]) Enqueue(msg Msg) {
	select {
	case <-loop.done:
	case loop.msgs.Add() <- msg:
	}
}
