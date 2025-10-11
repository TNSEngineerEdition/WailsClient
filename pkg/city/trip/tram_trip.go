package trip

import "github.com/TNSEngineerEdition/WailsClient/pkg/api"

type TramTrip struct {
	ID           uint
	Stops        []api.ResponseTramTripStop
	TripHeadSign string
}

func NewTramTrip(id uint, tripDetails *api.ResponseTramTrip) TramTrip {
	return TramTrip{
		ID:           id,
		Stops:        tripDetails.Stops,
		TripHeadSign: tripDetails.TripHeadSign,
	}
}

func TramTripsFromCityData(responseCityData *api.ResponseCityData) []TramRoute {
	tripID := uint(1)
	tramRoutes := make([]TramRoute, len(responseCityData.TramRoutes))

	for i, item := range responseCityData.TramRoutes {
		tramRoutes[i] = NewTramRoute(&item, &tripID)
	}

	return tramRoutes
}

func (t *TramTrip) GetScheduledTravelTime(start, end int) uint {
	return t.Stops[end].Time - t.Stops[start].Time
}
