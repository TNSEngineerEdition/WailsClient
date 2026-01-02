package passenger

import (
	"sync"

	"github.com/TNSEngineerEdition/WailsClient/pkg/consts"
)

type passengerStop struct {
	stopID     uint64
	passengers map[uint64]*Passenger
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
	ps.passengers[passenger.ID] = passenger
}

func (ps *passengerStop) despawnPassenger(passenger *Passenger) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.passengers, passenger.ID)
}

func (ps *passengerStop) loadPassengersToTram(tramID, time uint) []*Passenger {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	boardingPassengers := make([]*Passenger, 0, consts.MAX_PASSENGERS_CHANGE_RATE)
	for _, p := range ps.passengers {
		if p.TravelPlan.IsConnectionInPlan(ps.stopID, tramID) {
			boardingPassengers = append(boardingPassengers, p)
			p.saveNewTrip(tramID, time, ps.stopID, p.TravelPlan.GetConnectionEnd(tramID))
		}
		if len(boardingPassengers) >= consts.MAX_PASSENGERS_CHANGE_RATE {
			break
		}
	}

	for _, p := range boardingPassengers {
		delete(ps.passengers, p.ID)
	}

	return boardingPassengers
}
