/**
 *
 *	The server side of the Lynx application. Currently handles ~
 *
 *	 @author: Michael Bruce
 *	 @author: Max Kernchen
 *
 *	 @verison: 2/17/2016
 */

package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

/**
 * Function used to drive and test our server's functions
 */
func main() {

}

// ------------------------- CODE BELOW THIS LINE IS UNTESTED AND DANGEROUS ------------------------- \\

/**
 * Creates a welcomeSocket that listens for TCP connections - once someone connects a goroutine is spawned
 * to handle the request
 */
func listen() {

	welcomeSocket, wErr := net.Listen("tcp", ":8080") // Will later need to set port dynamically

	if wErr != nil {
		// handle error
	}

	var cErr error

	for cErr == nil {
		conn, err := welcomeSocket.Accept()
		if cErr != nil {
			// handle error
		}
		go handleFileRequest(conn)
	}

}

/**
 * Sends the meta.info file to the tracker. Gets the tracker IP from the client.
 */
func pushMeta() {
	trackerIP := client.GetTrackerIP()
	conn, err := net.Dial("tcp", trackerIP)
	if err != nil {
		return
	}

	sendFile("meta.info", conn)
}

/**
 * Handles a file request sent by another peer
 * @param net.Conn clientConn - The socket which the client is asking on
 */
func handleFileRequest(clientConn net.Conn) error {

	request, err := bufio.NewReader(clientConn).ReadString('\n') // Waits for a String ending in newline
	if err != nil {
		return err
	}

	fileReq := strings.Split(request, ":")[1] // Gets the name of requested file

	haveFile := client.HaveFile(fileReq)

	// Depending on if we have the file - we write back to our client accordingly
	if haveFile {
		bufio.NewWriter(clientConn).WriteString("YES\n")
		err = sendFile(fileReq, clientConn)
		if err != nil {
			return err
		}
	} else {
		bufio.NewWriter(clientConn).WriteString("NO\n")
	}

	return err
}

/**
 * Sends a file to a peer
 * @param string fileName - The name of the file to send to the peer
 * @param net.Conn clientConn - The socket over which we will send the file
 */
func sendFile(fileName string, clientConn net.Conn) error {

	fileToSend, err := os.Open(fileName)
	if err != nil {
		return err
	}

	n, err := io.Copy(fileToSend, clientConn)
	if err != nil {
		return err
	}
	fmt.Println(n, "this was sent")

	return fileToSend.Close()

}
