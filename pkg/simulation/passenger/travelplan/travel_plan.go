package travelplan

import (
	"fmt"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
)

type TravelStop struct {
	id             uint64
	transferToStop uint64
	connections    map[uint]*TravelConnection
}

type TravelConnection struct {
	id                      uint
	to                      uint64
	arrivalTime, travelTime uint
}

type TravelPlan struct {
	stops                  map[uint64]*TravelStop
	connections            map[uint]*TravelConnection
	startStopID, endStopID uint64
	spawnTime              uint
	c                      *city.City
}

func GetTravelPlan(strategy PassengerStrategy, startStopID uint64, spawnTime uint, c *city.City) (TravelPlan, uint64) {
	tp := TravelPlan{
		stops:       make(map[uint64]*TravelStop),
		connections: make(map[uint]*TravelConnection),
		startStopID: startStopID,
		spawnTime:   spawnTime,
		c:           c,
	}

	switch strategy {
	case RANDOM:
		tp.GenerateRandomTravelPlan()
	case ASAP:
		panic("ASAP strategy not implemented yet")
	case COMFORT:
		panic("COMFORT strategy not implemented yet")
	case SURE:
		panic("SURE strategy not implemented yet")
	default:
		panic("Unknown strategy name")
	}

	return tp, tp.endStopID
}

func (tp *TravelPlan) GetTransferStop(stopID uint64) uint64 {
	stop, ok := tp.stops[stopID]
	if !ok {
		panic(fmt.Sprintf("%d - there is no such stop in the travel plan", stopID))
	}
	return stop.transferToStop
}

func (tp *TravelPlan) GetConnectionEnd(tramID uint) uint64 {
	if _, ok := tp.connections[tramID]; !ok {
		panic(fmt.Sprintf("Connection %d not found", tramID))
	}
	return tp.connections[tramID].to
}

func (tp *TravelPlan) IsConnectionInPlan(stopID uint64, tramID uint) bool {
	if _, ok := tp.stops[stopID]; !ok {
		return false
	}
	if _, ok := tp.stops[stopID].connections[tramID]; ok {
		return true
	}
	return false
}

func (tp *TravelPlan) IsEndStopReached(stopID uint64) bool {
	stopsByID := tp.c.GetStopsByID()
	return stopID == tp.endStopID || stopsByID[stopID].GetGroupName() == stopsByID[tp.endStopID].GetGroupName()
}

func (tp *TravelPlan) addStop(stopID uint64) {
	if _, ok := tp.stops[stopID]; !ok {
		tp.stops[stopID] = &TravelStop{
			id:          stopID,
			connections: make(map[uint]*TravelConnection),
		}
	}
}

func (tp *TravelPlan) addConnection(from, to uint64, tripID, arrivalTime, travelTime uint) {
	if _, ok := tp.stops[from]; !ok {
		tp.addStop(from)
	}
	if _, ok := tp.stops[to]; !ok {
		tp.addStop(to)
	}

	conn := TravelConnection{
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
	if _, ok := tp.stops[from]; !ok {
		tp.addStop(from)
	}
	if _, ok := tp.stops[to]; !ok {
		tp.addStop(to)
	}

	fromNode := tp.stops[from]
	fromNode.transferToStop = to
}
