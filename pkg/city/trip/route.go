package trip

import "github.com/TNSEngineerEdition/WailsClient/pkg/api"

type Route struct {
	Name            string
	BackgroundColor string
	TextColor       string
	Trips           []Trip
	Variants        *map[string][]uint64
	routeDetails    *api.ResponseRoute
}

func NewRoute(tramRouteData *api.ResponseRoute, tripID *uint) Route {
	tramRoute := Route{
		Name:            tramRouteData.Name,
		BackgroundColor: tramRouteData.BackgroundColor,
		TextColor:       tramRouteData.TextColor,
		Trips:           make([]Trip, 0),
		Variants:        tramRouteData.Variants,
		routeDetails:    tramRouteData,
	}

	tramRoute.ResetTrips(tripID)

	return tramRoute
}

func (t *Route) ResetTrips(tripID *uint) {
	for _, item := range *t.routeDetails.Trips {
		t.Trips = append(t.Trips, NewTrip(*tripID, &item))
		*tripID += 1
	}
}
