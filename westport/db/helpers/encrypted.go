package helpers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql/driver"
	"encoding/base64"
	"fmt"
	"io"
)

var EncryptionKey [32]byte
// var EncryptionKey = []byte("passphrasewhichneedstobe32bytes!")

// EncryptedBytes is a custom type that automatically encrypts/decrypts
type EncryptedBytes struct {
	Data []byte
}

// Scan decrypts when reading from the database
func (e *EncryptedBytes) Scan(value any) error {
	if value == nil {
		e.Data = []byte{}
		return nil
	}

	var encrypted string
	switch v := value.(type) {
	case string:
		encrypted = v
	case []byte:
		encrypted = string(v)
	default:
		return fmt.Errorf("unexpected type for EncryptedString: %T", value)
	}

	if encrypted == "" {
		e.Data = []byte{}
		return nil
	}

	decrypted, err := decrypt(encrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt: %w", err)
	}

	e.Data = decrypted
	return nil
}

// Value encrypts when writing to the database
func (e EncryptedBytes) Value() (driver.Value, error) {
	if len(e.Data) == 0 {
		return []byte{}, nil
	}

	encrypted, err := encrypt(e.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt: %w", err)
	}

	return encrypted, nil
}

func encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(EncryptionKey[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decrypt(ciphertext string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return []byte{}, err
	}

	block, err := aes.NewCipher(EncryptionKey[:])
	if err != nil {
		return []byte{}, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return []byte{}, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return []byte{}, fmt.Errorf("ciphertext too short")
	}

	nonce, encrypted := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return []byte{}, err
	}

	return plaintext, nil
}

// String returns the decrypted value as a string
func (e EncryptedBytes) String() string {
	return string(e.Data)
}
