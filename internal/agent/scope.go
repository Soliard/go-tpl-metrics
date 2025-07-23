package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"time"

	"github.com/Soliard/go-tpl-metrics/internal/compressor"
	"github.com/Soliard/go-tpl-metrics/internal/signer"
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
			a.Logger.Info("stats collected")
			pollCounter = 0
		}

		if reportCounter >= int(a.reportInterval.Seconds()) {
			if err := a.reportMetricsBatch(); err != nil {
				a.Logger.Error("error while reporting metrics", zap.Error(err))
			}
			a.Logger.Warn("stats reported")
			reportCounter = 0
		}
	}
}

func (a *Agent) reportMetricsBatch() error {
	url, err := url.JoinPath(a.serverHostURL, "updates")
	if err != nil {
		return err
	}
	req := a.httpClient.R()

	body, err := json.Marshal(slices.Collect(maps.Values(a.collector.Metrics)))
	if err != nil {
		return fmt.Errorf("cant marshal metrics: %v", err)
	}
	compBody, err := compressor.CompressData(body)
	if err != nil {
		return fmt.Errorf("cant compress data: %v", err)
	}
	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	// resty позаботится о асептинге gzip и о расшифровке тела ответа из gzip
	req.Header.Set("Accept", "application/json")

	if a.hasSignKey() {
		signature := signer.Sign(body, a.signKey)
		req.Header.Set("HashSHA256", signer.EncodeSign(signature))
	}

	req.SetBody(compBody)

	res, err := req.Post(url)
	if err != nil {
		a.Logger.Error("error while send metrics to server",
			zap.Error(err))
		return err
	}

	//проверяем ответ
	if res.StatusCode() != http.StatusOK {
		a.Logger.Error("server returned not ok response for sended metrics",
			zap.Int("statuscode", res.StatusCode()))
		return errors.New("server returned not ok response for sended metrics")
	}

	return nil
}

func (a *Agent) reportMetrics() error {
	for _, value := range a.collector.Metrics {
		err := a.sendMetricJSON(value)
		if err != nil {
			return err
		}
	}

	a.Logger.Info("Metrics reported to the server")

	return nil
}

func (a *Agent) sendMetricJSON(metric *models.Metrics) error {
	url, err := url.JoinPath(a.serverHostURL, "update")
	if err != nil {
		return err
	}
	req := a.httpClient.R()

	buf, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("cant marshal metric: %v", err)
	}
	compressed, err := compressor.CompressData(buf)
	if err != nil {
		return fmt.Errorf("cant compress data: %v", err)
	}
	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	// resty позаботится о асептинге gzip и о расшифровке тела ответа из gzip
	req.Header.Set("Accept", "application/json")
	req.SetBody(compressed)

	res, err := req.Post(url)
	if err != nil {
		a.Logger.Error("error while send metric as json to server",
			zap.Any("metric", metric),
			zap.Error(err))
		return err
	}

	//проверяем ответ
	if res.StatusCode() != http.StatusOK {
		a.Logger.Error("server returned not ok response for sended metric from agent",
			zap.Any("metric", metric),
			zap.Int("statuscode", res.StatusCode()))
	}

	retMetric := models.Metrics{}
	err = json.Unmarshal(res.Body(), &retMetric)
	if err != nil {
		a.Logger.Error("cant unmarshal returned metric from server", zap.Error(err))
		return err
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
	url, err := url.JoinPath(a.serverHostURL, "update", metric.MType, metric.ID, value)
	if err != nil {
		return err
	}
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
