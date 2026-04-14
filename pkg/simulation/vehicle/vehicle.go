package vehicle

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

type Vehicle struct {
	ID                  uint
	pathIndex           int
	speed, length       float32
	lat, lon, azimuth   float32
	distToNextInterNode float32
	Route               *trip.Route
	TripDetails         tripDetails
	controlCenter       *controlcenter.ControlCenter
	blockedNodesBehind  []graph.GraphNode
	departureTime       uint
	isFinished          bool
	state               VehicleState
	prevState           VehicleState
	passengersInVehicle map[uint64]*passenger.Passenger
	passengersStore     *passenger.PassengersStore
	useNodeBlocking     bool
}

func NewVehicle(
	id uint,
	route *trip.Route,
	trip *trip.Trip,
	controlCenter *controlcenter.ControlCenter,
	passengersStore *passenger.PassengersStore,
) *Vehicle {
	startTime := uint(trip.Stops[0].Time)
	return &Vehicle{
		ID:                  id,
		length:              30,
		Route:               route,
		TripDetails:         newTripDetails(trip),
		departureTime:       startTime - uint(rand.IntN(11)) - 15,
		state:               StateTripNotStarted,
		controlCenter:       controlCenter,
		passengersStore:     passengersStore,
		passengersInVehicle: make(map[uint64]*passenger.Passenger),
		useNodeBlocking:     route.TransitType == api.Tram,
	}
}

type VehiclePositionChange struct {
	VehicleID uint         `json:"id"`
	Lat       float32      `json:"lat"`
	Lon       float32      `json:"lon"`
	Azimuth   float32      `json:"azimuth"`
	State     VehicleState `json:"state"`
	Delay     uint         `json:"delay"`
}

func (t *Vehicle) Advance(time uint, stopsByID map[uint64]*graph.GraphStop) (result VehiclePositionChange, update bool) {
	switch t.state {
	case StateTripNotStarted:
		result, update = t.onTripNotStarted(time, stopsByID)
	case StatePassengersLoading:
		t.onPassengersLoading(time)
	case StatePassengersUnloading:
		t.onPassengersUnloading(time)
	case StateTravelling, StateStopping:
		result, update = t.onTravelling(time)
	case StateTripFinished:
		result, update = t.onTripFinished()
	}

	result.Delay = t.TripDetails.getDelay(time)
	return
}

func (t *Vehicle) IsAtStop() bool {
	if t.state == StateStopped {
		return t.prevState == StatePassengersLoading || t.prevState == StatePassengersUnloading
	}

	return t.state == StatePassengersLoading || t.state == StatePassengersUnloading
}

func (t *Vehicle) getTravelPath() *controlcenter.Path {
	startStopID, endStopID := 0, 1
	if t.TripDetails.Index > 0 {
		startStopID, endStopID = t.TripDetails.Index-1, t.TripDetails.Index
	}

	previousStop := t.TripDetails.Trip.Stops[startStopID]
	nextStop := t.TripDetails.Trip.Stops[endStopID]

	return t.controlCenter.GetPath(previousStop.ID, nextStop.ID)
}

func (t *Vehicle) findNewLocation(path []graph.GraphNode, distanceToDrive float32) {
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

func (t *Vehicle) findIntermediateLocation(path []graph.GraphNode, remainingPart float32) {
	nextLat, nextLon := path[t.pathIndex+1].GetCoordinates()

	vectorLat := nextLat - t.lat
	vectorLon := nextLon - t.lon
	t.lat += vectorLat * remainingPart
	t.lon += vectorLon * remainingPart
}

func (t *Vehicle) setAzimuthAndDistanceToNextNode(path []graph.GraphNode) {
	neighbors := path[t.pathIndex].GetNeighbors()

	if nextNode, ok := neighbors[path[t.pathIndex+1].GetID()]; ok {
		t.azimuth = nextNode.Azimuth
		t.distToNextInterNode = nextNode.Distance
	}
}

func (t *Vehicle) getDistanceToNeighbor(v graph.GraphNode, u graph.GraphNode) float32 {
	if neighbor, ok := v.GetNeighbors()[u.GetID()]; ok {
		return neighbor.Distance
	} else if neighbor, ok := u.GetNeighbors()[v.GetID()]; ok {
		return neighbor.Distance
	} else {
		panic("Distance between nodes not found")
	}
}

func (t *Vehicle) nextNodeDistance(path []graph.GraphNode, i int) float32 {
	if i == t.pathIndex && t.distToNextInterNode > 0 {
		return t.distToNextInterNode
	}

	return t.getDistanceToNeighbor(path[i], path[i+1])
}

func (t *Vehicle) blockNodesBehind() {
	if !t.useNodeBlocking {
		return
	}

	if len(t.blockedNodesBehind) == 0 {
		return
	}
	idx := len(t.blockedNodesBehind) - 1

	// block current position of a vehicle marker
	u := t.blockedNodesBehind[idx]
	u.TryBlocking(t.ID)
	idx--

	// block nodes behind a vehicle marker simulating vehicle length
	distanceLeft := t.length
	for distanceLeft > 0 && idx >= 0 {
		v := t.blockedNodesBehind[idx]
		distanceLeft -= t.getDistanceToNeighbor(v, u)
		v.TryBlocking(t.ID)
		u = v
		idx--
	}

	// unblock (and remove from the slice) nodes left behind by a vehicle
	p := idx + 1
	for idx >= 0 {
		t.blockedNodesBehind[idx].Unblock(t.ID)
		idx--
	}
	t.blockedNodesBehind = t.blockedNodesBehind[p:]
}

func (t *Vehicle) unblockNodesBehind() {
	if !t.useNodeBlocking {
		return
	}

	for _, node := range t.blockedNodesBehind {
		node.Unblock(t.ID)
	}
}

func (t *Vehicle) unblockNodesAhead() {
	if !t.useNodeBlocking {
		return
	}

	path := t.getTravelPath()
	for i := t.pathIndex; i < len(path.Nodes)-1; i++ {
		path.Nodes[i+1].Unblock(t.ID)
	}
}

func (t *Vehicle) GetEstimatedArrival(stopIndex int, time uint) uint {
	if t.TripDetails.Index > stopIndex || t.TripDetails.Index == stopIndex && t.IsAtStop() {
		return t.TripDetails.Arrivals[stopIndex]
	}

	pathProgress := t.getTravelPath().GetProgressForIndex(t.pathIndex)

	if t.TripDetails.Index == 0 || stopIndex == 0 {
		lastDeparture := t.TripDetails.Trip.Stops[0].Time
		scheduledTravelTime := t.TripDetails.Trip.GetScheduledTravelTime(0, stopIndex)
		return lastDeparture + scheduledTravelTime
	}

	pathLeft := float64(1 - pathProgress)
	scheduledTravelTimeToNextStop := t.TripDetails.Trip.GetScheduledTravelTime(t.TripDetails.Index-1, t.TripDetails.Index)

	remainingTravelTimeToNextStop := uint(math.Round(float64(scheduledTravelTimeToNextStop) * pathLeft))
	estimatedArrivalAtNextStop := time + remainingTravelTimeToNextStop

	// Estimating arrival at next vehicle stop
	if t.TripDetails.Index == stopIndex {
		return estimatedArrivalAtNextStop
	}

	var estimatedPositiveDelay uint
	if estimatedArrivalAtNextStop > t.TripDetails.Trip.Stops[t.TripDetails.Index].Time {
		estimatedPositiveDelay = estimatedArrivalAtNextStop - t.TripDetails.Trip.Stops[t.TripDetails.Index].Time
	}

	return t.TripDetails.Trip.Stops[stopIndex].Time + estimatedPositiveDelay
}

// Guarantees smooth arrival and deceleration to another vehicle, stop or a section
// with a lower speed limit by solving a quadratic equation whose result is the new speed.
// Returns new speed.
func (t *Vehicle) handleDeceleration(targetDistance, targetSpeed, maxSpeed float32) float32 {
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

func (t *Vehicle) getBlockingDistance(speed float32) float32 {
	return speed + speed*speed/(2*MAX_ACCELERATION) + 2*t.length
}

func (t *Vehicle) extendReservedDistance(reservedDistance, neededDistance, distanceToNextNode float32) float32 {
	if reservedDistance+distanceToNextNode <= neededDistance {
		reservedDistance += distanceToNextNode
	} else {
		reservedDistance = neededDistance
	}
	return reservedDistance
}

func (t *Vehicle) updateSpeedAndReserveNodes(path *controlcenter.Path) (availableDistance float32) {
	currentMaxSpeed := path.MaxSpeeds[t.pathIndex]
	newSpeed := min(t.speed+MAX_ACCELERATION, currentMaxSpeed)

	if !t.useNodeBlocking {
		nextSpeed := newSpeed
		distance := (nextSpeed + t.speed) * 0.5
		t.speed = nextSpeed
		return distance
	}

	neededReserveAtCurrentSpeed := t.getBlockingDistance(t.speed)
	neededReserveIfAccel := t.getBlockingDistance(newSpeed)

	var reservedDistanceAtCurrentSpeed, reservedDistanceIfAccel float32
	var reservedDistanceAhead float32
	var distToStop, distToMaxSpeedChange float32
	var upcomingMaxSpeed float32
	var optimisticallyBlockedNodes []graph.GraphNode

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
			distToStop = 1e-3
			for _, blockedNode := range optimisticallyBlockedNodes {
				blockedNode.Unblock(t.ID)
			}

			break
		}

		if u.IsStop() {
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
		//in case we need to stop immediately after detecting blocked node we have to free optimistically blocked nodes
		if reservedDistanceAhead > neededReserveAtCurrentSpeed-2*t.length {
			optimisticallyBlockedNodes = append(optimisticallyBlockedNodes, u)
		}
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
			// handles situation when vehicle is waiting for free node
			nextSpeed = 0
		}
	}

	//this is the distance the vehicle will actually travel (consulting changing speed)
	distance := (nextSpeed + t.speed) * 0.5
	t.speed = nextSpeed

	return distance
}

func (t *Vehicle) getSpeed() uint8 {
	speedKPH := float64((t.speed * 18) / 5)
	return uint8(math.Round(speedKPH))
}

type VehicleDetails struct {
	Route           string                 `json:"route"`
	TripHeadSign    string                 `json:"trip_head_sign"`
	TripIndex       int                    `json:"trip_index"`
	Stops           []api.ResponseTripStop `json:"stops"`
	Arrivals        []uint                 `json:"arrivals"`
	Departures      []uint                 `json:"departures"`
	StopNames       []string               `json:"stop_names"`
	Speed           uint8                  `json:"speed"`
	State           VehicleState           `json:"state"`
	PassengersCount uint                   `json:"passengers_count"`
}

func (t *Vehicle) GetDetails(c *city.City, time uint) VehicleDetails {
	stopsByID := c.GetStopsByID()
	stopNames := make([]string, len(t.TripDetails.Trip.Stops))

	for i, stop := range t.TripDetails.Trip.Stops {
		stopNames[i] = stopsByID[stop.ID].GetName()
	}

	if t.state != StateTripFinished && t.TripDetails.Index < len(t.TripDetails.Arrivals) {
		t.TripDetails.Arrivals[t.TripDetails.Index] = t.GetEstimatedArrival(t.TripDetails.Index, time)
	}

	return VehicleDetails{
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

func (t *Vehicle) IsStopped() bool {
	return t.state == StateStopped || t.state == StateStopping
}

func (t *Vehicle) StopVehicle() {
	switch t.state {
	case StateTravelling, StateStopping:
		t.prevState = t.state
		t.state = StateStopping
	case StatePassengersLoading, StatePassengersUnloading:
		t.prevState = t.state
		t.state = StateStopped
		t.unblockNodesAhead()
	}
}

func (t *Vehicle) ResumeVehicle(currentTime uint) {
	switch t.prevState {
	case StatePassengersLoading:
		t.state = StatePassengersLoading
		if t.departureTime < currentTime {
			t.departureTime = currentTime + 1
		}
	case StatePassengersUnloading:
		t.state = StatePassengersUnloading
	default:
		t.state = StateTravelling
	}
}
