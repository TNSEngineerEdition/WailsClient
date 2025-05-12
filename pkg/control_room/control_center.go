package control_room

import (
	"container/heap"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/umahmood/haversine"
)

type GraphNode = city.GraphNode

type ControlCenter struct {
	city         *city.City
	cachedRoutes map[[2]uint64][]GraphNode
}

func CreateControlCenter(city *city.City) *ControlCenter {
	return &ControlCenter{
		city:         city,
		cachedRoutes: make(map[[2]uint64][]GraphNode),
	}
}

func (cc *ControlCenter) GetShortestPath(sourceID, destID uint64) []GraphNode {
	tramStopPair := [2]uint64{sourceID, destID}
	if path, ok := cc.cachedRoutes[tramStopPair]; ok {
		return path
	}

	tramStops := cc.city.GetStopsByID()
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
			path := cc.reconstructPath(predecessors, tramStops, currentID)
			cc.cachedRoutes[tramStopPair] = path
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
				hScore := cc.heuristic(tramStops[neighbor.ID], destNode)
				fScore := hScore + tentativeG
				heap.Push(openSet, &nodeRecord{ID: neighbor.ID, Priority: fScore})
			}

		}
	}
	return nil
}

func (cc *ControlCenter) heuristic(a, b *GraphNode) float64 {
	sourceCoords := haversine.Coord{Lat: float64(a.Latitude), Lon: float64(a.Longitude)}
	goalCoords := haversine.Coord{Lat: float64(b.Latitude), Lon: float64(b.Longitude)}
	_, km := haversine.Distance(sourceCoords, goalCoords)
	return km
}

func (cc *ControlCenter) reconstructPath(cameFrom map[uint64]uint64, stops map[uint64]*GraphNode, currentID uint64) []GraphNode {
	var path []GraphNode
	for {
		path = append([]GraphNode{*stops[currentID]}, path...)
		prev, ok := cameFrom[currentID]
		if !ok {
			break
		}
		currentID = prev
	}
	return path
}
