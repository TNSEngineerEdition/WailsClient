package travelplan

type PassengerStrategy uint8

const (
	RANDOM PassengerStrategy = iota
	ASAP
	COMFORT
	SURE
)
