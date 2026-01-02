package passenger

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"time"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/graph"
	"github.com/TNSEngineerEdition/WailsClient/pkg/consts"
	"github.com/TNSEngineerEdition/WailsClient/pkg/simulation/passenger/travelplan"
)

type passengerSpawn struct {
	passenger *Passenger
	stopID    uint64
}

type PassengersStore struct {
	passengers        []*Passenger
	passengerStops    map[uint64]*passengerStop
	passengersToSpawn map[uint][]passengerSpawn
	mu                sync.Mutex
}

func NewPassengersStore(c *city.City) *PassengersStore {
	stopsByID := c.GetStopsByID()

	store := &PassengersStore{
		passengers:        make([]*Passenger, 0, len(c.GetNodesByID())*50),
		passengerStops:    make(map[uint64]*passengerStop, len(stopsByID)),
		passengersToSpawn: make(map[uint][]passengerSpawn),
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

func (ps *PassengersStore) GenerateRandomPassengers(c *city.City) {
	timeBounds := c.GetTimeBounds()
	stopsByID := c.GetStopsByID()

	var counter uint64

	for startStopID := range stopsByID {
		for range 50 {
			// TODO: time's upper bound is set to 18360 (6:00:00) for presentation purposes
			//spawnTime := timeBounds.StartTime + uint(rand.IntN(int(timeBounds.EndTime-timeBounds.StartTime+1)))
			spawnTime := timeBounds.StartTime + uint(rand.IntN(int(18360-timeBounds.StartTime+1)))
			strategy := travelplan.RANDOM

			tp, endStopID := travelplan.GetTravelPlan(strategy, startStopID, spawnTime, c)

			if startStopID == endStopID {
				continue // no trips found
			}

			passenger := &Passenger{
				ID:          counter,
				strategy:    strategy,
				spawnTime:   spawnTime,
				startStopID: startStopID,
				endStopID:   endStopID,
				TravelPlan:  tp,
			}

			ps.passengersToSpawn[spawnTime] = append(ps.passengersToSpawn[spawnTime], passengerSpawn{
				passenger: passenger,
				stopID:    passenger.startStopID,
			})

			ps.passengers = append(ps.passengers, passenger)
			counter++
		}
	}
}

func (ps *PassengersStore) GeneratePassengersDueModel(c *city.City, passengerModel []byte) error {
	records, err := readPassengerCSV(passengerModel)
	if err != nil {
		return err
	}

	passengersToSpawn, err := buildPassengersToSpawn(c, records)
	if err != nil {
		return err
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.passengersToSpawn = passengersToSpawn

	return nil
}

func readPassengerCSV(passengerModel []byte) ([][]string, error) {
	reader := csv.NewReader(bytes.NewReader(passengerModel))
	reader.Comma = ','

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading passenger model csv: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("passenger model csv is empty")
	}

	header := records[0]
	if len(header) < 4 {
		return nil, fmt.Errorf("invalid header, expected 4 columns, got %d", len(header))
	}

	return records, nil
}

func buildPassengersToSpawn(c *city.City, records [][]string) (map[uint][]passengerSpawn, error) {
	stopsByName := c.GetStopsByName()
	result := make(map[uint][]passengerSpawn)

	for i, row := range records[1:] {

		lineNo := i + 2

		if len(row) < 4 {
			return nil, fmt.Errorf("line %d: expected 4 columns, got %d", lineNo, len(row))
		}

		startName := strings.TrimSpace(row[0])
		endName := strings.TrimSpace(row[1])
		spawnTimeStr := strings.TrimSpace(row[2])
		strategyStr := strings.TrimSpace(row[3])

		startStopsID, err := resolveStopsID(stopsByName, startName)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo, err)
		}

		endStopsID, err := resolveStopsID(stopsByName, endName)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo, err)
		}

		var spawnSeconds uint
		if t, err := time.Parse("15:04:05", spawnTimeStr); err == nil {
			spawnSeconds = uint(t.Hour()*3600 + t.Minute()*60 + t.Second())
		} else {
			return nil, fmt.Errorf("line %d: invalid spawn_time %q (expected HH:MM:SS)", lineNo, spawnTimeStr)
		}

		strategy := travelplan.PassengerStrategy(strings.ToUpper(strategyStr))

		// TODO: change when travelplans for strategies will be implemented
		switch strategy {
		case travelplan.RANDOM:
		case travelplan.ASAP, travelplan.COMFORT, travelplan.SURE:
			strategy = travelplan.RANDOM
		default:
			return nil, fmt.Errorf("line %d: unknown strategy %q", lineNo, strategyStr)
		}

		var (
			tp          travelplan.TravelPlan
			startStopID uint64
			endStopID   uint64
		)

		// TODO: change when other strategies will be added
		startStopID = startStopsID[rand.IntN(len(startStopsID))]
		endStopID = endStopsID[rand.IntN(len(endStopsID))]

		passenger := &Passenger{
			ID:          uint64(i),
			strategy:    strategy,
			spawnTime:   spawnSeconds,
			startStopID: startStopID,
			endStopID:   endStopID,
			TravelPlan:  tp,
		}

		result[spawnSeconds] = append(result[spawnSeconds], passengerSpawn{
			passenger: passenger,
			stopID:    passenger.startStopID,
		})
	}

	return result, nil
}

func resolveStopsID(stopsByName map[string]map[uint64]*graph.GraphTramStop, stopName string) ([]uint64, error) {
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
		transferStopID := p.TravelPlan.GetTransferStop(stopID)
		transferTime := time + consts.TRANSFER_TIME
		ps.passengersToSpawn[transferTime] = append(ps.passengersToSpawn[time], passengerSpawn{
			passenger: p,
			stopID:    transferStopID,
		})
	}
}
