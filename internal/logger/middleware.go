package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	// берём структуру для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

// LoggingMiddleware создает HTTP middleware для логирования запросов и ответов.
// Логирует URL, метод, время выполнения, размер ответа и статус код.
func LoggingMiddleware(log *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lw := loggingResponseWriter{
				ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
				responseData: &responseData{
					status: 0,
					size:   0,
				},
			}
			ctx := r.Context()
			loggerFromCtx := LoggerFromCtx(ctx, log)

			loggerFromCtx.Info("request info",
				zap.String("url", r.URL.String()),
				zap.String("method", r.Method),
			)

			ctx = CtxWithLogger(ctx, loggerFromCtx)
			next.ServeHTTP(&lw, r.WithContext(ctx))
			duration := time.Since(start)

			loggerFromCtx.Info("response info",
				zap.Duration("duration", duration),
				zap.Int("size", lw.responseData.size),
				zap.Int("status", lw.responseData.status),
			)
		})
	}
}
