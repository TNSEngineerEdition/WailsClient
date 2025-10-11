package api

import (
	"context"
	"encoding/json"
	"fmt"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

var ServerURL = "http://localhost:8000"

type APIClient struct {
	client *ClientWithResponses
}

func NewAPIClient() APIClient {
	client, err := NewClientWithResponses(ServerURL)
	if err != nil {
		panic(err)
	}

	return APIClient{client: client}
}

func jsonToError(value any) error {
	v, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return fmt.Errorf("%s", v)
}

type CityInfo struct {
	CityID            string               `json:"cityID"`
	CityConfiguration CityConfiguration    `json:"cityConfiguration"`
	AvailableDates    []openapi_types.Date `json:"availableDates"`
}

func (a *APIClient) GetCities() []CityInfo {
	response, err := a.client.CitiesCitiesGetWithResponse(context.Background())
	if err != nil {
		panic(err)
	}

	if response.JSON200 == nil {
		panic(response.Body)
	}

	cities := make([]CityInfo, 0)
	for cityID, cityConfiguration := range *response.JSON200 {
		cities = append(cities, CityInfo{
			CityID:            cityID,
			CityConfiguration: cityConfiguration.CityConfiguration,
			AvailableDates:    cityConfiguration.AvailableDates,
		})
	}

	return cities
}

func (a *APIClient) GetCityByID(cityID string, params *GetCityDataCitiesCityIdGetParams) (*ResponseCityData, error) {
	response, err := a.client.GetCityDataCitiesCityIdGetWithResponse(
		context.Background(), cityID, params,
	)
	if err != nil {
		return nil, err
	}

	if response.JSON200 != nil {
		return response.JSON200, nil
	}

	if response.JSON422 != nil {
		return nil, jsonToError(response.JSON422.Detail)
	}

	return nil, fmt.Errorf("%s", response.Body)
}
