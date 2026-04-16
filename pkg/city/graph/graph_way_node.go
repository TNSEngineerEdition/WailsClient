package graph

import "github.com/TNSEngineerEdition/WailsClient/pkg/api"

type GraphWayNode struct {
	NodeBlock
	Details api.ResponseGraphNode `json:"details"`
}

func (g *GraphWayNode) IsStop() bool {
	return false
}

func (g *GraphWayNode) GetID() uint64 {
	return g.Details.ID
}

func (g *GraphWayNode) GetCoordinates() (float32, float32) {
	return g.Details.Lat, g.Details.Lon
}

func (g *GraphWayNode) GetNeighbors() map[uint64]api.ResponseGraphEdge {
	return g.Details.Neighbors
}

func (g *GraphWayNode) UpdateMaxSpeed(neighborID uint64, maxSpeed float32) {
	neighbor := g.Details.Neighbors[neighborID]
	neighbor.MaxSpeed = maxSpeed
	g.Details.Neighbors[neighborID] = neighbor

}
