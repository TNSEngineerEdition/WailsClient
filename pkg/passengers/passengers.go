package passengers

import (
	"math/rand/v2"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
)

type Passenger struct {
	strategy               PassengerStrategy
	spawnTime              uint
	StartStopID, EndStopID uint64
}

func CreatePassengers(c *city.City) map[uint][]*Passenger {
	result := make(map[uint][]*Passenger)
	timeBounds := c.GetTimeBounds()
	tramStops := c.GetTramStops()

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
				strategy:    PassengerStrategy(rand.IntN(3)),
				spawnTime:   spawn,
				StartStopID: startStop.ID,
				EndStopID:   endStop.ID,
			}

			result[spawn] = append(result[spawn], passenger)
		}
	}
	return result
}
