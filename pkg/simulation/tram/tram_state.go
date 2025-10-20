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

var TramStates = []struct {
	Value  TramState
	TSName string
}{
	{StateTripNotStarted, "TRIP_NOT_STARTED"},
	{StatePassengerLoading, "PASSENGER_LOADING"},
	{StateTravelling, "TRAVELLING"},
	{StatePassengerUnloading, "PASSENGER_UNLOADING"},
	{StateTripFinished, "TRIP_FINISHED"},
	{StateStopping, "STOPPING"},
	{StateStopped, "STOPPED"},
}
