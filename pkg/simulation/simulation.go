package simulation

import (
	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/controlcenter"
)

type Simulation struct {
	city          *city.City
	trams         map[int]*tram
	controlCenter controlcenter.ControlCenter
}

func NewSimulation(city *city.City) Simulation {
	return Simulation{
		city: city,
	}
}

func (s *Simulation) ResetTrams() {
	s.trams = make(map[int]*tram, len(s.city.GetTramTrips()))
	for i, trip := range s.city.GetTramTrips() {
		s.trams[i] = newTram(i, &trip, &s.controlCenter)
	}
}

func (s *Simulation) FetchData(url string) {
	s.city.FetchCityData(url)
	s.controlCenter = controlcenter.NewControlCenter(s.city)
	s.ResetTrams()
}

type TramIdentifier struct {
	ID    int    `json:"id"`
	Route string `json:"route"`
}

func (s *Simulation) GetTramIDs() (result []TramIdentifier) {
	result = make([]TramIdentifier, 0, len(s.trams))
	for id, tram := range s.trams {
		result = append(result, TramIdentifier{
			ID:    id,
			Route: tram.trip.Route,
		})
	}
	return result
}

func (s *Simulation) AdvanceTrams(time uint) (result []TramPositionChange) {
	result = make([]TramPositionChange, 0)
	for _, tram := range s.trams {
		positionChange, update := tram.Advance(time, s.city.GetStopsByID())
		if update {
			result = append(result, positionChange)
		}
	}
	return result
}

func (s *Simulation) GetTramDetails(id int) TramDetails {
	if tram, ok := s.trams[id]; ok {
		return tram.GetDetails(s.city)
	}
	return TramDetails{}
}
