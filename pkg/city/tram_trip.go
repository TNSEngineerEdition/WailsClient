package city

type tramTripStop struct {
	ID   uint64 `json:"id"`
	Time uint   `json:"time"`
}

type TramTrip struct {
	Route        string         `json:"route"`
	TripHeadSign string         `json:"trip_head_sign"`
	Stops        []tramTripStop `json:"stops"`
}
