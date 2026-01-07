package travelplan

import (
	"cmp"

	"github.com/TNSEngineerEdition/WailsClient/pkg/api"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/trip"
	"github.com/TNSEngineerEdition/WailsClient/pkg/structs"
)

type fastestPQValue struct {
	tripID      uint
	stopID      uint64
	arrivalTime uint
	takenTrips  tripSequence
}

type fastestTravelPlan struct {
	abstractTravelPlanBuilder[fastestPQValue, uint]
	offsetBetweenTransfers uint
}

func GetFastestTravelPlan(
	currentCity *city.City,
	startStopIDs []uint64,
	endStopIDs structs.Set[uint64],
	spawnTime uint,
	offsetBetweenTransfers uint,
) (TravelPlan, bool) {
	ftp := fastestTravelPlan{
		abstractTravelPlanBuilder: NewAbstractTravelPlanBuilder[fastestPQValue](
			currentCity,
			startStopIDs,
			endStopIDs,
			spawnTime,
			spawnTime+MAX_TRAVEL_TIME,
			cmp.Compare[uint],
		),
		offsetBetweenTransfers: offsetBetweenTransfers,
	}

	// Inverse dependency to make abstraction work
	ftp.abstractTravelPlanBuilder.travelPlanBuilder = &ftp

	return ftp.buildTravelPlan()
}

func (ftp *fastestTravelPlan) handlePQValue(value *fastestPQValue) bool {
	for transferStopID := range ftp.currentCity.GetStopsInGroup(value.stopID) {
		startTime := value.arrivalTime + ftp.offsetBetweenTransfers
		endTime := value.arrivalTime + MAX_WAITING_TIME

		if value.stopID != transferStopID {
			startTime += TRANSFER_TIME
		}

		ftp.addTripsFromStop(transferStopID, startTime, endTime, value.takenTrips)
	}

	return false
}

func (ftp *fastestTravelPlan) getPQValueAndPriority(
	tramTrip *trip.TramTrip,
	stop *api.ResponseTramTripStop,
	takenTripsAfterStop *tripSequence,
) (fastestPQValue, uint) {
	value := fastestPQValue{
		tripID:      tramTrip.ID,
		stopID:      stop.ID,
		arrivalTime: stop.Time,
		takenTrips:  *takenTripsAfterStop,
	}

	priority := stop.Time - ftp.spawnTime

	return value, priority
}

func (ftp *fastestTravelPlan) onPathFound(takenTrips *tripSequence) {}
