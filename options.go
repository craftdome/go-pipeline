package pipeline

type Option[I, O any] func(p *Unit[I, O])

func WithWorkers[I, O any](n int) Option[I, O] {
	return func(u *Unit[I, O]) {
		u.workers = n
	}
}
