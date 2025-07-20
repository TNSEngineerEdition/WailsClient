package simulation

type TramState uint8

const (
	StateTripNotStarted TramState = iota
	StatePassengerTransfer
	StateTravelling
	StateTripFinished
)
