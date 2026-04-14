package trip

import "github.com/TNSEngineerEdition/WailsClient/pkg/api"

type Trip struct {
	ID           uint
	Stops        []api.ResponseTripStop
	TripHeadSign string
}

func NewTrip(id uint, tripDetails *api.ResponseTrip) Trip {
	return Trip{
		ID:           id,
		Stops:        tripDetails.Stops,
		TripHeadSign: tripDetails.TripHeadSign,
	}
}

func VehicleTripsFromCityData(responseCityData *api.ResponseCityData) []Route {
	tripID := uint(1)

	totalRoutes := len(responseCityData.TramRoutes) + len(responseCityData.BusRoutes)
	routes := make([]Route, 0, totalRoutes)

	for _, item := range responseCityData.TramRoutes {
		routes = append(routes, NewRoute(&item, &tripID, api.Tram))
	}

	for _, item := range responseCityData.BusRoutes {
		routes = append(routes, NewRoute(&item, &tripID, api.Bus))
	}

	return routes
}

func (t *Trip) GetScheduledTravelTime(start, end int) uint {
	return t.Stops[end].Time - t.Stops[start].Time
}
