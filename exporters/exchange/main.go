package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type rtx struct {
	Amount    float64   `json:"amount"`
	Timestamp time.Time `json:"timestamp"`
	From      []struct {
		Quotecurrency string  `json:"quotecurrency"`
		Mid           float64 `json:"mid"`
	} `json:"from"`
}

// default/quick convertion for this example
const (
	rtxapi = "https://xecdapi.xe.com/v1/convert_to.json/?to=USD&from=MXN&amount=1"
)

var (
	currentExchange = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "current_exchange_usd_to_mxn",
		Help: "Current exchange USD to MXN",
	})
)

func init() {
	prometheus.MustRegister(currentExchange)
}

func main() {

	ticker := time.NewTicker(1 * time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				req, err := http.NewRequest("GET", rtxapi, nil)
				if err != nil {
					fmt.Print(err.Error())
				}
				// set basic creds
				req.SetBasicAuth("<user>", "<key>")
				// Response
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					fmt.Print(err.Error())
				}
				defer resp.Body.Close()

				body, readErr := ioutil.ReadAll(resp.Body)
				if readErr != nil {
					log.Fatal(readErr)
				}

				// realtime exchange data
				rtxdata := rtx{}
				json.Unmarshal([]byte(body), &rtxdata)
				// just print the current exchange and timestamp
				fmt.Println("US Dollar - Mexican Peso Time stamp: ", rtxdata.Timestamp)
				fmt.Println("US Dollar - Mexican Peso exchange: ", rtxdata.From[0].Mid)
				// sent the data to prometheus
				currentExchange.Set(rtxdata.From[0].Mid)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	//Prometheus exporter port and metrics
	http.Handle("/metrics", prometheus.Handler())
	http.ListenAndServe(":8081", nil)
}
