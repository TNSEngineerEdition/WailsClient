package trip

import "github.com/TNSEngineerEdition/WailsClient/pkg/api"

type TramRoute struct {
	Name            string
	BackgroundColor string
	TextColor       string
	Trips           []TramTrip
	routeDetails    *api.ResponseTramRoute
}

func NewTramRoute(tramRouteData *api.ResponseTramRoute, tripID *uint) TramRoute {
	tramRoute := TramRoute{
		Name:            tramRouteData.Name,
		BackgroundColor: tramRouteData.BackgroundColor,
		TextColor:       tramRouteData.TextColor,
		Trips:           make([]TramTrip, 0),
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

func (t *TramRoute) AddRouteNamesToStopSet(routeSetByStopID *map[uint64]map[string]struct{}) {
	for _, trip := range t.Trips {
		for _, stop := range trip.Stops {
			if _, ok := (*routeSetByStopID)[stop.ID]; !ok {
				(*routeSetByStopID)[stop.ID] = make(map[string]struct{})
			}

			(*routeSetByStopID)[stop.ID][t.Name] = struct{}{}
		}
	}
}
