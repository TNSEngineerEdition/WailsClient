package graph

import (
	"fmt"

	"github.com/TNSEngineerEdition/WailsClient/pkg/api"
)

type GraphNode interface {
	NodeBlocker
	IsStop() bool
	GetID() uint64
	GetCoordinates() (float32, float32)
	GetNeighbors() map[uint64]api.ResponseGraphEdge
	UpdateMaxSpeed(neighborID uint64, maxSpeed float32)
}

func GraphNodesFromCityData(responseCityData *api.ResponseCityData) (map[uint64]GraphNode, error) {
	nodesByID := make(map[uint64]GraphNode, len(responseCityData.Graph))

	for _, nodeItem := range responseCityData.Graph {
		value, err := nodeItem.ValueByDiscriminator()
		if err != nil {
			return nil, err
		}

		switch node := value.(type) {
		case api.ResponseGraphStop:
			nodesByID[node.ID] = &GraphStop{Details: node}
		case api.ResponseGraphNode:
			nodesByID[node.ID] = &GraphRoadNode{Details: node}
		default:
			return nil, fmt.Errorf("Unrecognized node type: %s", node)
		}
	}

	return nodesByID, nil
}
