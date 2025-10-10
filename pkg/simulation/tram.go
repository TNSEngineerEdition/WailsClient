package simulation

import (
	"math"
	"math/rand/v2"

	"github.com/TNSEngineerEdition/WailsClient/pkg/api"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/graph"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/trip"
	"github.com/TNSEngineerEdition/WailsClient/pkg/controlcenter"
)

type tram struct {
	id                                 uint
	pathIndex                          int
	speed, length, distToNextInterNode float32
	latitude, longitude, azimuth       float32
	route                              *trip.TramRoute
	tripData                           tripData
	controlCenter                      *controlcenter.ControlCenter
	blockedNodesBehind                 []graph.GraphNode
	departureTime                      uint
	isFinished                         bool
	state                              TramState
}

const MAX_ACCELERATION = 1.5

func newTram(
	id uint,
	route *trip.TramRoute,
	trip *trip.TramTrip,
	controlCenter *controlcenter.ControlCenter,
) *tram {
	startTime := uint(trip.Stops[0].Time)
	return &tram{
		id:            id,
		length:        30,
		route:         route,
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

func (t *tram) findNewLocation(path []graph.GraphNode, distanceToDrive float32) {
	for distanceToDrive > 0 && t.pathIndex < len(path)-1 {
		if t.distToNextInterNode == 0 {
			t.setAzimuthAndDistanceToNextNode(path)
		}

		if t.distToNextInterNode <= distanceToDrive {
			distanceToDrive -= t.distToNextInterNode
			t.pathIndex++
			t.blockedNodesBehind = append(t.blockedNodesBehind, path[t.pathIndex])
			t.distToNextInterNode = 0
			t.latitude, t.longitude = path[t.pathIndex].GetCoordinates()
		} else {
			remainingPart := distanceToDrive / t.distToNextInterNode
			t.distToNextInterNode -= distanceToDrive
			t.findIntermediateLocation(path, remainingPart)
			distanceToDrive = 0
		}
	}
}

func (t *tram) findIntermediateLocation(path []graph.GraphNode, remainingPart float32) {
	currentLat, currentLon := path[t.pathIndex].GetCoordinates()
	nextLat, nextLon := path[t.pathIndex+1].GetCoordinates()

	vectorLat := nextLat - currentLat
	vectorLon := nextLon - currentLon
	t.latitude = currentLat + vectorLat*remainingPart
	t.longitude = nextLat + vectorLon*remainingPart
}

func (t *tram) setAzimuthAndDistanceToNextNode(path []graph.GraphNode) {
	neighbors := path[t.pathIndex].GetNeighbors()

	if nextNode, ok := neighbors[path[t.pathIndex+1].GetID()]; ok {
		t.azimuth = nextNode.Azimuth
		t.distToNextInterNode = nextNode.Distance
	}
}

func (t *tram) getDistanceToNeighbor(v graph.GraphNode, u graph.GraphNode) float32 {
	if neighbor, ok := v.GetNeighbors()[u.GetID()]; ok {
		return neighbor.Distance
	} else if neighbor, ok := u.GetNeighbors()[v.GetID()]; ok {
		return neighbor.Distance
	} else {
		panic("Distance between nodes not found")
	}
}

func (t *tram) nextNodeDistance(path []graph.GraphNode, i int) float32 {
	if i == t.pathIndex && t.distToNextInterNode > 0 {
		return t.distToNextInterNode
	}

	return t.getDistanceToNeighbor(path[i], path[i+1])
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

	if t.tripData.index == 0 || stopIndex == 0 || pathDistanceProgress == 0 {
		return lastDeparture + t.tripData.trip.GetScheduledTravelTime(t.tripData.index, stopIndex)
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
	TramID    uint    `json:"id"`
	Latitude  float32 `json:"lat"`
	Longitude float32 `json:"lon"`
	Azimuth   float32 `json:"azimuth"`
}

func (t *tram) onTripNotStarted(
	time uint,
	stopsByID map[uint64]*graph.GraphTramStop,
) (result TramPositionChange, update bool) {
	if time != t.departureTime {
		return
	}

	t.state = StatePassengerLoading
	t.tripData.saveArrival(time)
	t.azimuth = stopsByID[t.tripData.trip.Stops[0].ID].GetNeighbors()[0].Azimuth
	t.departureTime = t.tripData.trip.Stops[0].Time

	lat, lon := stopsByID[t.tripData.trip.Stops[0].ID].GetCoordinates()

	result = TramPositionChange{
		TramID:    t.id,
		Latitude:  lat,
		Longitude: lon,
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

// Guarantees smooth arrival and deceleration to another tram, stop or a section
// with a lower speed limit by solving a quadratic equation whose result is the new speed.
// Returns new speed.
func (t *tram) handleDeceleration(targetDistance, targetSpeed, maxSpeed float32) float32 {
	// (v0+v1target)/2 + v1target^2/(2a) = targetDistance =>
	// v1target^2 + v1target*a + v0*a - 2*a*targetDistance = 0
	A := 1.0
	B := float64(MAX_ACCELERATION)
	C := float64(MAX_ACCELERATION * (t.speed - 2*targetDistance))
	// sometimes delta < 0 due to numerical errors
	delta := max(0, B*B-4*A*C)
	v1target := float32((-B + math.Sqrt(delta)) / (2 * A))

	v1min := max(t.speed-MAX_ACCELERATION, targetSpeed) // do not go below target speed
	v1max := min(t.speed+MAX_ACCELERATION, maxSpeed)    // do not exceed max speed

	if v1target < v1min {
		return v1min
	}
	if v1target > v1max {
		return v1max
	}
	return v1target
}

func (t *tram) getBlockingDistance(speed float32) float32 {
	return speed + speed*speed/(2*MAX_ACCELERATION) + 2*t.length
}

func (t *tram) extendReservedDistance(reservedDistance, neededDistance, distanceToNextNode float32) float32 {
	if reservedDistance+distanceToNextNode <= neededDistance {
		reservedDistance += distanceToNextNode
	} else {
		reservedDistance = neededDistance
	}
	return reservedDistance
}

func (t *tram) updateSpeedAndReserveNodes(path *controlcenter.Path) (availableDistance float32) {
	currentMaxSpeed := path.MaxSpeeds[t.pathIndex]
	newSpeed := min(t.speed+MAX_ACCELERATION, currentMaxSpeed)

	neededReserveAtCurrentSpeed := t.getBlockingDistance(t.speed)
	neededReserveIfAccel := t.getBlockingDistance(newSpeed)

	var reservedDistanceAtCurrentSpeed, reservedDistanceIfAccel float32
	var reservedDistanceAhead float32
	var distToStop, distToMaxSpeedChange float32
	var upcomingMaxSpeed float32

	// reserve nodes ahead until we reach a stopping point or have enough reserved distance
	for i := t.pathIndex; i < len(path.Nodes)-1 && reservedDistanceIfAccel < neededReserveIfAccel; i++ {
		u := path.Nodes[i+1]
		distToNextNode := t.nextNodeDistance(path.Nodes, i)

		// set distance to upcoming speed limit change (if the speed limit is lower)
		if path.MaxSpeeds[i+1] < currentMaxSpeed && distToMaxSpeedChange == 0 {
			upcomingMaxSpeed = path.MaxSpeeds[i+1]
			distToMaxSpeedChange = reservedDistanceAhead
		}

		if !u.TryBlocking(t.id) {
			distToStop = reservedDistanceAhead
			break
		}

		if u.IsTramStop() {
			reservedDistanceAhead += distToNextNode
			distToStop = reservedDistanceAhead

			reservedDistanceIfAccel = t.extendReservedDistance(
				reservedDistanceIfAccel,
				neededReserveIfAccel,
				distToNextNode,
			)
			_ = t.extendReservedDistance(
				reservedDistanceAtCurrentSpeed,
				neededReserveAtCurrentSpeed,
				distToNextNode,
			)
			break
		}

		reservedDistanceAhead += distToNextNode

		reservedDistanceIfAccel = t.extendReservedDistance(
			reservedDistanceIfAccel,
			neededReserveIfAccel,
			distToNextNode,
		)
		reservedDistanceAtCurrentSpeed = t.extendReservedDistance(
			reservedDistanceAtCurrentSpeed,
			neededReserveAtCurrentSpeed,
			distToNextNode,
		)
	}

	var nextSpeed float32
	if distToStop > 0 {
		nextSpeed = t.handleDeceleration(distToStop, 0, currentMaxSpeed)
	} else if distToMaxSpeedChange > 0 {
		nextSpeed = t.handleDeceleration(distToMaxSpeedChange, upcomingMaxSpeed, currentMaxSpeed)
	} else {
		canAccelerate := (reservedDistanceIfAccel >= neededReserveIfAccel)

		if canAccelerate {
			nextSpeed = newSpeed
		} else {
			// handles situation when tram is waiting for free node
			nextSpeed = 0
		}
	}

	//this is the distance the tram will actually travel (consulting changing speed)
	distance := (nextSpeed + t.speed) * 0.5
	t.speed = nextSpeed

	return distance
}

func (t *tram) onTravelling(time uint) (result TramPositionChange, update bool) {
	path := t.getTravelPath()

	if t.distToNextInterNode == 0 {
		t.setAzimuthAndDistanceToNextNode(path.Nodes)
	}

	distanceToDrive := t.updateSpeedAndReserveNodes(path)

	t.findNewLocation(path.Nodes, distanceToDrive)
	t.blockNodesBehind()

	if t.pathIndex == len(path.Nodes)-1 {
		t.tripData.saveArrival(time)
		t.departureTime = max(
			t.tripData.trip.Stops[t.tripData.index].Time,
			time+uint(rand.IntN(11))+15,
		)
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

func (t *tram) Advance(time uint, stopsByID map[uint64]*graph.GraphTramStop) (result TramPositionChange, update bool) {
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
	Route        string                     `json:"route"`
	TripHeadSign string                     `json:"trip_head_sign"`
	TripIndex    int                        `json:"trip_index"`
	Stops        []api.ResponseTramTripStop `json:"stops"`
	Arrivals     []uint                     `json:"arrivals"`
	Departures   []uint                     `json:"departures"`
	StopNames    []string                   `json:"stop_names"`
	Speed        uint8                      `json:"speed"`
}

func (t *tram) GetDetails(c *city.City, time uint) TramDetails {
	stopsByID := c.GetStopsByID()
	stopNames := make([]string, len(t.tripData.trip.Stops))

	for i, stop := range t.tripData.trip.Stops {
		stopNames[i] = stopsByID[stop.ID].GetName()
	}

	t.tripData.arrivals[t.tripData.index] = t.getEstimatedArrival(t.tripData.index, time)

	return TramDetails{
		Route:        t.route.Name,
		TripHeadSign: t.tripData.trip.TripHeadSign,
		TripIndex:    t.tripData.index,
		Stops:        t.tripData.trip.Stops,
		Arrivals:     t.tripData.arrivals,
		Departures:   t.tripData.departures,
		StopNames:    stopNames,
		Speed:        uint8((t.speed * 18) / 5), // m/s -> km/h
	}
}
