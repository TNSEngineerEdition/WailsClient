package simulation

import (
	"math"
	"runtime"
	"slices"

	"github.com/TNSEngineerEdition/WailsClient/pkg/api"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/controlcenter"
	"github.com/TNSEngineerEdition/WailsClient/pkg/passengers"
)

type Simulation struct {
	apiClient             *api.APIClient
	city                  *city.City
	trams                 map[uint]*tram
	tramWorkersData       workerData[*tram, TramPositionChange]
	controlCenter         controlcenter.ControlCenter
	time                  uint
	initialPassengers     map[uint][]*passengers.Passenger
	passengersAtStopsByID *tramStops
}

func NewSimulation(apiClient *api.APIClient, city *city.City) Simulation {
	return Simulation{
		apiClient: apiClient,
		city:      city,
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

func (s *Simulation) getPassengersAt(time uint) []*passengers.Passenger {
	return s.initialPassengers[time]
}

func (s *Simulation) resetTrams() {
	s.trams = make(map[uint]*tram)
	for _, route := range s.city.GetTramRoutes() {
		for _, trip := range route.Trips {
			s.trams[trip.ID] = newTram(trip.ID, &route, &trip, &s.controlCenter)
		}
	}
}

func (s *Simulation) resetPassengers() {
	s.passengersAtStopsByID = newTramStops()
}

func (s *Simulation) ResetSimulation() {
	s.resetTrams()
	s.resetPassengers()
	s.city.Reset()
}

func (s *Simulation) FetchData(parameters SimulationFetchParameters) string {
	err := s.city.FetchCity(s.apiClient, parameters.CityID, &api.GetCityDataCitiesCityIdGetParams{
		Weekday: parameters.Weekday,
		Date:    parameters.Date,
	})
	if err != nil {
		return err.Error()
	}

	s.controlCenter = controlcenter.NewControlCenter(s.city)
	s.ResetSimulation()
	s.tramWorkersData.reset(len(s.trams))

	s.initialPassengers = passengers.CreatePassengers(s.city)

	if parameters.TramWorkerCount == 0 {
		// CPU count * 110% for more efficiency
		parameters.TramWorkerCount = uint(runtime.NumCPU()) * 11 / 10
	}

	for range parameters.TramWorkerCount {
		go s.tramWorker()
	}

	return ""
}

type TramIdentifier struct {
	ID    uint   `json:"id"`
	Route string `json:"route"`
}

func (s *Simulation) GetTramIDs() (result []TramIdentifier) {
	result = make([]TramIdentifier, 0, len(s.trams))
	for id, tram := range s.trams {
		result = append(result, TramIdentifier{
			ID:    id,
			Route: tram.route.Name,
		})
	}
	return result
}

func (s *Simulation) AdvanceTrams(time uint) (result []TramPositionChange) {
	s.time = time

	toSpawn := s.getPassengersAt(time)
	for _, p := range toSpawn {
		stopID := p.StartStopID
		s.passengersAtStopsByID.stops[stopID] = append(s.passengersAtStopsByID.stops[stopID], p)
	}

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

func (s *Simulation) GetTramDetails(id uint) TramDetails {
	if tram, ok := s.trams[id]; ok {
		return tram.GetDetails(s.city, s.time)
	}
	return TramDetails{}
}

type Arrival struct {
	Route        string `json:"route"`
	TripHeadSign string `json:"tripHeadSign"`
	Minutes      uint   `json:"time"`
}

func (s *Simulation) GetArrivalsForStop(stopID uint64, count int) []Arrival {
	plannedArrivals := s.city.GetPlannedArrivals(stopID)
	arrivals := make([]Arrival, 0)

	// Skip trams which have already departed for future iterations
	for i, arrival := range *plannedArrivals {
		if s.trams[arrival.TripID].tripData.index <= arrival.StopIndex {
			continue
		}

		*plannedArrivals = (*plannedArrivals)[i:]
		break
	}

	for _, arrival := range *plannedArrivals {
		if arrival.Time > s.time+30*60 {
			break
		}

		tram := s.trams[arrival.TripID]
		if tram.tripData.index > arrival.StopIndex {
			continue
		}

		var expectedTime uint
		if tram.tripData.index < arrival.StopIndex || !tram.isAtStop() {
			expectedTime = tram.getEstimatedArrival(arrival.StopIndex, s.time) - s.time
		}

		arrivals = append(arrivals, Arrival{
			Route:        tram.route.Name,
			TripHeadSign: tram.tripData.trip.TripHeadSign,
			Minutes:      uint(math.Ceil(float64(expectedTime) / 60)),
		})
	}

	slices.SortFunc(arrivals, func(a1, a2 Arrival) int {
		return int(a1.Minutes) - int(a2.Minutes)
	})

	return arrivals[:min(len(arrivals), count)]
}

func (s *Simulation) GetRoutePolylines(lineName string) controlcenter.RoutePolylines {
	return s.controlCenter.GetRoutePolylines(lineName)
}
