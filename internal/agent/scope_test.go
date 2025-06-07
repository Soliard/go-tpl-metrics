package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewAgent(t *testing.T) {
	agent := NewAgent("http://localhost:8080")

	if agent.collector == nil {
		t.Error("Expected collector to be initialized")
	}

	if agent.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}
}

func TestSendMetric(t *testing.T) {
	// Создаем тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/update/gauge/test/123.45" {
			t.Errorf("Expected path /update/gauge/test/123.45, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Создаем агента с тестовым сервером
	agent := NewAgent(server.URL)

	// Тестируем отправку метрики
	err := agent.sendMetric("gauge", "test", "123.45")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
