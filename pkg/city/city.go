package city

type City struct {
	cityData  CityData
	stopsById map[uint64]*GraphNode
}

func (c *City) FetchCityData(url string) {
	c.cityData.FetchCity(url)
	c.stopsById = c.cityData.GetStopsByID()
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
