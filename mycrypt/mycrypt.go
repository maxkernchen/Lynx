// Package mycrypt is a helper package that provides encryption / decryption using AES
// @author: Michael Bruce
// @author: Max Kernchen
// @verison: 3/25/2016
package mycrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// Encrypt - This function takes a key and a plain text byte slice and encrypts that slice using AES.
// @param []byte key - The key to be used for the encryption (AES requires only a single key
// for encryption / decryption)
// @param []byte text - The data that we would like encrypted.
// @returns []byte ciphertext - An encryted version of the data passed in.
// @returns error err - An error can be produced if a cipher cannot be created from the passed
// in key or if there is an issue reading from the passed in text. Otherwise it will be nil.
func Encrypt(key, text []byte) (ciphertext []byte, err error) {
	var block cipher.Block

	if block, err = aes.NewCipher(key); err != nil {
		return nil, err
	}

	ciphertext = make([]byte, aes.BlockSize+len(string(text)))

	// iv =  initialization vector
	iv := ciphertext[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], text)

	return
}

// Decrypt - This function takes a key and an encrypted byte slice and decrypts that slice using AES.
// @param []byte key - The key to be used for the encryption (AES requires only a single key
// for encryption / decryption)
// @param []byte ciphertext - The data that we would like decrypted.
// @returns []byte plaintext - A decryted version of the data passed in.
// @returns error err - An error can be produced if a cipher cannot be created from the passed
// in key or if there is an issue reading from the passed in text. Otherwise it will be nil.
func Decrypt(key, ciphertext []byte) (plaintext []byte, err error) {
	var block cipher.Block

	if block, err = aes.NewCipher(key); err != nil {
		return
	}

	if len(ciphertext) < aes.BlockSize {
		err = errors.New("ciphertext too short")
		return
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(ciphertext, ciphertext)

	plaintext = ciphertext

	return
}
