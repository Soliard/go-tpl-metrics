package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/Soliard/go-tpl-metrics/internal/compressor"
	"github.com/Soliard/go-tpl-metrics/internal/crypto"
	"github.com/Soliard/go-tpl-metrics/internal/signer"
	"github.com/Soliard/go-tpl-metrics/models"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

// Run запускает агент для сбора и отправки метрик.
// Создает горутины для сбора метрик и отправки данных с ограничением скорости.
func (a *Agent) Run(ctx context.Context) {
	jobs := make(chan []*models.Metrics, 10)
	sem := semaphore.NewWeighted(int64(a.requestRateLimit))

	var wg sync.WaitGroup

	// collectors
	for c := 1; c < 2; c++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			a.Collector(ctx, id, jobs)
		}(c)
	}

	for c := 1; c < 2; c++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			a.CollectorPS(ctx, id, jobs)
		}(c)
	}

	// senders
	for s := 1; s < 4; s++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			a.StartSender(ctx, jobs, sem)
		}(s)
	}

	// Ждем сигнал отмены
	<-ctx.Done()
	// Закрываем очередь заданий — отправители корректно дочитают
	close(jobs)
	// Дожидаемся завершения всех горутин
	wg.Wait()
}

// StartSender отправляет метрики на сервер с ограничением скорости.
// Использует семафор для контроля количества одновременных запросов.
func (a *Agent) StartSender(ctx context.Context, jobs <-chan []*models.Metrics, sem *semaphore.Weighted) {
	ticker := time.NewTicker(a.reportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			for j := range jobs {
				err := a.reportMetricsBatch(j)
				if err != nil {
					a.Logger.Error("failed to send metric while shutdown drainig", zap.Error(err))
				}
			}
			return

		case <-ticker.C:
			select {
			case <-ctx.Done():
				for j := range jobs {
					err := a.reportMetricsBatch(j)
					if err != nil {
						a.Logger.Error("failed to send metric while shutdown drainig", zap.Error(err))
					}
				}
				return

			case j, ok := <-jobs:
				if !ok {
					return
				}
				if err := sem.Acquire(ctx, 1); err != nil {
					a.Logger.Error("failed to acquire semaphore", zap.Error(err))
					continue
				}
				err := a.reportMetricsBatch(j)
				sem.Release(1)
				if err != nil {
					a.Logger.Error("error while sending metrics", zap.Error(err))
				}
			}
		}
	}
}

func (a *Agent) reportMetricsBatch(metrics []*models.Metrics) error {
	url, err := url.JoinPath(a.serverHostURL, "updates")
	if err != nil {
		return err
	}
	req := a.httpClient.R()

	body, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("cant marshal metrics: %v", err)
	}

	// Шифруем данные, если публичный ключ настроен
	if a.hasCryptoKey() {
		encryptedBody, err := crypto.EncryptHybrid(body, a.publicKey)
		if err != nil {
			return fmt.Errorf("cant encrypt data: %v", err)
		}
		body = encryptedBody
		a.Logger.Info("metrics encrypted successfully")
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
		signature := signer.Sign(compBody, a.signKey)
		req.Header.Set("HashSHA256", signer.EncodeSign(signature))
	}

	req.SetBody(compBody)

	res, err := req.Post(url)
	if err != nil {
		a.Logger.Error("error while send metrics to server",
			zap.Error(err),
			zap.String("recieved body", string(res.Body())))
		return err
	}

	//проверяем ответ
	if res.StatusCode() != http.StatusOK {
		a.Logger.Error("server returned not ok response for sended metrics",
			zap.Int("statuscode", res.StatusCode()),
			zap.String("recieved body", string(res.Body())))
		return errors.New("server returned not ok response for sended metrics")
	}

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

	// Шифруем данные, если публичный ключ настроен
	if a.hasCryptoKey() {
		encryptedBuf, err := crypto.EncryptHybrid(buf, a.publicKey)
		if err != nil {
			return fmt.Errorf("cant encrypt data: %v", err)
		}
		buf = encryptedBuf
		a.Logger.Info("metric encrypted successfully")
	}

	compressed, err := compressor.CompressData(buf)
	if err != nil {
		return fmt.Errorf("cant compress data: %v", err)
	}
	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	// resty позаботится о асептинге gzip и о расшифровке тела ответа из gzip
	req.Header.Set("Accept", "application/json")

	if a.hasSignKey() {
		signature := signer.Sign(compressed, a.signKey)
		req.Header.Set("HashSHA256", signer.EncodeSign(signature))
	}

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
