package agent

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Soliard/go-tpl-metrics/models"
)

func (a *Agent) Run() {
	ticker := time.NewTicker(time.Second) // минимальный интервал
	defer ticker.Stop()

	pollCounter := 0
	reportCounter := 0

	for {
		time.Sleep(time.Second)
		pollCounter++
		reportCounter++

		if pollCounter >= int(a.pollInterval.Seconds()) {
			if err := a.collector.Collect(); err != nil {
				fmt.Println(`Error while collection metrics:`, err)
			}
			pollCounter = 0
		}

		if reportCounter >= int(a.reportInterval.Seconds()) {
			if err := a.reportMetrics(); err != nil {
				fmt.Println(`error while reporting metrics:`, err)
			}
			reportCounter = 0
		}
	}
}

func (a *Agent) reportMetrics() error {
	for name, value := range a.collector.Gauges {
		err := a.sendMetric(models.Gauge, name, fmt.Sprintf("%v", value))
		if err != nil {
			return err
		}
	}

	for name, value := range a.collector.Counters {
		err := a.sendMetric(models.Counter, name, fmt.Sprintf("%v", value))
		if err != nil {
			return err
		}
	}

	fmt.Println("Metrics reported to the server")

	return nil
}

func (a *Agent) sendMetric(mtype string, id string, value string) error {
	url := fmt.Sprintf(`%s/update/%s/%s/%s`, a.serverHostURL, mtype, id, value)
	fmt.Printf(`[sendMetric] %s`, url)

	res, err := a.httpClient.R().
		SetHeader("Content-type", "text/plain").
		Post(url)
	if err != nil {
		return fmt.Errorf(`error while send request with metric: %v`, err)
	}

	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf(`metric sending request return not ok status code`)
	}

	return nil

}
