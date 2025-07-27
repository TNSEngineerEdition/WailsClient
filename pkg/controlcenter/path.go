package controlcenter

import (
	"container/heap"
	"fmt"
	"slices"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/umahmood/haversine"
)

type Path struct {
	Nodes    []*city.GraphNode
	Distance float32
}

func getShortestPath(city *city.City, stops stopPair) (result Path) {
	tramStops := city.GetStopsByID()

	nodesToProcess := &priorityQueue{}
	heap.Init(nodesToProcess)
	heap.Push(nodesToProcess, &nodeRecord{ID: stops.source})

	predecessors := make(map[uint64]uint64)
	tentativeDistFromSource := make(map[uint64]float32)
	visitedNodes := make(map[uint64]bool)

	for nodesToProcess.Len() > 0 {
		currentID := heap.Pop(nodesToProcess).(*nodeRecord).ID

		if currentID == stops.destination {
			result.Nodes = reconstructPath(predecessors, tramStops, currentID)
			result.Distance = getPathDistance(result.Nodes)
			return
		}

		if visitedNodes[currentID] {
			continue
		}

		visitedNodes[currentID] = true

		for _, neighbor := range tramStops[currentID].Neighbors {
			tentativeDistance := tentativeDistFromSource[currentID] + neighbor.Length
			cost, wasVisited := tentativeDistFromSource[neighbor.ID]

			if wasVisited && tentativeDistance >= cost {
				continue
			}

			predecessors[neighbor.ID] = currentID
			tentativeDistFromSource[neighbor.ID] = tentativeDistance

			heuristicDistance := getDistanceInMeters(
				tramStops[neighbor.ID], tramStops[stops.destination],
			)
			heap.Push(
				nodesToProcess,
				&nodeRecord{
					ID: neighbor.ID, Priority: heuristicDistance + tentativeDistance,
				},
			)
		}
	}

	panic(fmt.Sprintf("No path found between %d and %d nodes", stops.source, stops.destination))
}

func reconstructPath(
	predecessors map[uint64]uint64,
	stops map[uint64]*city.GraphNode,
	currentID uint64,
) (nodes []*city.GraphNode) {
	for {
		nodes = append(nodes, stops[currentID])

		if previousNodeID, ok := predecessors[currentID]; ok {
			currentID = previousNodeID
		} else {
			break
		}
	}

	slices.Reverse(nodes)
	return
}

func getDistanceInMeters(source, destination *city.GraphNode) float32 {
	_, kilometers := haversine.Distance(
		haversine.Coord{
			Lat: float64(source.Latitude),
			Lon: float64(source.Longitude),
		},
		haversine.Coord{
			Lat: float64(destination.Latitude),
			Lon: float64(destination.Longitude),
		},
	)

	return float32(kilometers * 1000)
}

func getPathDistance(nodes []*city.GraphNode) (result float32) {
	for i := 0; i < len(nodes)-1; i++ {
		result += getDistanceInMeters(nodes[i], nodes[i+1])
	}
	return
}
