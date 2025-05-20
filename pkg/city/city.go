package city

import (
	"slices"
)

type City struct {
	cityData         CityData
	stopsById        map[uint64]*GraphNode
	linesByStopID    map[uint64][]string
	arrivalsByStopID map[uint64][]Arrival
}

func (c *City) FetchCityData(url string) {
	c.cityData.FetchCity(url)
	c.stopsById = c.cityData.GetStopsByID()
	c.linesByStopID = c.cityData.GetLinesByStopID()
	c.arrivalsByStopID = c.cityData.GetArrivalsByStopID()
}

func (c *City) GetTramStops() []GraphNode {
	return c.cityData.GetTramStops()
}

func (c *City) GetTramTrips() []TramTrip {
	return c.cityData.TramTrips
}

func (c *City) GetBounds() LatLonBounds {
	return c.cityData.GetBounds()
}

func (c *City) GetStopsByID() map[uint64]*GraphNode {
	return c.stopsById
}

func (c *City) GetTimeBounds() TimeBounds {
	return c.cityData.GetTimeBounds()
}

func (c *City) GetLinesForStop(stopID uint64, chipPerRowSize int) []string {
	if lines, ok := c.linesByStopID[stopID]; ok {
		for start := 0; start < len(lines); start += chipPerRowSize {
			end := start + chipPerRowSize
			if end > len(lines) {
				end = len(lines)
			}
			slices.Reverse(lines[start:end])
		}
		return lines
	} else {
		return []string{}
	}
}

func (c *City) GetArrivalsForStop(stopID uint64, currentTime uint, numberOfArrivals int) (upcoming []Arrival) {
	upcoming = []Arrival{}
	for _, arrival := range c.arrivalsByStopID[stopID] {
		if arrival.ETA+30 >= currentTime {
			diff := arrival.ETA - currentTime
			var eta uint = 0
			if diff > 0 {
				eta = uint((diff + 59) / 60)
			}
			upcoming = append(upcoming, Arrival{
				Route:    arrival.Route,
				Headsign: arrival.Headsign,
				ETA:      eta,
			})
		}
		if len(upcoming) == numberOfArrivals {
			return
		}
	}
	return
}
