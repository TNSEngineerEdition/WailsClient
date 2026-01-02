package passenger

import "github.com/TNSEngineerEdition/WailsClient/pkg/simulation/passenger/travelplan"

type takenTrip struct {
	tramID                 uint
	tripSequence           int
	startStopID, endStopID uint64
	getOnTime, getOffTime  uint
}

type Passenger struct {
	ID                     uint64
	strategy               travelplan.PassengerStrategy
	spawnTime              uint
	startStopID, endStopID uint64
	TravelPlan             travelplan.TravelPlan
	TakenTrips             []takenTrip
}

func (p *Passenger) saveNewTrip(tramID, time uint, startStopID, endStopID uint64) {
	tripSequence := len(p.TakenTrips) + 1
	p.TakenTrips = append(p.TakenTrips, takenTrip{
		tramID:       tramID,
		tripSequence: tripSequence,
		getOnTime:    time,
		startStopID:  startStopID,
		endStopID:    endStopID,
	})
}

func (p *Passenger) saveGetOffTime(time uint) {
	lastTripIdx := len(p.TakenTrips) - 1
	if lastTripIdx < 0 {
		panic("Passenger have not taken any trips yet")
	}

	p.TakenTrips[lastTripIdx].getOffTime = time
}
