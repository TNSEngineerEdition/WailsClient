package simulation

import (
	"math/rand/v2"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/controlcenter"
)

type tram struct {
	id                  int
	speed               float32
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
	isFinished          bool
	state               TramState
	controlCenter       *controlcenter.ControlCenter
}

func newTram(id int, trip *city.TramTrip, controlCenter *controlcenter.ControlCenter) *tram {
	startTime := trip.Stops[0].Time
	return &tram{
		id:                 id,
		speed:              0,
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

	t.isFinished = true
	t.unblockNodesBehind()
	result = TramPositionChange{
		TramID: t.id,
	}
	update = true
	return
}

func (t *tram) handleTravelling(time uint) (result TramPositionChange, update bool) {
	// km/h -> m/s
	t.speed = float32(50*5) / float32(18)

	currentStop := t.trip.Stops[t.tripIndex]
	nextStop := t.trip.Stops[t.tripIndex+1]
	path := t.controlCenter.GetRouteBetweenNodes(currentStop.ID, nextStop.ID)

	if t.distToNextInterNode != 0 {
		t.setAzimuthAndDistanceToNextNode(path)
	}

	availableDistance := t.blockNodesAhead(path)
	distanceToDrive := min(t.speed, availableDistance)
	t.findNewLocation(path, distanceToDrive)
	t.blockNodesBehind()

	if t.pathIndex == len(path)-1 {
		t.tripIndex += 1
		t.pathIndex = 0
		t.departureTime = max(nextStop.Time, time+uint(rand.IntN(11))+15)
		t.state = StatePassengerTransfer
		t.speed = 0
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

func (t *tram) getDistanceToNeighbor(v *city.GraphNode, u *city.GraphNode) float32 {
	for _, neighbor := range v.Neighbors {
		if neighbor.ID == u.ID {
			return neighbor.Distance
		}
	}
	for _, neighbor := range u.Neighbors {
		if neighbor.ID == v.ID {
			return neighbor.Distance
		}
	}
	panic("Distance between nodes not found")
}

func (t *tram) blockNodesAhead(path []*city.GraphNode) (availableDistance float32) {
	deceleration := 1
	stoppingDistance := t.speed * t.speed / float32(2*deceleration)
	maxBlockingDistance := t.speed + stoppingDistance
	i := t.pathIndex

	for availableDistance < maxBlockingDistance && i < len(path)-1 {
		v := path[i]
		u := path[i+1]

		distanceToNextNode := t.getDistanceToNeighbor(v, u)
		if availableDistance+distanceToNextNode <= maxBlockingDistance {
			if !u.TryBlocking(t.id) {
				break
			}
			availableDistance += distanceToNextNode
			i++
		} else {
			availableDistance = maxBlockingDistance
			break
		}
	}

	return availableDistance
}

func (t *tram) blockNodesBehind() {
	if len(t.blockedNodesBehind) == 0 {
		return
	}
	idx := len(t.blockedNodesBehind) - 1

	// block current position of a tram marker
	u := t.blockedNodesBehind[idx]
	u.TryBlocking(t.id)
	idx--

	// block nodes behind a tram marker simulating tram length
	distanceLeft := t.length
	for distanceLeft > 0 && idx >= 0 {
		v := t.blockedNodesBehind[idx]
		distanceLeft -= t.getDistanceToNeighbor(v, u)
		v.TryBlocking(t.id)
		u = v
		idx--
	}

	// unblock (and remove from the slice) nodes left behind by a tram
	p := idx + 1
	for idx >= 0 {
		t.blockedNodesBehind[idx].Unblock(t.id)
		idx--
	}
	t.blockedNodesBehind = t.blockedNodesBehind[p:]
}

func (t *tram) unblockNodesBehind() {
	for _, node := range t.blockedNodesBehind {
		node.Unblock(t.id)
	}
}

func (t *tram) unblockWholePath() {
	t.unblockNodesBehind()
	if t.state != StateTravelling {
		return
	}
	currentStop := t.trip.Stops[t.tripIndex]
	nextStop := t.trip.Stops[t.tripIndex+1]
	path := t.controlCenter.GetRouteBetweenNodes(currentStop.ID, nextStop.ID)
	for _, node := range path {
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
		Speed:        uint8((t.speed * 18) / 5), // m/s -> km/h
		Delay:        0,
	}
}
