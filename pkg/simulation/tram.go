package simulation

import (
	"math"
	"math/rand/v2"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/controlcenter"
)

type tram struct {
	id, pathIndex                      int
	speed, length, distToNextInterNode float32
	latitude, longitude, azimuth       float32
	tripData                           tripData
	controlCenter                      *controlcenter.ControlCenter
	blockedNodesBehind                 []*city.GraphNode
	departureTime                      uint
	isFinished                         bool
	state                              TramState
}

func newTram(id int, trip *city.TramTrip, controlCenter *controlcenter.ControlCenter) *tram {
	startTime := trip.Stops[0].Time
	return &tram{
		id:            id,
		length:        30,
		tripData:      newTripData(trip),
		departureTime: startTime - uint(rand.IntN(11)) - 15,
		state:         StateTripNotStarted,
		controlCenter: controlCenter,
	}
}

func (t *tram) isAtStop() bool {
	return t.state == StatePassengerLoading ||
		t.state == StatePassengerUnloading
}

func (t *tram) getTravelPath() *controlcenter.Path {
	startStopID, endStopID := 0, 1
	if t.tripData.index > 0 {
		startStopID, endStopID = t.tripData.index-1, t.tripData.index
	}

	previousStop := t.tripData.trip.Stops[startStopID]
	nextStop := t.tripData.trip.Stops[endStopID]

	return t.controlCenter.GetPath(previousStop.ID, nextStop.ID)
}

func (t *tram) findNewLocation(path []*city.GraphNode, distanceToDrive float32) {
	for distanceToDrive > 0 && t.pathIndex < len(path)-1 {
		t.setAzimuthAndDistanceToNextNode(path)

		if t.distToNextInterNode <= distanceToDrive {
			distanceToDrive -= t.distToNextInterNode
			t.pathIndex++
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
	for _, neighbor := range path[t.pathIndex].Neighbors {
		if neighbor.ID == path[t.pathIndex+1].ID {
			t.azimuth = neighbor.Azimuth
			t.distToNextInterNode = neighbor.Distance
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

	path := t.controlCenter.GetPath(
		t.tripData.trip.Stops[t.tripData.index].ID,
		t.tripData.trip.Stops[t.tripData.index+1].ID,
	)

	for _, node := range path.Nodes {
		node.Unblock(t.id)
	}
}

func (t *tram) getEstimatedArrival(stopIndex int, time uint) uint {
	if t.tripData.index > stopIndex || t.tripData.index == stopIndex && t.isAtStop() {
		return t.tripData.arrivals[stopIndex]
	}

	// For not yet started trips, default to scheduled departure time
	lastDeparture := t.tripData.trip.Stops[0].Time
	if t.tripData.index > 0 {
		lastDeparture = t.tripData.departures[t.tripData.index-1]
	}

	pathDistanceProgress := t.getTravelPath().GetProgressForIndex(t.pathIndex)

	if t.tripData.index == 0 || stopIndex == 0 {
		return lastDeparture + t.tripData.trip.GetScheduledTravelTime(0, stopIndex)
	} else if t.tripData.index == stopIndex && pathDistanceProgress == 0 {
		return lastDeparture + t.tripData.trip.GetScheduledTravelTime(stopIndex-1, stopIndex)
	}

	timeSinceLastDeparture := float64(time - lastDeparture)
	estimatedTravelTimeToNextStop := uint(math.Round(timeSinceLastDeparture / float64(pathDistanceProgress)))
	estimatedArrivalToNextStop := lastDeparture + estimatedTravelTimeToNextStop

	if t.tripData.index == stopIndex {
		return estimatedArrivalToNextStop
	}

	var estimatedPositiveDelay uint
	if estimatedArrivalToNextStop > t.tripData.trip.Stops[t.tripData.index].Time {
		estimatedPositiveDelay = estimatedArrivalToNextStop - t.tripData.trip.Stops[t.tripData.index].Time
	}

	scheduledTravelTime := t.tripData.trip.GetScheduledTravelTime(t.tripData.index, stopIndex)
	return t.tripData.trip.Stops[t.tripData.index].Time + scheduledTravelTime + estimatedPositiveDelay
}

type TramPositionChange struct {
	TramID    int     `json:"id"`
	Latitude  float32 `json:"lat"`
	Longitude float32 `json:"lon"`
	Azimuth   float32 `json:"azimuth"`
}

func (t *tram) onTripNotStarted(
	time uint,
	stopsByID map[uint64]*city.GraphNode,
) (result TramPositionChange, update bool) {
	if time != t.departureTime {
		return
	}

	t.state = StatePassengerLoading
	t.tripData.saveArrival(time)
	t.azimuth = stopsByID[t.tripData.trip.Stops[0].ID].Neighbors[0].Azimuth
	t.departureTime = t.tripData.trip.Stops[0].Time

	result = TramPositionChange{
		TramID:    t.id,
		Latitude:  stopsByID[t.tripData.trip.Stops[0].ID].Latitude,
		Longitude: stopsByID[t.tripData.trip.Stops[0].ID].Longitude,
		Azimuth:   t.azimuth,
	}

	update = true

	return
}

func (t *tram) onPassengerLoading(time uint) {
	if time != t.departureTime {
		return
	}

	if len(t.tripData.trip.Stops) <= 1 {
		// Handle trips with a single stop
		t.state = StatePassengerUnloading
	} else {
		t.tripData.saveDeparture(time)
		t.pathIndex = 0
		t.state = StateTravelling
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

	t.isFinished = true
	t.unblockNodesBehind()

	result = TramPositionChange{
		TramID: t.id,
	}
	update = true

	return
}

func (t *tram) onTravelling(time uint) (result TramPositionChange, update bool) {
	t.speed = float32(50*5) / float32(18) // km/h -> m/s

	path := t.getTravelPath()
	if t.distToNextInterNode != 0 {
		t.setAzimuthAndDistanceToNextNode(path.Nodes)
	}

	availableDistance := t.blockNodesAhead(path.Nodes)
	distanceToDrive := min(t.speed, availableDistance)

	t.findNewLocation(path.Nodes, distanceToDrive)
	t.blockNodesBehind()

	if t.pathIndex == len(path.Nodes)-1 {
		t.tripData.saveArrival(time)
		t.departureTime = max(
			t.tripData.trip.Stops[t.tripData.index].Time,
			time+uint(rand.IntN(11))+15,
		)
		t.state = StatePassengerUnloading
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

func (t *tram) Advance(time uint, stopsByID map[uint64]*city.GraphNode) (result TramPositionChange, update bool) {
	switch t.state {
	case StateTripNotStarted:
		result, update = t.onTripNotStarted(time, stopsByID)
	case StatePassengerLoading:
		t.onPassengerLoading(time)
	case StatePassengerUnloading:
		t.onPassengerUnloading(time)
	case StateTravelling:
		result, update = t.onTravelling(time)
	case StateTripFinished:
		result, update = t.onTripFinished()
	}

	return
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

func (t *tram) GetDetails(c *city.City, time uint) TramDetails {
	stopsByID := c.GetStopsByID()
	stopNames := make([]string, len(t.tripData.trip.Stops))

	for i, stop := range t.tripData.trip.Stops {
		stopNames[i] = *stopsByID[stop.ID].Name
	}

	t.tripData.arrivals[t.tripData.index] = t.getEstimatedArrival(t.tripData.index, time)

	return TramDetails{
		Route:        t.tripData.trip.Route,
		TripHeadSign: t.tripData.trip.TripHeadSign,
		TripIndex:    t.tripData.index,
		Stops:        t.tripData.trip.Stops,
		Arrivals:     t.tripData.arrivals,
		Departures:   t.tripData.departures,
		StopNames:    stopNames,
		Speed:        uint8((t.speed * 18) / 5), // m/s -> km/h
	}
}
