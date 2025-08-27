package city

import (
	"encoding/json"
	"math"
	"net/http"
	"slices"

	"github.com/facette/natsort"
)

var ServerURL = "http://localhost:8000"

type CityData struct {
	TramTrackGraph []GraphNode `json:"tram_track_graph"`
	TramRoutes     []TramRoute `json:"tram_routes"`
	LastUpdated    string      `json:"last_updated"`
}

func (c *CityData) FetchCity(cityID string) {
	client := &http.Client{}

	resp, err := client.Get(ServerURL + "/cities/" + cityID)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		panic(err)
	}

	startTripID := uint(1)
	for i := range c.TramRoutes {
		c.TramRoutes[i].assignTripIDs(startTripID)
		startTripID += uint(len(c.TramRoutes[i].Trips))
	}
}

func (c *CityData) GetTramStops() (result []TramStop) {
	result = make([]TramStop, 0)

	for i := range c.TramTrackGraph {
		node := &c.TramTrackGraph[i]
		if node.IsTramStop() {
			result = append(result, node.getTramStopDetails())
		}
	}

	return result
}

func (c *CityData) GetNodesByID() (result map[uint64]*GraphNode) {
	result = make(map[uint64]*GraphNode, len(c.TramTrackGraph))
	for i := range c.TramTrackGraph {
		node := &c.TramTrackGraph[i]
		result[node.ID] = node
	}

	return result
}

func (c *CityData) GetStopsByID() map[uint64]*GraphNode {
	result := make(map[uint64]*GraphNode)

	for i := range c.TramTrackGraph {
		node := &c.TramTrackGraph[i]
		if node.IsTramStop() {
			result[node.ID] = node
		}
	}

	return result
}

type LatLonBounds struct {
	MinLat float32 `json:"minLat"`
	MinLon float32 `json:"minLon"`
	MaxLat float32 `json:"maxLat"`
	MaxLon float32 `json:"maxLon"`
}

func (c *CityData) GetBounds() LatLonBounds {
	minLat, minLon := float32(math.Inf(1)), float32(math.Inf(1))
	maxLat, maxLon := float32(math.Inf(-1)), float32(math.Inf(-1))

	for i := range c.TramTrackGraph {
		node := &c.TramTrackGraph[i]
		minLat = min(minLat, node.Latitude)
		minLon = min(minLon, node.Longitude)
		maxLat = max(maxLat, node.Latitude)
		maxLon = max(maxLon, node.Longitude)
	}

	return LatLonBounds{
		MinLat: minLat,
		MinLon: minLon,
		MaxLat: maxLat,
		MaxLon: maxLon,
	}
}

type TimeBounds struct {
	StartTime uint `json:"startTime"`
	EndTime   uint `json:"endTime"`
}

func (c *CityData) GetTimeBounds() (result TimeBounds) {
	result.StartTime = math.MaxUint

	for _, route := range c.TramRoutes {
		for _, trip := range route.Trips {
			result.StartTime = min(result.StartTime, trip.Stops[0].Time)
			result.EndTime = max(result.EndTime, trip.Stops[len(trip.Stops)-1].Time)
		}
	}

	result.StartTime -= 60
	return
}

func (c *CityData) getRouteSetsByStopID() map[uint64]map[string]struct{} {
	routeSetByStopID := make(map[uint64]map[string]struct{})

	for _, route := range c.TramRoutes {
		for _, trip := range route.Trips {
			for _, stop := range trip.Stops {
				if _, ok := routeSetByStopID[stop.ID]; !ok {
					routeSetByStopID[stop.ID] = make(map[string]struct{})
				}

				routeSetByStopID[stop.ID][route.Name] = struct{}{}
			}
		}
	}

	return routeSetByStopID
}

type RouteInfo struct {
	Name            string `json:"name"`
	TextColor       string `json:"text_color"`
	BackgroundColor string `json:"background_color"`
}

func (c *CityData) GetRoutesByStopID() map[uint64][]RouteInfo {
	routeSetByStopID := c.getRouteSetsByStopID()
	routeNamesByStopID := make(map[uint64][]string, len(routeSetByStopID))

	for stopID, routeSet := range routeSetByStopID {
		routes := make([]string, 0, len(routeSet))
		for r := range routeSet {
			routes = append(routes, r)
		}
		natsort.Sort(routes)
		routeNamesByStopID[stopID] = routes
	}

	routesByName := make(map[string]TramRoute, len(c.TramRoutes))
	for _, route := range c.TramRoutes {
		routesByName[route.Name] = route
	}

	routesByStopID := make(map[uint64][]RouteInfo, len(routeNamesByStopID))
	for stopID, routeNames := range routeNamesByStopID {
		for _, routeName := range routeNames {
			if route, ok := routesByName[routeName]; ok {
				routesByStopID[stopID] = append(routesByStopID[stopID], RouteInfo{
					Name:            route.Name,
					TextColor:       "#" + route.TextColor,
					BackgroundColor: "#" + route.BackgroundColor,
				})
			}
		}
	}

	return routesByStopID
}

type PlannedArrival struct {
	TripID    uint
	StopIndex int
	Time      uint
}

func (c *CityData) GetPlannedArrivals() map[uint64][]PlannedArrival {
	stops := c.GetStopsByID()
	plannedArrivals := make(map[uint64][]PlannedArrival, len(stops))

	for _, route := range c.TramRoutes {
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
