package tram

import (
	"bytes"
	"fmt"
)

func TramsToCSVBuffer(trams map[uint]*Tram) *bytes.Buffer {
	buffer := bytes.NewBufferString("tram_id,stop_id,stop_index,time,arrival_time,departure_time\n")

	for _, tram := range trams {
		for stopIndex, stop := range tram.TripDetails.Trip.Stops {
			fmt.Fprintf(
				buffer,
				"%d,%d,%d,%d,%d,%d\n",
				tram.ID,
				stop.ID,
				stopIndex,
				stop.Time,
				tram.TripDetails.Arrivals[stopIndex],
				tram.TripDetails.Departures[stopIndex],
			)
		}
	}

	return buffer
}
