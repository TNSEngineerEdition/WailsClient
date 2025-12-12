package passenger

import (
	"bytes"
	"encoding/csv"
	"fmt"
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

	return store
}

func (ps *PassengersStore) GetPassengerCountAtStop(stopID uint64) uint {
	return ps.PassengersAtStops[stopID].GetPassengerCount()
}

func (ps *PassengersStore) GeneratePassengers(c *city.City) {
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

func (ps *PassengersStore) GeneratePassengersDueModel(c *city.City, passengerModel []byte) error {
	reader := csv.NewReader(bytes.NewReader(passengerModel))
	reader.Comma = ','

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading passenger model csv: %w", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("passenger model csv is empty")
	}

	header := records[0]
	if len(header) < 4 {
		return fmt.Errorf("invalid header, expected at least 4 columns, got %d", len(header))
	}

	for i, row := range records[1:] {
		lineNo := i + 2

		if len(row) < 4 {
			return fmt.Errorf("line %d: expected 4 columns, got %d", lineNo, len(row))
		}

		// startName := strings.TrimSpace(row[0])
		// endName := strings.TrimSpace(row[1])
		// spawnTimeStr := strings.TrimSpace(row[2])
		// strategyStr := strings.TrimSpace(row[3])

		// t, err := time.Parse("15:04:05", spawnTimeStr)
		// if err != nil {
		//     return fmt.Errorf("line %d: invalid spawn_time %q: %w", lineNo, spawnTimeStr, err)
		// }

		// spawnSeconds := t.Hour()*3600 + t.Minute()*60 + t.Second()

		// spawn := PassengerSpawn{
		//     StartStopName: startName,
		//     EndStopName:   endName,
		//     SpawnSeconds:  spawnSeconds,
		//     Strategy:      SpawnStrategy(strategyStr),
		// }

		// spawns = append(spawns, spawn)
	}

	fmt.Printf("git")
	return nil
}

func (ps *PassengersStore) ResetPassengers() {
	for _, stop := range ps.PassengersAtStops {
		stop.passengers = stop.passengers[:0]
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
