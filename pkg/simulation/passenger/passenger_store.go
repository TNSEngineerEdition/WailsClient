package passenger

import (
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

			var startStopID, endStopID uint64
			var spawnTime uint

			//mock
			spawnTime = 18420 // 05:00:00

			startStopID = 2846212107 // grota 2
			endStopID = 1768224703   // bialucha 2

			//startStopID = 2419106061 // miodowa 2
			//endStopID = 2423789754   // pedzichow 1
			//spawnTime = 18000        // 05:00:00

			// startStopID = uint64(2420979790) // kampus -> cz.m.

			//startStopID = 12297835419 // jarzebiny -> centrum

			pg := NewPassengerGraph(startStopID, endStopID, spawnTime, c)
			tp := pg.getTravelPlan(SURE)

			stopsByID := pg.c.GetStopsByID()
			trips := pg.c.GetTripsByID()
			fmt.Println("\nTravel plan for strategy ", SURE)
			for _, stop := range tp.stops {
				fmt.Printf(" - stop: %s\n", stopsByID[stop.ID].GetName())
				if len(stop.arrivals) > 0 {
					fmt.Println("    arrivals:")
					for _, arrival := range stop.arrivals {
						fmt.Printf("        [%d] %s ... => %s\n", arrival.ID, stopsByID[arrival.to].GetName(), trips[arrival.ID].TripHeadSign)
					}
				}
				if len(stop.departures) > 0 {
					fmt.Println("    departures:")
					for _, departure := range stop.departures {
						fmt.Printf("        [%d] %s ... => %s\n", departure.ID, stopsByID[departure.to].GetName(), trips[departure.ID].TripHeadSign)
					}
				}
			}

			passenger := &Passenger{
				strategy:    PassengerStrategy(rand.IntN(3)),
				spawnTime:   spawn,
				StartStopID: startStop.ID,
				EndStopID:   endStop.ID,
				ID:          counter,
				TravelPlan:  tp,
			}

			ps.PassengersToSpawn[spawn] = append(ps.PassengersToSpawn[spawn], passenger)
			counter++
			return
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
