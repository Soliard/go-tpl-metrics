package agent

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Soliard/go-tpl-metrics/models"
	"go.uber.org/zap"
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
				a.Logger.Error("error while collection metrics", zap.Error(err))
			}
			pollCounter = 0
		}

		if reportCounter >= int(a.reportInterval.Seconds()) {
			if err := a.reportMetrics(); err != nil {
				a.Logger.Error("error while reporting metrics", zap.Error(err))
			}
			reportCounter = 0
		}
	}
}

func (a *Agent) reportMetrics() error {

	for _, value := range a.collector.Metrics {
		err := a.sendMetricJSON(value)
		if err != nil {
			return err
		}
	}

	a.Logger.Debug("Metrics reported to the server")

	return nil
}

func (a *Agent) sendMetricJSON(metric *models.Metrics) error {
	url := fmt.Sprintf(`%s/update`, a.serverHostURL)
	req := a.httpClient.R()
	req.SetHeader("Content-type", "application/json")
	req.SetBody(metric)
	res, err := req.Post(url)
	if err != nil {
		a.Logger.Error("error while send metric as json to server",
			zap.Any("metric", metric))
		return err
	}

	if res.StatusCode() != http.StatusOK {
		a.Logger.Error("server returned not ok response for sended metric from agent",
			zap.Any("metric", metric),
			zap.Int("statuscode", res.StatusCode()))
	}

	return nil
}

func (a *Agent) sendMetric(metric *models.Metrics) error {
	var value string
	if metric.MType == models.Counter {
		value = metric.StringifyDelta()
	} else {
		value = metric.StringifyValue()
	}
	url := fmt.Sprintf(`%s/update/%s/%s/%s`, a.serverHostURL, metric.MType, metric.ID, value)
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
