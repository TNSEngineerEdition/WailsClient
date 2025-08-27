package controlcenter

import (
	"fmt"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
)

type stopPair struct {
	source, destination uint64
}

type ControlCenter struct {
	city  *city.City
	paths map[stopPair]Path
}

func NewControlCenter(cityPointer *city.City) ControlCenter {
	c := ControlCenter{
		city:  cityPointer,
		paths: make(map[stopPair]Path),
	}

	for _, route := range cityPointer.GetTramRoutes() {
		for _, trip := range route.Trips {
			for i := 0; i < len(trip.Stops)-1; i++ {
				stopPair := stopPair{
					source:      trip.Stops[i].ID,
					destination: trip.Stops[i+1].ID,
				}

				if _, ok := c.paths[stopPair]; !ok {
					c.paths[stopPair] = getShortestPath(c.city, stopPair)
				}
			}
		}
	}

	return c
}

func (c *ControlCenter) GetPath(sourceNodeID, destinationNodeID uint64) *Path {
	if path, ok := c.paths[stopPair{source: sourceNodeID, destination: destinationNodeID}]; ok {
		return &path
	}

	panic(fmt.Sprintf("No path found between %d and %d nodes", sourceNodeID, destinationNodeID))
}
