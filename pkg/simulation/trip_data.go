package simulation

import "github.com/TNSEngineerEdition/WailsClient/pkg/city"

type tripData struct {
	trip                 *city.TramTrip
	index                int
	arrivals, departures []uint
}

func newTripData(trip *city.TramTrip) tripData {
	return tripData{
		trip:       trip,
		arrivals:   make([]uint, len(trip.Stops)),
		departures: make([]uint, len(trip.Stops)),
	}
}

func (t *tripData) saveArrival(time uint) {
	t.arrivals[t.index] = time
}

func (t *tripData) saveDeparture(time uint) {
	t.departures[t.index] = time
	t.index += 1
}
