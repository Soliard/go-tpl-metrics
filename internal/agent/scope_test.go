package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Soliard/go-tpl-metrics/cmd/agent/config"
	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/Soliard/go-tpl-metrics/models"
	"github.com/stretchr/testify/assert"
)

func setupTestAgent(serverHost string) *Agent {
	if serverHost == "" {
		serverHost = "http://localhost:8080"
	}
	config := config.Config{
		ServerHost:            serverHost,
		PollIntervalSeconds:   2,
		ReportIntervalSeconds: 10,
	}
	logger, err := logger.New("info")
	if err != nil {
		panic(err)
	}
	return New(&config, logger)
}

func TestNewAgent(t *testing.T) {
	agent := setupTestAgent("")

	if agent.collector == nil {
		t.Error("Expected collector to be initialized")
	}

	if agent.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}
}

func TestAgent_sendMetric(t *testing.T) {
	tests := []struct {
		name         string
		metric       *models.Metrics
		wantURL      string
		serverStatus int
		wantErr      bool
		wantErrMsg   string
	}{
		{
			name:         "successful gauge metric",
			metric:       models.NewGaugeMetric("testMetric", 123.45),
			wantURL:      fmt.Sprintf("/update/%s/testMetric/123.45", models.Gauge),
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "successful counter metric",
			metric:       models.NewCounterMetric("testMetric", 32),
			wantURL:      fmt.Sprintf("/update/%s/testMetric/32", models.Counter),
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "server error",
			metric:       models.NewGaugeMetric("testMetric", 123.45),
			wantURL:      fmt.Sprintf("/update/%s/testMetric/123.45", models.Gauge),
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
			wantErrMsg:   "metric sending request return not ok status code",
		},
		{
			name:         "server not found",
			metric:       models.NewGaugeMetric("testMetric", 123.45),
			wantURL:      fmt.Sprintf("/update/%s/testMetric/123.45", models.Gauge),
			serverStatus: http.StatusNotFound,
			wantErr:      true,
			wantErrMsg:   "metric sending request return not ok status code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.wantURL, r.URL.Path)
				w.WriteHeader(tt.serverStatus)
			}))
			defer server.Close()
			agent := setupTestAgent(server.URL)
			err := agent.sendMetric(tt.metric)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAgent_sendMetricJSON(t *testing.T) {
	tests := []struct {
		name         string
		metric       *models.Metrics
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful gauge metric",
			metric:       models.NewGaugeMetric("testMetric", 123.45),
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "server error",
			metric:       models.NewGaugeMetric("testMetric", 123.45),
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
		{
			name:         "gauge metric with max float64",
			metric:       models.NewGaugeMetric("maxFloat", 1.7976931348623157e+308),
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "gauge metric with min float64",
			metric:       models.NewGaugeMetric("minFloat", -1.7976931348623157e+308),
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "counter metric with max int64",
			metric:       models.NewCounterMetric("maxInt", 9223372036854775807),
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "counter metric with min int64",
			metric:       models.NewCounterMetric("minInt", -9223372036854775808),
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Простой сервер, который просто возвращает статус
			server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(tt.serverStatus)

				// Если успех - возвращаем метрику обратно
				if tt.serverStatus == http.StatusOK {
					body, _ := json.Marshal(tt.metric)
					res.Header().Set("Content-Type", "application/json")
					res.Write(body)
				}
			}))
			defer server.Close()

			agent := setupTestAgent(server.URL)
			err := agent.sendMetricJSON(tt.metric)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
