package tram

import (
	"math/rand/v2"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city/graph"
)

type TramState uint8

const (
	StateTripNotStarted TramState = iota
	StatePassengersBoarding
	StatePassengersDisembarking
	StateTravelling
	StateTripFinished
	StateStopping
	StateStopped
)

var TramStates = []struct {
	Value  TramState
	TSName string
}{
	{StateTripNotStarted, "TRIP_NOT_STARTED"},
	{StatePassengersBoarding, "PASSENGERS_BOARDING"},
	{StatePassengersDisembarking, "PASSENGERS_DISEMBARKING"},
	{StateTravelling, "TRAVELLING"},
	{StateTripFinished, "TRIP_FINISHED"},
	{StateStopping, "STOPPING"},
	{StateStopped, "STOPPED"},
}

func (t *Tram) onTripNotStarted(
	time uint,
	stopsByID map[uint64]*graph.GraphTramStop,
) (result TramPositionChange, update bool) {
	if time != t.departureTime {
		return
	}

	t.state = StatePassengersBoarding
	t.TripDetails.saveArrival(time)
	t.departureTime = t.TripDetails.Trip.Stops[0].Time

	// Set azimuth to any neighbor's azimuth
	for _, neighbor := range stopsByID[t.TripDetails.Trip.Stops[0].ID].GetNeighbors() {
		t.azimuth = neighbor.Azimuth
		break
	}

	lat, lon := stopsByID[t.TripDetails.Trip.Stops[0].ID].GetCoordinates()

	result = TramPositionChange{
		TramID:  t.ID,
		Lat:     lat,
		Lon:     lon,
		Azimuth: t.azimuth,
		State:   t.state,
	}

	update = true

	return
}

func (t *Tram) onPassengersBoarding(time uint) {
	isBoardingFinished := t.boardPassengers()

	if !isBoardingFinished || time < t.departureTime {
		return
	}

	t.TripDetails.saveDeparture(time)
	t.pathIndex = 0
	t.state = StateTravelling
}

func (t *Tram) onPassengersDisembarking(time uint) {
	_, isDisembarkingFinished := t.disembarkPassengers()

	if !isDisembarkingFinished {
		return
	}

	if t.TripDetails.Index == len(t.TripDetails.Trip.Stops)-1 {
		t.TripDetails.saveDeparture(time)
		t.state = StateTripFinished
	} else {
		t.state = StatePassengersBoarding
	}
}

func (t *Tram) onTravelling(time uint) (result TramPositionChange, update bool) {
	path := t.getTravelPath()

	if t.distToNextInterNode == 0 {
		t.setAzimuthAndDistanceToNextNode(path.Nodes)
	}

	distanceToDrive := t.updateSpeedAndReserveNodes(path)

	t.findNewLocation(path.Nodes, distanceToDrive)
	t.blockNodesBehind()

	if t.pathIndex == len(path.Nodes)-1 {
		t.TripDetails.saveArrival(time)
		t.departureTime = max(
			t.TripDetails.Trip.Stops[t.TripDetails.Index].Time,
			time+uint(rand.IntN(11))+15,
		)
		if t.state == StateStopping {
			t.prevState = StatePassengersDisembarking
			t.state = StateStopped
			t.unblockNodesAhead()
		} else {
			t.state = StatePassengersDisembarking
		}
	} else if t.state == StateStopping && t.speed <= 0.01 {
		t.prevState = StateTravelling
		t.state = StateStopped
		t.unblockNodesAhead()
	}

	result = TramPositionChange{
		TramID:  t.ID,
		Lat:     t.lat,
		Lon:     t.lon,
		Azimuth: t.azimuth,
		State:   t.state,
	}
	update = true

	return
}

func (t *Tram) onTripFinished() (result TramPositionChange, update bool) {
	if t.isFinished {
		return
	}

	t.isFinished = true
	t.unblockNodesBehind()

	result = TramPositionChange{
		TramID: t.ID,
		State:  t.state,
	}
	update = true

	return
}
