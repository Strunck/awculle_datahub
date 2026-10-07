package globals

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// function to query metrics from victoria-metrics

func (gs *GlobalState) QueryMetrics(query string) (result []map[string]any, err error) {

	// build the URL for the query
	url := fmt.Sprintf("%s/api/v1/query?query=%s", gs.Metrics.Tsdb_Url, query)

	// make the HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("📃 Error making HTTP GET request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// check if the response status is OK
	if resp.StatusCode != http.StatusOK {
		log.Printf("📃 Error: received non-OK HTTP status: %s", resp.Status)
		return nil, fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}

	// read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("📃 Error reading response body: %v", err)
		return nil, err
	}

	// parse the JSON response
	var jsonResponse struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string           `json:"resultType"`
			Result     []map[string]any `json:"result"`
		} `json:"data"`
	}

	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		log.Printf("📃 Error parsing JSON response: %v", err)
		return nil, err
	}

	if jsonResponse.Status != "success" {
		log.Printf("📃 Error: query failed with status: %s", jsonResponse.Status)
		return nil, fmt.Errorf("query failed with status: %s", jsonResponse.Status)
	}

	return jsonResponse.Data.Result, nil
}
