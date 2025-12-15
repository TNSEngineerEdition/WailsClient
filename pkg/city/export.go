package city

import (
	"encoding/json"
	"io"
)

func (c *City) CityDataToCSVBuffer(writer io.Writer) error {
	jsonCityData, err := json.Marshal(c.responseCityData)
	if err != nil {
		return err
	}

	_, err = writer.Write(jsonCityData)
	if err != nil {
		return err
	}

	return nil
}
