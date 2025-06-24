package logger

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const (
	ctxKeyLogger ctxKey = "logger"
)

type ctxKey string

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

func New(level string) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}
	// создаём новую конфигурацию логера
	cfg := zap.NewProductionConfig()
	// устанавливаем уровень
	cfg.Level = lvl
	// создаём логер на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	return zl, nil
}

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

func LoggingMiddleware(defLogger *zap.Logger) func(next http.Handler) http.Handler {
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
			loggerFromCtx := LoggerFromCtx(ctx, defLogger)

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

func CtxWithLogger(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxKeyLogger, l)
}

func LoggerFromCtx(ctx context.Context, defaultLogger *zap.Logger) *zap.Logger {
	loggerFromCtx, ok := ctx.Value(ctxKeyLogger).(*zap.Logger)
	if !ok {
		return defaultLogger
	}
	return loggerFromCtx
}
