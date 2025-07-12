package compressor

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"go.uber.org/zap"
)

// обертка для ответа чтобы увидеть его данные перед отпавкой клиенту
type bufferWriter struct {
	http.ResponseWriter
	headers    http.Header
	statusCode int
	bodyBuf    []byte
}

func (b *bufferWriter) Header() http.Header {
	return b.ResponseWriter.Header()
}

func (b *bufferWriter) WriteHeader(statusCode int) {
	b.statusCode = statusCode
}

func (b *bufferWriter) Write(p []byte) (int, error) {
	b.bodyBuf = append(b.bodyBuf, p...)
	return len(p), nil
}

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные
type compressReader struct {
	r  io.ReadCloser //оригинальный body запроса
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func GzipMiddleware(log *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			loggerFromCtx := logger.LoggerFromCtx(r.Context(), log)

			// проверяем, что клиент отправил серверу сжатые данные в формате gzip
			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				loggerFromCtx.Info("recieved body with supported compression")
				// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
				cr, err := newCompressReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				// меняем тело запроса на новое
				r.Body = cr
				defer cr.Close()
			}

			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(acceptEncoding, "gzip")
			if !supportsGzip {
				next.ServeHTTP(w, r)
				return
			}

			// Буферизуем ответ
			bw := &bufferWriter{ResponseWriter: w, headers: w.Header()}
			next.ServeHTTP(bw, r)

			contentType := bw.Header().Get("Content-Type")
			shouldCompress := strings.Contains(contentType, "html") ||
				strings.Contains(contentType, "json") ||
				strings.Contains(contentType, "xml")

			if shouldCompress {
				loggerFromCtx.Info("supports compression, response will be compressed")
				w.Header().Set("Content-Encoding", "gzip")
				if bw.statusCode != 0 {
					w.WriteHeader(bw.statusCode)
				}
				gz := gzip.NewWriter(w)
				defer gz.Close()
				gz.Write(bw.bodyBuf)
			} else {
				if bw.statusCode != 0 {
					w.WriteHeader(bw.statusCode)
				}
				w.Write(bw.bodyBuf)
			}
		})
	}
}
