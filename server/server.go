/**
 *
 *	The server side of the Lynx application. Currently handles
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

func listen() {

}

func pushMeta() {

}

/**
 * Handles a file request sent by another peer
 * @param net.Conn clientConn - The socket which the client is asking on
 */
func handleFileRequest(clientConn net.Conn) {

	request, err := bufio.NewReader(clientConn).ReadString('\n') // Waits for a String ending in newline

	fileReq := strings.Split(request, ":")[1] // Gets the name of requested file

	haveFile := client.HaveFile(fileReq)

	// Depending on if we have the file - we write back to our client accordingly
	if haveFile {
		bufio.NewWriter(clientConn).WriteString("YES\n")
		sendFile(fileReq, clientConn)
	} else {
		bufio.NewWriter(clientConn).WriteString("NO\n")
	}

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
