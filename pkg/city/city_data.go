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
	TramTrips      []TramTrip  `json:"tram_trips"`
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
}

func (c *CityData) GetTramStops() (result []TramStop) {
	result = make([]TramStop, 0)

	for i := range c.TramTrackGraph {
		node := &c.TramTrackGraph[i]
		if node.isTramStop() {
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
		if node.isTramStop() {
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

	for _, trip := range c.TramTrips {
		for _, stop := range trip.Stops {
			result.StartTime = min(result.StartTime, stop.Time)
			result.EndTime = max(result.EndTime, stop.Time)
		}
	}

	result.StartTime -= 60
	return
}

func (c *CityData) getLineSetsByStopID() map[uint64]map[string]struct{} {
	set := make(map[uint64]map[string]struct{})
	for _, trip := range c.TramTrips {
		for _, stop := range trip.Stops {
			if _, ok := set[stop.ID]; !ok {
				set[stop.ID] = make(map[string]struct{})
			}
			set[stop.ID][trip.Route] = struct{}{}
		}
	}
	return set
}

func (c *CityData) GetLinesByStopID() map[uint64][]string {
	routeSets := c.getLineSetsByStopID()
	linesByStopID := make(map[uint64][]string)
	for stopID, routeSet := range routeSets {
		routes := make([]string, 0, len(routeSet))
		for r := range routeSet {
			routes = append(routes, r)
		}
		natsort.Sort(routes)
		linesByStopID[stopID] = routes
	}
	return linesByStopID
}

type PlannedArrival struct {
	TramID    int
	StopIndex int
	Time      uint
}

func (c *CityData) GetPlannedArrivals() map[uint64][]PlannedArrival {
	stops := c.GetStopsByID()
	plannedArrivals := make(map[uint64][]PlannedArrival, len(stops))

	for tramID, trip := range c.TramTrips {
		for stopIndex, stop := range trip.Stops {
			plannedArrivals[stop.ID] = append(plannedArrivals[stop.ID], PlannedArrival{
				TramID:    tramID,
				StopIndex: stopIndex,
				Time:      stop.Time,
			})
		}
	}

	for _, arrivals := range plannedArrivals {
		slices.SortFunc(arrivals, func(a1, a2 PlannedArrival) int {
			return int(a1.Time) - int(a2.Time)
		})
	}

	return plannedArrivals
}
