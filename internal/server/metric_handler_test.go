package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Soliard/go-tpl-metrics/internal/store"
	"github.com/stretchr/testify/assert"
)

func TestUpdateClaimMetric(t *testing.T) {
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
			url:            "/update/gauge",
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
			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			UpdateClaimMetric(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// probably shouldnt done those
func Test_parseClaimMetricURL(t *testing.T) {
	type want struct {
		metricType  string
		metricName  string
		matricValue string
	}
	tests := []struct {
		name      string
		url       string
		want      want
		wantError bool
	}{
		{
			name:      "valid gauge metric",
			url:       "/update/gauge/testMetric/123.45",
			want:      want{metricType: "gauge", metricName: "testMetric", matricValue: "123.45"},
			wantError: false,
		},
		{
			name:      "valid counter metric",
			url:       "/update/counter/testCounter/42",
			want:      want{metricType: "counter", metricName: "testCounter", matricValue: "42"},
			wantError: false,
		},
		{
			name:      "invalid url format",
			url:       "/update/invalid",
			want:      want{metricType: "invalid", metricName: "", matricValue: ""},
			wantError: false,
		},
		{
			name:      "empty url",
			url:       "",
			want:      want{},
			wantError: true,
		},
		{
			name:      "invalid metric type",
			url:       "/update/invalid/testMetric/123",
			want:      want{metricType: "invalid", metricName: "testMetric", matricValue: "123"},
			wantError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMetricType, gotMetricName, gotMetricValue, err := parseClaimMetricURL(tt.url)
			if err != nil && !tt.wantError {
				assert.NoError(t, err)
			}
			got := want{metricType: gotMetricType, metricName: gotMetricName, matricValue: gotMetricValue}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_updateCounterMetric(t *testing.T) {
	mockStorage := store.NewStorage()
	storage = mockStorage

	tests := []struct {
		name       string
		metricName string
		value      string
		wantErr    bool
		wantValue  int64
	}{
		{
			name:       "valid counter value",
			metricName: "testCounter",
			value:      "42",
			wantErr:    false,
			wantValue:  42,
		},
		{
			name:       "invalid value format",
			metricName: "testCounter",
			value:      "not_a_number",
			wantErr:    true,
			wantValue:  0,
		},
		{
			name:       "negative value",
			metricName: "testCounter",
			value:      "-10",
			wantErr:    false,
			wantValue:  -10,
		},
		{
			name:       "zero value",
			metricName: "testCounter",
			value:      "0",
			wantErr:    false,
			wantValue:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage = store.NewStorage()
			storage = mockStorage

			err := updateCounterMetric(tt.metricName, tt.value)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "must be an integer")
			} else {
				assert.NoError(t, err)
				// Проверяем, что значение сохранилось в storage
				value, exists := mockStorage.GetCounter(tt.metricName)
				assert.True(t, exists)
				assert.Equal(t, tt.wantValue, value)
			}
		})
	}
}

func Test_updateCounterMetric_Accumulation(t *testing.T) {
	mockStorage := store.NewStorage()
	storage = mockStorage

	err := updateCounterMetric("testCounter", "10")
	assert.NoError(t, err)

	err = updateCounterMetric("testCounter", "20")
	assert.NoError(t, err)

	value, exists := mockStorage.GetCounter("testCounter")
	assert.True(t, exists)
	assert.Equal(t, int64(30), value) // 10 + 20
}

func Test_updateGaugeMetric(t *testing.T) {
	mockStorage := store.NewStorage()
	storage = mockStorage

	tests := []struct {
		name       string
		metricName string
		value      string
		wantErr    bool
		wantValue  float64
	}{
		{
			name:       "valid gauge value",
			metricName: "testGauge",
			value:      "123.45",
			wantErr:    false,
			wantValue:  123.45,
		},
		{
			name:       "invalid value format",
			metricName: "testGauge",
			value:      "not_a_number",
			wantErr:    true,
			wantValue:  0,
		},
		{
			name:       "negative value",
			metricName: "testGauge",
			value:      "-10.5",
			wantErr:    false,
			wantValue:  -10.5,
		},
		{
			name:       "zero value",
			metricName: "testGauge",
			value:      "0.0",
			wantErr:    false,
			wantValue:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage = store.NewStorage()
			storage = mockStorage

			err := updateGaugeMetric(tt.metricName, tt.value)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "must be a float")
			} else {
				assert.NoError(t, err)
				// Проверяем, что значение сохранилось в storage
				value, exists := mockStorage.GetGauge(tt.metricName)
				assert.True(t, exists)
				assert.Equal(t, tt.wantValue, value)
			}
		})
	}
}
