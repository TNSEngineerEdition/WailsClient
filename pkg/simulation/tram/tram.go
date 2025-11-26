package tram

import (
	"math"
	"math/rand/v2"

	"github.com/TNSEngineerEdition/WailsClient/pkg/api"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/graph"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/trip"
	"github.com/TNSEngineerEdition/WailsClient/pkg/controlcenter"
	"github.com/TNSEngineerEdition/WailsClient/pkg/simulation/passenger"
)

const MAX_ACCELERATION = 1.0

type Tram struct {
	ID                  uint
	pathIndex           int
	speed, length       float32
	lat, lon, azimuth   float32
	distToNextInterNode float32
	Route               *trip.TramRoute
	TripDetails         tripDetails
	controlCenter       *controlcenter.ControlCenter
	blockedNodesBehind  []graph.GraphNode
	departureTime       uint
	isFinished          bool
	state               TramState
	prevState           TramState
	passengersInTram    map[uint64]*passenger.Passenger
	passengersStore     *passenger.PassengersStore
}

func NewTram(
	id uint,
	route *trip.TramRoute,
	trip *trip.TramTrip,
	controlCenter *controlcenter.ControlCenter,
	passengersStore *passenger.PassengersStore,
) *Tram {
	startTime := uint(trip.Stops[0].Time)
	return &Tram{
		ID:               id,
		length:           30,
		Route:            route,
		TripDetails:      newTripDetails(trip),
		departureTime:    startTime - uint(rand.IntN(11)) - 15,
		state:            StateTripNotStarted,
		controlCenter:    controlCenter,
		passengersStore:  passengersStore,
		passengersInTram: make(map[uint64]*passenger.Passenger),
	}
}

func (t *Tram) Advance(time uint, stopsByID map[uint64]*graph.GraphTramStop) (result TramPositionChange, update bool) {
	switch t.state {
	case StateTripNotStarted:
		result, update = t.onTripNotStarted(time, stopsByID)
	case StatePassengersBoarding:
		t.onPassengersBoarding(time)
	case StatePassengersDisembarking:
		t.onPassengersDisembarking(time)
	case StateTravelling, StateStopping:
		result, update = t.onTravelling(time)
	case StateTripFinished:
		result, update = t.onTripFinished()
	}
	return
}

func (t *Tram) IsAtStop() bool {
	if t.state == StateStopped {
		return t.prevState == StatePassengersBoarding || t.prevState == StatePassengersDisembarking
	}

	return t.state == StatePassengersBoarding || t.state == StatePassengersDisembarking
}

func (t *Tram) getTravelPath() *controlcenter.Path {
	startStopID, endStopID := 0, 1
	if t.TripDetails.Index > 0 {
		startStopID, endStopID = t.TripDetails.Index-1, t.TripDetails.Index
	}

	previousStop := t.TripDetails.Trip.Stops[startStopID]
	nextStop := t.TripDetails.Trip.Stops[endStopID]

	return t.controlCenter.GetPath(previousStop.ID, nextStop.ID)
}

func (t *Tram) findNewLocation(path []graph.GraphNode, distanceToDrive float32) {
	for distanceToDrive > 0 && t.pathIndex < len(path)-1 {
		if t.distToNextInterNode == 0 {
			t.setAzimuthAndDistanceToNextNode(path)
		}

		if t.distToNextInterNode <= distanceToDrive {
			distanceToDrive -= t.distToNextInterNode
			t.pathIndex++
			t.blockedNodesBehind = append(t.blockedNodesBehind, path[t.pathIndex])
			t.distToNextInterNode = 0
			t.lat, t.lon = path[t.pathIndex].GetCoordinates()
		} else {
			remainingPart := distanceToDrive / t.distToNextInterNode
			t.distToNextInterNode -= distanceToDrive
			t.findIntermediateLocation(path, remainingPart)
			distanceToDrive = 0
		}
	}
}

func (t *Tram) findIntermediateLocation(path []graph.GraphNode, remainingPart float32) {
	currentLat, currentLon := path[t.pathIndex].GetCoordinates()
	nextLat, nextLon := path[t.pathIndex+1].GetCoordinates()

	vectorLat := nextLat - currentLat
	vectorLon := nextLon - currentLon
	t.lat = currentLat + vectorLat*remainingPart
	t.lon = currentLon + vectorLon*remainingPart
}

func (t *Tram) setAzimuthAndDistanceToNextNode(path []graph.GraphNode) {
	neighbors := path[t.pathIndex].GetNeighbors()

	if nextNode, ok := neighbors[path[t.pathIndex+1].GetID()]; ok {
		t.azimuth = nextNode.Azimuth
		t.distToNextInterNode = nextNode.Distance
	}
}

func (t *Tram) getDistanceToNeighbor(v graph.GraphNode, u graph.GraphNode) float32 {
	if neighbor, ok := v.GetNeighbors()[u.GetID()]; ok {
		return neighbor.Distance
	} else if neighbor, ok := u.GetNeighbors()[v.GetID()]; ok {
		return neighbor.Distance
	} else {
		panic("Distance between nodes not found")
	}
}

func (t *Tram) nextNodeDistance(path []graph.GraphNode, i int) float32 {
	if i == t.pathIndex && t.distToNextInterNode > 0 {
		return t.distToNextInterNode
	}

	return t.getDistanceToNeighbor(path[i], path[i+1])
}

func (t *Tram) blockNodesBehind() {
	if len(t.blockedNodesBehind) == 0 {
		return
	}
	idx := len(t.blockedNodesBehind) - 1

	// block current position of a tram marker
	u := t.blockedNodesBehind[idx]
	u.TryBlocking(t.ID)
	idx--

	// block nodes behind a tram marker simulating tram length
	distanceLeft := t.length
	for distanceLeft > 0 && idx >= 0 {
		v := t.blockedNodesBehind[idx]
		distanceLeft -= t.getDistanceToNeighbor(v, u)
		v.TryBlocking(t.ID)
		u = v
		idx--
	}

	// unblock (and remove from the slice) nodes left behind by a tram
	p := idx + 1
	for idx >= 0 {
		t.blockedNodesBehind[idx].Unblock(t.ID)
		idx--
	}
	t.blockedNodesBehind = t.blockedNodesBehind[p:]
}

func (t *Tram) unblockNodesBehind() {
	for _, node := range t.blockedNodesBehind {
		node.Unblock(t.ID)
	}
}

func (t *Tram) unblockNodesAhead() {
	path := t.getTravelPath()
	for i := t.pathIndex; i < len(path.Nodes)-1; i++ {
		path.Nodes[i+1].Unblock(t.ID)
	}
}

func (t *Tram) GetEstimatedArrival(stopIndex int, time uint) uint {
	if t.TripDetails.Index > stopIndex || t.TripDetails.Index == stopIndex && t.IsAtStop() {
		return t.TripDetails.Arrivals[stopIndex]
	}

	// For not yet started trips, default to scheduled departure time
	lastDeparture := t.TripDetails.Trip.Stops[0].Time
	if t.TripDetails.Index > 0 {
		lastDeparture = t.TripDetails.Departures[t.TripDetails.Index-1]
	}

	pathDistanceProgress := t.getTravelPath().GetProgressForIndex(t.pathIndex)

	if t.TripDetails.Index == 0 || stopIndex == 0 || pathDistanceProgress == 0 {
		return lastDeparture + t.TripDetails.Trip.GetScheduledTravelTime(t.TripDetails.Index, stopIndex)
	}

	timeSinceLastDeparture := float64(time - lastDeparture)
	estimatedTravelTimeToNextStop := uint(math.Round(timeSinceLastDeparture / float64(pathDistanceProgress)))
	estimatedArrivalToNextStop := lastDeparture + estimatedTravelTimeToNextStop

	if t.TripDetails.Index == stopIndex {
		return estimatedArrivalToNextStop
	}

	var estimatedPositiveDelay uint
	if estimatedArrivalToNextStop > t.TripDetails.Trip.Stops[t.TripDetails.Index].Time {
		estimatedPositiveDelay = estimatedArrivalToNextStop - t.TripDetails.Trip.Stops[t.TripDetails.Index].Time
	}

	scheduledTravelTime := t.TripDetails.Trip.GetScheduledTravelTime(t.TripDetails.Index, stopIndex)
	return t.TripDetails.Trip.Stops[t.TripDetails.Index].Time + scheduledTravelTime + estimatedPositiveDelay
}

type TramPositionChange struct {
	TramID  uint      `json:"id"`
	Lat     float32   `json:"lat"`
	Lon     float32   `json:"lon"`
	Azimuth float32   `json:"azimuth"`
	State   TramState `json:"state"`
}

// Guarantees smooth arrival and deceleration to another tram, stop or a section
// with a lower speed limit by solving a quadratic equation whose result is the new speed.
// Returns new speed.
func (t *Tram) handleDeceleration(targetDistance, targetSpeed, maxSpeed float32) float32 {
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

func (t *Tram) getBlockingDistance(speed float32) float32 {
	return speed + speed*speed/(2*MAX_ACCELERATION) + 2*t.length
}

func (t *Tram) extendReservedDistance(reservedDistance, neededDistance, distanceToNextNode float32) float32 {
	if reservedDistance+distanceToNextNode <= neededDistance {
		reservedDistance += distanceToNextNode
	} else {
		reservedDistance = neededDistance
	}
	return reservedDistance
}

func (t *Tram) updateSpeedAndReserveNodes(path *controlcenter.Path) (availableDistance float32) {
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

		if !u.TryBlocking(t.ID) {
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
			t.extendReservedDistance(
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

	if t.state == StateStopping && (distToStop == 0 || 1e-3 < distToStop) {
		distToStop = 1e-3
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

func (t *Tram) getSpeed() uint8 {
	speedKPH := float64((t.speed * 18) / 5)
	return uint8(math.Round(speedKPH))
}

type TramDetails struct {
	Route           string                     `json:"route"`
	TripHeadSign    string                     `json:"trip_head_sign"`
	TripIndex       int                        `json:"trip_index"`
	Stops           []api.ResponseTramTripStop `json:"stops"`
	Arrivals        []uint                     `json:"arrivals"`
	Departures      []uint                     `json:"departures"`
	StopNames       []string                   `json:"stop_names"`
	Speed           uint8                      `json:"speed"`
	State           TramState                  `json:"state"`
	PassengersCount uint                       `json:"passengers_count"`
}

func (t *Tram) GetPassengerCount() uint {
	return uint(len(t.passengersInTram))
}

func (t *Tram) GetDetails(c *city.City, time uint) TramDetails {
	stopsByID := c.GetStopsByID()
	stopNames := make([]string, len(t.TripDetails.Trip.Stops))

	for i, stop := range t.TripDetails.Trip.Stops {
		stopNames[i] = stopsByID[stop.ID].GetName()
	}

	if t.state != StateTripFinished && t.TripDetails.Index < len(t.TripDetails.Arrivals) {
		t.TripDetails.Arrivals[t.TripDetails.Index] = t.GetEstimatedArrival(t.TripDetails.Index, time)
	}

	return TramDetails{
		Route:           t.Route.Name,
		TripHeadSign:    t.TripDetails.Trip.TripHeadSign,
		TripIndex:       t.TripDetails.Index,
		Stops:           t.TripDetails.Trip.Stops,
		Arrivals:        t.TripDetails.Arrivals,
		Departures:      t.TripDetails.Departures,
		StopNames:       stopNames,
		Speed:           t.getSpeed(),
		State:           t.state,
		PassengersCount: t.GetPassengerCount(),
	}
}

func (t *Tram) IsStopped() bool {
	return t.state == StateStopped || t.state == StateStopping
}

func (t *Tram) StopTram() {
	switch t.state {
	case StateTravelling, StateStopping:
		t.prevState = t.state
		t.state = StateStopping
	case StatePassengersBoarding, StatePassengersDisembarking:
		t.prevState = t.state
		t.state = StateStopped
		t.unblockNodesAhead()
	}
}

func (t *Tram) ResumeTram(currentTime uint) {
	switch t.prevState {
	case StatePassengersBoarding:
		t.state = StatePassengersBoarding
		if t.departureTime < currentTime {
			t.departureTime = currentTime + 1
		}
	case StatePassengersDisembarking:
		t.state = StatePassengersDisembarking
	default:
		t.state = StateTravelling
	}
}
