package vehicle

import (
	"fmt"
	"io"
)

func VehiclesToCSVBuffer(vehicles map[uint]*Vehicle, writer io.Writer) error {
	writer.Write([]byte("vehicle_id,stop_id,stop_index,time,arrival_time,departure_time\n"))

	for _, vehicle := range vehicles {
		for stopIndex, stop := range vehicle.TripDetails.Trip.Stops {
			_, err := fmt.Fprintf(
				writer,
				"%d,%d,%d,%d,%d,%d\n",
				vehicle.ID,
				stop.ID,
				stopIndex,
				stop.Time,
				vehicle.TripDetails.Arrivals[stopIndex],
				vehicle.TripDetails.Departures[stopIndex],
			)

			if err != nil {
				return err
			}
		}
	}

	return nil
}
