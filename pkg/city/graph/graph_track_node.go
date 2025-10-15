package graph

import "github.com/TNSEngineerEdition/WailsClient/pkg/api"

type GraphTrackNode struct {
	NodeBlock
	details api.ResponseGraphNode
}

func (g *GraphTrackNode) IsTramStop() bool {
	return false
}

func (g *GraphTrackNode) GetID() uint64 {
	return g.details.ID
}

func (g *GraphTrackNode) GetCoordinates() (float32, float32) {
	return g.details.Lat, g.details.Lon
}

func (g *GraphTrackNode) GetNeighbors() map[uint64]api.ResponseGraphEdge {
	return g.details.Neighbors
}
