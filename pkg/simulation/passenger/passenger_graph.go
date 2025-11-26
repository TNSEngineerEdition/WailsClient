package passenger

import (
	"fmt"
	"math/rand"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/consts"
)

type travelStop struct {
	id          uint64
	connections map[uint]*travelConnection
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

	tp.endStopID = tp.addConnectionToStop(startStopID, spawnTime)

	return tp
}

func (tp *TravelPlan) checkIfConnectionIsInPlan(stopID uint64, tramID uint) bool {
	if _, ok := tp.stops[stopID]; !ok {
		return false
	}
	if _, ok := tp.stops[stopID].connections[tramID]; ok {
		return true
	}
	return false
}

func (tp *TravelPlan) GetConnectionEnd(tramID uint) uint64 {
	if _, ok := tp.connections[tramID]; !ok {
		panic(fmt.Sprintf("Connection %d not found", tramID))
	}
	return tp.connections[tramID].to
}

func (tp *TravelPlan) addConnectionToStop(fromStopID uint64, time uint) (toStopID uint64) {
	trips := tp.c.GetTripsByID()
	arrivals, ok := tp.c.GetPlannedArrivalsInTimeSpan(
		fromStopID,
		time,
		time+consts.MAX_WAITING_TIME,
	)

	if !ok {
		return fromStopID
	}

	filteredArrivals := make([]city.PlannedArrival, 0)
	for _, arrival := range *arrivals {
		if len(trips[arrival.TripID].Stops)-1 == arrival.StopIndex {
			continue // do not consider trams which are at their last stop
		}
		filteredArrivals = append(filteredArrivals, arrival)
	}

	if len(filteredArrivals) == 0 {
		return fromStopID
	}

	stopsLeft := 0
	var arrival city.PlannedArrival

	for stopsLeft == 0 {
		n := len(filteredArrivals)
		arrival = filteredArrivals[rand.Intn(n)]
		stopsTotal := len(trips[arrival.TripID].Stops)
		stopsLeft = stopsTotal - arrival.StopIndex - 1
	}

	// TODO: make passengers travel for as long as they want later, after the presentation
	//stopsToTravel := rand.Intn(stopsLeft) + 1 // we want to travel for at least 1 stop
	stopsLeft = min(stopsLeft, 1)             // force passengers to travel for no more than 2 stops
	stopsToTravel := rand.Intn(stopsLeft) + 1 // we want to travel for at least 1 stop
	trip := trips[arrival.TripID]

	toStopID = trip.Stops[arrival.StopIndex+stopsToTravel].ID
	toStopArrivalTime := trip.Stops[arrival.StopIndex+stopsToTravel].Time

	tp.addConnection(
		fromStopID,
		toStopID,
		arrival.TripID,
		arrival.Time,
		toStopArrivalTime-arrival.Time,
	)

	return
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
