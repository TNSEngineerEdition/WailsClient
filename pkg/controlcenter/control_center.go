package controlcenter

import (
	"fmt"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/graph"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/trip"
)

type stopPair struct {
	source, destination uint64
}

type Coordinates struct {
	Lat float32 `json:"lat"`
	Lon float32 `json:"lon"`
}

type RouteSegment struct {
	StopIDs  []uint64      `json:"stopIDs"`
	Polyline []Coordinates `json:"polyline"`
}

type ControlCenter struct {
	paths               map[stopPair]Path
	segmentsByRouteName map[string][]RouteSegment
}

func NewControlCenter(city *city.City) ControlCenter {
	controlCenter := ControlCenter{
		paths:               make(map[stopPair]Path),
		segmentsByRouteName: make(map[string][]RouteSegment),
	}

	tramRoutes := city.GetTramRoutes()
	nodesByID := city.GetNodesByID()

	for _, route := range tramRoutes {
		for _, trip := range route.Trips {
			controlCenter.addPathsFromTrip(&trip, &nodesByID)
		}
	}

	for _, route := range tramRoutes {
		if route.Variants == nil {
			continue
		}

		controlCenter.setSegmentsByRouteName(&route)
	}

	return controlCenter
}

func (c *ControlCenter) addPathsFromTrip(
	trip *trip.TramTrip,
	nodesByID *map[uint64]graph.GraphNode,
) {
	for i := 0; i < len(trip.Stops)-1; i++ {
		stopPair := stopPair{
			source:      trip.Stops[i].ID,
			destination: trip.Stops[i+1].ID,
		}

		if _, ok := c.paths[stopPair]; !ok {
			c.paths[stopPair] = getShortestPath(nodesByID, stopPair)
		}
	}
}

func getGraphNodes(route *trip.TramRoute) map[uint64]*city.Set[uint64] {
	nodes := make(map[uint64]*city.Set[uint64])

	for _, stopIDs := range *route.Variants {
		for _, stopID := range stopIDs {
			set := city.NewSet[uint64]()
			nodes[stopID] = &set
		}
	}

	return nodes
}

func getSegmentPathsForRoute(route *trip.TramRoute) [][]uint64 {
	inNodes, outNodes := getGraphNodes(route), getGraphNodes(route)

	for _, stopIDs := range *route.Variants {
		for i := 0; i < len(stopIDs)-1; i++ {
			inNodes[stopIDs[i+1]].Add(stopIDs[i])
			outNodes[stopIDs[i]].Add(stopIDs[i+1])
		}
	}

	var segmentStartNodes []uint64
	for node, neighbors := range inNodes {
		if neighbors.Len() != 1 {
			segmentStartNodes = append(segmentStartNodes, node)
		}
	}

	var segmentPaths [][]uint64
	for _, node := range segmentStartNodes {
		for nextNode := range outNodes[node].GetItems() {
			segment := []uint64{node}

			for inNodes[nextNode].Len() == 1 && outNodes[nextNode].Len() == 1 {
				segment = append(segment, nextNode)
				for node := range outNodes[nextNode].GetItems() {
					nextNode = node
				}
			}

			segment = append(segment, nextNode)
			segmentPaths = append(segmentPaths, segment)
		}
	}

	return segmentPaths
}

func (c *ControlCenter) setSegmentsByRouteName(route *trip.TramRoute) {
	segmentPaths := getSegmentPathsForRoute(route)

	for _, segment := range segmentPaths {
		var polyline []Coordinates

		for i := 0; i < len(segment)-1; i++ {
			path := c.GetPath(segment[i], segment[i+1])

			for _, node := range path.Nodes {
				lat, lon := node.GetCoordinates()
				polyline = append(polyline, Coordinates{Lat: lat, Lon: lon})
			}
		}

		c.segmentsByRouteName[route.Name] = append(c.segmentsByRouteName[route.Name], RouteSegment{
			StopIDs:  segment,
			Polyline: polyline,
		})
	}
}

func (c *ControlCenter) GetPath(sourceNodeID, destinationNodeID uint64) *Path {
	if path, ok := c.paths[stopPair{source: sourceNodeID, destination: destinationNodeID}]; ok {
		return &path
	}

	panic(fmt.Sprintf("No path found between %d and %d nodes", sourceNodeID, destinationNodeID))
}

func (c *ControlCenter) GetSegmentsForRoute(routeName string) []RouteSegment {
	if segments, ok := c.segmentsByRouteName[routeName]; ok {
		return segments
	}

	panic(fmt.Sprintf("Route %s not found", routeName))
}
