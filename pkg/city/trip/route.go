package trip

import "github.com/TNSEngineerEdition/WailsClient/pkg/api"

type Route struct {
	Name            string
	BackgroundColor string
	TextColor       string
	Trips           []Trip
	Variants        *map[string][]uint64
	routeDetails    *api.ResponseRoute
	TransitType     api.TransitType
}

func NewRoute(vehicleRouteData *api.ResponseRoute, tripID *uint, transitType api.TransitType) Route {
	vehicleRoute := Route{
		Name:            vehicleRouteData.Name,
		BackgroundColor: vehicleRouteData.BackgroundColor,
		TextColor:       vehicleRouteData.TextColor,
		Trips:           make([]Trip, 0),
		Variants:        vehicleRouteData.Variants,
		routeDetails:    vehicleRouteData,
		TransitType:     transitType,
	}

	vehicleRoute.ResetTrips(tripID)

	return vehicleRoute
}

func (t *Route) ResetTrips(tripID *uint) {
	for _, item := range *t.routeDetails.Trips {
		t.Trips = append(t.Trips, NewTrip(*tripID, &item))
		*tripID += 1
	}
}
