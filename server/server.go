// Package server - This package is the one responsible for sending data out.
// @author: Michael Bruce
// @author: Max Kernchen
// @verison: 2/17/2016
package server

import (
	"bufio"
	"bytes"
	"capstone/client"
	"capstone/lynxutil"
	"capstone/mycrypt"
	"compress/gzip"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
)

// Listen - Calls lynxutil to create a welcomeSocket that listens for TCP connections - once
// someone connects a goroutine is spawned to handle the request
func Listen() {
	lynxutil.Listen(handleFileRequest, lynxutil.ServerPort)
}

// handleFileRequest - Handles a file request sent by another peer - this involves checking to see
// if we have the file and, if so, sending the file.
// @param net.Conn conn - The socket which the client is asking on
// @return error - An error can be produced when trying to send a file or if there is incorrect
// syntax in the request - otherwise error will be nil.
func handleFileRequest(conn net.Conn) error {
	request, err := bufio.NewReader(conn).ReadString('\n') // Waits for a String ending in newline
	if err != nil {
		return err
	}

	// Will handle tracker request too
	// *Will handle receiving Meta as well
	tmpArr := strings.Split(request, ":")
	if len(tmpArr) != 2 {
		conn.Close()
		return errors.New("Invalid Request Syntax")
	}

	fileReq := tmpArr[1] // Gets the name of requested file
	fileReq = strings.TrimSpace(fileReq)

	fmt.Println("Asked for " + fileReq)

	haveFile := client.HaveFile(fileReq)
	fmt.Println(haveFile)

	// Depending on if we have the file - we write back to our client accordingly
	if haveFile {
		fmt.Fprintf(conn, "YES\n")    // Reply
		err = sendFile(fileReq, conn) // Sending The File
		if err != nil {
			return err
		}
	} else {
		fmt.Fprintf(conn, "NO\n") // Reply
	}
	fmt.Println("No Errors")
	return conn.Close()
}

// handleTrackerRequest - Handles a tracker request sent by another peer - this involves opening
// the meta.info file and passing the requesting peer the IP address stored inside.
// @param string request - The request the client made
// @param net.Conn conn - The socket which the client is asking on
// @return error - An error can be produced when trying to send a file or if there is incorrect
// syntax in the request - otherwise error will be nil.
func handleTrackerRequest(request string, conn net.Conn) error {
	tmpArr := strings.Split(request, ":")
	if len(tmpArr) != 2 {
		conn.Close()
		return errors.New("Invalid Request Syntax")
	}

	mPath := lynxutil.HomePath + tmpArr[1] + "meta.info"
	mPath = strings.TrimSpace(mPath)

	mFile, _ := os.Open(mPath)
	scanner := bufio.NewScanner(mFile)
	ip := (strings.Split(scanner.Text(), ":::"))[1] // Splits first line of metainfo and gets the IP
	fmt.Fprintf(conn, ip+"\n")

	return nil
}

// Helper function for handleRequest - handles the case where we are received meta.info file.
// @param net.Conn conn - The socket which the client is asking on
// @param string request - The request sent to tracker
// @return error - An error can be produced when trying to send a file or if there is incorrect
// syntax in the request - otherwise error will be nil.
func handlePush(request string, conn net.Conn) error {
	// Client syntax for push is "Meta_Push:<LynkName>\n"
	// So tmpArr[0] - Meta_Push | tmpArr[1] - <LynkName>
	tmpArr := strings.Split(request, ":")
	if len(tmpArr) != 2 {
		conn.Close()
		return errors.New("Invalid Request Syntax")
	}

	if client.IsDownloading(tmpArr[1]) {
		client.StopDownload(tmpArr[1])
	}

	metaPath := lynxutil.HomePath + tmpArr[1] + "/"
	err := os.Remove(metaPath)
	if err != nil {
		fmt.Println(err)
		return err
	}

	newMetainfo, err := os.Create(metaPath)
	if err != nil {
		fmt.Println(err)
		return err
	}

	bufIn, err := ioutil.ReadAll(conn)

	// Decrypt
	//key := []byte("abcdefghijklmnopqrstuvwxyz123456")
	key := []byte(lynxutil.PrivateKey)
	var plainFile []byte
	if plainFile, err = mycrypt.Decrypt(key, bufIn); err != nil {
		log.Fatal(err)
	}

	// Decompress
	r, _ := gzip.NewReader(bytes.NewBuffer(plainFile))
	bufOut, _ := ioutil.ReadAll(r)
	r.Read(bufOut)
	r.Close()

	fmt.Println(len(bufIn), "Bytes Received")
	newMetainfo.Write(bufOut)

	return nil // No errors if we reached this point
}

// Sends a file across the network to a peer.
// @param string fileName - The name of the file to send to the peer. It will have path from root
// of Lynx Directory.
// @param net.Conn conn - The socket over which we will send the file
// @return error - An error can be produced when trying open a file or write over
// the network - otherwise error will be nil.
func sendFile(fileName string, conn net.Conn) error {
	//fmt.Println(fileName)

	// Can use read when implementing chunking
	fBytes, err := ioutil.ReadFile(lynxutil.HomePath + fileName)

	// Begin Compression
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	gz.Write(fBytes)
	gz.Close()
	// End Compression

	// Begin Encryption
	var cipherFile []byte
	// The key length can be 32, 24, 16  bytes (OR in bits: 128, 192 or 256)
	//key := []byte("abcdefghijklmnopqrstuvwxyz123456")
	publicKey := lynxutil.Peer{IP: conn.LocalAddr().String()}.Key + lynxutil.PrivateKey
	key := []byte(publicKey)
	fmt.Println(publicKey)
	if cipherFile, err = mycrypt.Encrypt(key, b.Bytes()); err != nil {
		return err
	}
	// End Encryption

	n, err := conn.Write(cipherFile)
	if err != nil {
		return err
	}

	fmt.Println(n, "Bytes were sent")

	return nil // No Errors Occured If We Reached Here
}

// Sends the meta.info file to the tracker. Gets the tracker IP from the client.
// @param string metaPath - The meta.info path associated with the lynk we're interested in
// @return error - An error can be produced when trying to connect to the tracker
// over the network - otherwise error will be nil.
func pushMeta(metaPath string) error {
	trackerIP := client.GetTracker(metaPath)
	conn, err := net.Dial("tcp", trackerIP)
	if err != nil {
		fmt.Println(err)
		return err
	}

	lynkName := client.GetLynkName(metaPath)
	fmt.Fprintf(conn, "Meta_Push:"+lynkName+"\n") // Lets tracker know we are pushing

	err = sendFile(metaPath, conn)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return conn.Close()
}
