package travelplan

import (
	"github.com/TNSEngineerEdition/WailsClient/pkg/api"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/trip"
	"github.com/TNSEngineerEdition/WailsClient/pkg/structs"
)

type travelPlanBuilder[V, P any] interface {
	handlePQValue(value *V) bool
	getPQValueAndPriority(
		tramTrip *trip.TramTrip,
		stop *api.ResponseTramTripStop,
		takenTripsAfterStop *tripSequence,
	) (V, P)
	onPathFound(takenTripsAfterStop *tripSequence)
}

type abstractTravelPlanBuilder[V, P any] struct {
	travelPlanBuilder[V, P]
	currentCity        *city.City
	startStopIDs       []uint64
	endStopIDs         structs.Set[uint64]
	spawnTime          uint
	maxTravelTime      uint
	foundPaths         []tripSequence
	tripsPriorityQueue structs.PriorityQueue[V, P]
}

func NewAbstractTravelPlanBuilder[V, P any](
	currentCity *city.City,
	startStopIDs []uint64,
	endStopIDs structs.Set[uint64],
	spawnTime uint,
	maxTravelTime uint,
	priorityCompare func(left, right P) int,
) abstractTravelPlanBuilder[V, P] {
	return abstractTravelPlanBuilder[V, P]{
		currentCity:        currentCity,
		startStopIDs:       startStopIDs,
		endStopIDs:         endStopIDs,
		spawnTime:          spawnTime,
		maxTravelTime:      maxTravelTime,
		foundPaths:         make([]tripSequence, 0),
		tripsPriorityQueue: structs.NewPriorityQueue[V](priorityCompare),
	}
}

func (a *abstractTravelPlanBuilder[V, P]) initializePriorityQueue() {
	for _, startStopID := range a.startStopIDs {
		a.addTripsFromStop(startStopID, a.spawnTime, a.spawnTime+MAX_WAITING_TIME, newTripSequence(0))
	}
}

func (a *abstractTravelPlanBuilder[V, P]) addTripsFromStop(
	stopID uint64,
	startTime, endTime uint,
	takenTrips tripSequence,
) {
	arrivals := a.currentCity.GetPlannedArrivalsInTimeSpan(
		stopID,
		startTime,
		min(endTime, a.maxTravelTime),
	)

	for _, arrival := range arrivals {
		tripCount := takenTrips.tripCount()
		if tripCount > 0 && arrival.TripID == takenTrips.trips[tripCount-1].tripID {
			continue
		}

		a.addStopsAlongTrip(arrival, takenTrips)
	}
}

func (a *abstractTravelPlanBuilder[V, P]) addStopsAlongTrip(
	arrival city.PlannedArrival,
	takenTrips tripSequence,
) {
	if takenTrips.tripCount() >= MAX_TRIPS {
		return
	}

	stopsByID := a.currentCity.GetStopsByID()
	tramTrip := a.currentCity.GetTripByID(arrival.TripID)

	visitedStops := structs.NewSet[string]()
	visitedStops.Add(stopsByID[tramTrip.Stops[arrival.StopIndex].ID].GetGroupName())

	for _, stop := range tramTrip.Stops[arrival.StopIndex+1:] {
		stopGroupName := a.currentCity.GetStopByID(stop.ID).GetGroupName()
		visitedStops.Add(stopGroupName)

		if takenTrips.visitedStopNames.Includes(stopGroupName) {
			continue
		}

		// Transfer only on transfer stops, but allow ending
		// trips at end stops
		if !a.currentCity.IsTransferStop(stop.ID) && !a.endStopIDs.Includes(stop.ID) {
			continue
		}

		takenTripsAfterStop := takenTrips.extendTripRecords(
			tramTrip.ID,
			stop.Time,
			stop.Time-arrival.Time,
			tramTrip.Stops[arrival.StopIndex].ID,
			stop.ID,
			visitedStops,
		)

		if a.endStopIDs.Includes(stop.ID) {
			a.onPathFound(&takenTripsAfterStop)
			a.foundPaths = append(a.foundPaths, takenTripsAfterStop)
			break
		}

		value, priority := a.getPQValueAndPriority(tramTrip, &stop, &takenTripsAfterStop)
		a.tripsPriorityQueue.Push(value, priority)
	}
}

func (a *abstractTravelPlanBuilder[V, P]) buildTravelPlan() (TravelPlan, bool) {
	a.initializePriorityQueue()

	for a.tripsPriorityQueue.Len() > 0 {
		if len(a.foundPaths) >= MAX_PATHS {
			break
		}

		value := a.tripsPriorityQueue.Pop()
		if a.handlePQValue(&value) {
			break
		}
	}

	// If foundPaths is not updated, then no solution was found
	if len(a.foundPaths) == 0 {
		return TravelPlan{}, false
	}

	travelPlan := NewTravelPlan(
		a.foundPaths[0].getStartStopID(),
		a.endStopIDs,
		a.spawnTime,
	)

	for _, takenTripsInPath := range a.foundPaths {
		takenTripsInPath.addToTravelPlan(&travelPlan)
	}

	return travelPlan, true
}
