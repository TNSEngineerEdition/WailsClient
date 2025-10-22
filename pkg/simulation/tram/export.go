package tram

import (
	"fmt"
	"io"
)

func TramsToCSVBuffer(trams map[uint]*Tram, writer io.Writer) error {
	writer.Write([]byte("tram_id,stop_id,stop_index,time,arrival_time,departure_time\n"))

	for _, tram := range trams {
		for stopIndex, stop := range tram.TripDetails.Trip.Stops {
			_, err := fmt.Fprintf(
				writer,
				"%d,%d,%d,%d,%d,%d\n",
				tram.ID,
				stop.ID,
				stopIndex,
				stop.Time,
				tram.TripDetails.Arrivals[stopIndex],
				tram.TripDetails.Departures[stopIndex],
			)

			if err != nil {
				return err
			}
		}
	}

	return nil
}
