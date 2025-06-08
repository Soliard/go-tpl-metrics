package misc

const (
	defaultServerURL = "localhost:8080"
)

// GetServerURL возвращает URL сервера из переменной окружения или значение по умолчанию
func GetServerURL() string {
	return defaultServerURL
}
