package graph

import (
	"strings"
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

// TODO: Update this code after group name implementation on the backend
func (g *GraphTramStop) GetGroupName() string {
	r := []rune(strings.TrimSpace(g.Details.Name))
	n := len(r)
	if n < 3 {
		return ""
	}

	// stop name pattern: "[groupName] [twoDigitStopNumber]"
	if unicode.IsDigit(r[n-1]) && unicode.IsDigit(r[n-2]) && unicode.IsSpace(r[n-3]) {
		return strings.TrimSpace(string(r[:n-3]))
	}

	return string(r)
}

func (g *GraphTramStop) UpdateMaxSpeed(neighborID uint64, maxSpeed float32) {
	neighbor := g.Details.Neighbors[neighborID]
	neighbor.MaxSpeed = maxSpeed
	g.Details.Neighbors[neighborID] = neighbor
}
