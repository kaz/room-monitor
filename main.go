package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

type (
	RawData struct {
		ID           string                            `json:"id"`
		NewestEvents map[string]map[string]interface{} `json:"newest_events"`
	}
	Data struct {
		sensor string
		value  map[string]interface{}
	}
)

func main() {
	ch := make(chan Data)
	go writeRoutine(ch)
	measureRoutine(ch)
}

func measureRoutine(ch chan Data) {
	for {
		ds, err := measure()
		if err != nil {
			log.Printf("[NG] measure failed: %v", err)
		} else {
			log.Printf("[OK] measure succeeded")
			for _, d := range ds {
				ch <- d
			}
		}
		time.Sleep(30 * time.Second)
	}
}

func writeRoutine(ch chan Data) {
	for d := range ch {
		if err := write(d); err != nil {
			log.Printf("[NG] write failed: %v", err)
		} else {
			log.Printf("[OK] write succeeded")
		}
	}
}

func measure() ([]Data, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.nature.global/1/devices", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+os.Getenv("TOKEN"))

	c := http.Client{Timeout: 10 * time.Second}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	raw := []RawData{}
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&raw); err != nil {
		return nil, err
	}

	ds := []Data{}
	for _, entry := range raw {
		values := map[string]interface{}{}
		for k, d := range entry.NewestEvents {
			values[k] = d["val"]
		}
		ds = append(ds, Data{entry.ID, values})
	}

	return ds, nil
}

func write(d Data) error {
	c, err := client.NewHTTPClient(client.HTTPConfig{Addr: "http://influxdb:8086"})
	if err != nil {
		return err
	}
	defer c.Close()

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "room",
		Precision: "s",
	})
	if err != nil {
		return err
	}

	pt, err := client.NewPoint("room", map[string]string{"sensor": d.sensor}, d.value, time.Now())
	if err != nil {
		return err
	}

	bp.AddPoint(pt)
	return c.Write(bp)
}
