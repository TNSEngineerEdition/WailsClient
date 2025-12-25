package travelplan

import (
	"fmt"
	"math/rand/v2"

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

	switch strategy {
	case RANDOM:
		startStopID := startStopIDs[rand.IntN(len(startStopIDs))]
		travelPlan, ok = GetRandomTravelPlan(currentCity, startStopID, spawnTime)
	case COMFORT:
		travelPlan, ok = GetComfortTravelPlan(currentCity, startStopIDs, endStops, spawnTime)
	default:
		panic(fmt.Sprintf("Unknown strategy: %s", strategy))
	}

	return travelPlan, ok
}
