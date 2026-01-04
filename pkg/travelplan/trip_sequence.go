package travelplan

import "github.com/TNSEngineerEdition/WailsClient/pkg/structs"

type tripRecord struct {
	tripID                  uint
	arrivalTime, travelTime uint
	startStopID, endStopID  uint64
}

type tripSequence struct {
	trips            []*tripRecord
	visitedStopNames structs.Set[string]
}

func newTripSequence(tripCount int) tripSequence {
	return tripSequence{
		trips:            make([]*tripRecord, tripCount, tripCount+1),
		visitedStopNames: structs.NewSet[string](),
	}
}

func (t tripSequence) extendTripRecords(
	tripID uint,
	arrivalTime, travelTime uint,
	startStopID, endStopID uint64,
	visitedStops structs.Set[string],
) tripSequence {
	result := newTripSequence(len(t.trips))
	copy(result.trips, t.trips)

	result.visitedStopNames = t.visitedStopNames.Copy()
	for stopGroupName := range visitedStops.GetItems() {
		result.visitedStopNames.Add(stopGroupName)
	}

	result.trips = append(result.trips, &tripRecord{
		tripID:      tripID,
		arrivalTime: arrivalTime,
		travelTime:  travelTime,
		startStopID: startStopID,
		endStopID:   endStopID,
	})

	return result
}

func (t tripSequence) tripCount() uint {
	return uint(len(t.trips))
}

func (t tripSequence) getStartStopID() uint64 {
	return t.trips[0].startStopID
}

func (t tripSequence) addToTravelPlan(travelPlan *TravelPlan) {
	for i, takenTrip := range t.trips {
		if i > 0 {
			travelPlan.addTransfer(t.trips[i-1].endStopID, takenTrip.startStopID)
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
