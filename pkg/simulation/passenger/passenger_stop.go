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

func (ps *passengerStop) boardPassengersToTram(tramID uint) []*Passenger {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	counter := 0
	boardingPassengers := make([]*Passenger, 0, consts.MAX_PASSENGERS_CHANGE_RATE)
	restOfPassengers := make([]*Passenger, 0)

	for _, p := range ps.passengers {
		if p.TravelPlan.CheckIfConnectionIsInPlan(ps.stopID, tramID) {
			boardingPassengers = append(boardingPassengers, p)
			counter++
		} else {
			restOfPassengers = append(restOfPassengers, p)
		}
		if counter == consts.MAX_PASSENGERS_CHANGE_RATE {
			break
		}
	}

	ps.passengers = restOfPassengers

	return boardingPassengers
}
