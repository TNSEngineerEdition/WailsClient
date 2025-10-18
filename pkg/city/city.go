package city

import (
	"fmt"
	"slices"
)

type City struct {
	cityData             CityData
	nodesByID, stopsByID map[uint64]*GraphNode
	routesByStopID       map[uint64][]RouteInfo
	plannedArrivals      map[uint64][]PlannedArrival
}

func (c *City) FetchCityData(url string) {
	c.cityData.FetchCity(url)
	c.nodesByID = c.cityData.GetNodesByID()
	c.stopsByID = c.cityData.GetStopsByID()
	c.routesByStopID = c.cityData.GetRoutesByStopID()
	c.ResetPlannedArrivals()
}

func (c *City) ResetPlannedArrivals() {
	c.plannedArrivals = c.cityData.GetPlannedArrivals()
}

func (c *City) GetTramStops() []TramStop {
	return c.cityData.GetTramStops()
}

func (c *City) GetTramRoutes() []TramRoute {
	return c.cityData.TramRoutes
}

func (c *City) GetPlannedArrivals(stopID uint64) *[]PlannedArrival {
	if arrivals, ok := c.plannedArrivals[stopID]; ok {
		return &arrivals
	}

	panic(fmt.Sprintf("Stop ID %d not found", stopID))
}

func (c *City) GetBounds() LatLonBounds {
	return c.cityData.GetBounds()
}

func (c *City) GetNodesByID() map[uint64]*GraphNode {
	return c.nodesByID
}

func (c *City) GetStopsByID() map[uint64]*GraphNode {
	return c.stopsByID
}

func (c *City) GetTimeBounds() TimeBounds {
	return c.cityData.GetTimeBounds()
}

func (c *City) GetRoutesForStop(stopID uint64, chipPerRowSize int) []RouteInfo {
	routes, ok := c.routesByStopID[stopID]
	if !ok {
		return []RouteInfo{}
	}

	// Make a copy of the slice to avoid different results across multiple calls.
	processedRoutes := slices.Clone(routes)

	for start := 0; start < len(processedRoutes); start += chipPerRowSize {
		end := min(start+chipPerRowSize, len(processedRoutes))
		slices.Reverse(processedRoutes[start:end])
	}
	return processedRoutes
}

type CityRectangle struct {
	Bounds    LatLonBounds          `json:"bounds"`
	NodesByID map[uint64]*GraphNode `json:"nodes_by_id"`
}

func (c *City) GetCityRectangles() (cityRectangles []CityRectangle) {
	bounds := c.cityData.GetBounds()
	latDistance := bounds.MaxLat - bounds.MinLat
	lonDistance := bounds.MaxLon - bounds.MinLon

	const nRows int = 6 // based on lat
	const nCols int = 7 // based on lon

	rowSize := latDistance / float32(nRows)
	colSize := lonDistance / float32(nCols)

	for i := range nRows {
		for j := range nCols {
			bounds := LatLonBounds{
				MinLat: bounds.MinLat + (float32(i) * rowSize),
				MinLon: bounds.MinLon + (float32(j) * colSize),
				MaxLat: bounds.MinLat + (float32(i+1) * rowSize),
				MaxLon: bounds.MinLon + (float32(j+1) * colSize),
			}
			nodesByID := c.getNodesForBounds(bounds)

			if len(nodesByID) == 0 {
				continue
			}

			rect := CityRectangle{
				Bounds:    bounds,
				NodesByID: c.getNodesForBounds(bounds),
			}
			cityRectangles = append(cityRectangles, rect)
		}
	}

	return cityRectangles
}

func (c *City) getNodesForBounds(bounds LatLonBounds) (nodesByID map[uint64]*GraphNode) {
	nodesByID = make(map[uint64]*GraphNode)
	for _, node := range c.nodesByID {
		if c.isInBounds(node.Latitude, node.Longitude, bounds) {
			nodesByID[node.ID] = node

			// add neighbors right outside of bounds for the border nodes
			for neighborID := range node.Neighbors {
				neighbor := c.nodesByID[neighborID]
				if !c.isInBounds(neighbor.Latitude, neighbor.Longitude, bounds) {
					nodesByID[neighborID] = neighbor
				}
			}
		}
	}
	return nodesByID
}

func (c *City) isInBounds(lat, lon float32, bounds LatLonBounds) bool {
	return lat >= bounds.MinLat && lat <= bounds.MaxLat && lon >= bounds.MinLon && lon <= bounds.MaxLon
}

type Modifications struct {
	NeighborMaxSpeed map[uint64]float32 `json:"neighborMaxSpeed"`
}

func (c *City) UpdateTramTrackGraph(modifiedNodes map[uint64]Modifications) {
	for nodeID, mods := range modifiedNodes {
		if node, ok := c.nodesByID[nodeID]; ok {
			for neighborID, maxSpeed := range mods.NeighborMaxSpeed {
				if nodeNeighbor, ok := node.Neighbors[neighborID]; ok {
					nodeNeighbor.MaxSpeed = maxSpeed
					node.Neighbors[neighborID] = nodeNeighbor
				}
			}
		}
	}
}

func (c *City) UnblockGraph() {
	for _, node := range c.nodesByID {
		node.Unblock(0)
	}
}
