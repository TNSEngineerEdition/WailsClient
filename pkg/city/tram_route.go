package city

type TramTripStop struct {
	ID   uint64 `json:"id"`
	Time uint   `json:"time"`
}

type TramTrip struct {
	TripHeadSign string         `json:"trip_head_sign"`
	Stops        []TramTripStop `json:"stops"`
	ID           uint
}

type TramRoute struct {
	Name            string     `json:"name"`
	BackgroundColor string     `json:"background_color"`
	TextColor       string     `json:"text_color"`
	Trips           []TramTrip `json:"trips"`
}

func (t *TramRoute) assignTripIDs(startIndex uint) {
	for i := range t.Trips {
		t.Trips[i].ID = startIndex + uint(i)
	}
}

func (t *TramTrip) GetScheduledTravelTime(start, end int) uint {
	return t.Stops[end].Time - t.Stops[start].Time
}
