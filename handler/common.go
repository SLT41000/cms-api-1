package handler

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func ToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(v).Int(), 10)
	case uint, uint8, uint16, uint32, uint64:
		return strconv.FormatUint(reflect.ValueOf(v).Uint(), 10)
	case float32, float64:
		return strconv.FormatFloat(reflect.ValueOf(v).Float(), 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", value)
	}
}

func ToInt(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case int8, int16, int32, int64:
		return int(reflect.ValueOf(v).Int())
	case uint, uint8, uint16, uint32, uint64:
		return int(reflect.ValueOf(v).Uint())
	case float32, float64:
		return int(reflect.ValueOf(v).Float())
	case string:
		num, _ := strconv.Atoi(strings.TrimSpace(v))
		return num // Returns 0 if string isn't a number
	case bool:
		if v {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func deriveKey(passphrase string) []byte {
	hash := sha256.Sum256([]byte(passphrase))
	return hash[:] // 32 bytes
}

func encrypt(plaintext string) (string, error) {
	key := deriveKey(os.Getenv("SECRET_KEY"))

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)

	// base64-encode so it's printable
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decrypt(ciphertextBase64 string) (string, error) {
	key := deriveKey(os.Getenv("SECRET_KEY"))

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
