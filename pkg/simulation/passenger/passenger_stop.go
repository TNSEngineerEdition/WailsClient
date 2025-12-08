package passenger

import (
	"sync"

	"github.com/TNSEngineerEdition/WailsClient/pkg/consts"
)

type passengerStop struct {
	stopID     uint64
	passengers []*Passenger
	mu         sync.Mutex
}

func (ps *passengerStop) GetPassengerCount() uint {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return uint(len(ps.passengers))
}

func (ps *passengerStop) addPassengerToStop(passenger *Passenger) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.passengers = append(ps.passengers, passenger)
}

func (ps *passengerStop) despawnPassenger(passenger *Passenger) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	for i, p := range ps.passengers {
		if p.ID == passenger.ID {
			ps.passengers = append(ps.passengers[:i], ps.passengers[i+1:]...)
			return
		}
	}
}

func (ps *passengerStop) boardPassengersToTram(tramID uint) []*Passenger {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	counter := 0
	boardingPassengers := make([]*Passenger, 0, consts.MAX_PASSENGERS_CHANGE_RATE)
	restOfPassengers := make([]*Passenger, 0)

	for _, p := range ps.passengers {
		if p.TravelPlan.isConnectionInPlan(ps.stopID, tramID) && counter < consts.MAX_PASSENGERS_CHANGE_RATE {
			boardingPassengers = append(boardingPassengers, p)
			counter++
		} else {
			restOfPassengers = append(restOfPassengers, p)
		}
	}

	ps.passengers = restOfPassengers

	return boardingPassengers
}
