// The unit tests for our server
// @author: Michael Bruce
// @author: Max Kernchen
// @verison: 2/17/2016
package server

import (
	"bufio"
	"bytes"
	"capstone/lynxutil"
	"capstone/mycrypt"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"testing"
)

// Count of the # of successful tests.
var successful = 0

// Total # of the tests.
const total = 6

// Unit tests for listen, handle, and send functions as well as push meta
// @param *testing.T t - The wrapper for the test
func TestListenHandleSend(t *testing.T) {
	fmt.Println("\n----------------TestListen----------------")

	conn, err := net.Dial("tcp", "127.0.0.1:8080")

	if err != nil {
		t.Error(err.Error())
	} else {
		fmt.Println("Successfully Connected To Server")
		successful++
	}

	fmt.Println("\n----------------TestHandleRequest----------------")

	fmt.Fprintf(conn, "Do_You_Have_FileName:Tests/fake.txt\n")

	reply, err := bufio.NewReader(conn).ReadString('\n') // Waits for a String ending in newline
	reply = strings.TrimSpace(reply)

	if reply == "NO" {
		fmt.Println("Successfully Handled Invalid Request")
		successful++
	} else {
		t.Error("Test failed, expected to have server respond 'NO'. Got", reply)
	}
	conn.Close()

	conn, err = net.Dial("tcp", "127.0.0.1:8080")

	fmt.Fprintf(conn, "Do_You_Have_FileName:Tests/test.txt\n")

	reply, err = bufio.NewReader(conn).ReadString('\n') // Waits for a String ending in newline
	reply = strings.TrimSpace(reply)

	if reply == "YES" {
		fmt.Println("Successfully Handled Valid Request")
		successful++
	} else {
		t.Error("Test failed, expected to have server respond 'YES'. Got", reply)
	}

	fmt.Println("\n----------------TestSendFile----------------")

	file, err := os.Create("test.txt_ServerTest")
	if err != nil {
		t.Error(err.Error())
	}
	defer file.Close()

	bufIn := make([]byte, 512) // Set to 512 because we know this file is small
	_, err = conn.Read(bufIn)
	// Decrypt
	//key := []byte("abcdefghijklmnopqrstuvwxyz123456")
	key := []byte(lynxutil.PrivateKey)
	var plainFile []byte
	if plainFile, err = mycrypt.Decrypt(key, bufIn); err != nil {
		log.Fatal(err)
	}
	// Decompress
	r, err := gzip.NewReader(bytes.NewBuffer(plainFile))
	bufOut := make([]byte, 512) // Set to 512 because we know this file is small
	r.Read(bufOut)
	file.Write(bufOut)
	r.Close()

	if err != nil {
		t.Error(err.Error())
	} else {
		fmt.Println("Successfully Sent 'text.txt'")
		successful++
	}

	content, _ := ioutil.ReadFile("test.txt_ServerTest")
	s := string(content)

	if !strings.Contains(s, "test contents") {
		t.Error("File contents invalid. Got '" + s + "'")
	} else {
		fmt.Println("File contents valid.")
		successful++
	}

	fmt.Println("\n----------------TestPushMeta----------------")

	if err != nil {
		t.Error(err.Error())
	} else {
		fmt.Println("Successfully Pushed Meta")
		successful++
	}

	fmt.Println("\nSuccess on ", successful, "/", total, " tests.")
}
