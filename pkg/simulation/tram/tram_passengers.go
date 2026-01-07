package tram

import (
	"github.com/TNSEngineerEdition/WailsClient/pkg/simulation/passenger"
)

func (t *Tram) GetPassengerCount() uint {
	return uint(len(t.passengersInTram))
}

func (t *Tram) loadPassengers(time uint) bool {
	stopID := t.TripDetails.Trip.Stops[t.TripDetails.Index].ID
	boardedPassengers := t.passengersStore.LoadPassengers(stopID, t.ID, time)

	for _, p := range boardedPassengers {
		t.passengersInTram[p.ID] = p
	}

	// return true if loading is finished
	return len(boardedPassengers) < passenger.MAX_PASSENGERS_CHANGE_RATE
}

func (t *Tram) unloadPassengers(time uint) bool {
	stopID := t.TripDetails.Trip.Stops[t.TripDetails.Index].ID
	disembarkingPassengers := make([]*passenger.Passenger, 0, passenger.MAX_PASSENGERS_CHANGE_RATE)

	for _, p := range t.passengersInTram {
		if p.TravelPlan.GetConnectionDestination(t.ID) == stopID {
			disembarkingPassengers = append(disembarkingPassengers, p)
		}
		if len(disembarkingPassengers) == passenger.MAX_PASSENGERS_CHANGE_RATE {
			break
		}
	}

	for _, p := range disembarkingPassengers {
		delete(t.passengersInTram, p.ID)
	}

	t.passengersStore.UnloadPassengers(disembarkingPassengers, stopID, time)
	isUnloadingFinished := len(disembarkingPassengers) < passenger.MAX_PASSENGERS_CHANGE_RATE

	return isUnloadingFinished
}
