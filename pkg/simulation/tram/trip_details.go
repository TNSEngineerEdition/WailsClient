package tram

import (
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/trip"
)

type tripDetails struct {
	Trip                 *trip.TramTrip
	Index                int
	arrivals, departures []uint
}

func newTripDetails(trip *trip.TramTrip) tripDetails {
	return tripDetails{
		Trip:       trip,
		arrivals:   make([]uint, len(trip.Stops)),
		departures: make([]uint, len(trip.Stops)),
	}
}

func (t *tripDetails) saveArrival(time uint) {
	t.arrivals[t.Index] = time
}

func (t *tripDetails) saveDeparture(time uint) {
	t.departures[t.Index] = time
	t.Index += 1
}
