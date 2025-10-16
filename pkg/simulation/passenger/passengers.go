package passenger

import (
	"fmt"
	"math/rand/v2"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city/graph"
)

type Passenger struct {
	strategy               PassengerStrategy
	spawnTime              uint
	StartStopID, EndStopID uint64
	PassengerID            uint64
}

func CreatePassengers(c *city.City) map[uint][]*Passenger {
	result := make(map[uint][]*Passenger)
	timeBounds := c.GetTimeBounds()
	tramStops := c.GetStops()
	counter := uint64(0)

	for i := range tramStops {
		startStop := tramStops[i]
		for range 10 {
			var j int
			for {
				j = rand.IntN(len(tramStops))
				if j != i {
					break
				}
			}
			endStop := tramStops[j]
			// spawn := timeBounds.StartTime + uint(rand.IntN(int(timeBounds.EndTime-timeBounds.StartTime+1)))
			spawn := timeBounds.StartTime + uint(rand.IntN(100))

			passenger := &Passenger{
				strategy:    PassengerStrategy(rand.IntN(3)),
				spawnTime:   spawn,
				StartStopID: startStop.ID,
				EndStopID:   endStop.ID,
				PassengerID: counter,
			}

			result[spawn] = append(result[spawn], passenger)
			counter++
		}
	}
	return result
}

func (p *Passenger) StrategyName() string {
	switch p.strategy {
	case ASAP:
		return "ASAP"
	case COMFORT:
		return "COMFORT"
	case SURE:
		return "SURE"
	default:
		return "UNKNOWN"
	}
}

func (p *Passenger) PrintInfoWithCity(stopsByID map[uint64]*graph.GraphTramStop) {

	startName := ""
	if s, ok := stopsByID[p.StartStopID]; ok {
		startName = s.GetName()
	}

	endName := ""
	if e, ok := stopsByID[p.EndStopID]; ok {
		endName = e.GetName()
	}

	fmt.Printf(
		"Passenger { strategy: %s, spawnTime: %d, startStopID: %d (%s), endStopID: %d (%s), passengerID: %d }\n",
		p.StrategyName(),
		p.spawnTime,
		p.StartStopID, startName,
		p.EndStopID, endName,
		p.PassengerID,
	)
}
