package passenger

import "github.com/TNSEngineerEdition/WailsClient/pkg/simulation/passenger/travelplan"

type Passenger struct {
	ID                     uint64
	strategy               travelplan.PassengerStrategy
	spawnTime              uint
	startStopID, endStopID uint64
	TravelPlan             travelplan.TravelPlan
}
