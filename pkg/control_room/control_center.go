package control_room

import (
	"container/heap"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/umahmood/haversine"
)

type ControlCenter struct {
	city         *city.City
	cachedRoutes map[[2]uint64][]*city.GraphNode
}

func CreateControlCenter(cityData *city.City) ControlCenter {
	c := ControlCenter{
		city:         cityData,
		cachedRoutes: make(map[[2]uint64][]*city.GraphNode),
	}

	tramTrips := cityData.GetTramTrips()
	for _, tripData := range tramTrips {
		for idx := 0; idx < len(tripData.Stops)-1; idx++ {
			firstStop, secondStop := tripData.Stops[idx], tripData.Stops[idx+1]
			tramStopPair := [2]uint64{firstStop.ID, secondStop.ID}
			if _, ok := c.cachedRoutes[tramStopPair]; ok {
				continue
			}
			path := c.GetShortestPath(firstStop.ID, secondStop.ID)
			c.cachedRoutes[tramStopPair] = path
		}
	}
	return c
}

func (c *ControlCenter) GetShortestPath(sourceID, destID uint64) []*city.GraphNode {
	tramStopPair := [2]uint64{sourceID, destID}
	tramStops := c.city.GetStopsByID()
	destNode := tramStops[destID]

	openSet := &priorityQueue{}
	heap.Init(openSet)
	heap.Push(openSet, &nodeRecord{ID: sourceID, Priority: 0})

	predecessors := make(map[uint64]uint64)
	gScores := make(map[uint64]float64)
	visitedNodes := make(map[uint64]bool)

	for openSet.Len() > 0 {
		currentID := heap.Pop(openSet).(*nodeRecord).ID
		if currentID == destID {
			path := c.reconstructPath(predecessors, tramStops, currentID)
			c.cachedRoutes[tramStopPair] = path
			return path
		}

		if visitedNodes[currentID] {
			continue
		}

		visitedNodes[currentID] = true

		currentNode := tramStops[currentID]

		for _, neighbor := range currentNode.Neighbors {
			tentativeG := gScores[currentID] + float64(neighbor.Length)
			cost, wasVisited := gScores[neighbor.ID]
			if !wasVisited || tentativeG < cost {
				predecessors[neighbor.ID] = currentID
				gScores[neighbor.ID] = tentativeG
				hScore := c.heuristic(tramStops[neighbor.ID], destNode)
				fScore := hScore + tentativeG
				heap.Push(openSet, &nodeRecord{ID: neighbor.ID, Priority: fScore})
			}

		}
	}
	return nil
}

func (c *ControlCenter) heuristic(a, b *city.GraphNode) float64 {
	sourceCoords := haversine.Coord{Lat: float64(a.Latitude), Lon: float64(a.Longitude)}
	goalCoords := haversine.Coord{Lat: float64(b.Latitude), Lon: float64(b.Longitude)}
	_, km := haversine.Distance(sourceCoords, goalCoords)
	return km * 1000
}

func (c *ControlCenter) reconstructPath(predecessors map[uint64]uint64, stops map[uint64]*city.GraphNode, currentID uint64) (path []*city.GraphNode) {
	for {
		path = append([]*city.GraphNode{stops[currentID]}, path...)
		prev, ok := predecessors[currentID]
		if !ok {
			break
		}
		currentID = prev
	}
	return
}
