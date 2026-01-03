package travelplan

import (
	"cmp"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/structs"
)

const (
	COMFORT_MAX_WAITING_TIME = 30 * 60     // 30 minutes
	COMFORT_MAX_TRAVEL_TIME  = 3 * 60 * 60 // 2 hours
	COMFORT_MAX_TRIPS        = 5
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
	currentCity        *city.City
	endStopIDs         structs.Set[uint64]
	spawnTime          uint
	maxTravelTime      uint
	minTripCount       uint
	foundPaths         []tripSequence
	tripsPriorityQueue structs.PriorityQueue[comfortPQValue, comfortPQPriority]
}

func GetComfortTravelPlan(
	currentCity *city.City,
	startStopIDs []uint64,
	endStopIDs structs.Set[uint64],
	spawnTime uint,
) (TravelPlan, bool) {
	ctp := comfortTravelPlan{
		currentCity:   currentCity,
		endStopIDs:    endStopIDs,
		minTripCount:  COMFORT_MAX_TRIPS + 1,
		maxTravelTime: spawnTime + COMFORT_MAX_TRAVEL_TIME,
		spawnTime:     spawnTime,
		foundPaths:    make([]tripSequence, 0),
		tripsPriorityQueue: structs.NewPriorityQueue[comfortPQValue](
			func(left, right comfortPQPriority) int { return left.compare(right) },
		),
	}

	ctp.initializePriorityQueue(startStopIDs, spawnTime)
	ctp.findPathsToEndStops()

	return ctp.createTravelPlan()
}

func (ctp *comfortTravelPlan) updateTrips(takenTrips tripSequence) {
	// When updating the trip count, takenTrips.tripCount()
	// is never larger than ctp.minTripCount
	ctp.minTripCount = takenTrips.tripCount()

	ctp.foundPaths = append(ctp.foundPaths, takenTrips)
}

func (ctp *comfortTravelPlan) initializePriorityQueue(startStopIDs []uint64, spawnTime uint) {
	for _, startStopID := range startStopIDs {
		ctp.addTripsFromStop(startStopID, spawnTime, spawnTime+COMFORT_MAX_WAITING_TIME, newTripSequence(0))
	}
}

func (ctp *comfortTravelPlan) findPathsToEndStops() {
	for ctp.tripsPriorityQueue.Len() > 0 {
		value := ctp.tripsPriorityQueue.Pop()

		// Because priority queue orders entries by trip count,
		// when a larger than already discovered trip count is found,
		// it is no longer possible to find other solutions.
		if value.takenTrips.tripCount() >= ctp.minTripCount {
			break
		}

		if ctp.endStopIDs.Includes(value.stopID) {
			ctp.updateTrips(value.takenTrips)
			continue
		}

		if value.takenTrips.tripCount() == COMFORT_MAX_TRIPS {
			continue
		}

		for transferStopID := range ctp.currentCity.GetStopsInGroup(value.stopID) {
			ctp.addTripsFromStop(
				transferStopID,
				value.arrivalTime,
				value.arrivalTime+COMFORT_MAX_WAITING_TIME,
				value.takenTrips,
			)
		}
	}
}

func (ctp *comfortTravelPlan) addTripsFromStop(
	stopID uint64,
	startTime, endTime uint,
	takenTrips tripSequence,
) {
	arrivals := ctp.currentCity.GetPlannedArrivalsInTimeSpan(
		stopID,
		startTime,
		min(endTime, ctp.maxTravelTime),
	)

	for _, arrival := range arrivals {
		if takenTrips.tripCount() > 0 && arrival.TripID == takenTrips.trips[0].tripID {
			continue
		}

		ctp.addStopsAlongTrip(arrival, takenTrips)
	}
}

func (ctp *comfortTravelPlan) addStopsAlongTrip(
	arrival city.PlannedArrival,
	takenTrips tripSequence,
) {
	tramTrip := ctp.currentCity.GetTripByID(arrival.TripID)
	visitedStops := structs.NewSet[string]()
	visitedStops.Add(ctp.currentCity.GetStopByID(tramTrip.Stops[arrival.StopIndex].ID).GetGroupName())

	for _, stop := range tramTrip.Stops[arrival.StopIndex+1:] {
		stopGroupName := ctp.currentCity.GetStopByID(stop.ID).GetGroupName()
		visitedStops.Add(stopGroupName)

		if takenTrips.visitedStopNames.Includes(stopGroupName) {
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

		if ctp.endStopIDs.Includes(stop.ID) {
			ctp.updateTrips(takenTripsAfterStop)
			continue
		}

		ctp.tripsPriorityQueue.Push(
			comfortPQValue{
				tripID:      tramTrip.ID,
				stopID:      stop.ID,
				arrivalTime: stop.Time,
				takenTrips:  takenTripsAfterStop,
			},
			comfortPQPriority{
				tripCount:      takenTripsAfterStop.tripCount(),
				timeSinceSpawn: stop.Time - ctp.spawnTime,
			},
		)
	}
}

func (ctp comfortTravelPlan) createTravelPlan() (TravelPlan, bool) {
	// If minTripCount is not updated, then no solution is found
	if ctp.minTripCount > COMFORT_MAX_TRIPS {
		return TravelPlan{}, false
	}

	travelPlan := NewTravelPlan(
		ctp.foundPaths[0].getStartStopID(),
		ctp.endStopIDs,
		ctp.spawnTime,
	)

	for _, takenTripsInPath := range ctp.foundPaths {
		takenTripsInPath.addToTravelPlan(&travelPlan)
	}

	return travelPlan, true
}
