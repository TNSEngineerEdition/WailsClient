package simulation

import (
	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/control_room"
	"github.com/umahmood/haversine"
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
	// distanceToDrive := float32(50*10) / float32(36)
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

	// if currentStop.Time == time {
	// 	node, ok := stopsById[currentStop.ID]
	// 	if !ok {
	// 		panic(fmt.Sprintf("Tram stop with ID %d should exist", currentStop.ID))
	// 	}
	// 	result = TramPositionChange{
	// 		TramID:    t.id,
	// 		Latitude:  node.Latitude,
	// 		Longitude: node.Longitude,
	// 	}
	// 	update = true
	// 	return result, update
	// }

	nextTripIndex := t.tripIndex + 1
	if nextTripIndex < len(t.trip.Stops) {
		nextStop := t.trip.Stops[nextTripIndex]

		if time > currentStop.Time && time < nextStop.Time {
			path := c.GetRoutesBetweenNodes(currentStop.ID, nextStop.ID)
			if len(path) == 0 {
				return result, false
			}
			progress := float32(time-currentStop.Time) / float32(nextStop.Time-currentStop.Time)
			lat, lon := t.calculateNewLocation(path, progress)

			result = TramPositionChange{
				TramID:    t.id,
				Latitude:  lat,
				Longitude: lon,
			}
			update = true
			return result, update
		}

		if time >= nextStop.Time {
			t.tripIndex++
		}
	}

	return result, update
}

func (t *tram) calculateNewLocation(path []*city.GraphNode, progress float32) (lat, lon float32) {
	index := int(progress * float32(len(path)))
	if index >= len(path) {
		index = len(path) - 1
	}
	lat = path[index].Latitude
	lon = path[index].Longitude
	return
}

// func (t *tram) calculateNewLocation(currentPosition [2]float32, distanceToDrive float32, subTripIndex int, path []*city.GraphNode) (lat, lon float32, index int) {
// 	subTripNode := path[subTripIndex+1]
// 	distToNextNode := calculateDistance(currentPosition, subTripNode)
// 	// fmt.Println(distToNextNode)
// 	distanceToDrive -= distToNextNode
// 	subTripIndex += 1
// 	droven := distToNextNode
// 	for distanceToDrive > 0 && subTripIndex <= len(path)-2 {
// 		for _, neighbor := range path[subTripIndex].Neighbors {
// 			if neighbor.ID == path[subTripIndex+1].ID {
// 				distToNextNode = neighbor.Length
// 				break
// 			}
// 		}
// 		if distanceToDrive >= distToNextNode {
// 			distanceToDrive -= distToNextNode
// 			subTripIndex += 1
// 			droven += distToNextNode
// 		} else {
// 			remainingPart := distanceToDrive / distToNextNode
// 			// fmt.Println(remainingPart)
// 			droven += distanceToDrive
// 			lat, lon = calculateDistVector(path[subTripIndex], path[subTripIndex+1], remainingPart)
// 			lat += path[subTripIndex].Latitude
// 			lon += path[subTripIndex].Longitude
// 			distanceToDrive = 0
// 		}
// 	}
// 	index = subTripIndex
// 	// fmt.Println(currentPosition, lat, lon, index, distanceToDrive, droven)
// 	return
// }

func calculateDistVector(firstNode, secondNode *city.GraphNode, remainingPart float32) (subLat, subLot float32) {
	subLat = (secondNode.Latitude - firstNode.Latitude) / remainingPart
	subLot = (secondNode.Longitude - firstNode.Longitude) / remainingPart
	return
}

func calculateDistance(currentPosition [2]float32, goalNode *city.GraphNode) float32 {
	sourceCoords := haversine.Coord{Lat: float64(currentPosition[0]), Lon: float64(currentPosition[1])}
	goalCoords := haversine.Coord{Lat: float64(goalNode.Latitude), Lon: float64(goalNode.Longitude)}
	_, km := haversine.Distance(sourceCoords, goalCoords)
	return float32(km * 1000)
}
