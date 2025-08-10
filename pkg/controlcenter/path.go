package controlcenter

import (
	"container/heap"
	"fmt"
	"slices"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/umahmood/haversine"
)

type Path struct {
	Nodes             []*city.GraphNode
	DistancePrefixSum []float32
}

func (p *Path) GetProgressForIndex(index int) float32 {
	return p.DistancePrefixSum[index] / p.DistancePrefixSum[len(p.DistancePrefixSum)-1]
}

func getShortestPath(city *city.City, stops stopPair) (result Path) {
	nodesByID := city.GetNodesByID()

	nodesToProcess := &priorityQueue{}
	heap.Init(nodesToProcess)
	heap.Push(nodesToProcess, &nodeRecord{ID: stops.source})

	predecessors := make(map[uint64]uint64)
	tentativeDistFromSource := make(map[uint64]float32)
	visitedNodes := make(map[uint64]bool)

	for nodesToProcess.Len() > 0 {
		currentID := heap.Pop(nodesToProcess).(*nodeRecord).ID

		if currentID == stops.destination {
			result.Nodes = reconstructPath(predecessors, nodesByID, currentID)
			result.DistancePrefixSum = getPathDistancePrefixSum(result.Nodes)
			return
		}

		if visitedNodes[currentID] {
			continue
		}

		visitedNodes[currentID] = true

		for _, neighbor := range nodesByID[currentID].Neighbors {
			tentativeDistance := tentativeDistFromSource[currentID] + neighbor.Length
			cost, wasVisited := tentativeDistFromSource[neighbor.ID]

			if wasVisited && tentativeDistance >= cost {
				continue
			}

			predecessors[neighbor.ID] = currentID
			tentativeDistFromSource[neighbor.ID] = tentativeDistance

			heuristicDistance := getDistanceInMeters(
				nodesByID[neighbor.ID], nodesByID[stops.destination],
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
	nodesByID map[uint64]*city.GraphNode,
	currentID uint64,
) (nodes []*city.GraphNode) {
	for {
		nodes = append(nodes, nodesByID[currentID])

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

func getPathDistancePrefixSum(nodes []*city.GraphNode) []float32 {
	prefixSum := make([]float32, len(nodes))

	for i := 1; i < len(nodes); i++ {
		for _, neighbor := range nodes[i-1].Neighbors {
			if neighbor.ID != nodes[i].ID {
				continue
			}

			prefixSum[i] = neighbor.Length + prefixSum[i-1]
		}
	}

	return prefixSum
}
