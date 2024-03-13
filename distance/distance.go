package distance

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func GetCityName(latitude, longitude float64) (string, error) {
	// Replace with your preferred reverse geocoding API URL
	url := fmt.Sprintf("https://maps.googleapis.com/maps/api/geocode/json?latlng=%f,%f&key=AIzaSyDIv5Watl4g5PeDrQymI6YkeHaml8uIoj0", latitude, longitude)

	response, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error retrieving response: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	// Check for successful response
	status, ok := result["status"].(string)
	if !ok || status != "OK" {
		return "", fmt.Errorf("error in Google Maps API response: %s", result["error_message"].(string))
	}

	// Extract city name from the response
	results, ok := result["results"].([]interface{})
	if !ok || len(results) == 0 {
		return "", fmt.Errorf("no results found in the response")
	}

	addressComponents, ok := results[0].(map[string]interface{})["address_components"].([]interface{})
	if !ok || len(addressComponents) == 0 {
		return "", fmt.Errorf("no address components found in the response")
	}

	city := ""
	for _, component := range addressComponents {
		componentMap, ok := component.(map[string]interface{})
		if !ok {
			continue
		}
		types, ok := componentMap["types"].([]interface{})
		if !ok || len(types) == 0 {
			continue
		}
		if types[0] == "locality" || types[0] == "administrative_area_level_2" {
			city = componentMap["long_name"].(string)
			break
		}
	}

	if city == "" {
		return "", fmt.Errorf("city not found in the response")
	}

	return city, nil
}

func GetGeocodeLocation(s string) (float64, float64) {

	locationString := strings.ReplaceAll(s, " ", "+")
	url := ("https://maps.googleapis.com/maps/api/geocode/json?address=" + locationString + "&key=" + "AIzaSyDIv5Watl4g5PeDrQymI6YkeHaml8uIoj0")
	response, err := http.Get(url)
	if err != nil {
		panic("Error retreiving response")
	}

	body, err2 := io.ReadAll(response.Body)
	if err2 != nil {
		panic("Error retreiving response")
	}

	var longitude float64
	var latitude float64
	var values map[string]interface{}

	json.Unmarshal(body, &values)
	for _, v := range values["results"].([]interface{}) {
		for i2, v2 := range v.(map[string]interface{}) {
			if i2 == "geometry" {
				latitude = v2.(map[string]interface{})["location"].(map[string]interface{})["lat"].(float64)
				longitude = v2.(map[string]interface{})["location"].(map[string]interface{})["lng"].(float64)
				break
			}
		}
	}
	return latitude, longitude
}

const apiKey = "AIzaSyDIv5Watl4g5PeDrQymI6YkeHaml8uIoj0" // Replace with your actual API key

func GetDistance(origin, destination string) int {
	originString := strings.ReplaceAll(origin, " ", "+")
	destinationString := strings.ReplaceAll(destination, " ", "+")

	url := fmt.Sprintf("https://maps.googleapis.com/maps/api/directions/json?origin=%s&destination=%s&key=%s", originString, destinationString, apiKey)

	response, err := http.Get(url)
	if err != nil {
		panic("Error retrieving response")
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic("Error reading response body")
	}

	var directionsResponse map[string]interface{}
	if err := json.Unmarshal(body, &directionsResponse); err != nil {
		panic("Error unmarshalling JSON")
	}

	// Check if the API returned a valid response
	if status, ok := directionsResponse["status"].(string); !ok || status != "OK" {
		fmt.Println("Error in directions API response")
		return 0
	}

	// Extract and print route information
	routes := directionsResponse["routes"].([]interface{})
	if len(routes) > 0 {
		route := routes[0].(map[string]interface{})
		legs := route["legs"].([]interface{})
		if len(legs) > 0 {
			leg := legs[0].(map[string]interface{})
			distancestr := leg["distance"].(map[string]interface{})["text"].(string)
			//duration := leg["duration"].(map[string]interface{})["text"].(string)
			//fmt.Printf("Route from %s to %s:\n", origin, destination)
			//fmt.Printf("Distance: %s\n", distance)
			//fmt.Printf("Duration: %s\n", duration)
			distancestr = strings.ReplaceAll(distancestr, ",", "")
			distancestr = strings.ReplaceAll(distancestr, " ", "")
			if strings.HasSuffix(distancestr, "km") {
				distancestr = distancestr[:len(distancestr)-2]
			}
			distance, err := strconv.Atoi(distancestr)
			if err != nil {
				fmt.Println("Error converting string to integer:", err)
			}
			return distance
		}
	} else {
		fmt.Println("No routes found")
		return 0
	}
	return 0
}

func GetDirections(origin, destination string) {
	originString := strings.ReplaceAll(origin, " ", "+")
	destinationString := strings.ReplaceAll(destination, " ", "+")

	url := fmt.Sprintf("https://maps.googleapis.com/maps/api/directions/json?origin=%s&destination=%s&key=%s", originString, destinationString, apiKey)

	response, err := http.Get(url)
	if err != nil {
		panic("Error retrieving response")
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic("Error reading response body")
	}

	var directionsResponse map[string]interface{}
	if err := json.Unmarshal(body, &directionsResponse); err != nil {
		panic("Error unmarshalling JSON")
	}

	// Check if the API returned a valid response
	if status, ok := directionsResponse["status"].(string); !ok || status != "OK" {
		fmt.Println("Error in directions API response")
		return
	}

	// Extract and print route information
	routes := directionsResponse["routes"].([]interface{})
	if len(routes) > 0 {
		route := routes[0].(map[string]interface{})
		legs := route["legs"].([]interface{})
		if len(legs) > 0 {
			leg := legs[0].(map[string]interface{})
			distance := leg["distance"].(map[string]interface{})["text"].(string)
			duration := leg["duration"].(map[string]interface{})["text"].(string)
			fmt.Printf("Route from %s to %s:\n", origin, destination)
			fmt.Printf("Distance: %s\n", distance)
			fmt.Printf("Duration: %s\n", duration)
		}
	} else {
		fmt.Println("No routes found")
	}
}
