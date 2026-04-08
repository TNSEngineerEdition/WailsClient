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

func TramTripsFromCityData(responseCityData *api.ResponseCityData) []Route {
	tripID := uint(1)
	tramRoutes := make([]Route, len(responseCityData.TramRoutes))

	for i, item := range responseCityData.TramRoutes {
		tramRoutes[i] = NewRoute(&item, &tripID)
	}

	return tramRoutes
}

func (t *Trip) GetScheduledTravelTime(start, end int) uint {
	return t.Stops[end].Time - t.Stops[start].Time
}
