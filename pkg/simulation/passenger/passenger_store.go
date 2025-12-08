package passenger

import (
	"math/rand/v2"
	"sync"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/consts"
)

type passengerSpawn struct {
	passenger *Passenger
	stopID    uint64
}

type PassengersStore struct {
	passengerStops    map[uint64]*passengerStop
	passengersToSpawn map[uint][]passengerSpawn
	mu                sync.Mutex
}

func NewPassengersStore(c *city.City) *PassengersStore {
	stopsByID := c.GetStopsByID()

	store := &PassengersStore{
		passengerStops:    make(map[uint64]*passengerStop, len(stopsByID)),
		passengersToSpawn: make(map[uint][]passengerSpawn),
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
			// TODO: time's upper bound is set to 18360 (6:00:00) for presentation purposes
			//spawnTime := timeBounds.StartTime + uint(rand.IntN(int(timeBounds.EndTime-timeBounds.StartTime+1)))
			spawnTime := timeBounds.StartTime + uint(rand.IntN(int(18360-timeBounds.StartTime+1)))

			tp := GetRandomTravelPlan(startStopID, spawnTime, c)

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

			ps.passengersToSpawn[spawnTime] = append(ps.passengersToSpawn[spawnTime], passengerSpawn{
				passenger: passenger,
				stopID:    passenger.startStopID,
			})
			counter++
		}
	}
}

func (ps *PassengersStore) SpawnPassengersAtTime(time uint) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	spawnList := ps.passengersToSpawn[time]
	for _, entry := range spawnList {
		stop := ps.passengerStops[entry.stopID]
		stop.addPassengerToStop(entry.passenger)
	}
}

func (ps *PassengersStore) DespawnPassengersAtTime(time uint) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	despawnTime := time - consts.MAX_WAITING_TIME - consts.DESPAWN_TIME_OFFSET
	spawnList, ok := ps.passengersToSpawn[despawnTime]
	if !ok {
		return
	}

	for _, entry := range spawnList {
		stop := ps.passengerStops[entry.stopID]
		stop.despawnPassenger(entry.passenger)
	}
}

func (ps *PassengersStore) BoardPassengers(stopID uint64, tramID uint) []*Passenger {
	passengerStop := ps.passengerStops[stopID]
	return passengerStop.boardPassengersToTram(tramID)
}

func (ps *PassengersStore) DisembarkPassengers(passengers []*Passenger, stopID uint64, time uint) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	for _, p := range passengers {
		if p.TravelPlan.isEndStopReached(stopID) {
			continue
		}

		changeStopID := p.TravelPlan.stops[stopID].changeStopTo
		changeTime := time + consts.TRAM_CHANGE_TIME
		ps.passengersToSpawn[changeTime] = append(ps.passengersToSpawn[time], passengerSpawn{
			passenger: p,
			stopID:    changeStopID,
		})
	}
}
