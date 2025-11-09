package graph

import "github.com/TNSEngineerEdition/WailsClient/pkg/api"

type GraphTrackNode struct {
	NodeBlock
	Details api.ResponseGraphNode `json:"details"`
}

func (g *GraphTrackNode) IsTramStop() bool {
	return false
}

func (g *GraphTrackNode) GetID() uint64 {
	return g.Details.ID
}

func (g *GraphTrackNode) GetCoordinates() (float32, float32) {
	return g.Details.Lat, g.Details.Lon
}

func (g *GraphTrackNode) GetNeighbors() map[uint64]api.ResponseGraphEdge {
	return g.Details.Neighbors
}

func (g *GraphTrackNode) UpdateMaxSpeed(neighborID uint64, maxSpeed float32) {
	neighbor := g.Details.Neighbors[neighborID]
	neighbor.MaxSpeed = maxSpeed
	g.Details.Neighbors[neighborID] = neighbor

}
