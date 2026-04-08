package graph

import (
	"github.com/TNSEngineerEdition/WailsClient/pkg/api"
)

type GraphStop struct {
	NodeBlock
	Details api.ResponseGraphStop `json:"details"`
}

func (g *GraphStop) IsStop() bool {
	return true
}

func (g *GraphStop) GetDetails() api.ResponseGraphStop {
	return g.Details
}

func (g *GraphStop) GetID() uint64 {
	return g.Details.ID
}

func (g *GraphStop) GetCoordinates() (float32, float32) {
	return g.Details.Lat, g.Details.Lon
}

func (g *GraphStop) GetNeighbors() map[uint64]api.ResponseGraphEdge {
	return g.Details.Neighbors
}

func (g *GraphStop) GetName() string {
	return g.Details.Name
}

func (g *GraphStop) GetGroupName() string {
	if g.Details.StopGroupName == nil {
		return ""
	}
	return *g.Details.StopGroupName
}

func (g *GraphStop) UpdateMaxSpeed(neighborID uint64, maxSpeed float32) {
	neighbor := g.Details.Neighbors[neighborID]
	neighbor.MaxSpeed = maxSpeed
	g.Details.Neighbors[neighborID] = neighbor
}
