package passenger

import (
	"sync"
)

type passengerStop struct {
	passengers []*Passenger
	mu         sync.Mutex
}

func (ps *passengerStop) AddPassengerToStop(passenger *Passenger) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.passengers = append(ps.passengers, passenger)
}

func (ps *passengerStop) TakeAllFromStop(alreadyTakenSet map[uint64]struct{}) []*Passenger {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	boardingPassengers := make([]*Passenger, 0, len(ps.passengers))
	stayingPassengers := make([]*Passenger, 0, len(ps.passengers))
	for _, passenger := range ps.passengers {
		if _, taken := alreadyTakenSet[passenger.ID]; taken {
			stayingPassengers = append(stayingPassengers, passenger)
		} else {
			boardingPassengers = append(boardingPassengers, passenger)
		}
	}
	ps.passengers = stayingPassengers
	return boardingPassengers
}
