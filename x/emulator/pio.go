package emulator

// A PIO is a PIO peripheral.
type PIO int

// A StateMachine is a PIO state machine.
type StateMachine struct {
	PIO   PIO
	Index int
}

// StateMachineClaimer wraps the method that implements [PIO.ClaimStateMachine].
type StateMachineClaimer interface {
	// ClaimStateMachine claims an unused StateMachine.
	ClaimStateMachine(p PIO) (StateMachine, error)
}

// StateMachineUnclaimer wraps the method that implements [StateMachine.Unclaim].
type StateMachineUnclaimer interface {
	// Unclaim releases the state machine for use by other code.
	UnclaimStateMachine(sm StateMachine)
}

// Specify whether [PIO.ClaimStateMachine] will return an error unless emulated.
var MustClaimStateMachine = false

// ClaimStateMachine claims an unused StateMachine.
func (p PIO) ClaimStateMachine() (StateMachine, error) {
	e, ok := emulator.(StateMachineClaimer)
	switch {
	case ok:
		return e.ClaimStateMachine(p)
	case MustClaimStateMachine:
		return StateMachine{}, NotEmulatedError("PIO.ClaimStateMachine")
	default:
		return StateMachine{PIO: p}, nil
	}
}

// Specify whether [StateMachine.Unclaim] will panic unless emulated.
var MustUnclaimStateMachine = false

// Unclaim releases the state machine for use by other code.
func (sm StateMachine) Unclaim() {
	e, ok := emulator.(StateMachineUnclaimer)
	switch {
	case ok:
		e.UnclaimStateMachine(sm)
	case MustUnclaimStateMachine:
		panic(NotEmulatedError("StateMachine.Unclaim"))
	}
}
