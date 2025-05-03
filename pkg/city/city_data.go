package city

import (
	"encoding/json"
	"math"
	"net/http"
)

type CityData struct {
	TramTrackGraph []GraphNode `json:"tram_track_graph"`
	TramTrips      []TramTrip  `json:"tram_trips"`
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
