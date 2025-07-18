package simulation

import (
	"runtime"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/controlcenter"
)

type Simulation struct {
	city            *city.City
	trams           map[int]*tram
	tramWorkersData workerData[*tram, TramPositionChange]
	controlCenter   controlcenter.ControlCenter
	time            uint
}

func NewSimulation(city *city.City) Simulation {
	return Simulation{
		city: city,
	}
}

func (s *Simulation) tramWorker() {
	for tram := range s.tramWorkersData.inputChannel {
		positionChange, update := tram.Advance(s.time, s.city.GetStopsByID())
		if update {
			s.tramWorkersData.outputChannel <- positionChange
		}

		s.tramWorkersData.wg.Done()
	}
}

func (s *Simulation) ResetTrams() {
	s.trams = make(map[int]*tram, len(s.city.GetTramTrips()))
	for i, trip := range s.city.GetTramTrips() {
		s.trams[i] = newTram(i, &trip, &s.controlCenter)
	}
}

func (s *Simulation) FetchData(url string, tramWorkerCount uint) {
	s.city.FetchCityData(url)
	s.controlCenter = controlcenter.NewControlCenter(s.city)
	s.ResetTrams()
	s.tramWorkersData.reset(len(s.trams))

	if tramWorkerCount == 0 {
		tramWorkerCount = uint(runtime.NumCPU()) * 11 / 10
	}

	for range tramWorkerCount {
		go s.tramWorker()
	}
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
	s.time = time

	s.tramWorkersData.wg.Add(len(s.trams))
	for _, tram := range s.trams {
		s.tramWorkersData.inputChannel <- tram
	}

	s.tramWorkersData.wg.Wait()

	result = make([]TramPositionChange, 0)
	for range len(s.tramWorkersData.outputChannel) {
		result = append(result, <-s.tramWorkersData.outputChannel)
	}

	return result
}

func (s *Simulation) GetTramDetails(id int) TramDetails {
	if tram, ok := s.trams[id]; ok {
		return tram.GetDetails(s.city)
	}
	return TramDetails{}
}
