package controlcenter

import (
	"fmt"
	"sort"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
)

type stopPair struct {
	source, destination uint64
}

type ControlCenter struct {
	city  *city.City
	paths map[stopPair]Path
}

type RoutePolylines struct {
	Forward  [][2]float32 `json:"forward"`
	Backward [][2]float32 `json:"backward"`
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

// TODO: Temporary implementation of GetRoutePolylines and helper functions.
//
//	This part will be simplified after the TramRoute data structure is updated.
func (c *ControlCenter) coordsFromStopSequence(stopIDs []uint64) [][2]float32 {
	if len(stopIDs) < 2 {
		return nil
	}

	var out [][2]float32

	for i := 0; i+1 < len(stopIDs); i++ {
		key := stopPair{source: stopIDs[i], destination: stopIDs[i+1]}
		p := c.paths[key]
		pts := coordsFromPathNodes(&p)

		if len(out) > 0 && len(pts) > 0 && out[len(out)-1] == pts[0] {
			out = append(out, pts[1:]...)
		} else {
			out = append(out, pts...)
		}
	}

	return out
}

func coordsFromPathNodes(p *Path) [][2]float32 {
	out := make([][2]float32, 0, len(p.Nodes))
	for _, n := range p.Nodes {
		out = append(out, [2]float32{float32(n.Latitude), float32(n.Longitude)})
	}
	return out
}

func (c *ControlCenter) GetRoutePolylines(lineName string) RoutePolylines {
	routes := c.city.GetTramRoutes()

	var route *city.TramRoute
	for i := range routes {
		if routes[i].Name == lineName {
			r := routes[i]
			route = &r
			break
		}
	}

	type dirKey struct{ start, end uint64 }
	type dirAgg struct {
		key      dirKey
		count    int
		bestTrip city.TramTrip
	}

	agg := make(map[dirKey]*dirAgg)
	for _, trip := range route.Trips {
		key := dirKey{
			start: trip.Stops[0].ID,
			end:   trip.Stops[len(trip.Stops)-1].ID,
		}
		entry := agg[key]
		if entry == nil {
			entry = &dirAgg{key: key}
			agg[key] = entry
		}
		entry.count++
		if len(trip.Stops) > len(entry.bestTrip.Stops) {
			entry.bestTrip = trip
		}
	}

	list := make([]*dirAgg, 0, len(agg))
	for _, v := range agg {
		list = append(list, v)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].count == list[j].count {
			return len(list[i].bestTrip.Stops) > len(list[j].bestTrip.Stops)
		}
		return list[i].count > list[j].count
	})

	outA := stopsToIDs(list[0].bestTrip.Stops)
	outB := stopsToIDs(list[1].bestTrip.Stops)
	coordsA := c.coordsFromStopSequence(outA)
	coordsB := c.coordsFromStopSequence(outB)
	return RoutePolylines{Forward: coordsA, Backward: coordsB}
}

func stopsToIDs(stops []city.TramTripStop) []uint64 {
	ids := make([]uint64, len(stops))
	for i := range stops {
		ids[i] = stops[i].ID
	}
	return ids
}
