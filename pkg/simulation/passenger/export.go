package passenger

import (
	"fmt"
	"io"
)

func PassengersToCSVBuffer(passengers []*Passenger, writer io.Writer) error {
	writer.Write([]byte("passenger_id,time,strategy\n"))

	for _, passenger := range passengers {
		_, err := fmt.Fprintf(
			writer,
			"%d,%d,%s\n",
			passenger.ID,
			passenger.spawnTime,
			passenger.strategy,
		)
		if err != nil {
			return err
		}

	}

	return nil
}
