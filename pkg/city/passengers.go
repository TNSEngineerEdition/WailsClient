package city

import (
	"math/rand/v2"
)

type Passenger struct {
	strategy           PassangerStrategy
	spawnTime          uint
	StartStop, EndStop *GraphNode
}

type CityPassengers struct {
	city              *City
	initialPassengers map[uint][]*Passenger
}

func NewCityPassengers(c *City) *CityPassengers {
	return &CityPassengers{city: c}
}

func (cp *CityPassengers) CreatePassengers() {
	result := make(map[uint][]*Passenger)
	stopsMap := cp.city.GetStopsByID()
	timeBounds := cp.city.GetTimeBounds()

	tramStops := make([]*GraphNode, 0, len(stopsMap))
	for _, s := range stopsMap {
		tramStops = append(tramStops, s)
	}

	for i := range tramStops {
		startStop := tramStops[i]
		for n := 0; n < 10; n++ {
			var j int
			for {
				j = rand.IntN(len(tramStops))
				if j != i {
					break
				}
			}
			endStop := tramStops[j]
			spawn := timeBounds.StartTime + uint(rand.IntN(int(timeBounds.EndTime-timeBounds.StartTime+1)))
			passenger := &Passenger{
				strategy:  PassangerStrategy(rand.IntN(3)),
				spawnTime: spawn,
				StartStop: startStop,
				EndStop:   endStop,
			}

			result[spawn] = append(result[spawn], passenger)
		}
	}
	cp.initialPassengers = result
}
