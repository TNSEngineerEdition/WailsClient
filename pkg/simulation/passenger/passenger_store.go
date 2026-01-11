package passenger

import (
	"fmt"
	"sync"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/graph"
	"github.com/TNSEngineerEdition/WailsClient/pkg/travelplan"
)

const (
	WAIT_DESPAWN_TIME = travelplan.MAX_WAITING_TIME + 5*60
)

type passengerSpawn struct {
	passenger *Passenger
	stopID    uint64
}

type PassengersStore struct {
	passengers        []Passenger
	passengerStops    map[uint64]*passengerStop
	passengersToSpawn map[uint][]passengerSpawn
	mu                sync.Mutex
}

func NewPassengersStore(c *city.City, passengers []Passenger) *PassengersStore {
	stopsByID := c.GetStopsByID()

	store := &PassengersStore{
		passengers:        passengers,
		passengerStops:    make(map[uint64]*passengerStop, len(stopsByID)),
		passengersToSpawn: make(map[uint][]passengerSpawn),
	}

	for i, passenger := range store.passengers {
		store.passengersToSpawn[passenger.spawnTime] = append(store.passengersToSpawn[passenger.spawnTime], passengerSpawn{
			passenger: &store.passengers[i],
			stopID:    passenger.TravelPlan.GetStartStopID(),
		})
	}

	for id := range stopsByID {
		store.passengerStops[id] = &passengerStop{
			stopID:     id,
			passengers: make(map[uint64]*Passenger),
		}
	}

	return store
}

func (ps *PassengersStore) GetPassengerCountAtStop(stopID uint64) uint {
	return ps.passengerStops[stopID].GetPassengerCount()
}

func getStopIDsFromGroupName(stopsByName map[string]map[uint64]*graph.GraphTramStop, stopName string) ([]uint64, error) {
	if stopName == "" {
		return nil, fmt.Errorf("empty stop group name")
	}

	group, ok := stopsByName[stopName]
	if !ok || len(group) == 0 {
		return nil, fmt.Errorf("stop group %q not found", stopName)
	}

	ids := make([]uint64, 0, len(group))
	for id := range group {
		ids = append(ids, id)
	}

	return ids, nil
}

func (ps *PassengersStore) ResetPassengers() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	for _, stop := range ps.passengerStops {
		stop.mu.Lock()
		stop.passengers = make(map[uint64]*Passenger)
		stop.mu.Unlock()
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

	despawnTime := time - WAIT_DESPAWN_TIME
	spawnList, ok := ps.passengersToSpawn[despawnTime]
	if !ok {
		return
	}

	for _, entry := range spawnList {
		stop := ps.passengerStops[entry.stopID]
		stop.despawnPassenger(entry.passenger)
	}
}

func (ps *PassengersStore) LoadPassengers(stopID uint64, tramID, time uint) []*Passenger {
	passengerStop := ps.passengerStops[stopID]
	return passengerStop.loadPassengersToTram(tramID, time)
}

func (ps *PassengersStore) UnloadPassengers(passengers []*Passenger, stopID uint64, time uint) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	for _, p := range passengers {
		p.saveGetOffTime(time)

		if p.TravelPlan.IsEndStopReached(stopID) {
			continue
		}

		// transfer
		transferStopID := p.TravelPlan.GetConnectionTransferDestination(stopID)

		transferTime := time
		if transferStopID != stopID {
			transferTime += travelplan.TRANSFER_TIME
		}

		ps.passengersToSpawn[transferTime] = append(ps.passengersToSpawn[transferTime], passengerSpawn{
			passenger: p,
			stopID:    transferStopID,
		})
	}
}
