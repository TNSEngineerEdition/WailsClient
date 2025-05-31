package simulation

import (
	"math/rand/v2"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/controlcenter"
)

type tram struct {
	id                  int
	trip                *city.TramTrip
	tripIndex           int
	intermediateIndex   int
	latitude            float32
	longitude           float32
	azimuth             float32
	coveredDistance     float32
	distToNextInterNode float32
	departureTime       uint
	isFinished          bool
	state               TramState
	controlCenter       *controlcenter.ControlCenter
}

func newTram(id int, trip *city.TramTrip, controlCenter *controlcenter.ControlCenter) *tram {
	startTime := trip.Stops[0].Time
	return &tram{
		id:            id,
		trip:          trip,
		departureTime: startTime - uint(rand.IntN(11)) - 15,
		state:         StateTripNotStarted,
		controlCenter: controlCenter,
	}
}

type TramPositionChange struct {
	TramID    int     `json:"id"`
	Latitude  float32 `json:"lat"`
	Longitude float32 `json:"lon"`
	Azimuth   float32 `json:"azimuth"`
}

func (t *tram) handleTripNotStarted(
	time uint,
	stopsById map[uint64]*city.GraphNode,
) (result TramPositionChange, update bool) {
	if time == t.departureTime {
		t.state = StatePassengerTransfer
		t.azimuth = stopsById[t.trip.Stops[0].ID].Neighbors[0].Azimuth
		result = TramPositionChange{
			TramID:    t.id,
			Latitude:  stopsById[t.trip.Stops[0].ID].Latitude,
			Longitude: stopsById[t.trip.Stops[0].ID].Longitude,
			Azimuth:   t.azimuth,
		}
		t.departureTime = t.trip.Stops[0].Time
		update = true
	}
	return
}

func (t *tram) handlePassangerTransfer(time uint) {
	if time != t.departureTime {
		return
	}

	if t.tripIndex == len(t.trip.Stops)-1 {
		t.state = StateTripFinished
	} else {
		t.state = StateTravelling
	}
}

func (t *tram) handleTripFinished() (result TramPositionChange, update bool) {
	if t.isFinished {
		return
	}

	result = TramPositionChange{
		TramID: t.id,
	}
	update = true
	return
}

func (t *tram) handleTravelling(time uint, distanceToDrive float32) (result TramPositionChange, update bool) {
	currentStop := t.trip.Stops[t.tripIndex]
	nextStop := t.trip.Stops[t.tripIndex+1]
	path := t.controlCenter.GetRouteBetweenNodes(currentStop.ID, nextStop.ID)

	if t.distToNextInterNode != 0 {
		t.setAzimuthAndDistanceToNextNode(path)
	}

	t.findNewLocation(path, distanceToDrive)
	if t.intermediateIndex == len(path)-1 {
		t.tripIndex += 1
		t.intermediateIndex = 0
		t.departureTime = max(nextStop.Time, time+uint(rand.IntN(11))+15)
		t.state = StatePassengerTransfer
	}

	result = TramPositionChange{
		TramID:    t.id,
		Latitude:  t.latitude,
		Longitude: t.longitude,
		Azimuth:   t.azimuth,
	}
	update = true

	return
}

func (t *tram) Advance(time uint, stopsById map[uint64]*city.GraphNode) (result TramPositionChange, update bool) {
	//50 is the velocity, and *5/18 is used to convert velocity from km/h to m/s
	distanceToDrive := float32(50*5) / float32(18)
	switch t.state {
	case StateTripNotStarted:
		result, update = t.handleTripNotStarted(time, stopsById)

	case StatePassengerTransfer:
		t.handlePassangerTransfer(time)

	case StateTravelling:
		result, update = t.handleTravelling(time, distanceToDrive)

	case StateTripFinished:
		result, update = t.handleTripFinished()

	}

	return result, update
}

func (t *tram) findNewLocation(path []*city.GraphNode, distanceToDrive float32) {
	for distanceToDrive > 0 && t.intermediateIndex < len(path)-1 {
		t.setAzimuthAndDistanceToNextNode(path)
		if t.distToNextInterNode <= distanceToDrive {
			distanceToDrive -= t.distToNextInterNode
			t.coveredDistance += t.distToNextInterNode
			t.intermediateIndex += 1
			t.distToNextInterNode = 0
			t.latitude = path[t.intermediateIndex].Latitude
			t.longitude = path[t.intermediateIndex].Longitude
		} else {
			remainingPart := distanceToDrive / t.distToNextInterNode
			t.coveredDistance += distanceToDrive
			t.distToNextInterNode -= distanceToDrive
			t.findIntermediateLocation(path, remainingPart)
			distanceToDrive = 0
		}
	}
}

func (t *tram) findIntermediateLocation(path []*city.GraphNode, remainingPart float32) {
	vectorLat := path[t.intermediateIndex+1].Latitude - path[t.intermediateIndex].Latitude
	vectorLon := path[t.intermediateIndex+1].Longitude - path[t.intermediateIndex].Longitude
	t.latitude = path[t.intermediateIndex].Latitude + vectorLat*remainingPart
	t.longitude = path[t.intermediateIndex].Longitude + vectorLon*remainingPart
}

func (t *tram) setAzimuthAndDistanceToNextNode(path []*city.GraphNode) {
	for _, neigbor := range path[t.intermediateIndex].Neighbors {
		if neigbor.ID == path[t.intermediateIndex+1].ID {
			t.azimuth = neigbor.Azimuth
			t.distToNextInterNode = neigbor.Length
			return
		}
	}
}

type TramDetails struct {
	Route        string              `json:"route"`
	TripHeadSign string              `json:"trip_head_sign"`
	TripIndex    int                 `json:"trip_index"`
	Stops        []city.TramTripStop `json:"stops"`
	StopNames    []string            `json:"stop_names"`
	Speed        uint8               `json:"speed"`
	Delay        int                 `json:"delay"`
}

func (t *tram) GetDetails(c *city.City) TramDetails {
	stopsByID := c.GetStopsByID()
	stopNames := make([]string, len(t.trip.Stops))

	for i, stop := range t.trip.Stops {
		stopNames[i] = *stopsByID[stop.ID].Name
	}

	var tramSpeed uint8

	if t.state == StateTravelling {
		tramSpeed = 50
	}

	return TramDetails{
		Route:        t.trip.Route,
		TripHeadSign: t.trip.TripHeadSign,
		TripIndex:    t.tripIndex,
		Stops:        t.trip.Stops,
		StopNames:    stopNames,
		Speed:        tramSpeed,
		Delay:        0,
	}
}
