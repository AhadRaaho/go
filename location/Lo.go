package location

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"shipment.com/packages/models"
)

func GetLabelByID(id int) string {
	placeId := strconv.Itoa(id)
	url := fmt.Sprintf("https://api.stage.raaho.in/places2/v2/places/%s", placeId)

	response, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return ""
	}

	var apiResponse models.ApiResponse
	err = json.NewDecoder(response.Body).Decode(&apiResponse)
	if err != nil {
		return ""
	}

	if len(apiResponse.Data) == 0 {
		return ""
	}

	return apiResponse.Data[0].PlaceName
}
