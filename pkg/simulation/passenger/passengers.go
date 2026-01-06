package passenger

import (
	"log"
	"runtime"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/structs"
	"github.com/TNSEngineerEdition/WailsClient/pkg/travelplan"
)

type takenTrip struct {
	tramID                 uint
	tripSequence           int
	startStopID, endStopID uint64
	getOnTime, getOffTime  uint
}

type travelPlanWorkerInput struct {
	currentCity *city.City
	data        PassengerModelData
}

type Passenger struct {
	ID         uint64
	strategy   travelplan.TravelPlanStrategy
	spawnTime  uint
	TravelPlan travelplan.TravelPlan
	TakenTrips []takenTrip
}

func passengerWorker(state *structs.WorkerState[travelPlanWorkerInput, Passenger]) {
	for input := range state.InputChannel {
		travelPlan, ok := travelplan.GetTravelPlan(
			input.currentCity,
			input.data.strategy,
			input.data.startStopIDs,
			input.data.endStopIDs,
			input.data.spawnTime,
		)

		if !ok {
			log.Default().Printf("Travel plan couldn't be created for passenger %d", input.data.ID)
			state.WaitGroup.Done()
			continue
		}

		state.OutputChannel <- Passenger{
			ID:         input.data.ID,
			strategy:   input.data.strategy,
			spawnTime:  input.data.spawnTime,
			TravelPlan: travelPlan,
		}

		state.WaitGroup.Done()
	}
}

func PassengersFromModelData(
	currentCity *city.City,
	data []PassengerModelData,
	workerNumber uint,
) (passengers []Passenger) {
	workerState := structs.NewWorkerState[travelPlanWorkerInput, Passenger](len(data))

	if workerNumber == 0 {
		// CPU count * 110% for more efficiency
		workerNumber = uint(runtime.NumCPU()) * 11 / 10
	}

	for range workerNumber {
		go passengerWorker(workerState)
	}

	workerState.WaitGroup.Add(len(data))

	for _, data := range data {
		workerState.InputChannel <- travelPlanWorkerInput{
			currentCity: currentCity,
			data:        data,
		}
	}

	workerState.WaitGroup.Wait()

	for range len(workerState.OutputChannel) {
		passengers = append(passengers, <-workerState.OutputChannel)
	}

	workerState.Stop()

	return
}

func (p *Passenger) saveNewTrip(tramID, time uint, startStopID, endStopID uint64) {
	tripSequence := len(p.TakenTrips) + 1
	p.TakenTrips = append(p.TakenTrips, takenTrip{
		tramID:       tramID,
		tripSequence: tripSequence,
		getOnTime:    time,
		startStopID:  startStopID,
		endStopID:    endStopID,
	})
}

func (p *Passenger) saveGetOffTime(time uint) {
	lastTripIdx := len(p.TakenTrips) - 1
	if lastTripIdx < 0 {
		panic("Passenger have not taken any trips yet")
	}

	p.TakenTrips[lastTripIdx].getOffTime = time
}
