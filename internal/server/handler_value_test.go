package server

import (
	"context"
	"net/http"
	"testing"

	"github.com/Soliard/go-tpl-metrics/models"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValueViaURLHandler(t *testing.T) {
	ts, _ := setupTestServer(t)
	client := resty.New()
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
		resp, err := client.R().
			Execute(tt.method, ts.URL+tt.url)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode())
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
			resp, err := client.R().
				Execute(tt.method, ts.URL+tt.url)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedCode, resp.StatusCode())
			if tt.expectedCode == http.StatusOK {
				assert.Equal(t, tt.expectedValue, string(resp.Body()))
			}
		})
	}

}

func TestValueHandler(t *testing.T) {
	ts, s := setupTestServer(t)
	defer ts.Close()
	ctx := context.Background()
	client := resty.New()

	// Предварительная настройка данных
	s.UpdateMetric(ctx, models.NewCounterMetric("counter", 3))
	s.UpdateMetric(ctx, models.NewGaugeMetric("gauge", 22.1))

	tests := []struct {
		name       string
		metric     *models.Metrics
		wantMetric *models.Metrics
	}{
		{
			name:       "get gauge metric",
			metric:     &models.Metrics{ID: "gauge", MType: models.Gauge},
			wantMetric: &models.Metrics{ID: "gauge", MType: models.Gauge, Value: models.PFloat(22.1)},
		},
		{
			name:       "get counter metric",
			metric:     &models.Metrics{ID: "counter", MType: models.Counter},
			wantMetric: &models.Metrics{ID: "counter", MType: models.Counter, Delta: models.PInt(3)},
		},
		{
			name:       "get non existed metric",
			metric:     &models.Metrics{ID: "123", MType: models.Counter},
			wantMetric: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			returnedMetric := models.Metrics{}
			resp, err := client.R().
				SetBody(tt.metric).
				SetResult(&returnedMetric).
				Execute(http.MethodPost, ts.URL+`/value/`)

			require.NoError(t, err)
			if tt.wantMetric != nil {
				assert.Equal(t, http.StatusOK, resp.StatusCode())
				assert.NoError(t, err)
				assert.Equal(t, *tt.wantMetric, returnedMetric)
			} else {
				assert.NotEqual(t, http.StatusOK, resp.StatusCode())
			}

		})
	}
}
