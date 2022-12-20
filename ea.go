package ea

import (
	"context"

	"deedles.dev/xsync"
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
	stop  xsync.Stopper
	queue xsync.Queue[func()]
	m     M
}

func New[M Model[M]](initialModel M) *Loop[M] {
	return &Loop[M]{
		m: initialModel,
	}
}

// Model returns the current model that will be used to perform the
// next update. If the loop has stopped, it will return the final
// model.
func (loop *Loop[M]) Model() M {
	return loop.m
}

// Stop stops the loop. This should be called when the loop is no
// longer going to be used.
func (loop *Loop[M]) Stop() {
	loop.stop.Stop()
	loop.queue.Stop()
}

// do runs a Cmd, sending the result back into the Loop.
func (loop *Loop[M]) do(cmd Cmd) {
	loop.Enqueue(cmd())
}

// Updates yields functions that perform successive updates to the
// model. Calling these functions in a different order than how they
// are received will result in undefined behavior.
//
// This channel will be closed when the loop is stopped.
func (loop *Loop[M]) Updates() <-chan func() {
	return loop.queue.Get()
}

func (loop *Loop[M]) update(msg Msg) func() {
	return func() {
		switch msg := msg.(type) {
		case quitMsg:
			loop.Stop()

		case batchMsg:
			for _, cmd := range msg {
				go loop.do(cmd)
			}

		default:
			m, cmd := loop.m.Update(msg)
			loop.m = m
			if cmd != nil {
				go loop.do(cmd)
			}
		}
	}
}

// Enqueue adds msg to the Loop's internal queue of messages to be
// handled.
func (loop *Loop[M]) Enqueue(msg Msg) {
	select {
	case <-loop.stop.Done():
	case loop.queue.Add() <- loop.update(msg):
	}
}

// Run runs updates for loop until the loop is stopped or the context
// is canceled. It returns the final model produced by the loop.
func Run[M Model[M]](ctx context.Context, loop *Loop[M]) M {
	defer loop.Stop()

	for {
		select {
		case <-ctx.Done():
			return loop.Model()

		case update, ok := <-loop.Updates():
			if !ok {
				return loop.Model()
			}

			update()
		}
	}
}
