package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"

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

func (a *APIClient) GetCityByID(
	cityID string,
	params *GetCityDataCitiesCityIdGetParams,
) (*ResponseCityData, error) {
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

func getCustomScheduleMultipartBody(customSchedule []byte) (io.Reader, string, error) {
	var buffer bytes.Buffer

	writer := multipart.NewWriter(&buffer)

	filePartHeaders := textproto.MIMEHeader{}
	filePartHeaders.Set("Content-Disposition", `form-data; name="custom_schedule_file"; filename="schedule.zip"`)
	filePartHeaders.Set("Content-Type", "application/zip")

	fileWriter, err := writer.CreatePart(filePartHeaders)
	if err != nil {
		return nil, "", err
	}

	if _, err = fileWriter.Write(customSchedule); err != nil {
		return nil, "", err
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return &buffer, writer.FormDataContentType(), nil
}

func (a *APIClient) GetCityByIDWithCustomSchedule(
	cityID string,
	customSchedule []byte,
	params *GetCityDataWithCustomScheduleCitiesCityIdPostParams,
) (*ResponseCityData, error) {
	body, contentType, err := getCustomScheduleMultipartBody(customSchedule)
	if err != nil {
		return nil, err
	}

	response, err := a.client.GetCityDataWithCustomScheduleCitiesCityIdPostWithBodyWithResponse(
		context.Background(), cityID, params, contentType, body,
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
