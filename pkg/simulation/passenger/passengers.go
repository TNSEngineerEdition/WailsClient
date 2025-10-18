package passenger

import (
	"math/rand/v2"
	"sync"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
)

type Passenger struct {
	strategy               PassengerStrategy
	spawnTime              uint
	StartStopID, EndStopID uint64
	ID                     uint64
}

type passengersAtStop struct {
	passengers []*Passenger
	mu         sync.Mutex
}

type PassengersStore struct {
	PassengersAtStops map[uint64]*passengersAtStop
	PassengersToSpawn map[uint][]*Passenger
}

func NewPassengersStore(c *city.City) *PassengersStore {
	stopsByID := c.GetStopsByID()

	store := &PassengersStore{
		PassengersAtStops: make(map[uint64]*passengersAtStop, len(stopsByID)),
		PassengersToSpawn: make(map[uint][]*Passenger),
	}

	for id := range stopsByID {
		store.PassengersAtStops[id] = &passengersAtStop{
			passengers: make([]*Passenger, 0),
		}
	}

	store.generatePassengers(c)

	return store
}

func (ps *PassengersStore) generatePassengers(c *city.City) {
	timeBounds := c.GetTimeBounds()
	tramStops := c.GetStops()
	var counter uint64

	for i := range tramStops {
		startStop := tramStops[i]
		for range 10 {
			var j int
			for {
				j = rand.IntN(len(tramStops))
				if j != i {
					break
				}
			}
			endStop := tramStops[j]
			spawn := timeBounds.StartTime + uint(rand.IntN(int(timeBounds.EndTime-timeBounds.StartTime+1)))

			passenger := &Passenger{
				strategy:    PassengerStrategy(rand.IntN(3)),
				spawnTime:   spawn,
				StartStopID: startStop.ID,
				EndStopID:   endStop.ID,
				ID:          counter,
			}

			ps.PassengersToSpawn[spawn] = append(ps.PassengersToSpawn[spawn], passenger)
			counter++
		}
	}
}

func (s *PassengersStore) SpawnAtTime(time uint) {
	passengersToSpawn := s.PassengersToSpawn[time]

	for _, p := range passengersToSpawn {
		stop := s.PassengersAtStops[p.StartStopID]
		stop.mu.Lock()
		stop.passengers = append(stop.passengers, p)
		stop.mu.Unlock()
	}
}

func (s *PassengersStore) UnloadAllToStop(stopID uint64, passengers []*Passenger) {
	stop := s.PassengersAtStops[stopID]
	stop.mu.Lock()
	defer stop.mu.Unlock()
	stop.passengers = append(stop.passengers, passengers...)
}

func (s *PassengersStore) BoardAllFromStop(stopID uint64, alreadyBoardedIDS []uint64) []*Passenger {
	previousTakenSet := make(map[uint64]struct{}, len(alreadyBoardedIDS))
	for _, id := range alreadyBoardedIDS {
		previousTakenSet[id] = struct{}{}
	}

	stop := s.PassengersAtStops[stopID]
	stop.mu.Lock()
	defer stop.mu.Unlock()
	boardingPassengers := make([]*Passenger, 0, len(stop.passengers))
	stayingPassengers := make([]*Passenger, 0, len(stop.passengers))
	for _, passenger := range stop.passengers {
		if _, taken := previousTakenSet[passenger.ID]; taken {
			stayingPassengers = append(stayingPassengers, passenger)
		} else {
			boardingPassengers = append(boardingPassengers, passenger)
		}
	}
	stop.passengers = stayingPassengers
	return boardingPassengers
}
