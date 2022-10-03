package ea

import "context"

type Model[N any] interface {
	Update(Msg) (N, Cmd)
}

type Cmd func() Msg

type Msg any

type Loop[M Model[M]] struct {
	m M
}

func New[M Model[M]](model M) *Loop[M] {
	return &Loop[M]{
		m: model,
	}
}

func (loop *Loop[M]) coord(ctx context.Context) {
	panic("Not implemented.")
}

func (loop *Loop[M]) Run(ctx context.Context, cmd Cmd) Model[M] {
	panic("Not implemented.")
}

func (loop *Loop[M]) Send() chan<- Msg {
	panic("Not implemented.")
}
