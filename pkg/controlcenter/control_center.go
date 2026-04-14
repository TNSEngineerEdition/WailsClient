package controlcenter

import (
	"fmt"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/trip"
	"github.com/TNSEngineerEdition/WailsClient/pkg/structs"
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
	controlCenter.setPaths(city)

	tramRoutes := city.GetTramRoutes()
	for _, route := range tramRoutes {
		if route.Variants == nil {
			continue
		}

		controlCenter.setSegmentsByRouteName(&route)
	}

	return controlCenter
}

func (c *ControlCenter) setPaths(city *city.City) {
	nodesByID := city.GetNodesByID()
	paths := city.GetPaths()

	for startNodeID := range paths {
		for endNodeID := range paths[startNodeID] {
			stopPair := stopPair{
				source:      startNodeID,
				destination: endNodeID,
			}

			c.paths[stopPair] = pathFromNodeIDs(
				paths[startNodeID][endNodeID],
				&nodesByID,
			)
		}
	}
}

func getGraphNodes(route *trip.Route) map[uint64]*structs.Set[uint64] {
	nodes := make(map[uint64]*structs.Set[uint64])

	for _, stopIDs := range *route.Variants {
		for _, stopID := range stopIDs {
			set := structs.NewSet[uint64]()
			nodes[stopID] = &set
		}
	}

	return nodes
}

func getSegmentPathsForRoute(route *trip.Route) [][]uint64 {
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

func (c *ControlCenter) setSegmentsByRouteName(route *trip.Route) {
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
