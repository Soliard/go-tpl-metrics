package signer

import (
	"bytes"
	"io"
	"net/http"

	"go.uber.org/zap"
)

func VerifySignatureMiddleware(key []byte, log *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !SignKeyExists(key) {
				next.ServeHTTP(w, r)
				return
			}
			signatureHex := r.Header.Get("HashSHA256")
			if signatureHex == "" {
				log.Warn("recieved request without sign")
				http.Error(w, "missing signature", http.StatusBadRequest)
				return
			}
			signature, err := DecodeSign(signatureHex)
			if err != nil {
				http.Error(w, "invalid signature format", http.StatusBadRequest)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read body", http.StatusBadRequest)
				return
			}
			// восстановление тела
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			if !Verify(body, key, signature) {
				http.Error(w, "invalid signature", http.StatusBadRequest)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
