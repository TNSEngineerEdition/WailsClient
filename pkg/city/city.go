package city

import (
	"fmt"
	"slices"
)

type City struct {
	cityData             CityData
	nodesByID, stopsByID map[uint64]*GraphNode
	linesByStopID        map[uint64][]string
	plannedArrivals      map[uint64][]PlannedArrival
}

func (c *City) FetchCityData(url string) {
	c.cityData.FetchCity(url)
	c.nodesByID = c.cityData.GetNodesByID()
	c.stopsByID = c.cityData.GetStopsByID()
	c.linesByStopID = c.cityData.GetLinesByStopID()
	c.ResetPlannedArrivals()
}

func (c *City) ResetPlannedArrivals() {
	c.plannedArrivals = c.cityData.GetPlannedArrivals()
}

func (c *City) GetTramStops() []TramStop {
	return c.cityData.GetTramStops()
}

func (c *City) GetTramTrips() []TramTrip {
	return c.cityData.TramTrips
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

func (c *City) GetLinesForStop(stopID uint64, chipPerRowSize int) []string {
	lines, ok := c.linesByStopID[stopID]
	if !ok {
		return []string{}
	}

	// Make a copy of the slice to avoid different results across multiple calls.
	processedLines := slices.Clone(lines)

	for start := 0; start < len(processedLines); start += chipPerRowSize {
		end := min(start+chipPerRowSize, len(processedLines))
		slices.Reverse(processedLines[start:end])
	}
	return processedLines
}
