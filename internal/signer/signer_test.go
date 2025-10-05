package signer

import (
	"bytes"
	"testing"
)

func TestSign(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		key      []byte
		expected bool
	}{
		{
			name:     "valid signature",
			data:     []byte("test data"),
			key:      []byte("secret key"),
			expected: true,
		},
		{
			name:     "empty data",
			data:     []byte(""),
			key:      []byte("secret key"),
			expected: true,
		},
		{
			name:     "empty key",
			data:     []byte("test data"),
			key:      []byte(""),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature := Sign(tt.data, tt.key)
			if len(signature) != 32 { // SHA256 produces 32 bytes
				t.Errorf("Sign() signature length = %d, want 32", len(signature))
			}
		})
	}
}

func TestVerify(t *testing.T) {
	data := []byte("test data")
	key := []byte("secret key")
	signature := Sign(data, key)

	tests := []struct {
		name      string
		data      []byte
		key       []byte
		signature []byte
		want      bool
	}{
		{
			name:      "valid signature",
			data:      data,
			key:       key,
			signature: signature,
			want:      true,
		},
		{
			name:      "invalid signature",
			data:      data,
			key:       key,
			signature: []byte("invalid"),
			want:      false,
		},
		{
			name:      "wrong key",
			data:      data,
			key:       []byte("wrong key"),
			signature: signature,
			want:      false,
		},
		{
			name:      "wrong data",
			data:      []byte("wrong data"),
			key:       key,
			signature: signature,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Verify(tt.data, tt.key, tt.signature); got != tt.want {
				t.Errorf("Verify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncodeSign(t *testing.T) {
	tests := []struct {
		name     string
		sign     []byte
		expected string
	}{
		{
			name:     "valid signature",
			sign:     []byte{0x01, 0x02, 0x03, 0x04},
			expected: "01020304",
		},
		{
			name:     "empty signature",
			sign:     []byte{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodeSign(tt.sign)
			if got != tt.expected {
				t.Errorf("EncodeSign() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDecodeSign(t *testing.T) {
	tests := []struct {
		name        string
		sign        string
		expected    []byte
		expectError bool
	}{
		{
			name:        "valid hex string",
			sign:        "01020304",
			expected:    []byte{0x01, 0x02, 0x03, 0x04},
			expectError: false,
		},
		{
			name:        "empty string",
			sign:        "",
			expected:    []byte{},
			expectError: false,
		},
		{
			name:        "invalid hex string",
			sign:        "invalid",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeSign(tt.sign)
			if (err != nil) != tt.expectError {
				t.Errorf("DecodeSign() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !bytes.Equal(got, tt.expected) {
				t.Errorf("DecodeSign() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSignKeyExists(t *testing.T) {
	tests := []struct {
		name string
		key  []byte
		want bool
	}{
		{
			name: "non-empty key",
			key:  []byte("secret"),
			want: true,
		},
		{
			name: "empty key",
			key:  []byte(""),
			want: false,
		},
		{
			name: "nil key",
			key:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SignKeyExists(tt.key); got != tt.want {
				t.Errorf("SignKeyExists() = %v, want %v", got, tt.want)
			}
		})
	}
}
