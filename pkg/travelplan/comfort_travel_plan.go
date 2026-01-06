package travelplan

import (
	"cmp"

	"github.com/TNSEngineerEdition/WailsClient/pkg/api"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/trip"
	"github.com/TNSEngineerEdition/WailsClient/pkg/structs"
)

type comfortPQValue struct {
	tripID      uint
	stopID      uint64
	arrivalTime uint
	takenTrips  tripSequence
}

type comfortPQPriority struct {
	tripCount, timeSinceSpawn uint
}

func (c comfortPQPriority) compare(other comfortPQPriority) int {
	return cmp.Or(
		cmp.Compare(c.tripCount, other.tripCount),
		cmp.Compare(c.timeSinceSpawn, other.timeSinceSpawn),
	)
}

type comfortTravelPlan struct {
	abstractTravelPlanBuilder[comfortPQValue, comfortPQPriority]
	minTripCount uint
}

func GetComfortTravelPlan(
	currentCity *city.City,
	startStopIDs []uint64,
	endStopIDs structs.Set[uint64],
	spawnTime uint,
) (TravelPlan, bool) {
	ctp := comfortTravelPlan{
		abstractTravelPlanBuilder: NewAbstractTravelPlanBuilder[comfortPQValue](
			currentCity,
			startStopIDs,
			endStopIDs,
			spawnTime,
			spawnTime+MAX_TRAVEL_TIME,
			func(left, right comfortPQPriority) int { return left.compare(right) },
		),
		minTripCount: MAX_TRIPS + 1,
	}

	// Inverse dependency to make abstraction work
	ctp.abstractTravelPlanBuilder.travelPlanBuilder = &ctp

	return ctp.buildTravelPlan()
}

func (ctp *comfortTravelPlan) handlePQValue(value *comfortPQValue) bool {
	// Because priority queue orders entries by trip count,
	// when a larger than already discovered trip count is found,
	// it is no longer possible to find other solutions.
	if value.takenTrips.tripCount() >= ctp.minTripCount {
		return true
	}

	for transferStopID := range ctp.currentCity.GetStopsInGroup(value.stopID) {
		startTime, endTime := value.arrivalTime, value.arrivalTime+MAX_WAITING_TIME

		if value.stopID != transferStopID {
			startTime += TRANSFER_TIME
		}

		ctp.addTripsFromStop(transferStopID, startTime, endTime, value.takenTrips)
	}

	return false
}

func (ctp *comfortTravelPlan) getPQValueAndPriority(
	tramTrip *trip.TramTrip,
	stop *api.ResponseTramTripStop,
	takenTripsAfterStop *tripSequence,
) (comfortPQValue, comfortPQPriority) {
	value := comfortPQValue{
		tripID:      tramTrip.ID,
		stopID:      stop.ID,
		arrivalTime: stop.Time,
		takenTrips:  *takenTripsAfterStop,
	}

	priority := comfortPQPriority{
		tripCount:      takenTripsAfterStop.tripCount(),
		timeSinceSpawn: stop.Time - ctp.spawnTime,
	}

	return value, priority
}

func (ctp *comfortTravelPlan) onPathFound(takenTrips *tripSequence) {
	// When updating the trip count, takenTrips.tripCount()
	// is never larger than ctp.minTripCount
	ctp.minTripCount = takenTrips.tripCount()
}
