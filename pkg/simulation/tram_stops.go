package simulation

import (
	"github.com/TNSEngineerEdition/WailsClient/pkg/passengers"
)

type tramStops struct {
	stops map[uint64][]*passengers.Passenger
}

func newTramStops() *tramStops {
	return &tramStops{
		stops: make(map[uint64][]*passengers.Passenger),
	}
}
