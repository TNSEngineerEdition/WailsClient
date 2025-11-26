package passenger

type Passenger struct {
	ID                     uint64
	strategy               PassengerStrategy
	spawnTime              uint
	startStopID, endStopID uint64
	TravelPlan             TravelPlan
}
