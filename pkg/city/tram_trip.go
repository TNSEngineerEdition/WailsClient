package city

type TramTripStop struct {
	ID   uint64 `json:"id"`
	Time uint   `json:"time"`
}

type TramTrip struct {
	Route        string         `json:"route"`
	TripHeadSign string         `json:"trip_head_sign"`
	Stops        []TramTripStop `json:"stops"`
}

func (t *TramTrip) GetScheduledTravelTime(start, end int) uint {
	return t.Stops[end].Time - t.Stops[start].Time
}
