package passenger

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/graph"
	"github.com/TNSEngineerEdition/WailsClient/pkg/travelplan"
)

type PassengerModelData struct {
	ID           uint64
	startStopIDs []uint64
	endStopIDs   []uint64
	spawnTime    uint
	strategy     travelplan.TravelPlanStrategy
}

func GenerateRandomPassengers(currentCity *city.City) (passengers []PassengerModelData) {
	timeBounds := currentCity.GetTimeBounds()
	stopsByID := currentCity.GetStopsByID()

	// Start ID assignment from 1
	passengerID := uint64(1)

	for startStopID := range stopsByID {
		for range 500 {
			timeAfterStart := rand.IntN(int(timeBounds.EndTime - timeBounds.StartTime + 1))
			spawnTime := timeBounds.StartTime + uint(timeAfterStart)

			passengers = append(passengers, PassengerModelData{
				ID:           passengerID,
				startStopIDs: []uint64{startStopID},
				endStopIDs:   nil,
				spawnTime:    spawnTime,
				strategy:     travelplan.RANDOM,
			})

			passengerID++
		}
	}

	return
}

func GeneratePassengersFromModel(currentCity *city.City, passengerModel []byte) (passengers []PassengerModelData, error error) {
	records, err := readPassengerCSV(passengerModel)
	if err != nil {
		return
	}

	stopsByName := currentCity.GetStopsByName()
	for i, row := range records[1:] {
		passengerID := uint64(i + 1)

		data, err := getPassengerDataFromRow(passengerID, row, stopsByName)
		if err != nil {
			return nil, err
		}

		passengers = append(passengers, data)
	}

	return
}

func readPassengerCSV(passengerModel []byte) ([][]string, error) {
	reader := csv.NewReader(bytes.NewReader(passengerModel))
	reader.Comma = ','

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading passenger model csv: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("passenger model csv is empty")
	}

	header := records[0]
	if len(header) < 4 {
		return nil, fmt.Errorf("invalid header, expected 4 columns, got %d", len(header))
	}

	return records, nil
}

func getPassengerDataFromRow(
	passengerID uint64,
	row []string,
	stopsByName map[string]map[uint64]*graph.GraphTramStop,
) (PassengerModelData, error) {
	if len(row) < 4 {
		return PassengerModelData{}, fmt.Errorf("Passenger ID %d: expected 4 columns, got %d", passengerID, len(row))
	}

	startName := strings.TrimSpace(row[0])
	endName := strings.TrimSpace(row[1])
	spawnTimeStr := strings.TrimSpace(row[2])
	strategyStr := strings.TrimSpace(row[3])

	startStopIDs, err := getStopIDsFromGroupName(stopsByName, startName)
	if err != nil {
		return PassengerModelData{}, fmt.Errorf("Passenger ID %d: %w", passengerID, err)
	}

	endStopIDs, err := getStopIDsFromGroupName(stopsByName, endName)
	if err != nil {
		return PassengerModelData{}, fmt.Errorf("Passenger ID %d: %w", passengerID, err)
	}

	var spawnSeconds uint
	if t, err := time.Parse("15:04:05", spawnTimeStr); err == nil {
		spawnSeconds = uint(t.Hour()*3600 + t.Minute()*60 + t.Second())
	} else {
		return PassengerModelData{}, fmt.Errorf("Passenger ID %d: invalid spawn_time %q (expected HH:MM:SS)", passengerID, spawnTimeStr)
	}

	strategy := travelplan.TravelPlanStrategy(strings.ToUpper(strategyStr))

	data := PassengerModelData{
		ID:           passengerID,
		startStopIDs: startStopIDs,
		endStopIDs:   endStopIDs,
		spawnTime:    spawnSeconds,
		strategy:     strategy,
	}

	return data, nil
}
