package passenger

import (
	"fmt"
	"io"
)

func (ps *PassengersStore) PassengersToCSVBuffer(writer io.Writer) error {
	writer.Write([]byte("passenger_id,time,strategy\n"))

	for _, p := range ps.passengers {
		_, err := fmt.Fprintf(
			writer,
			"%d,%d,%s\n",
			p.ID,
			p.spawnTime,
			p.strategy,
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func (ps *PassengersStore) PassengerTripsToCSVBuffer(writer io.Writer) error {
	writer.Write([]byte("passenger_id,trip_sequence,tram_id,start_stop_id,get_on_time,end_stop_id,get_off_time\n"))

	for _, p := range ps.passengers {
		for _, t := range p.TakenTrips {
			_, err := fmt.Fprintf(
				writer,
				"%d,%d,%d,%d,%d,%d,%d\n",
				p.ID,
				t.tripSequence,
				t.tramID,
				t.startStopID,
				t.getOnTime,
				t.endStopID,
				t.getOffTime,
			)

			if err != nil {
				return err
			}
		}
	}

	return nil
}
