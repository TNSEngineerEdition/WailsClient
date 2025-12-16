package travelplan

type PassengerStrategy string

const (
	RANDOM  PassengerStrategy = "RANDOM"
	ASAP    PassengerStrategy = "ASAP"
	COMFORT PassengerStrategy = "COMFORT"
	SURE    PassengerStrategy = "SURE"
)
