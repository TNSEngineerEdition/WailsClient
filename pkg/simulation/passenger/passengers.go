package passenger

type Passenger struct {
	strategy               PassengerStrategy
	spawnTime              uint
	StartStopID, EndStopID uint64
	ID                     uint64
	TravelPlan             TravelPlan
}
