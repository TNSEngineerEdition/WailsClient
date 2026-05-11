package controlcenter

import (
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/graph"
)

type Path struct {
	Nodes         []graph.GraphNode
	MaxSpeeds     []float32
	TimePrefixSum []float32
}

func pathFromNodeIDs(nodeIDs []uint64, nodesByID *map[uint64]graph.GraphNode) (result Path) {
	result.Nodes = make([]graph.GraphNode, 0, len(nodeIDs))

	for _, nodeID := range nodeIDs {
		result.Nodes = append(result.Nodes, (*nodesByID)[nodeID])
	}

	result.MaxSpeeds = getMaxSpeeds(result.Nodes)
	result.TimePrefixSum = getPathTimePrefixSum(result.Nodes)

	return
}

func getMaxSpeeds(nodes []graph.GraphNode) []float32 {
	maxSpeeds := make([]float32, len(nodes))

	for i := 0; i < len(nodes)-1; i++ {
		neighbors := nodes[i].GetNeighbors()
		nextNode := neighbors[nodes[i+1].GetID()]
		maxSpeeds[i] = nextNode.MaxSpeed
	}

	if len(maxSpeeds) >= 2 {
		// max speed at the last node in path does not matter,
		// repeat the last known max speed
		maxSpeeds[len(maxSpeeds)-1] = maxSpeeds[len(maxSpeeds)-2]
	}

	return maxSpeeds
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

func (p *Path) GetProgressForIndex(index int) float32 {
	return p.TimePrefixSum[index] / p.TimePrefixSum[len(p.TimePrefixSum)-1]
}
