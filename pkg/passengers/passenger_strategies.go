package passengers

type PassengerStrategy uint8

const (
	ASAP PassengerStrategy = iota
	COMFORT
	SURE
)
