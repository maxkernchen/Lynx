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
	"net"
	"os"
	"strings"
)

// Listen - Creates a welcomeSocket that listens for TCP connections - once someone connects a
// goroutine is spawned to handle the request
// @param handler func(net.Conn) err - This is the function we want to use to handle a new
// connection
// @param func(net.Conn) error handler - This is the function we use to handle the requests we get
func Listen(handler func(net.Conn) error) {
	fmt.Println("Starting Server on Port " + lynxutil.ServerPort)

	welcomeSocket, wErr := net.Listen("tcp", ":"+lynxutil.ServerPort)
	if wErr != nil {
		fmt.Println("Could Not Create Server Welcome Socket - Aborting.")
		os.Exit(lynxutil.SockErr) // Cannot recover from not being able to generate welcomeSocket
	}

	var cErr error
	for cErr == nil {
		conn, cErr := welcomeSocket.Accept()
		if cErr != nil {
			// If a connection error occurs
			continue // To avoid calling handler
		}
		go handler(conn)
	}

}

// HandleFileRequest - Handles a file request sent by another peer - this involves checking to see
// if we have the file and, if so, sending the file.
// @param net.Conn conn - The socket which the client is asking on
// @return error - An error can be produced when trying to send a file or if there is incorrect
// syntax in the request - otherwise error will be nil.
func HandleFileRequest(conn net.Conn) error {
	request, err := bufio.NewReader(conn).ReadString('\n') // Waits for a String ending in newline
	if err != nil {
		return err
	}

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
	key := []byte("abcdefghijklmnopqrstuvwxyz123456")
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
