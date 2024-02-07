package common

import (
	"bytes"
	"crypto/aes"
	"encoding/hex"
	"errors"
)

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func AESEncryptECB(key, plaintext string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	msg := PKCS7Padding([]byte(plaintext), block.BlockSize())
	ciphertext := make([]byte, len(msg))

	for bs, be := 0, block.BlockSize(); bs < len(msg); bs, be = bs+block.BlockSize(), be+block.BlockSize() {
		block.Encrypt(ciphertext[bs:be], msg[bs:be])
	}

	return hex.EncodeToString(ciphertext), nil
}

func AESDecryptECB(key, ciphertext string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	decodedMsg, err := hex.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	if len(decodedMsg)%block.BlockSize() != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}

	plaintext := make([]byte, len(decodedMsg))

	for bs, be := 0, block.BlockSize(); bs < len(decodedMsg); bs, be = bs+block.BlockSize(), be+block.BlockSize() {
		block.Decrypt(plaintext[bs:be], decodedMsg[bs:be])
	}

	return string(PKCS7UnPadding(plaintext)), nil
}
