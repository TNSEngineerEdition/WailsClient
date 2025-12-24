package controlcenter

import (
	"fmt"
	"slices"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city/graph"
	"github.com/TNSEngineerEdition/WailsClient/pkg/structs"
	"github.com/umahmood/haversine"
)

type Path struct {
	Nodes         []graph.GraphNode
	MaxSpeeds     []float32
	TimePrefixSum []float32
}

func (p *Path) GetProgressForIndex(index int) float32 {
	return p.TimePrefixSum[index] / p.TimePrefixSum[len(p.TimePrefixSum)-1]
}

func getShortestPath(nodesByID *map[uint64]graph.GraphNode, stops stopPair) (result Path) {
	nodesToProcess := structs.NewPriorityQueue[uint64](func(left, right float32) bool { return left < right })
	nodesToProcess.Push(stops.source, 0)

	predecessors := make(map[uint64]uint64)
	tentativeDistFromSource := make(map[uint64]float32)
	visitedNodes := structs.NewSet[uint64]()

	for nodesToProcess.Len() > 0 {
		currentID := nodesToProcess.Pop()

		if currentID == stops.destination {
			result.Nodes = reconstructPath(predecessors, nodesByID, currentID)
			result.MaxSpeeds = getMaxSpeeds(result.Nodes)
			result.TimePrefixSum = getPathTimePrefixSum(result.Nodes)
			return
		}

		if visitedNodes.Includes(currentID) {
			continue
		}

		visitedNodes.Add(currentID)

		for _, neighbor := range (*nodesByID)[currentID].GetNeighbors() {
			tentativeDistance := tentativeDistFromSource[currentID] + neighbor.Distance
			cost, wasVisited := tentativeDistFromSource[neighbor.ID]

			if wasVisited && tentativeDistance >= cost {
				continue
			}

			predecessors[neighbor.ID] = currentID
			tentativeDistFromSource[neighbor.ID] = tentativeDistance

			heuristicDistance := getDistanceInMeters(
				(*nodesByID)[neighbor.ID], (*nodesByID)[stops.destination],
			)
			nodesToProcess.Push(neighbor.ID, heuristicDistance+tentativeDistance)
		}
	}

	panic(fmt.Sprintf("No path found between %d and %d nodes", stops.source, stops.destination))
}

func reconstructPath(
	predecessors map[uint64]uint64,
	nodesByID *map[uint64]graph.GraphNode,
	currentID uint64,
) (nodes []graph.GraphNode) {
	for {
		nodes = append(nodes, (*nodesByID)[currentID])

		if previousNodeID, ok := predecessors[currentID]; ok {
			currentID = previousNodeID
		} else {
			break
		}
	}

	slices.Reverse(nodes)
	return
}

func getMaxSpeeds(nodes []graph.GraphNode) []float32 {
	maxSpeeds := make([]float32, len(nodes))

	for i := 0; i < len(nodes)-1; i++ {
		neighbors := nodes[i].GetNeighbors()
		nextNode := neighbors[nodes[i+1].GetID()]
		maxSpeeds[i] = nextNode.MaxSpeed
	}

	// max speed at the last node in path does not matter,
	// repeat the last known max speed
	maxSpeeds[len(maxSpeeds)-1] = maxSpeeds[len(maxSpeeds)-2]

	return maxSpeeds
}

func getDistanceInMeters(source, destination graph.GraphNode) float32 {
	sourceLat, sourceLon := source.GetCoordinates()
	destLat, destLon := destination.GetCoordinates()

	_, kilometers := haversine.Distance(
		haversine.Coord{
			Lat: float64(sourceLat),
			Lon: float64(sourceLon),
		},
		haversine.Coord{
			Lat: float64(destLat),
			Lon: float64(destLon),
		},
	)

	return float32(kilometers * 1000)
}

func getPathTimePrefixSum(nodes []graph.GraphNode) []float32 {
	prefixSum := make([]float32, len(nodes))

	for i := 1; i < len(nodes); i++ {
		neighbors := nodes[i-1].GetNeighbors()
		nextNode := neighbors[nodes[i].GetID()]
		prefixSum[i] = nextNode.Distance/nextNode.MaxSpeed + prefixSum[i-1]
	}

	return prefixSum
}
