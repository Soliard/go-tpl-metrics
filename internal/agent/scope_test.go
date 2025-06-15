package agent

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Soliard/go-tpl-metrics/models"
	"github.com/stretchr/testify/assert"
)

func setupTestAgent(serverHost string) *Agent {
	if serverHost == "" {
		serverHost = "http://localhost:8080"
	}
	config := Config{
		ServerHost:     serverHost,
		PollInterval:   2,
		ReportInterval: 10,
	}

	return NewAgent(config)
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
		metricType   string
		metricID     string
		metricValue  string
		wantURL      string
		serverStatus int
		wantErr      bool
		wantErrMsg   string
	}{
		{
			name:         "successful gauge metric",
			metricType:   models.Gauge,
			metricID:     "testMetric",
			metricValue:  "123.45",
			wantURL:      fmt.Sprintf("/update/%s/testMetric/123.45", models.Gauge),
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "successful counter metric",
			metricType:   models.Counter,
			metricID:     "testMetric",
			metricValue:  "32",
			wantURL:      fmt.Sprintf("/update/%s/testMetric/32", models.Counter),
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "server error",
			metricType:   models.Gauge,
			metricID:     "testMetric",
			metricValue:  "123.45",
			wantURL:      fmt.Sprintf("/update/%s/testMetric/123.45", models.Gauge),
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
			wantErrMsg:   "metric sending request return not ok status code",
		},
		{
			name:         "server not found",
			metricType:   models.Gauge,
			metricID:     "testMetric",
			metricValue:  "123.45",
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
			err := agent.sendMetric(tt.metricType, tt.metricID, tt.metricValue)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
