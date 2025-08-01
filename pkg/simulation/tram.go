package simulation

import (
	"math/rand/v2"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/controlcenter"
)

type tram struct {
	id                  int
	speed               uint8
	length              float32
	trip                *city.TramTrip
	tripIndex           int
	pathIndex           int
	blockedNodesBehind  []*city.GraphNode
	latitude            float32
	longitude           float32
	azimuth             float32
	distToNextInterNode float32
	departureTime       uint
	state               TramState
	controlCenter       *controlcenter.ControlCenter
}

func newTram(id int, trip *city.TramTrip, controlCenter *controlcenter.ControlCenter) *tram {
	startTime := trip.Stops[0].Time
	return &tram{
		id:                 id,
		speed:              50,
		length:             30,
		trip:               trip,
		departureTime:      startTime - uint(rand.IntN(11)) - 15,
		blockedNodesBehind: make([]*city.GraphNode, 0),
		state:              StateTripNotStarted,
		controlCenter:      controlCenter,
	}
}

type TramPositionChange struct {
	TramID    int     `json:"id"`
	Latitude  float32 `json:"lat"`
	Longitude float32 `json:"lon"`
	Azimuth   float32 `json:"azimuth"`
}

func (t *tram) Advance(time uint, stopsById map[uint64]*city.GraphNode) (result TramPositionChange, update bool) {
	switch t.state {
	case StateTripNotStarted:
		result, update = t.handleTripNotStarted(time, stopsById)

	case StatePassengerTransfer:
		t.handlePassangerTransfer(time)

	case StateTravelling:
		result, update = t.handleTravelling(time)

	case StateTripFinished:
		result, update = t.handleTripFinished()
	}

	return result, update
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
	t.speed = 0

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
	t.unblockAllBlockedNodes()
	result = TramPositionChange{
		TramID: t.id,
	}
	update = true
	return
}

func (t *tram) handleTravelling(time uint) (result TramPositionChange, update bool) {
	//50 is the velocity, and *5/18 is used to convert velocity from km/h to m/s
	t.speed = 50
	distanceToDrive := float32(t.speed*5) / float32(18)

	currentStop := t.trip.Stops[t.tripIndex]
	nextStop := t.trip.Stops[t.tripIndex+1]
	path := t.controlCenter.GetRouteBetweenNodes(currentStop.ID, nextStop.ID)

	if t.distToNextInterNode != 0 {
		t.setAzimuthAndDistanceToNextNode(path)
	}

	t.findNewLocation(path, distanceToDrive)
	t.blockNodesBehind()

	if t.pathIndex == len(path)-1 {
		t.tripIndex += 1
		t.pathIndex = 0
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

func (t *tram) findNewLocation(path []*city.GraphNode, distanceToDrive float32) {
	for distanceToDrive > 0 && t.pathIndex < len(path)-1 {
		t.setAzimuthAndDistanceToNextNode(path)

		if path[t.pathIndex+1].IsBlocked() {
			break
		}

		if t.distToNextInterNode <= distanceToDrive {
			distanceToDrive -= t.distToNextInterNode
			t.pathIndex += 1
			t.blockedNodesBehind = append(t.blockedNodesBehind, path[t.pathIndex])
			t.distToNextInterNode = 0
			t.latitude = path[t.pathIndex].Latitude
			t.longitude = path[t.pathIndex].Longitude
		} else {
			remainingPart := distanceToDrive / t.distToNextInterNode
			t.distToNextInterNode -= distanceToDrive
			t.findIntermediateLocation(path, remainingPart)
			distanceToDrive = 0
		}
	}
}

func (t *tram) findIntermediateLocation(path []*city.GraphNode, remainingPart float32) {
	vectorLat := path[t.pathIndex+1].Latitude - path[t.pathIndex].Latitude
	vectorLon := path[t.pathIndex+1].Longitude - path[t.pathIndex].Longitude
	t.latitude = path[t.pathIndex].Latitude + vectorLat*remainingPart
	t.longitude = path[t.pathIndex].Longitude + vectorLon*remainingPart
}

func (t *tram) setAzimuthAndDistanceToNextNode(path []*city.GraphNode) {
	for _, neigbor := range path[t.pathIndex].Neighbors {
		if neigbor.ID == path[t.pathIndex+1].ID {
			t.azimuth = neigbor.Azimuth
			t.distToNextInterNode = neigbor.Distance
			return
		}
	}
}

func (t *tram) getDistanceBetweenNeighboringNodes(v *city.GraphNode, u *city.GraphNode) float32 {
	for _, neigbor := range v.Neighbors {
		if neigbor.ID == u.ID {
			return neigbor.Distance
		}
	}
	for _, neigbor := range u.Neighbors {
		if neigbor.ID == v.ID {
			return neigbor.Distance
		}
	}
	return 0
}

func (t *tram) blockNodesBehind() {
	if len(t.blockedNodesBehind) == 0 {
		return
	}

	idx := len(t.blockedNodesBehind) - 1
	u := t.blockedNodesBehind[idx]
	u.Block(t.id)
	idx--
	distanceLeft := t.length

	for distanceLeft > 0 && idx >= 0 {
		v := t.blockedNodesBehind[idx]
		distanceLeft -= t.getDistanceBetweenNeighboringNodes(v, u)
		v.Block(t.id)
		u = v
		idx--
	}

	p := idx + 1
	for idx >= 0 {
		t.blockedNodesBehind[idx].Unblock(t.id)
		idx--
	}
	t.blockedNodesBehind = t.blockedNodesBehind[p:]
}

func (t *tram) unblockAllBlockedNodes() {
	for _, node := range t.blockedNodesBehind {
		node.Unblock(t.id)
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

	return TramDetails{
		Route:        t.trip.Route,
		TripHeadSign: t.trip.TripHeadSign,
		TripIndex:    t.tripIndex,
		Stops:        t.trip.Stops,
		StopNames:    stopNames,
		Speed:        t.speed,
		Delay:        0,
	}
}
