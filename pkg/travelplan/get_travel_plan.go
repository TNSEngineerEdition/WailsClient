package travelplan

import (
	"fmt"
	"math/rand/v2"
	"slices"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/structs"
)

func GetTravelPlan(
	currentCity *city.City,
	strategy TravelPlanStrategy,
	startStopIDs, endStopIDs []uint64,
	spawnTime uint,
) (TravelPlan, bool) {
	var (
		travelPlan TravelPlan
		ok         bool
	)

	endStops := structs.NewSet[uint64]()
	for _, stopID := range endStopIDs {
		endStops.Add(stopID)
	}

	// Return empty travel plan if start stop is the same as end stop
	if slices.ContainsFunc(startStopIDs, endStops.Includes) {
		return TravelPlan{}, false
	}

	switch strategy {
	case RANDOM:
		startStopID := startStopIDs[rand.IntN(len(startStopIDs))]
		travelPlan, ok = GetRandomTravelPlan(currentCity, startStopID, spawnTime)
	case COMFORT:
		travelPlan, ok = GetComfortTravelPlan(currentCity, startStopIDs, endStops, spawnTime)
	case ASAP:
		travelPlan, ok = GetFastestTravelPlan(currentCity, startStopIDs, endStops, spawnTime, 0)
	case SURE:
		travelPlan, ok = GetFastestTravelPlan(currentCity, startStopIDs, endStops, spawnTime, 5*60) // 5 minutes
	default:
		panic(fmt.Sprintf("Unknown strategy: %s", strategy))
	}

	return travelPlan, ok
}
