package travelplan

import (
	"cmp"
	"math"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/structs"
)

const (
	COMFORT_MAX_WAITING_TIME = 30 * 60     // 30 minutes
	COMFORT_MAX_TRAVEL_TIME  = 3 * 60 * 60 // 3 hours
	COMFORT_MAX_TRIPS        = 5
)

type takenTrip struct {
	tripID                  uint
	startStopID, endStopID  uint64
	arrivalTime, travelTime uint
}

type comfortPQValue struct {
	tripID      uint
	stopID      uint64
	arrivalTime uint
	tripCount   uint8
	takenTrips  []takenTrip
}

type comfortPQPriority struct {
	tripCount      uint8
	timeSinceSpawn uint
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
	minTripCount       uint8
	foundPaths         [][]takenTrip
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
		minTripCount:  math.MaxUint8,
		maxTravelTime: spawnTime + COMFORT_MAX_TRAVEL_TIME,
		spawnTime:     spawnTime,
		tripsPriorityQueue: structs.NewPriorityQueue[comfortPQValue](
			func(left, right comfortPQPriority) int { return left.compare(right) },
		),
	}

	ctp.initializePriorityQueue(startStopIDs, spawnTime)
	ctp.findPathsToEndStops()

	return ctp.createTravelPlan()
}

func (ctp *comfortTravelPlan) updateTrips(takenTrips []takenTrip) {
	ctp.minTripCount = uint8(len(takenTrips))
	ctp.foundPaths = append(ctp.foundPaths, takenTrips)
}

func (ctp *comfortTravelPlan) initializePriorityQueue(startStopIDs []uint64, spawnTime uint) {
	for _, startStopID := range startStopIDs {
		ctp.addTripsFromStop(startStopID, spawnTime, spawnTime+COMFORT_MAX_WAITING_TIME, 1, []takenTrip{}, 0)
	}
}

func (ctp *comfortTravelPlan) findPathsToEndStops() {
	for ctp.tripsPriorityQueue.Len() > 0 {
		value := ctp.tripsPriorityQueue.Pop()

		// Because priority queue orders entries by trip count,
		// when a larger than already discovered trip count is found,
		// it is no longer possible to find other solutions.
		if value.tripCount > ctp.minTripCount {
			break
		}

		if ctp.endStopIDs.Includes(value.stopID) {
			ctp.updateTrips(value.takenTrips)
			continue
		}

		if value.tripCount == COMFORT_MAX_TRIPS {
			continue
		}

		for transferStopID := range ctp.currentCity.GetStopsInGroup(value.stopID) {
			ctp.addTripsFromStop(
				transferStopID,
				value.arrivalTime,
				value.arrivalTime+COMFORT_MAX_WAITING_TIME,
				value.tripCount+1,
				value.takenTrips,
				value.tripID,
			)
		}
	}
}

func (ctp *comfortTravelPlan) addTripsFromStop(stopID uint64, startTime, endTime uint, tripCount uint8, takenTrips []takenTrip, lastTripID uint) {
	arrivals := ctp.currentCity.GetPlannedArrivalsInTimeSpan(
		stopID,
		startTime,
		min(endTime, ctp.maxTravelTime),
	)

	for _, arrival := range arrivals {
		if arrival.TripID == lastTripID {
			continue
		}

		ctp.addStopsAlongTrip(arrival, tripCount, takenTrips)
	}
}

func (ctp *comfortTravelPlan) addStopsAlongTrip(arrival city.PlannedArrival, tripCount uint8, takenTrips []takenTrip) {
	tramTrip := ctp.currentCity.GetTripByID(arrival.TripID)

	for _, stop := range tramTrip.Stops[arrival.StopIndex+1:] {
		takenTrips = append(takenTrips[:], takenTrip{
			tripID:      tramTrip.ID,
			startStopID: tramTrip.Stops[arrival.StopIndex].ID,
			endStopID:   stop.ID,
			arrivalTime: stop.Time,
			travelTime:  stop.Time - arrival.Time,
		})

		if ctp.endStopIDs.Includes(stop.ID) {
			ctp.updateTrips(takenTrips)
			ctp.foundPaths = append(ctp.foundPaths, takenTrips)
			continue
		}

		ctp.tripsPriorityQueue.Push(
			comfortPQValue{
				tripID:      tramTrip.ID,
				stopID:      stop.ID,
				arrivalTime: stop.Time,
				tripCount:   tripCount,
				takenTrips:  takenTrips,
			},
			comfortPQPriority{
				tripCount:      tripCount,
				timeSinceSpawn: stop.Time - ctp.spawnTime,
			},
		)
	}
}

func (ctp comfortTravelPlan) createTravelPlan() (TravelPlan, bool) {
	// If minTripCount is not updated, then no solution is found
	if ctp.minTripCount == math.MaxUint8 {
		return TravelPlan{}, false
	}

	travelPlan := NewTravelPlan(
		ctp.foundPaths[0][0].startStopID,
		ctp.endStopIDs,
		ctp.spawnTime,
	)

	for _, takenTripsInPath := range ctp.foundPaths {
		for i, takenTrip := range takenTripsInPath {
			if i > 0 && takenTripsInPath[i-1].endStopID != takenTrip.startStopID {
				travelPlan.addTransfer(takenTripsInPath[i-1].endStopID, takenTrip.startStopID)
			}

			travelPlan.addConnection(
				takenTrip.startStopID,
				takenTrip.endStopID,
				takenTrip.tripID,
				takenTrip.arrivalTime,
				takenTrip.travelTime,
			)
		}
	}

	return travelPlan, true
}
