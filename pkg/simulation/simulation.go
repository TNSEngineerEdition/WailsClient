package simulation

import (
	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/controlcenter"
)

type Simulation struct {
	city          *city.City
	trams         []*tram
	controlCenter controlcenter.ControlCenter
}

func NewSimulation(city *city.City) Simulation {
	return Simulation{
		city: city,
	}
}

func (s *Simulation) ResetTrams() {
	s.trams = make([]*tram, len(s.city.GetTramTrips()))
	for i, trip := range s.city.GetTramTrips() {
		s.trams[i] = newTram(i, &trip, &s.controlCenter)
	}
}

func (s *Simulation) FetchData(url string) {
	s.city.FetchCityData(url)
	s.controlCenter = controlcenter.NewControlCenter(s.city)
	s.ResetTrams()
}

func (s *Simulation) GetTramIDs() (result []TramBasic) {
	result = make([]TramBasic, len(s.trams))
	for i, tram := range s.trams {
		result[i] = TramBasic{
			ID:    tram.id,
			Route: tram.trip.Route,
		}
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
	var myTram *tram

	for _, tram := range s.trams {
		if tram.id == id {
			myTram = tram
			break
		}
	}

	return myTram.GetDetails(s.city)
}
