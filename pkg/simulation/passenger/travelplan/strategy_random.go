package travelplan

import (
	"math/rand"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/trip"
	"github.com/TNSEngineerEdition/WailsClient/pkg/consts"
)

/*

Random strategy selects one random arrival and then picks one stop on the selected arrival's route.
This becomes a travel plan.

Optionally, for some passengers the travel plan includes one tram change on a transfer stop

*/

func (tp *TravelPlan) GenerateRandomTravelPlan() {
	isPassengerChangingStops := rand.Float32() < consts.TRANSFER_PROBABILITY

	// direct trip
	if !isPassengerChangingStops {
		tp.endStopID, _ = tp.findConnectionToStop(tp.startStopID, tp.spawnTime, false)
		return
	}

	// trip with transfer
	stopID, time := tp.findConnectionToStop(tp.startStopID, tp.spawnTime, true)
	if !tp.c.IsTransferStop(stopID) {
		tp.endStopID = stopID
		return
	}

	// select random stop to transfer to
	stops := tp.c.GetStopsInGroup(stopID)
	n := len(stops)
	a := rand.Intn(n)
	i := 0
	var transferStopID uint64

	for handleStopID := range stops {
		if i == a {
			transferStopID = handleStopID
			break
		}
		i++
	}
	tp.addTransfer(stopID, transferStopID)
	tp.endStopID, _ = tp.findConnectionToStop(transferStopID, time+consts.TRANSFER_TIME, false)
}

func (tp *TravelPlan) findConnectionToStop(fromStopID uint64, time uint, goToTransferStop bool) (uint64, uint) {
	arrival, stopsLeft := tp.getRandomArrivalFromStop(fromStopID, time)
	if arrival == nil {
		return fromStopID, time
	}

	trips := tp.c.GetTripsByID()
	trip := trips[arrival.TripID]

	// passenger wants to travel to a transfer stop
	if goToTransferStop {
		if toStopID, travelTime, found := tp.findTransferStop(trip, arrival); found {
			tp.addConnection(
				fromStopID,
				toStopID,
				arrival.TripID,
				arrival.Time,
				travelTime,
			)
			return toStopID, arrival.Time + travelTime
		}
	}

	// default: travel to a random stop
	toStopID, travelTime := tp.selectRandomStop(trip, arrival, stopsLeft)
	tp.addConnection(
		fromStopID,
		toStopID,
		arrival.TripID,
		arrival.Time,
		travelTime,
	)
	return toStopID, arrival.Time + travelTime
}

func (tp *TravelPlan) findTransferStop(trip *trip.TramTrip, arrival *city.PlannedArrival) (uint64, uint, bool) {
	transferStops := make([]struct {
		stopID      uint64
		arrivalTime uint
	}, 0)

	currentStopIndex := arrival.StopIndex + 1
	for currentStopIndex < len(trip.Stops) {
		currStopID := trip.Stops[currentStopIndex].ID
		if tp.c.IsTransferStop(currStopID) {
			transferStops = append(transferStops, struct {
				stopID      uint64
				arrivalTime uint
			}{
				stopID:      currStopID,
				arrivalTime: trip.Stops[currentStopIndex].Time,
			})
		}
		currentStopIndex++
	}

	if len(transferStops) == 0 {
		return 0, 0, false
	}

	destination := transferStops[rand.Intn(len(transferStops))]
	return destination.stopID, destination.arrivalTime - arrival.Time, true
}

func (tp *TravelPlan) selectRandomStop(trip *trip.TramTrip, arrival *city.PlannedArrival, stopsLeft int) (uint64, uint) {
	stopsToTravel := rand.Intn(stopsLeft) + 1 // Travel for at least 1 stop
	toStopIndex := arrival.StopIndex + stopsToTravel
	toStopID := trip.Stops[toStopIndex].ID
	travelTime := trip.Stops[toStopIndex].Time - arrival.Time
	return toStopID, travelTime
}

func (tp *TravelPlan) getRandomArrivalFromStop(stopID uint64, time uint) (arrival *city.PlannedArrival, stopsLeft int) {
	trips := tp.c.GetTripsByID()
	arrivals := tp.c.GetPlannedArrivalsInTimeSpan(stopID, time, time+consts.MAX_WAITING_TIME)
	if arrivals == nil {
		return nil, 0
	}

	filteredArrivals := make([]city.PlannedArrival, 0)
	for _, arrival := range arrivals {
		if len(trips[arrival.TripID].Stops)-1 == arrival.StopIndex {
			continue // skip trams being at their last stop
		}
		filteredArrivals = append(filteredArrivals, arrival)
	}

	if len(filteredArrivals) == 0 {
		return nil, 0
	}

	// try up to 10 times to find a valid arrival with stops left
	for range 10 {
		arrival = &filteredArrivals[rand.Intn(len(filteredArrivals))]
		stopsTotal := len(trips[arrival.TripID].Stops)
		stopsLeft = stopsTotal - arrival.StopIndex - 1
		if stopsLeft > 0 {
			return arrival, stopsLeft
		}
	}

	return nil, 0
}
