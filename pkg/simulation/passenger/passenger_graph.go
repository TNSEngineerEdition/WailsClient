package passenger

import (
	"fmt"
	"math/rand"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/consts"
)

type travelStop struct {
	id           uint64
	changeStopTo uint64
	connections  map[uint]*travelConnection
}

type travelConnection struct {
	id                      uint
	to                      uint64
	arrivalTime, travelTime uint
}

type TravelPlan struct {
	stops                  map[uint64]*travelStop
	connections            map[uint]*travelConnection
	startStopID, endStopID uint64
	spawnTime              uint
	c                      *city.City
}

func GetTravelPlan(startStopID uint64, spawnTime uint, c *city.City) TravelPlan {
	tp := TravelPlan{
		stops:       make(map[uint64]*travelStop),
		connections: make(map[uint]*travelConnection),
		startStopID: startStopID,
		spawnTime:   spawnTime,
		c:           c,
	}

	isPassengerChangingStops := rand.Float32() < 0.5

	// direct trip
	if !isPassengerChangingStops {
		tp.endStopID, _ = tp.findConnectionToStop(startStopID, spawnTime, false)
		return tp
	}

	// trip with tram change
	stopID, time := tp.findConnectionToStop(startStopID, spawnTime, true)
	if !c.IsInterchangeStop(stopID) {
		tp.endStopID = stopID
		return tp
	}

	stops := c.GetStopsInGroup(stopID)
	n := len(stops)
	a := rand.Intn(n)
	i := 0
	var changeStopID uint64

	for handleStopID := range stops {
		if i == a {
			changeStopID = handleStopID
			break
		}
		i++
	}
	tp.addStopChange(stopID, changeStopID)
	tp.endStopID, _ = tp.findConnectionToStop(changeStopID, time, false)

	return tp
}

func (tp *TravelPlan) GetConnectionEnd(tramID uint) uint64 {
	if _, ok := tp.connections[tramID]; !ok {
		panic(fmt.Sprintf("Connection %d not found", tramID))
	}
	return tp.connections[tramID].to
}

func (tp *TravelPlan) isConnectionInPlan(stopID uint64, tramID uint) bool {
	if _, ok := tp.stops[stopID]; !ok {
		return false
	}
	if _, ok := tp.stops[stopID].connections[tramID]; ok {
		return true
	}
	return false
}

func (tp *TravelPlan) isEndStopReached(stopID uint64) bool {
	stopsByID := tp.c.GetStopsByID()
	return stopID == tp.endStopID || stopsByID[stopID].GetGroupName() == stopsByID[tp.endStopID].GetGroupName()
}

func (tp *TravelPlan) getRandomArrivalFromStop(stopID uint64, time uint) (arrival *city.PlannedArrival, stopsLeft int) {
	trips := tp.c.GetTripsByID()
	arrivals, ok := tp.c.GetPlannedArrivalsInTimeSpan(
		stopID,
		time,
		time+consts.MAX_WAITING_TIME,
	)

	if !ok {
		return nil, 0
	}

	filteredArrivals := make([]city.PlannedArrival, 0)
	for _, arrival := range *arrivals {
		if len(trips[arrival.TripID].Stops)-1 == arrival.StopIndex {
			continue // do not consider trams which are at their last stop
		}
		filteredArrivals = append(filteredArrivals, arrival)
	}

	if len(filteredArrivals) == 0 {
		return nil, 0
	}

	loopCounter := 0

	for stopsLeft == 0 {
		if loopCounter >= 10 {
			return nil, 0
		}

		n := len(filteredArrivals)
		arrival = &filteredArrivals[rand.Intn(n)]
		stopsTotal := len(trips[arrival.TripID].Stops)
		stopsLeft = stopsTotal - arrival.StopIndex - 1
		loopCounter++
	}

	return arrival, stopsLeft
}

func (tp *TravelPlan) findConnectionToStop(fromStopID uint64, time uint, goToInterchangeStop bool) (toStopID uint64, arrivalTime uint) {
	arrival, stopsLeft := tp.getRandomArrivalFromStop(fromStopID, time)
	if arrival == nil {
		return fromStopID, time
	}

	trips := tp.c.GetTripsByID()
	trip := trips[arrival.TripID]

	var travelTime uint
	var interchangeStopsFound bool

	if goToInterchangeStop {
		interchangeStops := make([]struct {
			stopID      uint64
			arrivalTime uint
		}, 0)
		currStopIndex := arrival.StopIndex + 1

		for currStopIndex < len(trip.Stops) {
			currStopID := trip.Stops[currStopIndex].ID
			if tp.c.IsInterchangeStop(currStopID) {
				interchangeStops = append(interchangeStops, struct {
					stopID      uint64
					arrivalTime uint
				}{
					stopID:      currStopID,
					arrivalTime: trip.Stops[currStopIndex].Time,
				})
			}
			currStopIndex++
		}

		n := len(interchangeStops)
		if n > 0 {
			destination := interchangeStops[rand.Intn(n)]
			toStopID = destination.stopID
			travelTime = destination.arrivalTime - arrival.Time
			interchangeStopsFound = true
		}
	}

	if !goToInterchangeStop || !interchangeStopsFound {
		stopsToTravel := rand.Intn(stopsLeft) + 1 // we want to travel for at least 1 stop
		toStopIndex := arrival.StopIndex + stopsToTravel
		toStopID = trip.Stops[toStopIndex].ID
		travelTime = trip.Stops[toStopIndex].Time - arrival.Time
	}

	tp.addConnection(
		fromStopID,
		toStopID,
		arrival.TripID,
		arrival.Time,
		travelTime,
	)

	return toStopID, arrival.Time + travelTime
}

func (tp *TravelPlan) addStop(stopID uint64) {
	if _, ok := tp.stops[stopID]; !ok {
		tp.stops[stopID] = &travelStop{
			id:          stopID,
			connections: make(map[uint]*travelConnection),
		}
	}
}

func (tp *TravelPlan) addConnection(from, to uint64, tripID, arrivalTime, travelTime uint) {
	if _, ok := tp.stops[from]; !ok {
		tp.addStop(from)
	}
	if _, ok := tp.stops[to]; !ok {
		tp.addStop(to)
	}

	conn := travelConnection{
		id:          tripID,
		to:          to,
		arrivalTime: arrivalTime,
		travelTime:  travelTime,
	}

	fromNode := tp.stops[from]
	fromNode.connections[tripID] = &conn
	tp.stops[from] = fromNode

	tp.connections[tripID] = &conn
}

func (tp *TravelPlan) addStopChange(from, to uint64) {
	if _, ok := tp.stops[from]; !ok {
		tp.addStop(from)
	}
	if _, ok := tp.stops[to]; !ok {
		tp.addStop(to)
	}

	fromNode := tp.stops[from]
	fromNode.changeStopTo = to
}
