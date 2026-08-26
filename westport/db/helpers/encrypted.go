package helpers

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

var EncryptionKey [32]byte

var encryptionSalt = []byte("west-port-v1")

func SetEncryptionPassword(password []byte) {
	key := argon2.IDKey(password, encryptionSalt, 3, 64*1024, 4, 32)
	copy(EncryptionKey[:], key)
}

// EncryptedBytes is a custom type that automatically encrypts/decrypts
type EncryptedBytes []byte

// Scan decrypts when reading from the database
func (e *EncryptedBytes) Scan(value any) error {
	if value == nil {
		*e = []byte{}
		return nil
	}

	var encrypted []byte
	switch v := value.(type) {
	case string:
		encrypted = []byte(v)
	case []byte:
		encrypted = v
	default:
		return fmt.Errorf("unexpected type for EncryptedString: %T", value)
	}

	if len(encrypted) == 0 {
		return nil
	}

	decrypted, err := Decrypt(string(encrypted))
	if err != nil {
		return fmt.Errorf("failed to decrypt: %w", err)
	}

	*e = decrypted
	return nil
}

// Value encrypts when writing to the database
func (e EncryptedBytes) Value() (driver.Value, error) {
	if len(e) == 0 {
		return []byte{}, nil
	}

	encrypted, err := Encrypt(e)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt: %w", err)
	}

	return []byte(encrypted), nil
}

func Encrypt(plaintext []byte) (string, error) {
	aead, err := chacha20poly1305.NewX(EncryptionKey[:])
	if err != nil {
		return "", err
	}

	// Generate a random nonce (24 bytes for XChaCha20)
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	// Encrypt and authenticate
	ciphertext := aead.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(ciphertext string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return []byte{}, err
	}

	aead, err := chacha20poly1305.NewX(EncryptionKey[:])
	if err != nil {
		return []byte{}, err
	}

	nonceSize := aead.NonceSize()
	if len(data) < nonceSize {
		return []byte{}, fmt.Errorf("ciphertext too short")
	}

	nonce, encrypted := data[:nonceSize], data[nonceSize:]
	plaintext, err := aead.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return []byte{}, err
	}

	return plaintext, nil
}

// String returns the decrypted value as a string
func (e EncryptedBytes) String() string {
	return string(e)
}
