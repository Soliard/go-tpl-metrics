package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func Sign(data []byte, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

func Verify(data, key, signature []byte) bool {
	expected := Sign(data, key)
	return hmac.Equal(expected, signature)
}

func EncodeSign(sign []byte) string {
	return hex.EncodeToString(sign)
}

func DecodeSign(sign string) ([]byte, error) {
	s, err := hex.DecodeString(sign)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func SignKeyExists(k []byte) bool {
	return len(k) > 0
}
