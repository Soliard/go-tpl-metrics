package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Soliard/go-tpl-metrics/cmd/server/config"
	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/Soliard/go-tpl-metrics/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T) (*httptest.Server, *MetricsService) {
	storage := store.NewStorage()
	config := config.Config{ServerHost: "localhost:8080"}
	logger, err := logger.New("info")
	if err != nil {
		panic(err)
	}
	service := NewMetricsService(storage, &config, logger)
	router := MetricRouter(service)
	return httptest.NewServer(router), service
}

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestUpdateHandler(t *testing.T) {
	ts, _ := setupTestServer(t)
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
		{
			name:           "malformed URL",
			method:         http.MethodPost,
			url:            "/update",
			expectedStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, _ := testRequest(t, ts, tt.method, tt.url)
			defer res.Body.Close()
			assert.Equal(t, tt.expectedStatus, res.StatusCode)
		})
	}
}

func TestValueHandler(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	// Предварительная настройка данных
	updateTests := []struct {
		method string
		url    string
	}{
		{http.MethodPost, "/update/gauge/testMetric/123.45"},
		{http.MethodPost, "/update/counter/testCounter/42"},
	}

	for _, tt := range updateTests {
		res, _ := testRequest(t, ts, tt.method, tt.url)
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	}
	// Тестирование получения значений
	tests := []struct {
		name          string
		method        string
		url           string
		expectedValue string
		expectedCode  int
	}{
		{
			name:          "get gauge metric",
			method:        http.MethodGet,
			url:           "/value/gauge/testMetric",
			expectedValue: "123.45",
			expectedCode:  http.StatusOK,
		},
		{
			name:          "get counter metric",
			method:        http.MethodGet,
			url:           "/value/counter/testCounter",
			expectedValue: "42",
			expectedCode:  http.StatusOK,
		},
		{
			name:          "non-existent metric",
			method:        http.MethodGet,
			url:           "/value/gauge/nonExistent",
			expectedValue: "",
			expectedCode:  http.StatusNotFound,
		},
		{
			name:          "invalid method",
			method:        http.MethodPost,
			url:           "/value/gauge/testMetric",
			expectedValue: "",
			expectedCode:  http.StatusMethodNotAllowed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, body := testRequest(t, ts, tt.method, tt.url)
			defer res.Body.Close()
			assert.Equal(t, tt.expectedCode, res.StatusCode)
			if tt.expectedCode == http.StatusOK {
				assert.Equal(t, tt.expectedValue, body)
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

			err := service.updateCounterMetric(tt.metricName, &tt.value)

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
	err := s.updateCounterMetric("testCounter", &value1)
	assert.NoError(t, err)

	var value2 int64 = 20
	err = s.updateCounterMetric("testCounter", &value2)
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

			err := s.updateGaugeMetric(tt.metricName, &tt.delta)

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
