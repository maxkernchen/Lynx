// The unit tests for our mycrypt helper functions
// @author: Michael Bruce
// @author: Max Kernchen
// @verison: 2/17/2016
package mycrypt

import (
	"fmt"
	"testing"
)

// Count of the # of successful tests.
var successful = 0

// Total # of the tests.
const total = 3

// Unit tests for our Encrypt and Decrypt functions.
// @param *testing.T t - The wrapper for the test
func TestFileCopy(t *testing.T) {
	fmt.Println("\n----------------TestEncrypt----------------")

	var ciphertext, plaintext []byte
	var err error

	// The key length can be 32, 24, 16  bytes (OR in bits: 128, 192 or 256)
	key := []byte("longer means more possible keys ")
	plaintext = []byte("This is the unecrypted data. Referring to it as plain text.")

	if ciphertext, err = Encrypt(key, plaintext); err != nil {
		t.Error("Test failed, expected no errors. Got ", err)
	} else {
		fmt.Println("Successfully Encrypted File")
		successful++
	}

	fmt.Println("\n----------------TestDecrypt----------------")

	if plaintext, err = Decrypt(key, ciphertext); err != nil {
		t.Error("Test failed, expected no errors. Got ", err)
	} else {
		fmt.Println("Successfully Decrypted File")
		successful++
	}

	if string(plaintext) != "This is the unecrypted data. Referring to it as plain text." {
		t.Error("Test failed, expected 'This is the unecrypted data. Referring to it as plain text.' Got '" + string(plaintext) + "'")
	} else {
		fmt.Println("Decrypted File Contents Valid")
		successful++
	}

	fmt.Println("\nSuccess on ", successful, "/", total, " tests.")
}
