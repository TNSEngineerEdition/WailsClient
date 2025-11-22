package passenger

import (
	"fmt"
	"slices"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/graph"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/trip"
)

type node struct {
	stopID              uint64
	edgesOut            map[uint]*edge
	edgesIn             map[uint]*edge
	visitCounter        uint
	earliestArrivalTime uint
	earliestTram        uint
}

type edge struct {
	tramID        uint
	from          uint64
	to            uint64
	travelTime    uint
	fromStopIndex int
}

type PassengerGraph struct {
	startStopID uint64
	endStopID   uint64
	spawnTime   uint
	nodes       map[uint64]*node
	c           *city.City
}

const MAX_WAITING_TIME = 30 * 60    // 30 min
const MAX_TRAVEL_TIME = 2 * 60 * 60 // 2 hrs
const MIN_CHANGE_TIME = 0
const MAX_FORWARD_BFS_LIMIT_OFFSET = 5 // BFS limit + offset

func NewPassengerGraph(startStopID, endStopID uint64, spawnTime uint, c *city.City) PassengerGraph {
	stopsByID := c.GetStopsByID()

	startStop := stopsByID[startStopID]
	endStop := stopsByID[endStopID]

	fmt.Printf("\n==> Hello! I will spawn at %d\n", spawnTime)
	fmt.Printf("I will travel from %s [%d] to %s [%d]\n\n", startStop.GetName(), startStopID, endStop.GetName(), endStopID)

	pg := PassengerGraph{
		startStopID: startStopID,
		endStopID:   endStopID,
		spawnTime:   spawnTime,
		nodes:       make(map[uint64]*node),
		c:           c,
	}

	visitedStopsForward := pg.forwardBFS()
	visitedStopsBackward := pg.backwardBFS()
	pg.cleanup(visitedStopsForward, visitedStopsBackward)

	fmt.Println("\n\nStops in graph after")
	for _, node := range pg.nodes {
		fmt.Println(stopsByID[node.stopID].GetName())
	}

	return pg
}

type bfsState struct {
	stopID       uint64
	time         uint
	visitCounter uint
}

type visitState struct {
	visited             bool
	earliestArrivalTime uint
}

func (pg *PassengerGraph) forwardBFS() map[uint64]visitState {
	stopsByID := pg.c.GetStopsByID()
	endStop := stopsByID[pg.endStopID]
	trips := pg.c.GetTripsByID()

	queue := []bfsState{{stopID: pg.startStopID, time: pg.spawnTime, visitCounter: 0}}

	isTargetAlreadyReached := false
	BFSLimit := uint(100)

	visitedStops := make(map[uint64]visitState)
	for _, stop := range stopsByID {
		visitedStops[stop.GetID()] = visitState{
			visited:             false,
			earliestArrivalTime: 93600, // 02:00:00
		}
	}

	for len(queue) > 0 {
		state := queue[0]
		queue = queue[1:]

		// end conditions
		killBFS := state.visitCounter > BFSLimit                                          // prevent unnecessary looping over the tram network
		isTravelTimeExceeded := state.time > pg.spawnTime+MAX_TRAVEL_TIME                 // exceeded total travel time
		isStopVisited := visitedStops[state.stopID].visited                               // true if stop is already visited
		isStopVisitedLater := visitedStops[state.stopID].earliestArrivalTime > state.time // true if stop was visited "later" than the current's iteration time

		if killBFS || isTravelTimeExceeded || (isStopVisited && !isStopVisitedLater) {
			continue
		}

		visitedStops[state.stopID] = visitState{
			visited:             true,
			earliestArrivalTime: state.time,
		}

		isTargetReached := state.stopID == pg.endStopID || stopsByID[state.stopID].GetGroupName() == endStop.GetGroupName()
		if isTargetReached {
			if !isTargetAlreadyReached {
				BFSLimit = state.visitCounter + MAX_FORWARD_BFS_LIMIT_OFFSET
				isTargetAlreadyReached = true
			}
			continue
		}

		// iterate over every tram arrival in the upcoming MAX_WAITING_TIME starting from current state.time
		arrivals := *pg.c.GetPlannedArrivalsInTimeSpan(state.stopID, state.time, state.time+MAX_WAITING_TIME)
		for _, arrival := range arrivals {
			trip := trips[arrival.TripID]

			if len(trip.Stops) == arrival.StopIndex+1 {
				continue // in case we reached the final stop of a trip
			}

			nextStop := trip.Stops[arrival.StopIndex+1]

			pg.addEdge(
				state.stopID,
				nextStop.ID,
				arrival.TripID,
				nextStop.Time-arrival.Time,
				state.visitCounter,
				state.time,
				nextStop.Time,
				arrival.StopIndex,
			)

			queue = append(queue, bfsState{
				stopID:       nextStop.ID,
				time:         nextStop.Time,
				visitCounter: state.visitCounter + 1,
			})
		}
	}

	return visitedStops
}

func (pg *PassengerGraph) backwardBFS() map[uint64]visitState {
	stopsByID := pg.c.GetStopsByID()
	stopsByName := pg.c.GetStopsByName()
	endStop := stopsByID[pg.endStopID]

	trips := pg.c.GetTripsByID()

	visitedStops := make(map[uint64]visitState)
	for stopID := range stopsByID {
		visitedStops[stopID] = visitState{
			visited:             false,
			earliestArrivalTime: 93600, // 02:00:00
		}
	}

	queue := []bfsState{}
	for stopID := range stopsByName[endStop.GetGroupName()] {
		queue = append(queue, bfsState{stopID: stopID})
	}

	for len(queue) > 0 {
		state := queue[0]
		queue = queue[1:]

		isStopVisited := visitedStops[state.stopID].visited
		if isStopVisited {
			continue
		}

		visitedStops[state.stopID] = visitState{visited: true}

		isStartReached := state.stopID == pg.startStopID
		if isStartReached {
			continue
		}

		// look for potential changes between trams at interchange point (>2 stops in one group)
		// we only consider nodes not yet present in the graph
		// as they were not processed and potential changes can be found there
		if _, ok := pg.nodes[state.stopID]; !ok {
			stopGroupName := stopsByID[state.stopID].GetGroupName()

			if len(stopsByName[stopGroupName]) < 3 {
				continue
			}

			earliestArrivalTime := pg.getEarliestArrivalForStopsGroup(stopGroupName).time
			arrivals := *pg.c.GetPlannedArrivalsInTimeSpan(
				state.stopID,
				earliestArrivalTime+MIN_CHANGE_TIME,
				earliestArrivalTime+MIN_CHANGE_TIME+MAX_WAITING_TIME,
			)

			for _, arrival := range arrivals {
				trip := pg.c.GetTripsByID()[arrival.TripID]

				if len(trip.Stops) <= arrival.StopIndex+1 {
					continue // in case here is the final stop of a trip
				}

				nextStop := trip.Stops[arrival.StopIndex+1]
				if _, ok := pg.nodes[nextStop.ID]; !ok {
					continue // we only care about trams that are going where we just came from (so towards passenger's target stop)
				}
				if _, ok := pg.nodes[nextStop.ID].edgesOut[arrival.TripID]; !ok {
					continue // we only care about trams that we already have in the graph
				}

				visitedStops[state.stopID] = visitState{visited: true}

				pg.addEdge(state.stopID, nextStop.ID, arrival.TripID, nextStop.Time-arrival.Time, 0, arrival.Time, nextStop.Time, arrival.StopIndex)
			}
			continue // because we do not want to go deeper when we discover a "new" tram stop in the stops group
		}

		// look for other nodes that can be potentially added to the graph (like new tram stop in a stops group)
		//
		// example:
		// passenger wants to go from Poczta Glowna to Rondo Mogilskie,
		// so we have Teatr Slowackiego 03 in graph (lines 10, 52)
		// but we might also try to take line 3, get off at Teatr Slowackiego 03 and then change to 4 or 14 departing from Teatr Slowackiego 02
		// current stop is Lubicz and all mentioned trams departure from there (except 3),
		// but we do not have Teatr Slowackiego 02 in the graph
		// - this for loop solves this issue by adding this stop to the queue
		for _, edgeOut := range pg.nodes[state.stopID].edgesOut {
			trip := trips[edgeOut.tramID]

			if edgeOut.fromStopIndex == 0 {
				continue
			}

			prevStop := trip.Stops[edgeOut.fromStopIndex-1]
			if visitedStops[prevStop.ID].visited {
				continue // node is already in the graph
			}

			queue = append(queue, bfsState{stopID: prevStop.ID})
		}

		// go backwards to previous nodes in graph
		for _, edgeIn := range pg.nodes[state.stopID].edgesIn {
			prevStopID := edgeIn.from
			queue = append(queue, bfsState{stopID: prevStopID})
		}
	}

	return visitedStops
}

func (pg *PassengerGraph) cleanup(visitedStopsForward, visitedStopsBackward map[uint64]visitState) {
	for nodeID, node := range pg.nodes {
		if visitedStopsForward[nodeID].visited && visitedStopsBackward[nodeID].visited {
			continue // if node was visited in both forward and backward BFS, it means it is relevant
		}

		for edgeID, edgeIn := range node.edgesIn {
			delete(pg.nodes[edgeIn.from].edgesOut, edgeID) // delete in neighbor
			delete(node.edgesIn, edgeID)                   // delete in this node
		}

		for edgeID, edgeOut := range node.edgesOut {
			delete(pg.nodes[edgeOut.to].edgesIn, edgeID) // delete in neighbor
			delete(node.edgesOut, edgeID)                // delete in this node
		}

		delete(pg.nodes, nodeID) // delete node from graph
	}
}

type earliestArrival struct {
	stopID uint64
	tramID uint
	time   uint
}

func (pg *PassengerGraph) getEarliestArrivalForStopsGroup(groupName string) earliestArrival {
	stopsGroup := pg.c.GetStopsByName()
	earliestArrival := earliestArrival{time: 93600}

	for stopID := range stopsGroup[groupName] {
		node, ok := pg.nodes[stopID]
		if !ok {
			continue
		}
		if node.earliestArrivalTime < earliestArrival.time {
			earliestArrival.stopID = stopID
			earliestArrival.tramID = node.earliestTram
			earliestArrival.time = node.earliestArrivalTime
		}
	}

	return earliestArrival
}

func (pg *PassengerGraph) addEdge(from uint64, to uint64, tramID uint, travelTime uint, visitCounter uint, fromTime, toTime uint, fromStopIndex int) {
	if _, ok := pg.nodes[from]; !ok {
		pg.nodes[from] = &node{
			stopID:              from,
			edgesOut:            make(map[uint]*edge),
			edgesIn:             make(map[uint]*edge),
			visitCounter:        visitCounter,
			earliestArrivalTime: fromTime,
		}
	}
	if _, ok := pg.nodes[to]; !ok {
		pg.nodes[to] = &node{
			stopID:              to,
			edgesOut:            make(map[uint]*edge),
			edgesIn:             make(map[uint]*edge),
			visitCounter:        visitCounter + 1,
			earliestArrivalTime: toTime,
			earliestTram:        tramID,
		}
	}

	e := edge{
		tramID:        tramID,
		from:          from,
		to:            to,
		travelTime:    travelTime,
		fromStopIndex: fromStopIndex,
	}

	fromNode := pg.nodes[from]
	fromNode.edgesOut[tramID] = &e
	pg.nodes[from] = fromNode

	toNode := pg.nodes[to]
	toNode.edgesIn[tramID] = &e

	if toTime < toNode.earliestArrivalTime {
		toNode.earliestArrivalTime = toTime
		toNode.earliestTram = tramID
	}

	pg.nodes[to] = toNode
}

func (pg *PassengerGraph) findFastestConnection(
	stopsByID map[uint64]*graph.GraphTramStop,
	stopsByName map[string]map[uint64]*graph.GraphTramStop,
	trips map[uint]*trip.TramTrip,
) {
	type TravelPath struct {
		from, to uint64
		tramID   uint
	}
	path := make([]TravelPath, 0)

	fmt.Println("\nWyznaczanie najkrotszej trasy")

	// find the stop with the earliest arrivals
	endStopName := stopsByID[pg.endStopID].GetGroupName()
	endStopsByName := stopsByName[endStopName]
	earliestArrivalTime := uint(99999)
	var currentStop *node

	for stopID := range endStopsByName {
		if _, ok := pg.nodes[stopID]; !ok {
			continue
		}
		arrivalTime := pg.nodes[stopID].earliestArrivalTime
		if arrivalTime < earliestArrivalTime {
			earliestArrivalTime = arrivalTime
			currentStop = pg.nodes[stopID]
		}
	}

	fmt.Println("\nWyznaczono poczatkowy przystanek")

	if currentStop == nil {
		panic("End stop not reached apparently")
	}

	for currentStop.stopID != pg.startStopID {
		tramID := currentStop.earliestTram
		from := currentStop.edgesIn[tramID].from

		p := TravelPath{
			from:   from,
			to:     currentStop.stopID,
			tramID: tramID,
		}
		path = append(path, p)

		currentStop = pg.nodes[from]
	}

	slices.Reverse(path)

	fmt.Println("\n=> Travel plan:")
	for _, p := range path {
		fmt.Printf("\t%s -> %s by %d to %s\n", stopsByID[p.from].GetName(), stopsByID[p.to].GetName(), p.tramID, trips[p.tramID].TripHeadSign)
	}
}
