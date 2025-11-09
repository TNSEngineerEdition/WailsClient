package city

import (
	"math"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city/graph"
)

type LatLonBounds struct {
	MinLat float32 `json:"minLat"`
	MinLon float32 `json:"minLon"`
	MaxLat float32 `json:"maxLat"`
	MaxLon float32 `json:"maxLon"`
}

func GetBoundsFromNodes(nodes map[uint64]graph.GraphNode) LatLonBounds {
	minLat, minLon := float32(math.Inf(1)), float32(math.Inf(1))
	maxLat, maxLon := float32(math.Inf(-1)), float32(math.Inf(-1))

	for _, node := range nodes {
		lat, lon := node.GetCoordinates()
		minLat = min(minLat, lat)
		minLon = min(minLon, lon)
		maxLat = max(maxLat, lat)
		maxLon = max(maxLon, lon)
	}

	return LatLonBounds{
		MinLat: minLat,
		MinLon: minLon,
		MaxLat: maxLat,
		MaxLon: maxLon,
	}
}

func (b *LatLonBounds) isInBounds(lat, lon float32) bool {
	return lat >= b.MinLat && lat <= b.MaxLat && lon >= b.MinLon && lon <= b.MaxLon
}
