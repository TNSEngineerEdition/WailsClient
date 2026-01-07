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
}

func (t *tripDetails) getDelay(time uint) uint {
	if t.Index == 0 {
		if time < t.Trip.Stops[0].Time {
			return 0
		}
		return time - t.Trip.Stops[0].Time
	}

	departureDelay := t.Departures[t.Index-1] - t.Trip.Stops[t.Index-1].Time
	if time < t.Trip.Stops[t.Index].Time {
		return departureDelay
	}
	nextStopDelay := time - t.Trip.Stops[t.Index].Time

	return max(departureDelay, nextStopDelay)
}
