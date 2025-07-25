package simulation

import (
	"math/rand/v2"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/controlcenter"
)

type tram struct {
	id                  int
	tripData            tripData
	controlCenter       *controlcenter.ControlCenter
	intermediateIndex   int
	latitude            float32
	longitude           float32
	azimuth             float32
	coveredDistance     float32
	distToNextInterNode float32
	departureTime       uint
	isFinished          bool
	state               TramState
}

func newTram(id int, trip *city.TramTrip, controlCenter *controlcenter.ControlCenter) *tram {
	startTime := trip.Stops[0].Time
	return &tram{
		id:            id,
		tripData:      newTripData(trip),
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

func (t *tram) onTripNotStarted(
	time uint,
	stopsById map[uint64]*city.GraphNode,
) (result TramPositionChange, update bool) {
	if time != t.departureTime {
		return
	}

	t.state = StatePassengerLoading
	t.tripData.saveArrival(time)
	t.azimuth = stopsById[t.tripData.trip.Stops[0].ID].Neighbors[0].Azimuth
	t.departureTime = t.tripData.trip.Stops[0].Time

	result = TramPositionChange{
		TramID:    t.id,
		Latitude:  stopsById[t.tripData.trip.Stops[0].ID].Latitude,
		Longitude: stopsById[t.tripData.trip.Stops[0].ID].Longitude,
		Azimuth:   t.azimuth,
	}

	update = true

	return
}

func (t *tram) onPassengerLoading(time uint) {
	if time != t.departureTime {
		return
	}

	t.tripData.saveDeparture(time)
	if t.tripData.index < len(t.tripData.trip.Stops) {
		t.state = StateTravelling
	} else {
		t.state = StatePassengerUnloading
	}
}

func (t *tram) onPassengerUnloading(time uint) {
	if t.tripData.index == len(t.tripData.trip.Stops)-1 {
		t.tripData.saveDeparture(time)
		t.state = StateTripFinished
	} else {
		t.state = StatePassengerLoading
	}
}

func (t *tram) onTripFinished() (result TramPositionChange, update bool) {
	if t.isFinished {
		return
	}

	result = TramPositionChange{
		TramID: t.id,
	}
	update = true
	return
}

func (t *tram) onTravelling(time uint, distanceToDrive float32) (result TramPositionChange, update bool) {
	previousStop := t.tripData.trip.Stops[t.tripData.index-1]
	nextStop := t.tripData.trip.Stops[t.tripData.index]
	path := t.controlCenter.GetRouteBetweenNodes(previousStop.ID, nextStop.ID)

	if t.distToNextInterNode != 0 {
		t.setAzimuthAndDistanceToNextNode(path)
	}

	t.findNewLocation(path, distanceToDrive)
	if t.intermediateIndex == len(path)-1 {
		// Tram arrived to the next stop
		t.tripData.saveArrival(time)
		t.intermediateIndex = 0
		t.departureTime = max(nextStop.Time, time+uint(rand.IntN(11))+15)
		t.state = StatePassengerUnloading
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
	// 50 is the velocity, and *5/18 is used to convert velocity from km/h to m/s
	distanceToDrive := float32(50*5) / float32(18)

	switch t.state {
	case StateTripNotStarted:
		result, update = t.onTripNotStarted(time, stopsById)
	case StatePassengerLoading:
		t.onPassengerLoading(time)
	case StatePassengerUnloading:
		t.onPassengerUnloading(time)
	case StateTravelling:
		result, update = t.onTravelling(time, distanceToDrive)
	case StateTripFinished:
		result, update = t.onTripFinished()
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
	for _, neighbor := range path[t.intermediateIndex].Neighbors {
		if neighbor.ID == path[t.intermediateIndex+1].ID {
			t.azimuth = neighbor.Azimuth
			t.distToNextInterNode = neighbor.Length
			return
		}
	}
}

type TramDetails struct {
	Route        string              `json:"route"`
	TripHeadSign string              `json:"trip_head_sign"`
	TripIndex    int                 `json:"trip_index"`
	Stops        []city.TramTripStop `json:"stops"`
	Arrivals     []uint              `json:"arrivals"`
	Departures   []uint              `json:"departures"`
	StopNames    []string            `json:"stop_names"`
	Speed        uint8               `json:"speed"`
}

func (t *tram) GetDetails(c *city.City) TramDetails {
	stopsByID := c.GetStopsByID()
	stopNames := make([]string, len(t.tripData.trip.Stops))

	for i, stop := range t.tripData.trip.Stops {
		stopNames[i] = *stopsByID[stop.ID].Name
	}

	var tramSpeed uint8
	if t.state == StateTravelling {
		tramSpeed = 50
	}

	return TramDetails{
		Route:        t.tripData.trip.Route,
		TripHeadSign: t.tripData.trip.TripHeadSign,
		TripIndex:    t.tripData.index,
		Stops:        t.tripData.trip.Stops,
		Arrivals:     t.tripData.arrivals,
		Departures:   t.tripData.departures,
		StopNames:    stopNames,
		Speed:        tramSpeed,
	}
}
