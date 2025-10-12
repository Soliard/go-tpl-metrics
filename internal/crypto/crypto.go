// internal/crypto/crypto.go - исправленная версия

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
)

// CryptoKey представляет криптографический ключ
type CryptoKey struct {
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}

// EncryptedData представляет зашифрованные данные с метаданными
type EncryptedData struct {
	EncryptedAESKey string `json:"aes_key"` // AES ключ, зашифрованный RSA (base64)
	EncryptedData   string `json:"data"`    // Данные, зашифрованные AES (base64)
}

// LoadPublicKey загружает публичный ключ из файла
func LoadPublicKey(keyPath string) (*rsa.PublicKey, error) {
	if keyPath == "" {
		return nil, nil // Ключ не указан
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	pubKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pubKey, nil
}

// LoadPrivateKey загружает приватный ключ из файла
func LoadPrivateKey(keyPath string) (*rsa.PrivateKey, error) {
	if keyPath == "" {
		return nil, nil // Ключ не указан
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

// EncryptHybrid шифрует данные с помощью гибридного метода (RSA + AES)
func EncryptHybrid(data []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	if publicKey == nil {
		return data, nil // Шифрование отключено
	}

	// Генерируем случайный AES ключ (32 байта для AES-256)
	aesKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return nil, fmt.Errorf("failed to generate AES key: %v", err)
	}

	// Создаем AES-GCM шифр
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %v", err)
	}

	// Генерируем случайный nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %v", err)
	}

	// Шифруем данные AES-GCM
	encryptedData := gcm.Seal(nonce, nonce, data, nil)

	// Шифруем AES ключ RSA публичным ключом
	hash := sha256.New()
	encryptedAESKey, err := rsa.EncryptOAEP(hash, rand.Reader, publicKey, aesKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt AES key with RSA: %v", err)
	}

	// Создаем структуру с зашифрованными данными
	encrypted := &EncryptedData{
		EncryptedAESKey: base64.StdEncoding.EncodeToString(encryptedAESKey),
		EncryptedData:   base64.StdEncoding.EncodeToString(encryptedData),
	}

	// Сериализуем в JSON для передачи
	return json.Marshal(encrypted)
}

// DecryptHybrid расшифровывает данные с помощью гибридного метода
func DecryptHybrid(encryptedData []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	if privateKey == nil {
		return encryptedData, nil // Расшифровка отключена
	}

	// Десериализуем структуру
	encrypted, err := DeserializeEncryptedData(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize encrypted data: %v", err)
	}

	// Декодируем base64 данные
	encryptedAESKeyBytes, err := base64.StdEncoding.DecodeString(encrypted.EncryptedAESKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode AES key: %v", err)
	}

	encryptedDataBytes, err := base64.StdEncoding.DecodeString(encrypted.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted data: %v", err)
	}

	// Расшифровываем AES ключ RSA приватным ключом
	hash := sha256.New()
	aesKey, err := rsa.DecryptOAEP(hash, rand.Reader, privateKey, encryptedAESKeyBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt AES key with RSA: %v", err)
	}

	// Создаем AES-GCM шифр
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %v", err)
	}

	// Извлекаем nonce из зашифрованных данных
	nonceSize := gcm.NonceSize()
	if len(encryptedDataBytes) < nonceSize {
		return nil, errors.New("encrypted data too short")
	}

	nonce, ciphertext := encryptedDataBytes[:nonceSize], encryptedDataBytes[nonceSize:]

	// Расшифровываем данные
	decryptedData, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data with AES: %v", err)
	}

	return decryptedData, nil
}

// DeserializeEncryptedData десериализует JSON данные в EncryptedData
func DeserializeEncryptedData(data []byte) (*EncryptedData, error) {
	var encrypted EncryptedData
	err := json.Unmarshal(data, &encrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// Проверяем, что поля не пустые
	if encrypted.EncryptedAESKey == "" {
		return nil, errors.New("encrypted AES key is empty")
	}
	if encrypted.EncryptedData == "" {
		return nil, errors.New("encrypted data is empty")
	}

	return &encrypted, nil
}

// Encrypt - старая функция для обратной совместимости
func Encrypt(data []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	return EncryptHybrid(data, publicKey)
}

// Decrypt - старая функция для обратной совместимости
func Decrypt(encryptedData []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	return DecryptHybrid(encryptedData, privateKey)
}
