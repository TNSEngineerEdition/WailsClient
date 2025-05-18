package city

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
)

type CityData struct {
	TramTrackGraph []GraphNode `json:"tram_track_graph"`
	TramTrips      []TramTrip  `json:"tram_trips"`
	LastUpdated    string      `json:"last_updated"`
}

func (c *CityData) FetchCity(url string) {
	client := &http.Client{}

	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		panic(err)
	}
}

func (c *CityData) GetTramStops() (result []GraphNode) {
	result = make([]GraphNode, 0)

	for _, node := range c.TramTrackGraph {
		if node.isTramStop() {
			result = append(result, node)
		}
	}

	return result
}

func (c *CityData) GetStopsByID() (result map[uint64]*GraphNode) {
	result = make(map[uint64]*GraphNode, len(c.TramTrackGraph))
	for _, node := range c.TramTrackGraph {
		result[node.ID] = &node
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

	for _, node := range c.TramTrackGraph {
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

	return result
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
		sort.Slice(routes, func(i, j int) bool {
			ri, err1 := strconv.Atoi(routes[i])
			rj, err2 := strconv.Atoi(routes[j])
			if err1 == nil && err2 == nil {
				return ri < rj
			}
			return routes[i] < routes[j]
		})
		linesByStopID[stopID] = routes
	}
	return linesByStopID
}

type Arrival struct {
	Route     string
	Headsign  string
	Departure uint
}

func (c *CityData) getRawArrivalsByStopID() map[uint64][]Arrival {
	arrivalsByStopID := make(map[uint64][]Arrival)
	seen := make(map[uint64]map[[2]string]struct{})

	for _, trip := range c.TramTrips {
		for _, s := range trip.Stops {
			stopID := s.ID
			key := [2]string{trip.Route, fmt.Sprint(s.Time)}

			if seen[stopID] == nil {
				seen[stopID] = make(map[[2]string]struct{})
			}
			if _, ok := seen[stopID][key]; ok {
				continue
			}
			seen[stopID][key] = struct{}{}

			arrival := Arrival{
				Route:     trip.Route,
				Headsign:  trip.TripHeadSign,
				Departure: s.Time,
			}
			arrivalsByStopID[stopID] = append(arrivalsByStopID[stopID], arrival)
		}
	}
	return arrivalsByStopID
}

func (c *CityData) GetArrivalsByStopID() map[uint64][]Arrival {
	arrivalsByStopID := c.getRawArrivalsByStopID()
	for stopID, arrivals := range arrivalsByStopID {
		sort.Slice(arrivals, func(i, j int) bool {
			return arrivals[i].Departure < arrivals[j].Departure
		})
		arrivalsByStopID[stopID] = arrivals
	}
	return arrivalsByStopID
}
