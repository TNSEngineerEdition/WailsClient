package simulation

import (
	"fmt"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
)

type tram struct {
	id         int
	trip       *city.TramTrip
	tripIndex  int
	isFinished bool
}

func newTram(id int, trip *city.TramTrip) *tram {
	return &tram{
		id:         id,
		trip:       trip,
		tripIndex:  -1,
		isFinished: false,
	}
}

type TramPositionChange struct {
	TramID    int     `json:"id"`
	Latitude  float32 `json:"lat"`
	Longitude float32 `json:"lon"`
}

func (t *tram) Advance(time uint, stopsById map[uint64]*city.GraphNode) (result TramPositionChange, update bool) {
	nextTripIndex := t.tripIndex + 1

	if nextTripIndex == len(t.trip.Stops) && !t.isFinished {
		result = TramPositionChange{
			TramID:    t.id,
			Latitude:  0,
			Longitude: 0,
		}
		update = true

		t.isFinished = true
	} else if nextTripIndex < len(t.trip.Stops) && t.trip.Stops[nextTripIndex].Time == time {
		tramStop, ok := stopsById[t.trip.Stops[nextTripIndex].ID]
		if !ok {
			panic(fmt.Sprintf("Tram stop with ID %d should exist", t.trip.Stops[nextTripIndex].ID))
		}

		result = TramPositionChange{
			TramID:    t.id,
			Latitude:  tramStop.Latitude,
			Longitude: tramStop.Longitude,
		}
		update = true

		t.tripIndex = nextTripIndex
	}

	return result, update
}

type TramDetails struct {
	Route        string              `json:"route"`
	TripHeadSign string              `json:"trip_head_sign"`
	Stops        []city.TramTripStop `json:"stops"`
	StopNames    []string            `json:"stop_names"`
	Speed        uint8               `json:"speed"`
}

func (t *tram) GetDetails(c *city.City) TramDetails {
	stopsByID := c.GetStopsByID()
	var stopNames []string

	for _, stop := range t.trip.Stops {
		stopNames = append(stopNames, *stopsByID[stop.ID].Name)
	}

	return TramDetails{
		Route:        t.trip.Route,
		TripHeadSign: t.trip.TripHeadSign,
		Stops:        t.trip.Stops,
		StopNames:    stopNames,
		Speed:        50,
	}
}
