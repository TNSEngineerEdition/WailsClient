package passenger

import (
	"fmt"
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

	earliestGettingOff := uint(9999999)
	var earliestPassenger *Passenger

	for startStopID := range stopsByID {
		for range 10 {
			// TODO: time's upper bound is set to 25200 for presentation purposes
			//spawnTime := timeBounds.StartTime + uint(rand.IntN(int(timeBounds.EndTime-timeBounds.StartTime+1)))
			spawnTime := timeBounds.StartTime + uint(rand.IntN(int(25200-timeBounds.StartTime+1)))

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

			for _, conn := range passenger.TravelPlan.stops[startStopID].connections {
				finishArrTime := conn.arrivalTime + conn.travelTime
				if finishArrTime < earliestGettingOff {
					earliestGettingOff = finishArrTime
					earliestPassenger = passenger
				}
			}

			// 2846212107 - Grota-Roweckiego -> centre
			// printStopID := uint64(2846212107)
			// if startStopID == printStopID {
			// 	printStopName := stopsByID[printStopID].GetName()
			// 	fmt.Printf("\n%s\n", printStopName)
			// 	fmt.Printf("  passenger spawn at %d\n", spawnTime)
			// 	fmt.Printf("  connections:\n")
			// 	for _, conn := range passenger.TravelPlan.stops[startStopID].connections {
			// 		fmt.Printf(
			// 			"    - [%d] => %s, getting off at %s, tram arriving at %d\n",
			// 			conn.id,
			// 			c.GetTripsByID()[conn.id].TripHeadSign,
			// 			stopsByID[conn.to].GetName(),
			// 			conn.arrivalTime,
			// 		)
			// 	}
			// }
		}
	}

	fmt.Println("Earliest get off happens at ", earliestGettingOff)
	fmt.Println("Passenger spawns at ", stopsByID[earliestPassenger.startStopID].GetName())
	fmt.Println("Passenger goes to ", stopsByID[earliestPassenger.endStopID].GetName())
}

func (ps *PassengersStore) SpawnPassengersAtTime(time uint) {
	passengersToSpawn := ps.passengersToSpawn[time]
	for _, p := range passengersToSpawn {
		stop := ps.passengerStops[p.startStopID]
		stop.addPassengerToStop(p)
	}
}

func (ps *PassengersStore) BoardPassengers(stopID uint64, tramID uint) (boardingPassengers []*Passenger) {
	passengerStop := ps.passengerStops[stopID]
	boardingPassengers = passengerStop.boardPassengersToTram(tramID)
	if len(boardingPassengers) > 0 {
		fmt.Println(tramID, "is having", len(boardingPassengers), "new passengers!")
	}
	return boardingPassengers
}

func (ps *PassengersStore) DisembarkPassengers(stopID uint64, disembarkingPassengers []*Passenger) {
	// TODO: this function should be responsible for the tram changes
	for _, p := range disembarkingPassengers {
		if p.endStopID == stopID {
			continue // passenger reached destination
		}
	}
}

func (ps *PassengersStore) UnloadAllToStop(stopID uint64, passengers []*Passenger) {
	stop := ps.passengerStops[stopID]
	for _, p := range passengers {
		stop.addPassengerToStop(p)
	}
}
