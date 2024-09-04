package pipeline

import (
	"context"
	"runtime"
	"sync"
	"time"
)

type ErrorData[I any] struct {
	Err  error
	Data I
}

func (e ErrorData[I]) Error() string {
	return e.Err.Error()
}

type NextUnit[I any] interface {
	Input() chan<- I
}

type Unit[I, O any] struct {
	workers int

	sharedStates []State

	OnExecute func(I) (O, error)

	StartedAt time.Time
	EndedAt   time.Time

	inputChannel  chan I
	outputChannel chan O
	errorsChannel chan ErrorData[I]
	doneChannel   chan struct{}

	nextUnit NextUnit[O]
}

func NewUnit[I, O any](opts ...Option[I, O]) *Unit[I, O] {
	u := &Unit[I, O]{
		workers:      runtime.NumCPU(),
		inputChannel: make(chan I),
		doneChannel:  make(chan struct{}, 1),
	}

	for _, opt := range opts {
		opt(u)
	}

	u.sharedStates = make([]State, u.workers)
	u.outputChannel = make(chan O, u.workers)
	u.errorsChannel = make(chan ErrorData[I], u.workers)

	return u
}

// Input Channel for receiving input data.
func (u *Unit[I, O]) Input() chan<- I {
	return u.inputChannel
}

// Output Channel for receiving output data.
// Throw panic if the unit has next unit set.
func (u *Unit[I, O]) Output() <-chan O {
	if u.HasNextUnit() {
		panic("Next unit has been defined. You cannot use this output.")
	}
	return u.outputChannel
}

func (u *Unit[I, O]) Errors() <-chan ErrorData[I] {
	return u.errorsChannel
}

func (u *Unit[I, O]) Start() error {
	if u.OnExecute == nil {
		return ErrExecuteMethodNotSet
	}
	if !u.EndedAt.IsZero() {
		return ErrUnitDoneAlready
	}
	if !u.StartedAt.IsZero() {
		return ErrUnitStartedAlready
	}

	u.StartedAt = time.Now()

	var wg sync.WaitGroup
	wg.Add(u.workers)
	for i := 0; i < u.workers; i++ {
		go func() {
			defer wg.Done()

			for task := range u.inputChannel {
				u.execute(i, task)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(u.outputChannel)
		close(u.errorsChannel)
		u.doneChannel <- struct{}{}
		u.EndedAt = time.Now()
	}()

	return nil
}

func (u *Unit[I, O]) Stop(ctx context.Context) error {
	if u.StartedAt.IsZero() {
		return ErrUnitNotStarted
	}
	if !u.EndedAt.IsZero() {
		return ErrUnitStoppedAlready
	}

	close(u.inputChannel)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-u.doneChannel:
		return nil
	}
}

func (u *Unit[I, O]) GetState() *State {
	state := &State{}
	for i := 0; i < len(u.sharedStates); i++ {
		state.Total.Add(u.sharedStates[i].Total.Load())
		state.Done.Add(u.sharedStates[i].Done.Load())
		state.Errors.Add(u.sharedStates[i].Errors.Load())
	}

	return state
}

func (u *Unit[I, O]) execute(worker int, i I) {
	u.sharedStates[worker].Total.Add(1)

	if o, err := u.OnExecute(i); err != nil {
		u.sharedStates[worker].Errors.Add(1)
		u.errorsChannel <- ErrorData[I]{
			Err:  err,
			Data: i,
		}
	} else {
		u.sharedStates[worker].Done.Add(1)

		if u.HasNextUnit() {
			u.nextUnit.Input() <- o
		} else {
			u.outputChannel <- o
		}
	}
}

// SetNextUnit
// Deletes the output channel and redirects a result to the next unit input
func (u *Unit[I, O]) SetNextUnit(unit NextUnit[O]) {
	u.nextUnit = unit
	u.outputChannel = nil
}

func (u *Unit[I, O]) HasNextUnit() bool {
	return u.nextUnit != nil
}
