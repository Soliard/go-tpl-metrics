package agent

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Soliard/go-tpl-metrics/models"
)

type Agent struct {
	serverHostURL  string
	collector      *StatsCollector
	httpClient     *http.Client
	pollInterval   time.Duration
	reportInterval time.Duration
}

func NewAgent(serverHostURL string) *Agent {
	return &Agent{
		serverHostURL:  serverHostURL,
		collector:      NewStatsCollector(),
		httpClient:     &http.Client{},
		pollInterval:   time.Second * 2,
		reportInterval: time.Second * 10,
	}
}

func (agent *Agent) Run() {
	for {
		if err := agent.collector.Collect(); err != nil {
			fmt.Println(`Error while collection metrics:`, err)
		}

		polCount, ok := agent.collector.counters["PollCount"]
		if ok {
			if polCount%5 == 0 {
				if err := agent.reportMetrics(); err != nil {
					fmt.Println(`error while reporting metrics:`, err)
				}
			}
		} else {
			fmt.Println(`cannot get pol count from counter metrics`)
		}

		time.Sleep(agent.pollInterval)
	}
}

func (agent *Agent) reportMetrics() error {
	for name, value := range agent.collector.Gauges {
		err := agent.sendMetric(models.Gauge, name, fmt.Sprintf("%v", value))
		if err != nil {
			return err
		}
	}

	for name, value := range agent.collector.counters {
		err := agent.sendMetric(models.Counter, name, fmt.Sprintf("%v", value))
		if err != nil {
			return err
		}
	}

	return nil
}

func (agent *Agent) sendMetric(mtype string, mid string, mvalue string) error {
	url := fmt.Sprintf(`%s/update/%s/%s/%s`, agent.serverHostURL, mtype, mid, mvalue)
	fmt.Printf(`send metric url - %s`, url)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf(`metric send request creation error: %v`, err)
	}

	req.Header.Set("Content-type", "text/plain")
	res, err := agent.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf(`error while send request with metric: %v`, err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf(`metric sending request return not ok status code`)
	}

	return nil

}
