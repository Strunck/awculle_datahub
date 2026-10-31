package globals

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// function to query metrics from victoria-metrics

func (gs *GlobalState) QueryMetrics(query string) (result []map[string]any, err error) {

	// build the URL for the query
	url := fmt.Sprintf("%s/api/v1/query?query=%s", gs.Cfg.Metrics.Tsdb_Url, query)
	// log.Printf("📃 DEBUG Query URL: %s", url)

	// make the HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("📃 Error making HTTP GET request: %v,\nFULL URL: %s", err, url)
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

// validateDateString checks if the given date string is valid in the format "YYYY-MM-DD", and if it can be resolved as a valid date.
func validateDateString(dateStr string) bool {
	_, err := time.Parse("2006-01-02", dateStr)

	if err != nil {
		log.Printf("📃 Error: invalid date string: %v", err)
		return false
	} else {
		return true
	}
}

func checkDateRange(firstDate string, checkDate string, lastDate string) bool {
	first, err1 := time.Parse("2006-01-02", firstDate)
	check, err2 := time.Parse("2006-01-02", checkDate)
	last, err3 := time.Parse("2006-01-02", lastDate)

	if err1 != nil || err2 != nil || err3 != nil {
		log.Printf("📃 Error: invalid date string in date range: First: '%s', Check: '%s', Last: '%s'", firstDate, checkDate, lastDate)
		return false
	}

	if check.Before(first) || check.After(last) {
		log.Printf("📃 Error: check date '%s' is out of range ('%s' - '%s')", checkDate, firstDate, lastDate)
		return false
	}

	return true
}

func composeDeltaQueryString(metric string, short string, ort string, startDate string, endDate string) string {
	stepStr := convertToStep(startDate, endDate)
	q := fmt.Sprintf("delta(%s{short='%s',ort='%s'})%s&time=%s", metric, short, ort, stepStr, endDate)
	return q
}

func convertToStep(startDate string, endDate string) string {
	const layout = "2006-01-02"
	start, err1 := time.Parse(layout, startDate)
	end, err2 := time.Parse(layout, endDate)

	if err1 != nil || err2 != nil {
		log.Printf("📃 Error: invalid date string in convertToStep: Start: '%s', End: '%s'", startDate, endDate)
		return ""
	}

	duration := end.Sub(start)
	days := int(duration.Hours() / 24)
	return fmt.Sprintf("&step=%dd", days)
}
