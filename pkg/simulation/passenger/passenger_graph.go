package passenger

import (
	"fmt"
	"slices"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/graph"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/trip"
	// "github.com/TNSEngineerEdition/WailsClient/pkg/city/graph"
	// "github.com/TNSEngineerEdition/WailsClient/pkg/city/trip"
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
	nodes       map[uint64]*node
	c           *city.City
}

func NewPassengerGraph(startStopID, endStopID uint64, spawnTime uint, c *city.City) PassengerGraph {
	stops := c.GetStopsByID()
	stopGroups := c.GetStopsByName()

	startStop := stops[startStopID]
	endStop := stops[endStopID]

	endStopGroupName := endStop.GetGroupName()

	fmt.Printf("\n==> Hello! I will spawn at %d\n", spawnTime)
	fmt.Printf("I will travel from %s [%d] to %s [%d]\n\n", startStop.GetName(), startStopID, endStop.GetName(), endStopID)

	graph := PassengerGraph{
		startStopID: startStopID,
		endStopID:   endStopID,
		nodes:       make(map[uint64]*node),
		c:           c,
	}

	const MAX_WAITING_TIME = 30 * 60    // 30 min
	const MAX_TRAVEL_TIME = 2 * 60 * 60 // 2 hrs
	const MIN_CHANGE_TIME = 0

	trips := c.GetTripsByID()

	/////// ========= ///////
	///////    BFS    ///////
	/////// ========= ///////

	type bfsState struct {
		stopID       uint64
		time         uint
		visitCounter uint
	}

	type visitedState struct {
		visited             bool
		earliestArrivalTime uint
	}

	visitedStops := make(map[uint64]visitedState)
	for _, stop := range stops {
		visitedStops[stop.GetID()] = visitedState{
			visited:             false,
			earliestArrivalTime: 93600, // 02:00:00
		}
	}

	targetAlreadyReached := false
	BFSLimit := uint(100)

	queue := []bfsState{{stopID: startStopID, time: spawnTime, visitCounter: 0}}

	for len(queue) > 0 {
		state := queue[0]
		queue = queue[1:]

		// end conditions
		killBFS := state.visitCounter > BFSLimit
		exceededTravelTime := state.time > spawnTime+MAX_TRAVEL_TIME
		alreadyVisitedStop := visitedStops[state.stopID].visited
		isAlreadyVisitedLater := visitedStops[state.stopID].earliestArrivalTime > state.time

		if killBFS || exceededTravelTime || (alreadyVisitedStop && !isAlreadyVisitedLater) {
			continue
		}

		visitedStops[state.stopID] = visitedState{
			visited:             true,
			earliestArrivalTime: state.time,
		}

		reachedTarget := state.stopID == endStopID || stops[state.stopID].GetGroupName() == endStopGroupName
		if reachedTarget {
			if !targetAlreadyReached {
				BFSLimit = state.visitCounter + 5
				targetAlreadyReached = true
			}

			continue
		}

		// get arrivals on current stop
		arrivals := *c.GetPlannedArrivalsInTimeSpan(state.stopID, state.time, state.time+MAX_WAITING_TIME)

		for _, arrival := range arrivals {
			// add next stop for each arrival to new state
			trip := trips[arrival.TripID]

			if len(trip.Stops) == arrival.StopIndex+1 {
				continue // in case we reached the final stop of a trip
			}

			nextStop := trip.Stops[arrival.StopIndex+1]
			graph.addEdge(state.stopID, nextStop.ID, arrival.TripID, nextStop.Time-arrival.Time, state.visitCounter, state.time, nextStop.Time, arrival.StopIndex)

			// skip if the stop was already visited
			if visitedStops[nextStop.ID].visited {
				continue
			}

			queue = append(queue, bfsState{
				stopID:       nextStop.ID,
				time:         nextStop.Time,
				visitCounter: state.visitCounter + 1,
			})
		}
	}

	////// =========== ///////
	////// REVERSE BFS ///////
	////// =========== ///////
	// travel the graph backwards to mark possible interchange stops

	visitedStopsReverse := make(map[uint64]visitedState)

	for stopID := range stops {
		visitedStopsReverse[stopID] = visitedState{visited: false}
	}

	for stopID := range stopGroups[endStopGroupName] {
		queue = append(queue, bfsState{stopID: stopID})
	}

	for len(queue) > 0 {
		// opis algorytmu
		// cofamy sie w grafie biorąc pod uwage przyjazdy
		// cel - chcemy dorzucić do grafu przystanki, na których można się przesiąść i dotrzeć z nich do celu

		// gdy natrafiamy na stopID, ktorego nie ma (ale są inne z jego grupy!), a mamy podanego poprzednika
		// to sprawdzamy kiedy mamy najwczesniejszy przyjazd (+ czas na przesiadke)
		// pobieramy odjazdy z tego przystanku
		// patrzymy ktore tramID sa w nodes[succStopID]
		// dla tych tramID tworzymy krawedzie i ofc dodajemy node
		// TO BEDZIE NASZA PRZESIADKA

		state := queue[0]
		queue = queue[1:]

		if visitedStopsReverse[state.stopID].visited {
			continue // already visited
		}

		visitedStopsReverse[state.stopID] = visitedState{visited: true}

		//fmt.Println("Analyzing stop: ", stops[state.stopID].GetName())

		if state.stopID == startStopID {
			continue // reached end
		}

		// changing between trams
		if _, ok := graph.nodes[state.stopID]; !ok {
			if state.stopID == 2423481568 {
				continue
			}
			groupName := stops[state.stopID].GetGroupName()

			if len(stopGroups[groupName]) < 3 {
				continue
			}

			earliestArrival := graph.getEarliestArrivalForStopsGroup(groupName).time

			arrivals := *c.GetPlannedArrivalsInTimeSpan(
				state.stopID,
				earliestArrival+MIN_CHANGE_TIME,
				earliestArrival+MIN_CHANGE_TIME+MAX_WAITING_TIME,
			)

			// fmt.Println("New tram stop: ", stops[state.stopID].GetName())
			// fmt.Println("   => Odjazdy")
			// fmt.Println(arrivals)

			// bierzemy tylko te arrivals, dla ktorych next stop jest w naszym grafie
			for _, arrival := range arrivals {
				trip := trips[arrival.TripID]

				// fmt.Println("   => StopIndex check")
				if len(trip.Stops) <= arrival.StopIndex+1 {
					continue // in case we reached the final stop of a trip
				}

				// fmt.Println("   => Assinging next stop")
				nextStop := trips[arrival.TripID].Stops[arrival.StopIndex+1]

				// fmt.Println("   => Checking if next stop is in the graph, next stop: ", stops[nextStop.ID].GetName())
				if _, ok := graph.nodes[nextStop.ID]; !ok {
					continue
				}

				// check if this trip is in the node
				// fmt.Println("   => Check if this trip is actually in the next node, this trip is to: ", trip.TripHeadSign)
				if _, ok := graph.nodes[nextStop.ID].edgesOut[arrival.TripID]; !ok {
					continue
				}

				// fmt.Printf("      ==> Trying to add edge between %s -> %s\n", stops[state.stopID].GetName(), stops[nextStop.ID].GetName())
				visitedStops[state.stopID] = visitedState{visited: true}
				graph.addEdge(state.stopID, nextStop.ID, arrival.TripID, nextStop.Time-arrival.Time, 0, arrival.Time, nextStop.Time, arrival.StopIndex)

			}
			continue
		}

		// look for other nodes that can be potentially in graph
		for _, edgeOut := range graph.nodes[state.stopID].edgesOut {
			trip := trips[edgeOut.tramID]
			stopIndex := edgeOut.fromStopIndex
			if stopIndex == 0 {
				continue
			}
			prevStop := trip.Stops[stopIndex-1]
			if visitedStops[prevStop.ID].visited {
				continue // this means that this node exists in the graph
			}
			queue = append(queue, bfsState{stopID: prevStop.ID})
		}

		// go backwards to previousNodes in graph
		for _, edgeIn := range graph.nodes[state.stopID].edgesIn {
			prevStopID := edgeIn.from
			queue = append(queue, bfsState{stopID: prevStopID})
		}
	}

	fmt.Println("\n\nStops in graph before")
	for _, node := range graph.nodes {
		fmt.Println(stops[node.stopID].GetName())
	}

	// fmt.Println("\n\nStops without edges in")
	// for _, node := range graph.nodes {
	// 	if len(node.edgesIn) == 0 {
	// 		fmt.Println(stops[node.stopID].GetName())
	// 	}
	// }

	////// =========== ///////
	//////   CLEANUP   ///////
	////// =========== ///////

	// zostawiamy tylko przystanki odwiedzone w obu etapach :)

	for nodeID, node := range graph.nodes {
		if visitedStops[nodeID].visited && visitedStopsReverse[nodeID].visited {
			continue
		}

		// delete edges - also in neighbors

		// edges in
		for edgeID, edgeIn := range node.edgesIn {
			// delete in neighbor
			delete(graph.nodes[edgeIn.from].edgesOut, edgeID)
			// delete in this node
			delete(node.edgesIn, edgeID)
		}

		// edges out
		for edgeID, edgeOut := range node.edgesOut {
			// delete in neighbor
			delete(graph.nodes[edgeOut.to].edgesIn, edgeID)
			// delete in this node
			delete(node.edgesOut, edgeID)
		}

		// delete node from graph
		delete(graph.nodes, nodeID)
	}

	fmt.Println("\n\nStops in graph after")
	for _, node := range graph.nodes {
		fmt.Println(stops[node.stopID].GetName())
	}

	return graph
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

	// 1️⃣ pobierz kopię node'a, zmodyfikuj, potem zapisz z powrotem
	fromNode := pg.nodes[from]
	fromNode.edgesOut[tramID] = &e
	pg.nodes[from] = fromNode

	// 2️⃣ to samo dla node’a docelowego
	toNode := pg.nodes[to]
	toNode.edgesIn[tramID] = &e

	if toTime < toNode.earliestArrivalTime {
		toNode.earliestArrivalTime = toTime
		toNode.earliestTram = tramID
	}

	pg.nodes[to] = toNode
}

func (pg *PassengerGraph) findFastestConnection(
	stops map[uint64]*graph.GraphTramStop,
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
	endStopName := stops[pg.endStopID].GetGroupName()
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
		fmt.Printf("\t%s -> %s by %d to %s\n", stops[p.from].GetName(), stops[p.to].GetName(), p.tramID, trips[p.tramID].TripHeadSign)
	}
}
