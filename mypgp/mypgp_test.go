// The unit tests for our mycrypt helper functions
// @author: Michael Bruce
// @author: Max Kernchen
// @verison: 2/17/2016
package mypgp

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"testing"
	"time"
)

// Count of the # of successful tests.
var successful = 0

// Total # of the tests.
const total = 3

// The name of the current user
var currentusr, _ = user.Current()

// The home path for Lynx
var homePath = currentusr.HomeDir + "/Lynx/"

// Unit tests for creating, encoding, and decoding using OpenPGP public @ private keys.
// @param *testing.T t - The wrapper for the test
func TestPGP(t *testing.T) {
	fmt.Println("\n----------------TestCreateKey----------------")

	config := Config{Expiry: 365 * 24 * time.Hour}
	key, err := CreateKey("JohnDoe", "test key", "test@example.com", &config)

	if err != nil {
		t.Error("Test failed, expected no errors. Got ", err)
	} else {
		fmt.Println("Successfully Created File")
		successful++
	}

	fmt.Println("\n----------------TestEncodePublicKey----------------")

	output, err := key.Armor()
	fmt.Printf("%s\n", output)

	encodeTest(output)

	if err != nil {
		t.Error("Test failed, expected no errors. Got ", err)
	} else {
		fmt.Println("Successfully Encoded File")
		successful++
	}

	fmt.Println("\n----------------TestDecodeWithPrivateKey----------------")

	output, err = key.ArmorPrivate(&config)
	if err != nil {
		t.Error("Test failed, expected no errors. Got ", err)
	} else {
		fmt.Println("Successfully Encoded File")
		successful++
	}
	fmt.Printf("%s\n", output)

	decodeTest(output)

	fmt.Println("\nSuccess on ", successful, "/", total, " tests.")
}

// Helper function for testing decoding of a key.
// @params string - This is our private key
func decodeTest(key string) {
	var privateKey []byte
	privateKey = []byte(key)
	passphrase := []byte{}

	toDecrypt, err := os.OpenFile(homePath+"testE.txt", os.O_RDONLY, 0660)
	if err != nil {
		log.Fatal(err)
	}

	destination, err := os.OpenFile(homePath+"testU2.txt", os.O_WRONLY, 0660)
	if err != nil {
		log.Fatal(err)
	}

	// Decrypt...
	err = Decode(privateKey, passphrase, toDecrypt, destination)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Decrypted file!")
}

// Helper function for testing encoding of a key.
// @params string - This is our public key
func encodeTest(key string) {
	var publicKey []byte
	publicKey = []byte(key)

	toEncrypt, err := os.OpenFile(homePath+"testU.txt", os.O_RDONLY, 0660)
	if err != nil {
		log.Fatal(err)
	}

	destination, err := os.OpenFile(homePath+"testE.txt", os.O_WRONLY, 0660)
	if err != nil {
		log.Fatal(err)
	}

	// Encrypt...
	err = Encode(publicKey, toEncrypt, destination)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Encrypted file!")
}
