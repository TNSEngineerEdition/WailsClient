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
	initialPassengers    map[uint][]*Passenger
}

func (c *City) FetchCityData(url string) {
	c.cityData.FetchCity(url)
	c.nodesByID = c.cityData.GetNodesByID()
	c.stopsByID = c.cityData.GetStopsByID()
	c.routesByStopID = c.cityData.GetRoutesByStopID()
	c.initialPassengers = c.loadInitialPassengers()
}

func (c *City) loadInitialPassengers() map[uint][]*Passenger {
	cp := NewCityPassengers(c)
	cp.CreatePassengers()
	return cp.initialPassengers
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

func (c *City) GetPassengersAt(t uint) []*Passenger {
	return c.initialPassengers[t]
}

func (c *City) UnblockGraph() {
	for _, node := range c.nodesByID {
		node.Unblock(0)
	}
}

func (c *City) ResetPassengers() {
	for _, stop := range c.stopsByID {
		stop.AwaitingPassengers = make([]*Passenger, 0)
	}
}
