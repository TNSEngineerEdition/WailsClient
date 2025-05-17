package control_room

import (
	"container/heap"
	"fmt"
	"slices"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/umahmood/haversine"
)

type RouteBetweenNodes struct {
	SourceID uint64            `json:"sourceId"`
	DestID   uint64            `json:"destId"`
	Route    []*city.GraphNode `json:"route"`
}

type ControlCenter struct {
	city               *city.City
	routesBetweenNodes map[[2]uint64][]*city.GraphNode
}

func (c *ControlCenter) GetRoutesBetweenNodes(sourceID, destID uint64) (path []*city.GraphNode) {
	key := [2]uint64{sourceID, destID}
	path, _ = c.routesBetweenNodes[key]
	return
}

func CreateControlCenter(cityPointer *city.City) ControlCenter {
	c := ControlCenter{
		city:               cityPointer,
		routesBetweenNodes: make(map[[2]uint64][]*city.GraphNode),
	}

	tramTrips := cityPointer.GetTramTrips()
	for _, tripData := range tramTrips {
		for i := 0; i < len(tripData.Stops)-1; i++ {
			firstStop, secondStop := tripData.Stops[i], tripData.Stops[i+1]
			tramStopPair := [2]uint64{firstStop.ID, secondStop.ID}
			if _, ok := c.routesBetweenNodes[tramStopPair]; ok {
				continue
			}

			c.routesBetweenNodes[tramStopPair] = c.GetShortestPath(firstStop.ID, secondStop.ID)
		}
	}
	return c
}

func (c *ControlCenter) GetShortestPath(sourceID, destID uint64) []*city.GraphNode {
	tramStopPair := [2]uint64{sourceID, destID}
	tramStops := c.city.GetStopsByID()
	destNode := tramStops[destID]

	nodesToProcess := &priorityQueue{}
	heap.Init(nodesToProcess)
	heap.Push(nodesToProcess, &nodeRecord{ID: sourceID})

	predecessors := make(map[uint64]uint64)
	tentativeDistFromSource := make(map[uint64]float32)
	visitedNodes := make(map[uint64]bool)

	for nodesToProcess.Len() > 0 {
		currentID := heap.Pop(nodesToProcess).(*nodeRecord).ID
		if currentID == destID {
			c.routesBetweenNodes[tramStopPair] = c.reconstructPath(predecessors, tramStops, currentID)
			return c.routesBetweenNodes[tramStopPair]
		}

		if visitedNodes[currentID] {
			continue
		}

		visitedNodes[currentID] = true

		currentNode := tramStops[currentID]

		for _, neighbor := range currentNode.Neighbors {
			tentativeDist := tentativeDistFromSource[currentID] + neighbor.Length
			cost, wasVisited := tentativeDistFromSource[neighbor.ID]
			if wasVisited && tentativeDist >= cost {
				continue
			}

			predecessors[neighbor.ID] = currentID
			tentativeDistFromSource[neighbor.ID] = tentativeDist
			heuristicDistance := c.heuristic(tramStops[neighbor.ID], destNode)
			expectedDistFromSrcToDest := heuristicDistance + tentativeDist
			heap.Push(nodesToProcess, &nodeRecord{ID: neighbor.ID, Priority: expectedDistFromSrcToDest})
		}
	}

	panic(fmt.Sprintf("No path found between %d and %d nodes", sourceID, destID))
}

func (c *ControlCenter) heuristic(a, b *city.GraphNode) float32 {
	sourceCoords := haversine.Coord{Lat: float64(a.Latitude), Lon: float64(a.Longitude)}
	goalCoords := haversine.Coord{Lat: float64(b.Latitude), Lon: float64(b.Longitude)}
	_, km := haversine.Distance(sourceCoords, goalCoords)
	return float32(km * 1000)
}

func (c *ControlCenter) reconstructPath(
	predecessors map[uint64]uint64,
	stops map[uint64]*city.GraphNode,
	currentID uint64,
) (path []*city.GraphNode) {
	for {
		path = append(path, stops[currentID])
		prev, ok := predecessors[currentID]
		if !ok {
			break
		}
		currentID = prev
	}

	slices.Reverse(path)
	return
}
