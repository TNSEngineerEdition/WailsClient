package city

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

func (c *City) GetLinesForStop(stopID uint64) []string {
	if lines, ok := c.linesByStopID[stopID]; ok {
		return lines
	}
	return []string{}
}

func (c *City) GetArrivalsForStop(stopID uint64, currentTime uint) (upcoming []Arrival) {
	for _, arrival := range c.arrivalsByStopID[stopID] {
		if arrival.Departure >= currentTime {
			upcoming = append(upcoming, arrival)
		}
	}
	return
}
