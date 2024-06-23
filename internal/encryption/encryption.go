package encryption 

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"bytes"
)

func Encrypt(password string, key []byte) (string, error) {
    plaintext := []byte(password)

    padding := aes.BlockSize - len(plaintext)%aes.BlockSize
    padtext := bytes.Repeat([]byte{byte(padding)}, padding)
    plaintext = append(plaintext, padtext...)

    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }

    ciphertext := make([]byte, aes.BlockSize+len(plaintext))
    iv := ciphertext[:aes.BlockSize]
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return "", err
    }

    mode := cipher.NewCBCEncrypter(block, iv)
    mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

    return base64.URLEncoding.EncodeToString(ciphertext), nil
}


func Decrypt(ciphertextStr string, key []byte) (string, error) {
    ciphertext, err := base64.URLEncoding.DecodeString(ciphertextStr)
    if err != nil {
        return "", err
    }

    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }

    if len(ciphertext) < aes.BlockSize {
        return "", fmt.Errorf("ciphertext too short")
    }
    iv := ciphertext[:aes.BlockSize]
    ciphertext = ciphertext[aes.BlockSize:]

    mode := cipher.NewCBCDecrypter(block, iv)
    mode.CryptBlocks(ciphertext, ciphertext)

    // Remove PKCS#7 padding
    padLen := int(ciphertext[len(ciphertext)-1])
    if padLen > len(ciphertext) || padLen > aes.BlockSize {
        return "", fmt.Errorf("padding size is invalid")
    }
    ciphertext = ciphertext[:len(ciphertext)-padLen]

    return string(ciphertext), nil
}
