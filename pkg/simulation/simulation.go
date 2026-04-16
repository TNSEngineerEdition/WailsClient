package simulation

import (
	"archive/zip"
	"context"
	"fmt"
	"math"
	"os"
	"runtime"
	"slices"
	"time"

	"github.com/TNSEngineerEdition/WailsClient/pkg/api"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/controlcenter"
	"github.com/TNSEngineerEdition/WailsClient/pkg/simulation/passenger"
	"github.com/TNSEngineerEdition/WailsClient/pkg/simulation/vehicle"
	"github.com/TNSEngineerEdition/WailsClient/pkg/structs"
	"github.com/oapi-codegen/runtime/types"
	wails_runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type Simulation struct {
	apiClient           *api.APIClient
	city                *city.City
	ctx                 context.Context
	vehicles            map[uint]*vehicle.Vehicle
	vehicleWorkersState *structs.WorkerState[*vehicle.Vehicle, vehicle.VehiclePositionChange]
	controlCenter       controlcenter.ControlCenter
	time                uint
	passengersStore     *passenger.PassengersStore
}

func NewSimulation(apiClient *api.APIClient, city *city.City) Simulation {
	return Simulation{
		apiClient: apiClient,
		city:      city,
	}
}

func (s *Simulation) SetContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *Simulation) vehicleWorker(state *structs.WorkerState[*vehicle.Vehicle, vehicle.VehiclePositionChange]) {
	for vehicle := range state.InputChannel {
		positionChange, update := vehicle.Advance(s.time, s.city.GetStopsByID())
		if update {
			state.OutputChannel <- positionChange
		}

		state.WaitGroup.Done()
	}
}

func (s *Simulation) resetVehicles() {
	vehicles := make(map[uint]*vehicle.Vehicle)

	for _, route := range s.city.GetVehicleRoutes() {
		for _, trip := range route.Trips {
			vehicles[trip.ID] = vehicle.NewVehicle(trip.ID, &route, &trip, &s.controlCenter, s.passengersStore)
		}
	}

	s.vehicles = vehicles
}

func (s *Simulation) ResetSimulation() {
	s.passengersStore.ResetPassengers()
	s.resetVehicles()
	s.city.Reset()
}

type SimulationParameters struct {
	CityID         string       `json:"cityID"`
	Weekday        *api.Weekday `json:"weekday,omitempty"`
	Date           *types.Date  `json:"date,omitempty"`
	CustomSchedule []byte       `json:"customSchedule,omitempty"`
	PassengerModel []byte       `json:"passengerModel,omitempty"`
}

func (s *Simulation) InitializeCity(parameters SimulationParameters) string {
	err := s.city.FetchCity(
		s.apiClient,
		parameters.CityID,
		&city.FetchCityParams{
			Weekday: parameters.Weekday,
			Date:    parameters.Date,
		},
		parameters.CustomSchedule,
	)

	if err != nil {
		return err.Error()
	}

	var passengerModelData []passenger.PassengerModelData
	if len(parameters.PassengerModel) == 0 {
		passengerModelData = passenger.GenerateRandomPassengers(s.city)
	} else {
		passengerModelData, err = passenger.GeneratePassengersFromModel(s.city, parameters.PassengerModel)
	}

	passengers := passenger.PassengersFromModelData(s.city, passengerModelData, 0)

	if err != nil {
		return err.Error()
	}

	s.passengersStore = passenger.NewPassengersStore(s.city, passengers)

	return ""
}

func (s *Simulation) InitializeSimulation(vehicleWorkerCount uint) string {
	if s.city.CityID == "" {
		panic("City data is not fetched")
	}

	s.controlCenter = controlcenter.NewControlCenter(s.city)
	s.ResetSimulation()

	if s.vehicleWorkersState != nil {
		s.vehicleWorkersState.Stop()
	}

	s.vehicleWorkersState = structs.NewWorkerState[*vehicle.Vehicle, vehicle.VehiclePositionChange](len(s.vehicles))

	if vehicleWorkerCount == 0 {
		// CPU count * 110% for more efficiency
		vehicleWorkerCount = uint(runtime.NumCPU()) * 11 / 10
	}

	for range vehicleWorkerCount {
		go s.vehicleWorker(s.vehicleWorkersState)
	}

	return ""
}

type VehicleIdentifier struct {
	ID    uint   `json:"id"`
	Route string `json:"route"`
}

func (s *Simulation) GetVehicleIDs() (result []VehicleIdentifier) {
	result = make([]VehicleIdentifier, 0, len(s.vehicles))
	for id, vehicle := range s.vehicles {
		result = append(result, VehicleIdentifier{
			ID:    id,
			Route: vehicle.Route.Name,
		})
	}
	return result
}

func (s *Simulation) AdvanceVehicles(time uint) (result []vehicle.VehiclePositionChange) {
	s.time = time

	s.passengersStore.DespawnPassengersAtTime(time)
	s.passengersStore.SpawnPassengersAtTime(time)

	s.vehicleWorkersState.WaitGroup.Add(len(s.vehicles))
	for _, vehicle := range s.vehicles {
		s.vehicleWorkersState.InputChannel <- vehicle
	}

	s.vehicleWorkersState.WaitGroup.Wait()

	result = make([]vehicle.VehiclePositionChange, 0)
	for range len(s.vehicleWorkersState.OutputChannel) {
		result = append(result, <-s.vehicleWorkersState.OutputChannel)
	}

	return result
}

func (s *Simulation) GetVehicleDetails(id uint) vehicle.VehicleDetails {
	if vehicle, ok := s.vehicles[id]; ok {
		return vehicle.GetDetails(s.city, s.time)
	}

	panic(fmt.Sprintf("Vehicle with ID %d not found", id))
}

func (s *Simulation) StopResumeVehicle(id uint) vehicle.VehicleDetails {
	vehicle, ok := s.vehicles[id]
	if !ok {
		panic(fmt.Sprintf("StopResumeVehicle: vehicle with ID %d not found", id))
	}

	if vehicle.IsStopped() {
		vehicle.ResumeVehicle(s.time)
	} else {
		vehicle.StopVehicle()
	}

	return vehicle.GetDetails(s.city, s.time)
}

type Arrival struct {
	Route        string `json:"route"`
	TripHeadSign string `json:"tripHeadSign"`
	Minutes      uint   `json:"time"`
	VehicleID    uint   `json:"id"`
}

func (s *Simulation) GetArrivalsForStop(stopID uint64, count int) []Arrival {
	plannedArrivals := s.city.GetPlannedArrivals(stopID)
	arrivals := make([]Arrival, 0)

	if plannedArrivals == nil {
		return arrivals
	}

	// Skip vehicles which have already departed for future iterations
	for i, arrival := range *plannedArrivals {
		if s.vehicles[arrival.TripID].TripDetails.Index <= arrival.StopIndex {
			continue
		}

		*plannedArrivals = (*plannedArrivals)[i:]
		break
	}

	for _, arrival := range *plannedArrivals {
		if arrival.Time > s.time+30*60 {
			break
		}

		vehicle := s.vehicles[arrival.TripID]
		if vehicle.TripDetails.Index > arrival.StopIndex {
			continue
		}

		var expectedTime uint
		if vehicle.TripDetails.Index < arrival.StopIndex || !vehicle.IsAtStop() {
			expectedTime = vehicle.GetEstimatedArrival(arrival.StopIndex, s.time) - s.time
		}

		arrivals = append(arrivals, Arrival{
			Route:        vehicle.Route.Name,
			TripHeadSign: vehicle.TripDetails.Trip.TripHeadSign,
			Minutes:      uint(math.Ceil(float64(expectedTime) / 60)),
			VehicleID:    vehicle.ID,
		})
	}

	slices.SortFunc(arrivals, func(a1, a2 Arrival) int {
		return int(a1.Minutes) - int(a2.Minutes)
	})

	return arrivals[:min(len(arrivals), count)]
}

func (s *Simulation) GetSegmentsForRoute(routeName string) []controlcenter.RouteSegment {
	return s.controlCenter.GetSegmentsForRoute(routeName)
}

func (s *Simulation) GetPassengerCountAtStop(stopID uint64) uint {
	return s.passengersStore.GetPassengerCountAtStop(stopID)
}

func (s *Simulation) GetPassengerCountOnRoute(routeName string) (count uint) {
	for _, vehicle := range s.vehicles {
		if vehicle.Route.Name == routeName {
			count += vehicle.GetPassengerCount()
		}
	}
	return
}

func (s *Simulation) ExportToFile() string {
	filename, err := wails_runtime.SaveFileDialog(s.ctx, wails_runtime.SaveDialogOptions{
		DefaultFilename:      fmt.Sprintf("%s-%d.zip", s.city.CityID, time.Now().Unix()),
		CanCreateDirectories: true,
		Filters: []wails_runtime.FileFilter{
			{DisplayName: "ZIP file", Pattern: "*.zip"},
		},
	})
	if err != nil {
		return err.Error()
	}

	file, err := os.Create(filename)
	if err != nil {
		return err.Error()
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// city data
	if cityDataZipFileWriter, err := zipWriter.Create("city_data.json"); err != nil {
		return err.Error()
	} else if err := s.city.CityDataToJSONBuffer(cityDataZipFileWriter); err != nil {
		return err.Error()
	}

	// vehicles
	if vehicleZipFileWriter, err := zipWriter.Create("vehicles.csv"); err != nil {
		return err.Error()
	} else if err := vehicle.VehiclesToCSVBuffer(s.vehicles, vehicleZipFileWriter); err != nil {
		return err.Error()
	}

	// passengers
	if passengerZipFileWriter, err := zipWriter.Create("passengers.csv"); err != nil {
		return err.Error()
	} else if err := s.passengersStore.PassengersToCSVBuffer(passengerZipFileWriter); err != nil {
		return err.Error()
	}

	// passenger trips
	if passengerTripsZipFileWriter, err := zipWriter.Create("passenger_trips.csv"); err != nil {
		return err.Error()
	} else if err := s.passengersStore.PassengerTripsToCSVBuffer(passengerTripsZipFileWriter); err != nil {
		return err.Error()
	}

	return ""
}
