// Package signer предоставляет утилиты для подписи и проверки данных.
// Использует HMAC-SHA256 для создания и проверки подписей.
package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Sign создает HMAC-SHA256 подпись для данных с использованием указанного ключа.
func Sign(data []byte, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

// Verify проверяет HMAC-SHA256 подпись данных.
// Возвращает true если подпись корректна, false в противном случае.
func Verify(data, key, signature []byte) bool {
	expected := Sign(data, key)
	return hmac.Equal(expected, signature)
}

// EncodeSign кодирует подпись в hex строку для передачи в HTTP заголовках.
func EncodeSign(sign []byte) string {
	return hex.EncodeToString(sign)
}

// DecodeSign декодирует hex строку обратно в байты подписи.
func DecodeSign(sign string) ([]byte, error) {
	s, err := hex.DecodeString(sign)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// SignKeyExists проверяет, что ключ для подписи не пустой.
func SignKeyExists(k []byte) bool {
	return len(k) > 0
}
