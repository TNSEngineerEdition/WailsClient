package tram

import (
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/trip"
)

type tripDetails struct {
	Trip                 *trip.TramTrip
	Index                int
	Arrivals, Departures []uint
}

func newTripDetails(trip *trip.TramTrip) tripDetails {
	return tripDetails{
		Trip:       trip,
		Arrivals:   make([]uint, len(trip.Stops)),
		Departures: make([]uint, len(trip.Stops)),
	}
}

func (t *tripDetails) saveArrival(time uint) {
	t.Arrivals[t.Index] = time
}

func (t *tripDetails) saveDeparture(time uint) {
	t.Departures[t.Index] = time
	t.Index += 1
}
