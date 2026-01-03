package travelplan

import (
	"fmt"

	"github.com/TNSEngineerEdition/WailsClient/pkg/structs"
)

type travelConnection struct {
	id, arrivalTime, travelTime uint
	to                          uint64
}

type travelStop struct {
	id             uint64
	transferToStop uint64
	connections    map[uint]*travelConnection
}

type TravelPlan struct {
	stops       map[uint64]*travelStop
	connections map[uint]*travelConnection
	startStopID uint64
	endStopIDs  structs.Set[uint64]
	spawnTime   uint
}

func NewTravelPlan(startStopID uint64, endStopIDs structs.Set[uint64], spawnTime uint) TravelPlan {
	return TravelPlan{
		stops:       make(map[uint64]*travelStop),
		connections: make(map[uint]*travelConnection),
		startStopID: startStopID,
		endStopIDs:  endStopIDs,
		spawnTime:   spawnTime,
	}
}

func (tp TravelPlan) GetStartStopID() uint64 {
	return tp.startStopID
}

func (tp TravelPlan) GetConnectionTransferDestination(stopID uint64) uint64 {
	if stop, ok := tp.stops[stopID]; ok {
		return stop.transferToStop
	} else {
		panic(fmt.Sprintf("%d - there is no such stop in the travel plan", stopID))
	}
}

func (tp TravelPlan) GetConnectionDestination(tramID uint) uint64 {
	if _, ok := tp.connections[tramID]; ok {
		return tp.connections[tramID].to
	} else {
		panic(fmt.Sprintf("Connection %d not found", tramID))
	}
}

func (tp TravelPlan) ContainsConnection(stopID uint64, tramID uint) bool {
	if _, ok := tp.stops[stopID]; !ok {
		return false
	}

	if _, ok := tp.stops[stopID].connections[tramID]; ok {
		return true
	}

	return false
}

func (tp TravelPlan) IsEndStopReached(stopID uint64) bool {
	return tp.endStopIDs.Includes(stopID)
}

func (tp *TravelPlan) addStop(stopID uint64) {
	if _, ok := tp.stops[stopID]; !ok {
		tp.stops[stopID] = &travelStop{
			id:          stopID,
			connections: make(map[uint]*travelConnection),
		}
	}
}

func (tp *TravelPlan) addConnection(from, to uint64, tripID, arrivalTime, travelTime uint) {
	tp.addStop(from)
	tp.addStop(to)

	conn := travelConnection{
		id:          tripID,
		to:          to,
		arrivalTime: arrivalTime,
		travelTime:  travelTime,
	}

	fromNode := tp.stops[from]
	fromNode.connections[tripID] = &conn
	tp.stops[from] = fromNode

	tp.connections[tripID] = &conn
}

func (tp *TravelPlan) addTransfer(from, to uint64) {
	tp.addStop(from)
	tp.addStop(to)

	fromNode := tp.stops[from]
	fromNode.transferToStop = to
}
