package tram

import (
	"github.com/TNSEngineerEdition/WailsClient/pkg/consts"
	"github.com/TNSEngineerEdition/WailsClient/pkg/simulation/passenger"
)

func (t *Tram) boardPassengers() bool {
	stopID := t.TripDetails.Trip.Stops[t.TripDetails.Index].ID
	boardedPassengers := t.passengersStore.BoardPassengers(stopID, t.ID)

	for _, p := range boardedPassengers {
		t.passengersInTram[p.ID] = p
	}

	// return true if boarding is finished
	return len(boardedPassengers) < consts.MAX_PASSENGERS_CHANGE_RATE
}

func (t *Tram) disembarkPassengers(time uint) bool {
	stopID := t.TripDetails.Trip.Stops[t.TripDetails.Index].ID
	counter := 0
	disembarkingPassengers := make([]*passenger.Passenger, 0, consts.MAX_PASSENGERS_CHANGE_RATE)

	for _, p := range t.passengersInTram {
		if p.TravelPlan.GetConnectionEnd(t.ID) == stopID {
			disembarkingPassengers = append(disembarkingPassengers, p)
			counter++
		}
		if counter == consts.MAX_PASSENGERS_CHANGE_RATE {
			break
		}
	}

	for _, p := range disembarkingPassengers {
		delete(t.passengersInTram, p.ID)
	}

	t.passengersStore.DisembarkPassengers(disembarkingPassengers, stopID, time)
	isDisembarkingFinished := len(disembarkingPassengers) < consts.MAX_PASSENGERS_CHANGE_RATE

	return isDisembarkingFinished
}
