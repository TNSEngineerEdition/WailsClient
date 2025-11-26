package passenger

import (
	"math/rand/v2"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
)

type PassengersStore struct {
	passengerStops    map[uint64]*passengerStop
	passengersToSpawn map[uint][]*Passenger
}

func NewPassengersStore(c *city.City) *PassengersStore {
	stopsByID := c.GetStopsByID()

	store := &PassengersStore{
		passengerStops:    make(map[uint64]*passengerStop, len(stopsByID)),
		passengersToSpawn: make(map[uint][]*Passenger),
	}

	for id := range stopsByID {
		store.passengerStops[id] = &passengerStop{
			stopID:     id,
			passengers: make([]*Passenger, 0),
		}
	}

	store.generatePassengers(c)

	return store
}

func (ps *PassengersStore) GetPassengerCountAtStop(stopID uint64) uint {
	return ps.passengerStops[stopID].GetPassengerCount()
}

func (ps *PassengersStore) generatePassengers(c *city.City) {
	timeBounds := c.GetTimeBounds()
	stopsByID := c.GetStopsByID()
	var counter uint64

	for startStopID := range stopsByID {
		for range 50 {
			// TODO: time's upper bound is set to 18000 (5:00:00) for presentation purposes
			//spawnTime := timeBounds.StartTime + uint(rand.IntN(int(timeBounds.EndTime-timeBounds.StartTime+1)))
			spawnTime := timeBounds.StartTime + uint(rand.IntN(int(18000-timeBounds.StartTime+1)))

			tp := GetTravelPlan(startStopID, spawnTime, c)

			if startStopID == tp.endStopID {
				continue // no trips found
			}

			passenger := &Passenger{
				ID:          counter,
				strategy:    PassengerStrategy(rand.IntN(3)),
				spawnTime:   spawnTime,
				startStopID: startStopID,
				endStopID:   tp.endStopID,
				TravelPlan:  tp,
			}

			ps.passengersToSpawn[spawnTime] = append(ps.passengersToSpawn[spawnTime], passenger)
			counter++
		}
	}
}

func (ps *PassengersStore) SpawnPassengersAtTime(time uint) {
	passengersToSpawn := ps.passengersToSpawn[time]
	for _, p := range passengersToSpawn {
		stop := ps.passengerStops[p.startStopID]
		stop.addPassengerToStop(p)
	}
}

func (ps *PassengersStore) BoardPassengers(stopID uint64, tramID uint) []*Passenger {
	passengerStop := ps.passengerStops[stopID]
	return passengerStop.boardPassengersToTram(tramID)
}

func (ps *PassengersStore) UnloadAllToStop(stopID uint64, passengers []*Passenger) {
	stop := ps.passengerStops[stopID]
	for _, p := range passengers {
		stop.addPassengerToStop(p)
	}
}
