package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Soliard/go-tpl-metrics/cmd/server/config"
	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/Soliard/go-tpl-metrics/internal/store"
	"github.com/Soliard/go-tpl-metrics/models"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T) (*httptest.Server, *MetricsService) {
	storage := store.NewMemoryStorage()
	config := config.Config{ServerHost: "localhost:8080"}
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
	client := resty.New()
	defer ts.Close()
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
				got, exists := service.GetMetric(tt.metric.ID)
				assert.True(t, exists)
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

			err := service.UpdateCounter(tt.metricName, &tt.value)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "must be an integer")
			} else {
				assert.NoError(t, err)
				// Проверяем, что значение сохранилось в storage
				metric, exists := service.GetMetric(tt.metricName)
				assert.True(t, exists)
				assert.Equal(t, tt.wantValue, *metric.Delta)
			}
		})
	}
}

func Test_updateCounterMetric_Accumulation(t *testing.T) {
	_, s := setupTestServer(t)

	var value1 int64 = 10
	err := s.UpdateCounter("testCounter", &value1)
	assert.NoError(t, err)

	var value2 int64 = 20
	err = s.UpdateCounter("testCounter", &value2)
	assert.NoError(t, err)

	metric, exists := s.GetMetric("testCounter")
	assert.True(t, exists)
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

			err := s.UpdateGauge(tt.metricName, &tt.delta)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Проверяем, что значение сохранилось в storage
				metric, exists := s.GetMetric(tt.metricName)
				assert.True(t, exists)
				assert.Equal(t, tt.wantValue, *metric.Value)
			}
		})
	}
}
