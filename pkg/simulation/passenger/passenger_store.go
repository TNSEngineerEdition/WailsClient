package passenger

import (
	"math/rand/v2"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
)

type PassengersStore struct {
	PassengersAtStops map[uint64]*passengerStop
	PassengersToSpawn map[uint][]*Passenger
}

func NewPassengersStore(c *city.City) *PassengersStore {
	stopsByID := c.GetStopsByID()

	store := &PassengersStore{
		PassengersAtStops: make(map[uint64]*passengerStop, len(stopsByID)),
		PassengersToSpawn: make(map[uint][]*Passenger),
	}

	for id := range stopsByID {
		store.PassengersAtStops[id] = &passengerStop{
			passengers: make([]*Passenger, 0),
		}
	}

	store.generatePassengers(c)

	return store
}

func (ps *PassengersStore) GetPassengerCountAtStop(stopID uint64) uint {
	return ps.PassengersAtStops[stopID].GetPassengerCount()
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

func (ps *PassengersStore) SpawnAtTime(time uint) {
	passengersToSpawn := ps.PassengersToSpawn[time]

	for _, p := range passengersToSpawn {
		stop := ps.PassengersAtStops[p.StartStopID]
		stop.AddPassengerToStop(p)
	}
}

func (ps *PassengersStore) UnloadAllToStop(stopID uint64, passengers []*Passenger) {
	stop := ps.PassengersAtStops[stopID]
	for _, p := range passengers {
		stop.AddPassengerToStop(p)
	}
}

func (ps *PassengersStore) BoardAllFromStop(stopID uint64, alreadyBoardedIDS []uint64) []*Passenger {
	// alreadyTakenSet is for temporary usage -> currently trams board passengers and
	// drop them at the next stop; they must not board the same passenger again
	// during the same stop visit
	//TODO: remove when passenger strategy is implemented
	alreadyTakenSet := make(map[uint64]struct{}, len(alreadyBoardedIDS))
	for _, id := range alreadyBoardedIDS {
		alreadyTakenSet[id] = struct{}{}
	}

	stop := ps.PassengersAtStops[stopID]
	return stop.TakeAllFromStop(alreadyTakenSet)
}
