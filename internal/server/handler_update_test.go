package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Soliard/go-tpl-metrics/internal/compressor"
	"github.com/Soliard/go-tpl-metrics/internal/config"
	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/Soliard/go-tpl-metrics/internal/store"
	"github.com/Soliard/go-tpl-metrics/models"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T) (*httptest.Server, *MetricsService) {
	storage := store.NewMemoryStorage()
	config := config.ServerConfig{ServerHost: "localhost:8080"}
	logger, err := logger.New("info")
	require.NoError(t, err)
	service := NewMetricsService(storage, &config, logger)
	router := MetricRouter(service)
	return httptest.NewServer(router), service
}

func TestUpdateViaURLHandler(t *testing.T) {
	ts, _ := setupTestServer(t)
	client := resty.New()
	defer ts.Close()
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
	}{
		{
			name:           "valid gauge metric",
			method:         http.MethodPost,
			url:            "/update/gauge/testMetric/123.45",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid counter metric",
			method:         http.MethodPost,
			url:            "/update/counter/testCounter/42",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid method",
			method:         http.MethodGet,
			url:            "/update/gauge/testMetric/123.45",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "empty metric name",
			method:         http.MethodPost,
			url:            "/update/gauge//123.3",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid metric type",
			method:         http.MethodPost,
			url:            "/update/invalid/metricName/32",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing metric value",
			method:         http.MethodPost,
			url:            "/update/gauge/testMetric",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid gauge value",
			method:         http.MethodPost,
			url:            "/update/gauge/testMetric/not_a_number",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid counter value",
			method:         http.MethodPost,
			url:            "/update/counter/testCounter/not_a_number",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "without_id",
			method:         http.MethodPost,
			url:            "/update/counter/",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "empty metric type",
			method:         http.MethodPost,
			url:            "/update//testMetric/123.45",
			expectedStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.R().
				Execute(tt.method, ts.URL+tt.url)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode())
		})
	}
}

func TestUpdateHandler(t *testing.T) {
	ts, service := setupTestServer(t)
	defer ts.Close()
	client := resty.New()
	ctx := context.Background()
	tests := []struct {
		name           string
		metric         *models.Metrics
		method         string
		wantMetric     *models.Metrics
		expectedStatus int
	}{
		{
			name:           "valid gauge metric",
			method:         http.MethodPost,
			metric:         models.NewGaugeMetric("testMetric", 123.45),
			wantMetric:     models.NewGaugeMetric("testMetric", 123.45),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "second valid gauge metric",
			method:         http.MethodPost,
			metric:         models.NewGaugeMetric("testMetric", 123.45),
			wantMetric:     models.NewGaugeMetric("testMetric", 123.45),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid counter metric",
			method:         http.MethodPost,
			metric:         models.NewCounterMetric("testCounter", 42),
			wantMetric:     models.NewCounterMetric("testCounter", 42),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "second valid counter metric",
			method:         http.MethodPost,
			metric:         models.NewCounterMetric("testCounter", 42),
			wantMetric:     models.NewCounterMetric("testCounter", 84),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid method",
			method:         http.MethodGet,
			metric:         models.NewCounterMetric("testCounter", 42),
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.R().
				SetBody(tt.metric).
				Execute(tt.method, ts.URL+"/update")
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode())

			// Если ожидаем успех, проверим, что метрика сохранилась
			if tt.expectedStatus == http.StatusOK && tt.wantMetric != nil {
				got, err := service.GetMetric(ctx, tt.metric.ID)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantMetric.ID, got.ID)
				assert.Equal(t, tt.wantMetric.MType, got.MType)
				if got.MType == models.Gauge {
					assert.Equal(t, *tt.wantMetric.Value, *got.Value)
				} else if got.MType == models.Counter {
					assert.Equal(t, *tt.wantMetric.Delta, *got.Delta)
				}
			}
		})
	}
}

func Test_updateCounterMetric(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		value      int64
		wantErr    bool
		wantValue  int64
	}{
		{
			name:       "valid counter value",
			metricName: "testCounter",
			value:      42,
			wantErr:    false,
			wantValue:  42,
		},
		{
			name:       "negative value",
			metricName: "testCounter",
			value:      -10,
			wantErr:    false,
			wantValue:  -10,
		},
		{
			name:       "zero value",
			metricName: "testCounter",
			value:      0,
			wantErr:    false,
			wantValue:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, service := setupTestServer(t)
			ctx := context.Background()
			_, err := service.UpdateMetric(ctx, models.NewCounterMetric(tt.metricName, tt.value))

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "must be an integer")
			} else {
				assert.NoError(t, err)
				// Проверяем, что значение сохранилось в storage
				metric, err := service.GetMetric(ctx, tt.metricName)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantValue, *metric.Delta)
			}
		})
	}
}

func Test_updateCounterMetric_Accumulation(t *testing.T) {
	ctx := context.Background()
	_, s := setupTestServer(t)

	_, err := s.UpdateMetric(ctx, models.NewCounterMetric("testCounter", 10))
	assert.NoError(t, err)

	_, err = s.UpdateMetric(ctx, models.NewCounterMetric("testCounter", 20))
	assert.NoError(t, err)

	metric, err := s.GetMetric(ctx, "testCounter")
	assert.NoError(t, err)
	assert.Equal(t, int64(30), *metric.Delta) // 10 + 20
}

func Test_updateGaugeMetric(t *testing.T) {

	tests := []struct {
		name       string
		metricName string
		delta      float64
		wantErr    bool
		wantValue  float64
	}{
		{
			name:       "valid gauge value",
			metricName: "testGauge",
			delta:      123.45,
			wantErr:    false,
			wantValue:  123.45,
		},
		{
			name:       "negative value",
			metricName: "testGauge",
			delta:      -10.5,
			wantErr:    false,
			wantValue:  -10.5,
		},
		{
			name:       "zero value",
			metricName: "testGauge",
			delta:      0.0,
			wantErr:    false,
			wantValue:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, s := setupTestServer(t)
			ctx := context.Background()
			_, err := s.UpdateMetric(ctx, models.NewGaugeMetric(tt.metricName, tt.delta))

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Проверяем, что значение сохранилось в storage
				metric, err := s.GetMetric(ctx, tt.metricName)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantValue, *metric.Value)
			}
		})
	}
}

func Test_UpdatesHandler(t *testing.T) {
	server, service := setupTestServer(t)
	client := resty.New()
	tests := []struct {
		name       string
		metrics    []*models.Metrics
		wantStatus int
		wantSaved  []*models.Metrics // Ожидаемые метрики, которые должны сохраниться
	}{
		{
			name: "valid metrics",
			metrics: []*models.Metrics{
				models.NewGaugeMetric("test1", 42.0),
				models.NewCounterMetric("test2", 7),
			},
			wantStatus: 200,
			wantSaved: []*models.Metrics{
				models.NewGaugeMetric("test1", 42.0),
				models.NewCounterMetric("test2", 7),
			},
		},
		{
			name: "second valid metrics",
			metrics: []*models.Metrics{
				models.NewGaugeMetric("test1", 55.5),
				models.NewCounterMetric("test2", 7),
			},
			wantStatus: 200,
			wantSaved: []*models.Metrics{
				models.NewGaugeMetric("test1", 55.5),
				models.NewCounterMetric("test2", 14),
			},
		},
		{
			name:       "empty metrics",
			metrics:    []*models.Metrics{},
			wantStatus: 200,
			wantSaved:  []*models.Metrics{},
		},
		{
			name: "badrequest metrics",
			metrics: []*models.Metrics{
				{ID: "ttt", MType: "bad", Value: models.PFloat(3.0)},
				models.NewCounterMetric("test2", 7),
			},
			wantStatus: 400,
			wantSaved:  []*models.Metrics{},
		},
		// Можно добавить ещё кейсы
	}

	url, err := url.JoinPath(server.URL, "updates")
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.metrics)
			assert.NoError(t, err)
			compBody, err := compressor.CompressData(body)
			assert.NoError(t, err)
			res, err := client.R().
				SetHeader("Content-type", "application/json").
				SetHeader("Content-Encoding", "gzip").
				SetHeader("Accept", "application/json").
				SetBody(compBody).
				Post(url)
			assert.NoError(t, err)

			require.Equal(t, tt.wantStatus, res.StatusCode())

			for _, wantMetric := range tt.wantSaved {
				got, err := service.GetMetric(context.Background(), wantMetric.ID)
				require.NoError(t, err)
				require.Equal(t, wantMetric, got)
			}
		})
	}
}
