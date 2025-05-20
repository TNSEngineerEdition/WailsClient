package simulation

import (
	"math/rand/v2"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/control_room"
	"github.com/umahmood/haversine"
)

type tram struct {
	id                int
	trip              *city.TramTrip
	tripIndex         int
	intermediateIndex int
	Latitude          float32
	Longitude         float32
	coveredDistance   float32
	departureTime     uint
	isFinished        bool
	state             TramState
	c                 control_room.ControlCenter
}

type TramState uint8

const (
	StateTripNotStarted TramState = iota
	StatePassengerTransfer
	StateTravelling
	StateTripFinished
)

func newTram(id int, trip *city.TramTrip) *tram {
	startTime := trip.Stops[0].Time
	return &tram{
		id:                id,
		trip:              trip,
		tripIndex:         0,
		intermediateIndex: 0,
		coveredDistance:   0,
		departureTime:     startTime - uint(rand.IntN(11)) - 15,
		state:             StateTripNotStarted,
	}
}

type TramPositionChange struct {
	TramID    int     `json:"id"`
	Latitude  float32 `json:"lat"`
	Longitude float32 `json:"lon"`
}

func (t *tram) Advance(time uint, stopsById map[uint64]*city.GraphNode, c control_room.ControlCenter) (result TramPositionChange, update bool) {
	distanceToDrive := float32(50*5) / float32(18)
	switch t.state {
	case StateTripNotStarted:

		if time == t.departureTime {
			t.state = StatePassengerTransfer
			result = TramPositionChange{
				TramID:    t.id,
				Latitude:  stopsById[t.trip.Stops[0].ID].Latitude,
				Longitude: stopsById[t.trip.Stops[0].ID].Longitude,
			}
			t.departureTime = t.trip.Stops[0].Time
			update = true
		}

	case StatePassengerTransfer:

		if time == t.departureTime {
			if t.tripIndex == len(t.trip.Stops)-1 {
				t.state = StateTripFinished
			} else {
				t.state = StateTravelling
			}
		}

	case StateTravelling:

		currentStop := t.trip.Stops[t.tripIndex]
		nextStop := t.trip.Stops[t.tripIndex+1]
		path := c.GetRouteBetweenNodes(currentStop.ID, nextStop.ID)
		t.findNewLocation(path, distanceToDrive)
		if t.intermediateIndex == len(path)-1 {
			t.tripIndex += 1
			t.intermediateIndex = 0
			t.departureTime = max(nextStop.Time, time+uint(rand.IntN(11))+15)
			t.state = StatePassengerTransfer
		}
		result = TramPositionChange{
			TramID:    t.id,
			Latitude:  t.Latitude,
			Longitude: t.Longitude,
		}
		update = true

	case StateTripFinished:

		if !t.isFinished {
			result = TramPositionChange{
				TramID:    t.id,
				Latitude:  0,
				Longitude: 0,
			}
			update = true
			t.isFinished = true
		}

	}
	return result, update
}

func (t *tram) findNewLocation(path []*city.GraphNode, distanceToDrive float32) {
	for distanceToDrive > 0 && t.intermediateIndex < len(path)-1 {
		distToNextNode := t.distanceToNextNode(path)
		if distToNextNode <= distanceToDrive {
			distanceToDrive -= distToNextNode
			t.coveredDistance += distToNextNode
			t.intermediateIndex += 1
			t.Latitude = path[t.intermediateIndex].Latitude
			t.Longitude = path[t.intermediateIndex].Longitude
		} else {
			remainingPart := distanceToDrive / distToNextNode
			t.coveredDistance += distanceToDrive
			t.findIntermediateLocation(path, remainingPart)
			distanceToDrive = 0
		}
	}
}

func (t *tram) findIntermediateLocation(path []*city.GraphNode, remainingPart float32) {
	t.Latitude = path[t.intermediateIndex].Latitude + ((path[t.intermediateIndex+1].Latitude - path[t.intermediateIndex].Latitude) * remainingPart)
	t.Longitude = path[t.intermediateIndex].Longitude + ((path[t.intermediateIndex+1].Longitude - path[t.intermediateIndex].Longitude) * remainingPart)
}

func (t *tram) distanceToNextNode(path []*city.GraphNode) float32 {
	sourceCoords := haversine.Coord{Lat: float64(t.Latitude), Lon: float64(t.Longitude)}
	goalCoords := haversine.Coord{Lat: float64(path[t.intermediateIndex+1].Latitude), Lon: float64(path[t.intermediateIndex+1].Longitude)}
	_, km := haversine.Distance(sourceCoords, goalCoords)
	return float32(km * 1000)
}
