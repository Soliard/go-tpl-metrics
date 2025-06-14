package agent

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Soliard/go-tpl-metrics/models"
)

func (agent *Agent) Run() {
	ticker := time.NewTicker(time.Second) // минимальный интервал
	defer ticker.Stop()

	pollCounter := 0
	reportCounter := 0

	for {
		<-ticker.C
		pollCounter++
		reportCounter++

		if pollCounter >= int(agent.pollInterval.Seconds()) {
			if err := agent.collector.Collect(); err != nil {
				fmt.Println(`Error while collection metrics:`, err)
			}
			pollCounter = 0
		}

		if reportCounter >= int(agent.reportInterval.Seconds()) {
			if err := agent.reportMetrics(); err != nil {
				fmt.Println(`error while reporting metrics:`, err)
			}
			reportCounter = 0
		}
	}
}

func (agent *Agent) reportMetrics() error {
	for name, value := range agent.collector.Gauges {
		err := agent.sendMetric(models.Gauge, name, fmt.Sprintf("%v", value))
		if err != nil {
			return err
		}
	}

	for name, value := range agent.collector.Counters {
		err := agent.sendMetric(models.Counter, name, fmt.Sprintf("%v", value))
		if err != nil {
			return err
		}
	}

	fmt.Println("Metrics reported to the server")

	return nil
}

func (agent *Agent) sendMetric(mtype string, id string, value string) error {
	url := fmt.Sprintf(`%s/update/%s/%s/%s`, agent.serverHostURL, mtype, id, value)
	fmt.Printf(`[sendMetric] %s`, url)

	res, err := agent.httpClient.R().
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
