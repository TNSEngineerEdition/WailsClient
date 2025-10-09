package tram

type TramState uint8

const (
	StateTripNotStarted TramState = iota
	StatePassengerLoading
	StateTravelling
	StatePassengerUnloading
	StateTripFinished
	StateStopping
	StateStopped
)
