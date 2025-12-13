package travelplan

import (
	"math/rand"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/consts"
)

/*

Random strategy selects one random arrival and then picks one stop on the selected arrival's route.
This becomes a travel plan.

Optionally, for some passengers the travel plan includes one tram change on an interchange stop

*/

func (tp *TravelPlan) GenerateRandomTravelPlan() {
	isPassengerChangingStops := rand.Float32() < consts.TRAM_CHANGE_PROBABILITY

	// direct trip
	if !isPassengerChangingStops {
		tp.endStopID, _ = tp.findConnectionToStop(tp.startStopID, tp.spawnTime, false)
		return
	}

	// trip with tram change
	stopID, time := tp.findConnectionToStop(tp.startStopID, tp.spawnTime, true)
	if !tp.c.IsInterchangeStop(stopID) {
		tp.endStopID = stopID
		return
	}

	// select random stop to change to
	stops := tp.c.GetStopsInGroup(stopID)
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
	tp.endStopID, _ = tp.findConnectionToStop(changeStopID, time+consts.TRAM_CHANGE_TIME, false)
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
		currentStopIndex := arrival.StopIndex + 1

		for currentStopIndex < len(trip.Stops) {
			currStopID := trip.Stops[currentStopIndex].ID
			if tp.c.IsInterchangeStop(currStopID) {
				interchangeStops = append(interchangeStops, struct {
					stopID      uint64
					arrivalTime uint
				}{
					stopID:      currStopID,
					arrivalTime: trip.Stops[currentStopIndex].Time,
				})
			}
			currentStopIndex++
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

func (tp *TravelPlan) getRandomArrivalFromStop(stopID uint64, time uint) (arrival *city.PlannedArrival, stopsLeft int) {
	trips := tp.c.GetTripsByID()
	arrivals := tp.c.GetPlannedArrivalsInTimeSpan(
		stopID,
		time,
		time+consts.MAX_WAITING_TIME,
	)

	if arrivals == nil {
		return nil, 0
	}

	filteredArrivals := make([]city.PlannedArrival, 0)
	for _, arrival := range arrivals {
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
