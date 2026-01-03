package travelplan

import (
	"math/rand"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/trip"
	"github.com/TNSEngineerEdition/WailsClient/pkg/consts"
	"github.com/TNSEngineerEdition/WailsClient/pkg/structs"
)

/*

Random strategy selects one random arrival and then picks one stop on the selected arrival's route.
This becomes a travel plan.

Optionally, for some passengers the travel plan includes one tram change on a transfer stop

*/

type randomTravelPlan struct {
	TravelPlan
	currentCity *city.City
}

func GetRandomTravelPlan(currentCity *city.City, startStopID uint64, spawnTime uint) (TravelPlan, bool) {
	rtp := randomTravelPlan{
		TravelPlan:  NewTravelPlan(startStopID, structs.NewSet[uint64](), spawnTime),
		currentCity: currentCity,
	}

	isPassengerChangingStops := rand.Float32() < consts.TRANSFER_PROBABILITY

	// direct trip
	if !isPassengerChangingStops {
		endStopID, _ := rtp.findConnectionToStop(startStopID, spawnTime, false)
		rtp.endStopIDs.Add(endStopID)
		return rtp.TravelPlan, startStopID != endStopID
	}

	// trip with transfer
	intermediateStopID, time := rtp.findConnectionToStop(startStopID, spawnTime, true)
	if !currentCity.IsTransferStop(intermediateStopID) {
		rtp.endStopIDs.Add(intermediateStopID)
		return rtp.TravelPlan, true
	}

	// select random stop to transfer to
	stops := currentCity.GetStopsInGroup(intermediateStopID)
	chosenStopID, i := rand.Intn(len(stops)), 0

	var transferStopID uint64
	for handleStopID := range stops {
		if i == chosenStopID {
			transferStopID = handleStopID
			break
		}
		i++
	}

	rtp.TravelPlan.addTransfer(intermediateStopID, transferStopID)
	endStopID, _ := rtp.findConnectionToStop(transferStopID, time+consts.TRANSFER_TIME, false)
	rtp.endStopIDs.Add(endStopID)

	return rtp.TravelPlan, true
}

func (rtp *randomTravelPlan) findConnectionToStop(fromStopID uint64, time uint, goToTransferStop bool) (uint64, uint) {
	arrival, stopsLeft := rtp.getRandomArrivalFromStop(fromStopID, time)
	if arrival == nil {
		return fromStopID, time
	}

	trips := rtp.currentCity.GetTripsByID()
	trip := trips[arrival.TripID]

	// passenger wants to travel to a transfer stop
	if toStopID, travelTime, ok := rtp.findTransferStop(trip, arrival); goToTransferStop && ok {
		rtp.addConnection(
			fromStopID,
			toStopID,
			arrival.TripID,
			arrival.Time,
			travelTime,
		)
		return toStopID, arrival.Time + travelTime
	}

	// travel to a random stop
	toStopID, travelTime := rtp.selectRandomStop(trip, arrival, stopsLeft)
	rtp.addConnection(
		fromStopID,
		toStopID,
		arrival.TripID,
		arrival.Time,
		travelTime,
	)

	return toStopID, arrival.Time + travelTime
}

func (rtp *randomTravelPlan) findTransferStop(trip *trip.TramTrip, arrival *city.PlannedArrival) (uint64, uint, bool) {
	transferStops := make([]struct {
		stopID      uint64
		arrivalTime uint
	}, 0)

	currentStopIndex := arrival.StopIndex + 1
	for currentStopIndex < len(trip.Stops) {
		currStopID := trip.Stops[currentStopIndex].ID
		if rtp.currentCity.IsTransferStop(currStopID) {
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

func (rtp *randomTravelPlan) selectRandomStop(trip *trip.TramTrip, arrival *city.PlannedArrival, stopsLeft int) (uint64, uint) {
	stopsToTravel := rand.Intn(stopsLeft) + 1 // Travel for at least 1 stop
	toStopIndex := arrival.StopIndex + stopsToTravel
	toStopID := trip.Stops[toStopIndex].ID
	travelTime := trip.Stops[toStopIndex].Time - arrival.Time
	return toStopID, travelTime
}

func (rtp *randomTravelPlan) getRandomArrivalFromStop(stopID uint64, time uint) (arrival *city.PlannedArrival, stopsLeft int) {
	trips := rtp.currentCity.GetTripsByID()
	arrivals := rtp.currentCity.GetPlannedArrivalsInTimeSpan(stopID, time, time+consts.MAX_WAITING_TIME)
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
