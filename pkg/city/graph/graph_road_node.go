package graph

import "github.com/TNSEngineerEdition/WailsClient/pkg/api"

type GraphRoadNode struct {
	NodeBlock
	Details api.ResponseGraphNode `json:"details"`
}

func (g *GraphRoadNode) IsStop() bool {
	return false
}

func (g *GraphRoadNode) GetID() uint64 {
	return g.Details.ID
}

func (g *GraphRoadNode) GetCoordinates() (float32, float32) {
	return g.Details.Lat, g.Details.Lon
}

func (g *GraphRoadNode) GetNeighbors() map[uint64]api.ResponseGraphEdge {
	return g.Details.Neighbors
}

func (g *GraphRoadNode) UpdateMaxSpeed(neighborID uint64, maxSpeed float32) {
	neighbor := g.Details.Neighbors[neighborID]
	neighbor.MaxSpeed = maxSpeed
	g.Details.Neighbors[neighborID] = neighbor

}
