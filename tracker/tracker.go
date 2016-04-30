// Package tracker - This package is responsible for communicating with and getting peers connected
// in the Lynx network.
// @author: Michael Bruce
// @author: Max Kernchen
// @verison: 2/17/2016
package tracker

import (
	"bufio"
	"bytes"
	"capstone/lynxutil"
	"capstone/mycrypt"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// An array of tLynks this tracker presides over
var tLynks []lynxutil.Lynk

// Function that deletes an entry from a lynk's peers array.
// @param Peer peerToDelete - This is the peer struct we want to delete
// @param string lynkName - The lynk we want to delete it from
func deletePeer(peerToDelete lynxutil.Peer, lynkName string) {
	lynk := lynxutil.GetLynk(tLynks, lynkName)

	i := 0
	for i < len(lynk.Peers) {
		if peerToDelete.IP == lynk.Peers[i].IP && peerToDelete.Port == lynk.Peers[i].Port {
			lynk.Peers = append(lynk.Peers[:i], lynk.Peers[i+1:]...)
		}
		i++
	}
}

// Deletes the current swarm.info and replaces it with a new version that
// accurately reflects the array of Peers after they have been modified
// @return error - An error can be produced when issues arise from trying to create
// or remove the swarm file - otherwise error will be nil.
func updateSwarminfo(swarmPath string) error {
	parseSwarminfo(swarmPath)

	err := os.Remove(swarmPath)
	if err != nil {
		fmt.Println(err)
		return err
	}

	newSwarmInfo, err := os.Create(swarmPath)
	if err != nil {
		fmt.Println(err)
		return err
	}

	lynkName := getTLynkName(swarmPath)
	lynk := lynxutil.GetLynk(tLynks, lynkName)

	i := 0
	for i < len(lynk.Peers) {
		newSwarmInfo.WriteString(lynk.Peers[i].IP + ":::" + lynk.Peers[i].Port + "\n")
		i++
	}

	return newSwarmInfo.Close()
}

// Parses the information in swarm.info file and places each entry into a Peer
// struct and appends that struct to the array of peers
// @param string swarmPath - The path to the swarminfo file
// @return error - An error can be produced when issues arise from trying to access
// the swarm file or from an invalid swarm file type - otherwise error will be nil.
func parseSwarminfo(swarmPath string) error {
	lynkName := getTLynkName(swarmPath)
	lynk := lynxutil.GetLynk(tLynks, lynkName)

	//fmt.Println(lynk)
	lynk.Peers = nil // Resets peers array

	swarmFile, err := os.Open(swarmPath)
	if err != nil {
		return err
	} else if !strings.Contains(swarmPath, "swarm.info") {
		return errors.New("Invalid File Type")
	}

	scanner := bufio.NewScanner(swarmFile)
	tempPeer := lynxutil.Peer{}

	// Scan each line
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text()) // Trim helps with errors in \n
		split := strings.Split(line, ":::")

		tempPeer.IP = split[0]
		tempPeer.Port = split[1]
		lynk.Peers = append(lynk.Peers, tempPeer)
	}

	//fmt.Println(lynk.Peers)
	return swarmFile.Close()
}

// Adds a peer to the swarm.info file
// @param string addPath - the path of the file to be added
// @param string swarmPath - the path of the swarminfo file
// @return error - An error can be produced when issues arise from trying to access
// the swarm file or if the file to be added already exists in the swarm file - otherwise
// error will be nil.
func addToSwarminfo(addPeer lynxutil.Peer, swarmPath string) error {
	swarmFile, err := os.OpenFile(swarmPath, os.O_APPEND|os.O_WRONLY, 0644) // Opens for appending
	if err != nil {
		return err
	}

	lynkName := getTLynkName(swarmPath)
	lynk := lynxutil.GetLynk(tLynks, lynkName)
	if lynk == nil {
		tLynks = append(tLynks, lynxutil.Lynk{Name: lynkName})
		lynk = lynxutil.GetLynk(tLynks, lynkName)
	}

	parseSwarminfo(swarmPath)

	i := 0
	for i < len(lynk.Peers) {
		if lynk.Peers[i].IP == addPeer.IP && lynk.Peers[i].Port == addPeer.Port {
			return errors.New("Can't Add Duplicates To Swarminfo")
		}
		i++
	}

	// Write to swarminfo file using ::: to IP and Port
	swarmFile.WriteString(addPeer.IP + ":::" + addPeer.Port + "\n")

	return swarmFile.Close()
}

// Listen - Creates a welcomeSocket that listens for TCP connections - once someone connects a
// goroutine is spawned to handle the request
func Listen() {
	fmt.Println("Starting Tracker on Port " + lynxutil.TrackerPort)

	welcomeSocket, wErr := net.Listen("tcp", ":"+lynxutil.TrackerPort)

	if wErr != nil {
		fmt.Println("Could Not Create Server Welcome Socket - Aborting.")
		os.Exit(lynxutil.SockErr) // Cannot recover from not being able to generate welcomeSocket
	}

	var cErr error
	for cErr == nil {
		conn, cErr := welcomeSocket.Accept()
		if cErr != nil {
			// handle error
			continue // To avoid calling handleRequest
		}
		go handleRequest(conn)
	}
}

// Handles a request / push sent by a client, can either be a swarm or meta request or a push
// of an updated meta.info file - also adds the requesting client to the swarm.info file
// @param net.Conn conn - The socket which the client is asking on
// @return error - An error can be produced when trying to send a file or if there is incorrect
// syntax in the request - otherwise error will be nil.
func handleRequest(conn net.Conn) error {
	request, err := bufio.NewReader(conn).ReadString('\n') // Waits for a String ending in newline
	if err != nil {
		return err
	}

	request = strings.TrimSpace(request)
	fmt.Println("REQUEST:" + request)
	// Makes sure we are dealing with a request and not a push
	if !strings.Contains(request, "Meta_Push:") {
		// Client syntax for request is "X_Request:<IP>:<Port>:<LynkName>\n"
		// So tmpArr[0] - X_Request | tmpArr[1] - <IP> | tmpArr[2] - <Port> | tmpArr[3] - <LynkName>
		tmpArr := strings.Split(request, ":")
		if len(tmpArr) != 4 {
			conn.Close()
			return errors.New("Invalid Request Syntax")
		}

		fileToSend := ""
		// Checks to see if we are dealing w/ a Swarm or Meta Request
		swarmPath := lynxutil.HomePath + tmpArr[3] + "/" + tmpArr[3] + "_Tracker/" + "swarm.info"
		if tmpArr[0] == "Swarm_Request" {
			fileToSend = swarmPath
		} else if tmpArr[0] == "Meta_Request" {
			fileToSend = lynxutil.HomePath + tmpArr[3] + "/meta.info"
		} else {
			conn.Close()
			return errors.New("Invalid Request Syntax")
		}

		tmpPeer := lynxutil.Peer{IP: strings.TrimSpace(tmpArr[1]), Port: strings.TrimSpace(tmpArr[2])}
		fmt.Println("New Peer: ", tmpPeer)

		err = sendFile(fileToSend, conn) // Sending The file
		if err != nil {
			conn.Close()
			return err
		}

		addToSwarminfo(tmpPeer, swarmPath) // So we only add peer to swarmlist on success
		fmt.Println("No Errors")
	} else { // We are receiving a meta.info file
		// Client syntax for push is "Meta_Push:<LynkName>\n"
		// So tmpArr[0] - Meta_Push | tmpArr[1] - <LynkName>
		tmpArr := strings.Split(request, ":")
		metaPath := lynxutil.HomePath + tmpArr[3] + "/" + tmpArr[3] + "_Tracker/" + "meta.info"
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
		key := []byte("abcdefghijklmnopqrstuvwxyz123456")
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
	}

	return conn.Close()
}

// Sends a file to a peer.
// @param string fileName - The name of the file to send to the peer
// @param net.Conn conn - The socket over which we will send the file
// @return error - An error can be produced when trying open a file or write over
// the network - otherwise error will be nil.
func sendFile(fileName string, conn net.Conn) error {
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

// CreateSwarm - Creates a new swarm.info upon clicking of create button in gui
// @param string name - the name of the lynk
func CreateSwarm(name string) {
	p1 := lynxutil.Peer{IP: "", Port: "8080"}

	currentuser, err := user.Current()
	trackerDir := currentuser.HomeDir + "/Lynx/" + name + "/" + name + "_Tracker"
	os.Mkdir(trackerDir, 0755)

	_, err = os.Create(trackerDir + "/swarm.info")
	if err != nil {
		fmt.Println(err)
	}

	p1.IP = lynxutil.GetIP()
	addToSwarminfo(p1, trackerDir+"/swarm.info")

	lynxutil.FileCopy(currentuser.HomeDir+"/Lynx/"+name+"/meta.info", trackerDir+"/meta.info")
}

// Function which visits each tracker directory within the Lynx root
// @param path: the path where the root directory is located
// @param file: each file within the root or inner directories
// @param err: any error we way encoutner along the way
func visitTrackers(path string, file os.FileInfo, err error) error {
	base := strings.TrimPrefix(path, lynxutil.HomePath)
	split := strings.Split(base, "/")

	// Checks that there is directory beneath another directory and has _tracker
	if file.IsDir() && len(split) == 2 && strings.Contains(split[1], "_Tracker") {
		fmt.Println(file.Name())
		lynkName := strings.TrimSuffix(file.Name(), "_Tracker")
		tLynks = append(tLynks, lynxutil.Lynk{Name: lynkName})
	}

	return nil
}

// Function init runs before main and allows us to setup our tracker properly.
func init() {
	filepath.Walk(lynxutil.HomePath, visitTrackers)
}

// Helper function that returns a lynk's name give it's swarm.info filepath.
// @param string metaPath - The meta.info path associated with the lynk we're interested in
// @returns string - The lynk name
func getTLynkName(swarmPath string) string {
	tmpStr := strings.TrimSuffix(strings.TrimPrefix(swarmPath, lynxutil.HomePath), "/swarm.info")
	split := strings.Split(tmpStr, "/")
	fmt.Println(split[0])
	return split[0]
}
