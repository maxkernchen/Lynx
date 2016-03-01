/**
 *
 *	The server side of the Lynx application. Currently handles ~ sending files, handling file requests,
 *  listening for client connections.
 *
 *	 @author: Michael Bruce
 *	 @author: Max Kernchen
 *
 *	 @verison: 2/17/2016
 */

package main

import (
	"bufio"
	"capstone/client"
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
	listen()
}

// ------------------------- CODE BELOW THIS LINE IS UNTESTED AND DANGEROUS ------------------------- \\

/**
 * Creates a welcomeSocket that listens for TCP connections - once someone connects a goroutine is spawned
 * to handle the request
 */
func listen() {

	fmt.Println("Starting Server on Port 8080")

	welcomeSocket, wErr := net.Listen("tcp", ":8080") // Will later need to set port dynamically

	if wErr != nil {
		// handle error
	}

	var cErr error

	for cErr == nil {
		conn, cErr := welcomeSocket.Accept()
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
 * @param net.Conn conn - The socket which the client is asking on
 */
func handleFileRequest(conn net.Conn) error {

	request, err := bufio.NewReader(conn).ReadString('\n') // Waits for a String ending in newline
	if err != nil {
		return err
	}

	// NEED TO CHECK PROPER FORMAT BEFORE ACCESSING INDEX 1
	fileReq := strings.Split(request, ":")[1] // Gets the name of requested file
	fileReq = strings.TrimSpace(fileReq)

	fmt.Println("Asked for " + fileReq)

	haveFile := client.HaveFile(fileReq)
	//writer   := bufio.NewWriter(conn)
	fmt.Println(haveFile)

	// Depending on if we have the file - we write back to our client accordingly
	if haveFile {
		//bufio.NewWriter(conn).WriteString("YES\n")
		fmt.Fprintf(conn, "YES\n")
		err = sendFile(fileReq, conn)
		if err != nil {
			return err
		}
	} else {
		//bufio.NewWriter(conn).WriteString("NO\n")
		fmt.Fprintf(conn, "NO\n")
	}
	fmt.Println("No Errors")
	return conn.Close()
}

/**
 * Sends a file to a peer
 * @param string fileName - The name of the file to send to the peer
 * @param net.Conn conn - The socket over which we will send the file
 */
func sendFile(fileName string, conn net.Conn) error {
	fileName = "../client/" + fileName // Need to change this - move files to a different directory
	fmt.Println(fileName)

	fileToSend, err := os.Open(fileName)
	if err != nil {
		return err
	}

	n, err := io.Copy(conn, fileToSend)
	if err != nil {
		return err
	}

	fmt.Println(n, "bytes were sent")

	return fileToSend.Close()

}
