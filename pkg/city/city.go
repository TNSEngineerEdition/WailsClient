package city

import (
	"fmt"
	"math"
	"slices"

	"github.com/TNSEngineerEdition/WailsClient/pkg/api"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/graph"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/trip"
	"github.com/facette/natsort"
	"github.com/oapi-codegen/runtime/types"
)

type City struct {
	CityID          string
	tramRoutes      []trip.TramRoute
	nodesByID       map[uint64]graph.GraphNode
	stopsByID       map[uint64]*graph.GraphTramStop
	routesByStopID  map[uint64][]RouteInfo
	plannedArrivals map[uint64][]PlannedArrival
	bounds          LatLonBounds
}

type FetchCityParams struct {
	Weekday *api.Weekday
	Date    *types.Date
}

func (c *City) FetchCity(
	apiClient *api.APIClient,
	cityID string,
	parameters *FetchCityParams,
	customSchedule []byte,
) error {
	var responseCityData *api.ResponseCityData
	var err error

	if len(customSchedule) == 0 {
		responseCityData, err = apiClient.GetCityByID(
			cityID,
			&api.GetCityDataCitiesCityIdGetParams{
				Weekday: parameters.Weekday,
				Date:    parameters.Date,
			},
		)
	} else {
		responseCityData, err = apiClient.GetCityByIDWithCustomSchedule(
			cityID,
			customSchedule,
			&api.GetCityDataWithCustomScheduleCitiesCityIdPostParams{
				Weekday: parameters.Weekday,
			},
		)
	}

	if err != nil {
		return err
	}

	c.CityID = cityID

	c.tramRoutes = trip.TramTripsFromCityData(responseCityData)

	if nodesByID, err := graph.GraphNodesFromCityData(responseCityData); err == nil {
		c.nodesByID = nodesByID
	} else {
		return err
	}

	c.stopsByID = make(map[uint64]*graph.GraphTramStop)
	for nodeID, node := range c.nodesByID {
		switch v := node.(type) {
		case *graph.GraphTramStop:
			c.stopsByID[nodeID] = v
		}
	}

	c.CityID = cityID
	c.routesByStopID = c.GetRoutesByStopID()
	c.Reset()

	c.bounds = GetBoundsFromNodes(c.nodesByID)

	return nil
}

func (c *City) Reset() {
	for _, node := range c.nodesByID {
		node.ForceUnblock()
	}

	c.plannedArrivals = c.GetInitialPlannedArrivals()
}

func (c *City) GetNodesByID() map[uint64]graph.GraphNode {
	return c.nodesByID
}

func (c *City) GetStopsByID() map[uint64]*graph.GraphTramStop {
	return c.stopsByID
}

func (c *City) GetStops() []api.ResponseGraphTramStop {
	result := make([]api.ResponseGraphTramStop, 0, len(c.stopsByID))

	for _, stop := range c.stopsByID {
		result = append(result, stop.GetDetails())
	}

	return result
}

func (c *City) GetTramRoutes() []trip.TramRoute {
	return c.tramRoutes
}

func (c *City) GetBounds() LatLonBounds {
	return c.bounds
}

type RouteInfo struct {
	Name            string `json:"name"`
	TextColor       string `json:"text_color"`
	BackgroundColor string `json:"background_color"`
}

func (c *City) GetRoutesByStopID() map[uint64][]RouteInfo {
	routeSetByStopID := make(map[uint64]map[string]struct{})

	for _, route := range c.tramRoutes {
		route.AddRouteNamesToStopSet(&routeSetByStopID)
	}

	routeNamesByStopID := make(map[uint64][]string, len(routeSetByStopID))

	for stopID, routeSet := range routeSetByStopID {
		routes := make([]string, 0, len(routeSet))
		for r := range routeSet {
			routes = append(routes, r)
		}

		natsort.Sort(routes)
		routeNamesByStopID[stopID] = routes
	}

	routesByName := make(map[string]trip.TramRoute, len(c.tramRoutes))
	for _, route := range c.tramRoutes {
		routesByName[route.Name] = route
	}

	routesByStopID := make(map[uint64][]RouteInfo, len(routeNamesByStopID))
	for stopID, routeNames := range routeNamesByStopID {
		for _, routeName := range routeNames {
			route := routesByName[routeName]

			routesByStopID[stopID] = append(routesByStopID[stopID], RouteInfo{
				Name:            route.Name,
				TextColor:       "#" + route.TextColor,
				BackgroundColor: "#" + route.BackgroundColor,
			})
		}
	}

	return routesByStopID
}

type PlannedArrival struct {
	TripID    uint
	StopIndex int
	Time      uint
}

func (c *City) GetInitialPlannedArrivals() map[uint64][]PlannedArrival {
	plannedArrivals := make(map[uint64][]PlannedArrival, len(c.stopsByID))

	for _, route := range c.tramRoutes {
		for _, trip := range route.Trips {
			for stopIndex, stop := range trip.Stops {
				plannedArrivals[stop.ID] = append(plannedArrivals[stop.ID], PlannedArrival{
					TripID:    trip.ID,
					StopIndex: stopIndex,
					Time:      stop.Time,
				})
			}
		}
	}

	for _, arrivals := range plannedArrivals {
		slices.SortFunc(arrivals, func(a1, a2 PlannedArrival) int {
			return int(a1.Time) - int(a2.Time)
		})
	}

	return plannedArrivals
}

func (c *City) GetPlannedArrivals(stopID uint64) *[]PlannedArrival {
	if arrivals, ok := c.plannedArrivals[stopID]; ok {
		return &arrivals
	}

	panic(fmt.Sprintf("Stop ID %d not found", stopID))
}

type TimeBounds struct {
	StartTime uint `json:"startTime"`
	EndTime   uint `json:"endTime"`
}

func (c *City) GetTimeBounds() (result TimeBounds) {
	result.StartTime = math.MaxInt

	for _, route := range c.tramRoutes {
		for _, trip := range route.Trips {
			result.StartTime = min(result.StartTime, trip.Stops[0].Time)
			result.EndTime = max(result.EndTime, trip.Stops[len(trip.Stops)-1].Time)
		}
	}

	result.StartTime = max(result.StartTime-60, 0)
	return
}

func (c *City) GetRoutesForStop(stopID uint64, chipPerRowSize int) []RouteInfo {
	routes, ok := c.routesByStopID[stopID]
	if !ok {
		return []RouteInfo{}
	}

	// Make a copy of the slice to avoid different results across multiple calls.
	processedRoutes := slices.Clone(routes)

	for start := 0; start < len(processedRoutes); start += chipPerRowSize {
		end := min(start+chipPerRowSize, len(processedRoutes))
		slices.Reverse(processedRoutes[start:end])
	}
	return processedRoutes
}

type CityRectangle struct {
	Bounds    LatLonBounds               `json:"bounds"`
	NodesByID map[uint64]graph.GraphNode `json:"nodes_by_id"`
}

func (c *City) GetCityRectangles() []CityRectangle {
	//TODO: make these values based on city bounds
	const (
		nRows int = 6 // based on lat
		nCols int = 7 // based on lon
	)

	latDistance := c.bounds.MaxLat - c.bounds.MinLat
	lonDistance := c.bounds.MaxLon - c.bounds.MinLon

	rowSize := latDistance / float32(nRows)
	colSize := lonDistance / float32(nCols)

	rects := make([]CityRectangle, 0, nRows*nCols)
	for i := range nRows {
		for j := range nCols {
			rectBounds := LatLonBounds{
				MinLat: c.bounds.MinLat + float32(i)*rowSize,
				MaxLat: c.bounds.MinLat + float32(i+1)*rowSize,
				MinLon: c.bounds.MinLon + float32(j)*colSize,
				MaxLon: c.bounds.MinLon + float32(j+1)*colSize,
			}
			rects = append(rects, CityRectangle{
				Bounds:    rectBounds,
				NodesByID: make(map[uint64]graph.GraphNode),
			})
		}
	}

	for _, node := range c.nodesByID {
		lat, lon := node.GetCoordinates()

		for i := range rects {
			if rects[i].Bounds.isInBounds(lat, lon) {
				rects[i].NodesByID[node.GetID()] = node

				// add neighbors right outside of bounds for the border nodes
				for neighborID := range node.GetNeighbors() {
					neighbor := c.nodesByID[neighborID]
					neiLat, neiLon := neighbor.GetCoordinates()
					if !rects[i].Bounds.isInBounds(neiLat, neiLon) {
						rects[i].NodesByID[neighborID] = neighbor
					}
				}
			}
		}
	}

	filteredRects := make([]CityRectangle, 0, len(rects))
	for _, r := range rects {
		if len(r.NodesByID) > 0 {
			filteredRects = append(filteredRects, r)
		}
	}

	return filteredRects
}

type Modifications struct {
	NeighborMaxSpeed map[uint64]float32 `json:"neighborMaxSpeed"`
}

func (c *City) UpdateTramTrackGraph(modifiedNodes map[uint64]Modifications) {
	for nodeID, mods := range modifiedNodes {
		node, ok := c.nodesByID[nodeID]
		if !ok {
			continue
		}
		for neighborID, maxSpeed := range mods.NeighborMaxSpeed {
			node.UpdateMaxSpeed(neighborID, maxSpeed)
		}
	}
}

func (c *City) UnblockGraph() {
	for _, node := range c.nodesByID {
		node.Unblock(0)
	}
}
