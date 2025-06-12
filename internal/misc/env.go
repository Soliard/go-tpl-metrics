package misc

const (
	defaultServerURL = "localhost:8080"
	defaultAgentURL  = "http://localhost:8081"
)

// GetServerURL возвращает URL сервера из переменной окружения или значение по умолчанию
func GetServerURL() string {
	return defaultServerURL
}

func GetAgentURL() string {
	return defaultAgentURL
}
