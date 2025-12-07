package graph

import (
	"unicode"

	"github.com/TNSEngineerEdition/WailsClient/pkg/api"
)

type GraphTramStop struct {
	NodeBlock
	Details api.ResponseGraphTramStop `json:"details"`
}

func (g *GraphTramStop) IsTramStop() bool {
	return true
}

func (g *GraphTramStop) GetDetails() api.ResponseGraphTramStop {
	return g.Details
}

func (g *GraphTramStop) GetID() uint64 {
	return g.Details.ID
}

func (g *GraphTramStop) GetCoordinates() (float32, float32) {
	return g.Details.Lat, g.Details.Lon
}

func (g *GraphTramStop) GetNeighbors() map[uint64]api.ResponseGraphEdge {
	return g.Details.Neighbors
}

func (g *GraphTramStop) GetName() string {
	return g.Details.Name
}

func (g *GraphTramStop) GetGroupName() string {
	runeName := []rune(g.Details.Name)
	n := len(runeName)

	if n < 3 {
		return ""
	}

	// stop name pattern: "[groupName] [twoDigitStopNumber]"
	if unicode.IsDigit(runeName[n-1]) && unicode.IsDigit(runeName[n-2]) && unicode.IsSpace(runeName[n-3]) {
		return g.Details.Name[:n-3]
	}

	return g.Details.Name
}

func (g *GraphTramStop) UpdateMaxSpeed(neighborID uint64, maxSpeed float32) {
	neighbor := g.Details.Neighbors[neighborID]
	neighbor.MaxSpeed = maxSpeed
	g.Details.Neighbors[neighborID] = neighbor
}
