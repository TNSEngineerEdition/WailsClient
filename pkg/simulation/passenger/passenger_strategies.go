package passenger

import "fmt"

type PassengerStrategy uint8

const (
	SURE PassengerStrategy = iota
	COMFORT
	ASAP
)

type travelStop struct {
	ID         uint64
	arrivals   map[uint]*travelConnection
	departures map[uint]*travelConnection
}

type travelConnection struct {
	ID         uint
	to         uint64
	travelTime uint
}

type TravelPlan struct {
	stops map[uint64]*travelStop
}

func (pg *PassengerGraph) getTravelPlan(strategy PassengerStrategy) TravelPlan {
	tp := TravelPlan{stops: make(map[uint64]*travelStop)}
	tp.addStop(pg.startStopID)

	minTramChanges := uint(99)
	for _, edgeOut := range pg.nodes[pg.startStopID].edgesOut {
		minTramChanges = min(minTramChanges, pg.dfs(&tp, pg.startStopID, pg.startStopID, pg.startStopID, edgeOut.tramID, 0))
	}

	switch strategy {
	case SURE:
		return tp
	default:
		panic(fmt.Sprintf("%d: strategy not recognised\n", strategy))
	}
}

func (pg *PassengerGraph) dfs(tp *TravelPlan, fromStopID, prevStopID, currentStopID uint64, tramID, tramChanges uint) uint {
	stopsByID := pg.c.GetStopsByID()
	node := pg.nodes[currentStopID]

	// stick with the same tram as long as we can
	if trip, ok := node.edgesOut[tramID]; ok {
		return pg.dfs(tp, fromStopID, currentStopID, trip.to, tramID, tramChanges)
	}

	currentTime := node.edgesIn[tramID].toArrivalTime
	travelTime := currentTime - pg.nodes[fromStopID].edgesOut[tramID].fromArrivalTime

	tp.addStop(currentStopID)
	tp.addConnection(
		fromStopID,
		currentStopID,
		tramID,
		travelTime,
	)

	currentStopGroupName := stopsByID[currentStopID].GetGroupName()
	if currentStopGroupName == stopsByID[pg.endStopID].GetGroupName() {
		return tramChanges // if the target stop (group) is reached
	}

	minTramChanges := uint(99)
	prevStopGroupName := stopsByID[prevStopID].GetGroupName()

	arrivals := pg.getArrivalsFromTimeForGroupExcept(currentStopGroupName, prevStopGroupName, currentTime)
	for _, arrival := range arrivals {
		minTramChanges = min(minTramChanges, pg.dfs(
			tp,
			arrival.stopID,
			arrival.stopID,
			arrival.stopID,
			arrival.tramID,
			tramChanges+1,
		))
	}

	return minTramChanges
}

func (tp *TravelPlan) addStop(stopID uint64) {
	if _, ok := tp.stops[stopID]; !ok {
		tp.stops[stopID] = &travelStop{
			ID:         stopID,
			arrivals:   make(map[uint]*travelConnection),
			departures: make(map[uint]*travelConnection),
		}
	}
}

func (tp *TravelPlan) addConnection(from, to uint64, tripID, travelTime uint) {
	if _, ok := tp.stops[from]; !ok {
		tp.addStop(from)
	}
	if _, ok := tp.stops[to]; !ok {
		tp.addStop(to)
	}

	conn := travelConnection{
		ID:         tripID,
		to:         to,
		travelTime: travelTime,
	}

	fromNode := tp.stops[from]
	departue, ok := fromNode.departures[tripID]

	// if connection with the same ID already exists
	// check the travel time - do not update if time is longer (long ride on a singe tram is preferred)
	if ok {
		if departue.travelTime >= travelTime {
			return
		}
	}

	fromNode.departures[tripID] = &conn
	tp.stops[from] = fromNode

	toNode := tp.stops[to]
	toNode.arrivals[tripID] = &conn
	tp.stops[to] = toNode
}
