package simulation

import (
	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/control_room"
	"github.com/umahmood/haversine"
	"golang.org/x/exp/rand"
)

type tram struct {
	id              int
	trip            *city.TramTrip
	tripIndex       int
	subTripIndex    int
	isFinished      bool
	Position        [2]float32
	CoveredDistance float32
	departureTime   uint
	c               control_room.ControlCenter
}

func newTram(id int, trip *city.TramTrip) *tram {
	return &tram{
		id:              id,
		trip:            trip,
		tripIndex:       0,
		subTripIndex:    0,
		isFinished:      false,
		CoveredDistance: 0,
		departureTime:   0,
	}
}

type TramPositionChange struct {
	TramID    int     `json:"id"`
	Latitude  float32 `json:"lat"`
	Longitude float32 `json:"lon"`
}

func (t *tram) Advance(time uint, stopsById map[uint64]*city.GraphNode, c *control_room.ControlCenter) (result TramPositionChange, update bool) {
	distanceToDrive := float32(50*10) / float32(36)
	if t.tripIndex == len(t.trip.Stops) {
		if !t.isFinished {
			t.isFinished = true
			result = TramPositionChange{
				TramID:    t.id,
				Latitude:  0,
				Longitude: 0,
			}
			update = true
		}
		return result, update
	}

	currentStop := t.trip.Stops[t.tripIndex]

	nextTripIndex := t.tripIndex + 1
	if nextTripIndex < len(t.trip.Stops) && time >= t.departureTime {
		nextStop := t.trip.Stops[nextTripIndex]
		path := c.GetRoutesBetweenNodes(currentStop.ID, nextStop.ID)
		lat, lon, index := t.calculateNewLocation(path, distanceToDrive, t.Position, t.subTripIndex)
		t.Position = [2]float32{lat, lon}
		t.subTripIndex = index
		if t.subTripIndex == len(path)-1 {
			t.tripIndex += 1
			t.subTripIndex = 0
			t.departureTime = max(nextStop.Time, time+uint(rand.Intn(11))+15)
		}
		result = TramPositionChange{
			TramID:    t.id,
			Latitude:  lat,
			Longitude: lon,
		}
		update = true
		return result, update
	}

	return result, false
}

func (t *tram) calculateNewLocation(path []*city.GraphNode, distanceToDrive float32, position [2]float32, tripSubIndex int) (lat, lon float32, index int) {
	currPosition := position
	for distanceToDrive > 0 && tripSubIndex < len(path)-1 {
		distToNextNode := calculateDistance(currPosition, path[tripSubIndex+1])
		if distToNextNode <= distanceToDrive {
			distanceToDrive -= distToNextNode
			tripSubIndex += 1
			lat = path[tripSubIndex].Latitude
			lon = path[tripSubIndex].Longitude
			currPosition = [2]float32{lat, lon}
		} else {
			remainingPart := distanceToDrive / distToNextNode
			lat, lon = calculateDistVector(path[tripSubIndex], path[tripSubIndex+1], remainingPart)
			currPosition = [2]float32{lat, lon}
			distanceToDrive = 0
		}
	}
	index = tripSubIndex
	return
}

func calculateDistVector(firstNode, secondNode *city.GraphNode, remainingPart float32) (subLat, subLot float32) {
	subLat = firstNode.Latitude + ((secondNode.Latitude - firstNode.Latitude) * remainingPart)
	subLot = firstNode.Longitude + ((secondNode.Longitude - firstNode.Longitude) * remainingPart)
	return
}

func calculateDistance(currentPosition [2]float32, goalNode *city.GraphNode) float32 {
	sourceCoords := haversine.Coord{Lat: float64(currentPosition[0]), Lon: float64(currentPosition[1])}
	goalCoords := haversine.Coord{Lat: float64(goalNode.Latitude), Lon: float64(goalNode.Longitude)}
	_, km := haversine.Distance(sourceCoords, goalCoords)
	return float32(km * 1000)
}
