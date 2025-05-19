package simulation

import (
	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/control_room"
	"github.com/umahmood/haversine"
	"golang.org/x/exp/rand"
)

type tram struct {
	id                int
	trip              *city.TramTrip
	tripIndex         int
	intermediateIndex int
	isFinished        bool
	Latitude          float32
	Longitude         float32
	coveredDistance   float32
	departureTime     uint
	c                 control_room.ControlCenter
}

func newTram(id int, trip *city.TramTrip) *tram {
	startTime := trip.Stops[0].Time
	return &tram{
		id:                id,
		trip:              trip,
		tripIndex:         0,
		intermediateIndex: 0,
		isFinished:        false,
		coveredDistance:   0,
		departureTime:     startTime,
	}
}

type TramPositionChange struct {
	TramID    int     `json:"id"`
	Latitude  float32 `json:"lat"`
	Longitude float32 `json:"lon"`
}

func (t *tram) Advance(time uint, stopsById map[uint64]*city.GraphNode, c *control_room.ControlCenter) (result TramPositionChange, update bool) {
	distanceToDrive := float32(50*5) / float32(18)
	if time >= t.departureTime && !t.isFinished {

		if t.tripIndex == len(t.trip.Stops)-1 && time == t.departureTime {
			result = TramPositionChange{
				TramID:    t.id,
				Latitude:  0,
				Longitude: 0,
			}
			update = true
			t.isFinished = true
			return
		}

		currentStop := t.trip.Stops[t.tripIndex]
		nextTripIndex := t.tripIndex + 1

		nextStop := t.trip.Stops[nextTripIndex]
		path := c.GetRouteBetweenNodes(currentStop.ID, nextStop.ID)
		t.calculateNewLocation(path, distanceToDrive)
		if t.intermediateIndex == len(path)-1 {
			t.tripIndex += 1
			t.intermediateIndex = 0
			t.departureTime = max(nextStop.Time, time+uint(rand.Intn(11))+15)
		}
		result = TramPositionChange{
			TramID:    t.id,
			Latitude:  t.Latitude,
			Longitude: t.Longitude,
		}
		update = true
	}
	return
}

func (t *tram) calculateNewLocation(path []*city.GraphNode, distanceToDrive float32) {
	for distanceToDrive > 0 && t.intermediateIndex < len(path)-1 {
		distToNextNode := calculateDistance(t.Latitude, t.Longitude, path[t.intermediateIndex+1])
		if distToNextNode <= distanceToDrive {
			distanceToDrive -= distToNextNode
			t.coveredDistance += distToNextNode
			t.intermediateIndex += 1
			t.Latitude = path[t.intermediateIndex].Latitude
			t.Longitude = path[t.intermediateIndex].Longitude
		} else {
			remainingPart := distanceToDrive / distToNextNode
			t.coveredDistance += distanceToDrive
			t.Latitude, t.Longitude = calculateDistVector(path[t.intermediateIndex], path[t.intermediateIndex+1], remainingPart)
			distanceToDrive = 0
		}
	}
}

func calculateDistVector(firstNode, secondNode *city.GraphNode, remainingPart float32) (subLat, subLot float32) {
	subLat = firstNode.Latitude + ((secondNode.Latitude - firstNode.Latitude) * remainingPart)
	subLot = firstNode.Longitude + ((secondNode.Longitude - firstNode.Longitude) * remainingPart)
	return
}

func calculateDistance(lat, lon float32, goalNode *city.GraphNode) float32 {
	sourceCoords := haversine.Coord{Lat: float64(lat), Lon: float64(lon)}
	goalCoords := haversine.Coord{Lat: float64(goalNode.Latitude), Lon: float64(goalNode.Longitude)}
	_, km := haversine.Distance(sourceCoords, goalCoords)
	return float32(km * 1000)
}
