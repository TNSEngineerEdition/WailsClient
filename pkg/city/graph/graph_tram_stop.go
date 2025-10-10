package graph

import "github.com/TNSEngineerEdition/WailsClient/pkg/api"

type GraphTramStop struct {
	NodeBlock
	details api.ResponseGraphTramStop
}

func (g *GraphTramStop) IsTramStop() bool {
	return true
}

func (g *GraphTramStop) GetDetails() api.ResponseGraphTramStop {
	return g.details
}

func (g *GraphTramStop) GetID() uint64 {
	return g.details.ID
}

func (g *GraphTramStop) GetCoordinates() (float32, float32) {
	return g.details.Lat, g.details.Lon
}

func (g *GraphTramStop) GetNeighbors() map[uint64]api.ResponseGraphEdge {
	return g.details.Neighbors
}

func (g *GraphTramStop) GetName() string {
	return g.details.Name
}
