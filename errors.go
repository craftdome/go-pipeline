package pipeline

import "errors"

var (
	ErrExecuteMethodNotSet = errors.New("OnExecute() method must be set before Start() invokes")
	ErrUnitDoneAlready     = errors.New("you cannot start unit after stopping")
	ErrUnitStartedAlready  = errors.New("unit already started")
	ErrUnitNotStarted      = errors.New("unit not started")
	ErrUnitStoppedAlready  = errors.New("unit already stopped")
)
