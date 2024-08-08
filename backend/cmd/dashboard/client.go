package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	g "github.com/tdadadavid/analytics"
)


func getMetric(what g.QueryType) ([]g.Metric, error) {
	data:= g.MetricData{
		What: what,
		SiteID: siteID,
		Start: uint32(start),
		End: uint32(end),
	}

	b, err := json.Marshal(data)
	if err != nil {
    log.Printf("error decoding JSON response: %v", err)
    if e, ok := err.(*json.SyntaxError); ok {
        log.Printf("syntax error at byte offset %d", e.Offset)
    }
    log.Printf("JSON response: %q", b)
    return nil, err
}

	c := g.GetConfig()
	req, err := http.NewRequest("POST",  c.GoTrackerHost+"/stats", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var metrics []g.Metric
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		fmt.Println("error from decoding: ", err)
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		fmt.Println("error from API: ", string(b))
		return nil, err
	}

	return metrics, nil
}