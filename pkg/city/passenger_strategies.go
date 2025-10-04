package city

type PassangerStrategy uint8

const (
	ASAP PassangerStrategy = iota
	COMFORT
	SURE
)
