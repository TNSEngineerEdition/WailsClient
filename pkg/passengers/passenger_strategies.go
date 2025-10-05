package passengers

type PassangerStrategy uint8

const (
	ASAP PassangerStrategy = iota
	COMFORT
	SURE
)
