package simulation

import (
	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/control_room"
)

type Simulation struct {
	city  *city.City
	trams []*tram
	c     control_room.ControlCenter
}

func NewSimulation(city *city.City) Simulation {
	return Simulation{
		city: city,
	}
}

func (s *Simulation) ResetTrams() {
	s.trams = make([]*tram, len(s.city.GetTramTrips()))
	for i, trip := range s.city.GetTramTrips() {
		s.trams[i] = newTram(i, &trip)
	}
}

func (s *Simulation) FetchData(url string) {
	s.city.FetchCityData(url)
	s.c = control_room.CreateControlCenter(s.city)
	s.ResetTrams()
}

func (s *Simulation) GetTramIDs() (result []int) {
	result = make([]int, len(s.trams))
	for i, tram := range s.trams {
		result[i] = tram.id
	}

	return result
}

func (s *Simulation) AdvanceTrams(time uint) (result []TramPositionChange) {
	result = make([]TramPositionChange, 0)
	for _, tram := range s.trams {
		positionChange, update := tram.Advance(time, s.city.GetStopsByID(), s.c)
		if update {
			result = append(result, positionChange)
		}
	}

	return result
}
