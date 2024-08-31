package pipeline

import (
	"fmt"
	"sync/atomic"
)

type State struct {
	Total  atomic.Uint32
	Done   atomic.Uint32
	Errors atomic.Uint32
	_      [52]byte
}

func (s *State) String() string {
	return fmt.Sprintf("Total: %d, Done: %d, Errors: %d", s.Total.Load(), s.Done.Load(), s.Errors.Load())
}
