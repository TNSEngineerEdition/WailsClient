package trip

import "github.com/TNSEngineerEdition/WailsClient/pkg/api"

type TramRoute struct {
	Name            string
	BackgroundColor string
	TextColor       string
	Trips           []TramTrip
	Variants        *map[string][]uint64
	routeDetails    *api.ResponseTramRoute
}

func NewTramRoute(tramRouteData *api.ResponseTramRoute, tripID *uint) TramRoute {
	tramRoute := TramRoute{
		Name:            tramRouteData.Name,
		BackgroundColor: tramRouteData.BackgroundColor,
		TextColor:       tramRouteData.TextColor,
		Trips:           make([]TramTrip, 0),
		Variants:        tramRouteData.Variants,
		routeDetails:    tramRouteData,
	}

	tramRoute.ResetTrips(tripID)

	return tramRoute
}

func (t *TramRoute) ResetTrips(tripID *uint) {
	for _, item := range *t.routeDetails.Trips {
		t.Trips = append(t.Trips, NewTramTrip(*tripID, &item))
		*tripID += 1
	}
}
