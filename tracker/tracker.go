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
	"net/textproto"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// An array of tLynks this tracker presides over
var tLynks []lynxutil.Lynk

// Function that deletes an entry from a lynk's peers array and the swarm.info file.
// @param string peerToDelete - This is the peer struct we want to delete - uses the IP address
// @param string lynkName - The lynk we want to delete it from
func deletePeer(peerToDelete, lynkName string) {
	lynk := lynxutil.GetLynk(tLynks, lynkName)

	i := 0
	for i < len(lynk.Peers) {
		if peerToDelete == lynk.Peers[i].IP {
			lynk.Peers = append(lynk.Peers[:i], lynk.Peers[i+1:]...)
		}
		i++
	}

	swarmPath := lynxutil.HomePath + lynkName + "/" + lynkName + "_Tracker/" + "swarm.info"

	os.Remove(swarmPath)
	newSwarmInfo, _ := os.Create(swarmPath)

	i = 0
	for i < len(lynk.Peers) {
		newSwarmInfo.WriteString(lynk.Peers[i].IP + ":::" + lynk.Peers[i].Port + "\n")
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

// Listen - Calls lynxutil to create a welcomeSocket that listens for TCP connections - once
// someone connects a goroutine is spawned to handle the request
func Listen() {
	lynxutil.Listen(handleRequest, lynxutil.TrackerPort)
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

	if strings.Contains(request, "Meta_Push:") { // We are receiving a meta.info file
		handlePush(request, conn)
		notifyPeers(request)
	} else if strings.Contains(request, "Disconnect:") {
		// tmpArr[0] - Disconnect | tmpArr[1] - <IP> | tmpArr[2] - <LynkName>
		tmpArr := strings.Split(request, ":")
		deletePeer(tmpArr[1], tmpArr[2])
	} else { // We are receiving a pull request
		handlePull(request, conn)
	}
	return conn.Close()
}

// Helper function for handleRequest - handles the case where a client is requesting a meta.info
// or swarm.info file.
// @param net.Conn conn - The socket which the client is asking on
// @param string request - The request sent to tracker
// @return error - An error can be produced when trying to send a file or if there is incorrect
// syntax in the request - otherwise error will be nil.
func handlePull(request string, conn net.Conn) error {
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
	err := sendFile(fileToSend, conn) // Sending The file
	if err != nil {
		conn.Close()
		return err
	}

	addToSwarminfo(tmpPeer, swarmPath) // So we only add peer to swarmlist on success
	return nil                         // No errors if we reached this point
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
	metaPath := lynxutil.HomePath + tmpArr[1] + "/" + tmpArr[1] + "_Tracker/" + "meta.info"
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

	if err != nil {
		log.Fatal("Tracker:", err)
	}

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

// Helper function for handleRequest - handles the case where we update peers after receiving a new
// meta.info file
// @param string request - The request sent to tracker
// @return error - An error can be produced when trying to send a file or if there is incorrect
// syntax in the request - otherwise error will be nil.
func notifyPeers(request string) error {
	// So tmpArr[0] - Meta_Push | tmpArr[1] - <LynkName>
	tmpArr := strings.Split(request, ":")
	metaPath := lynxutil.HomePath + tmpArr[1] + "/" + tmpArr[1] + "_Tracker/" + "meta.info"
	swarmPath := lynxutil.HomePath + tmpArr[1] + "/" + tmpArr[1] + "_Tracker/" + "swarm.info"

	// Opens the swarm file for the specific Lynk and notifies all of the listed peers
	swarmFile, _ := os.Open(swarmPath)
	r := bufio.NewReader(swarmFile)
	tp := textproto.NewReader(r)
	line, e := tp.ReadLine()
	for e == nil {
		peerArray := strings.Split(line, ":::")
		// [0] is IP / [1 ]is Port
		pConn, _ := net.Dial("tcp", peerArray[0]+":"+peerArray[1])
		fmt.Fprintf(pConn, "Meta_Push:"+tmpArr[1]+"\n")

		fBytes, err := ioutil.ReadFile(metaPath)
		fmt.Println("fBytes: ", string(fBytes))

		// Begin Compression
		var b bytes.Buffer
		gz := gzip.NewWriter(&b)
		gz.Write(fBytes)
		gz.Close()
		// End Compression

		// Begin Encryption
		var cipherFile []byte
		publicKey := lynxutil.Peer{IP: pConn.LocalAddr().String()}.Key + lynxutil.PrivateKey
		key := []byte(publicKey)
		if cipherFile, err = mycrypt.Encrypt(key, b.Bytes()); err != nil {
			return err
		}
		// End Encryption

		n, err := pConn.Write(cipherFile)
		if err != nil {
			fmt.Println("CONNECTION ERROR:", err)
			return err
		}

		fmt.Println("TRACKER SENT", n, "BYTES TO PEER")

		pConn.Close()
		line, e = tp.ReadLine()
	}

	return nil // No errors if we reached this point
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
	path = strings.Replace(path, "\\", "/", -1) // Switches windows \ to unix /
	base := strings.TrimPrefix(path, lynxutil.HomePath)
	split := strings.Split(base, "/")

	// Checks that there is directory beneath another directory and has _tracker
	if file.IsDir() && len(split) == 2 && strings.Contains(split[1], "_Tracker") {
		fmt.Println(file.Name())
		lynkName := strings.TrimSuffix(file.Name(), "_Tracker")
		tLynks = append(tLynks, lynxutil.Lynk{Name: lynkName})
		// Need to populate Peers here.
	}

	return nil
}

// Function init runs before main and allows us to setup our tracker properly.
func init() {
	filepath.Walk(lynxutil.HomePath, visitTrackers)
}

// Helper function that returns a lynk's name give it's swarm.info filepath.
// @param string swarmPath - The swarm.info path associated with the lynk we're interested in
// @returns string - The lynk name
func getTLynkName(swarmPath string) string {
	tmpStr := strings.TrimSuffix(strings.TrimPrefix(swarmPath, lynxutil.HomePath), "/swarm.info")
	split := strings.Split(tmpStr, "/")
	//fmt.Println(split[0])
	return split[0]
}

// BroadcastNewIP - This function broadcasts a tracker's new IP address to all of its peers
// @param string swarmPath - The swarm.info path associated with the lynk we're interested in
func BroadcastNewIP(swarmPath string) {
	// Can update Meta here if needed
	lynkName := getTLynkName(swarmPath)
	lynk := lynxutil.GetLynk(tLynks, lynkName)

	i := 0
	for i < len(lynk.Peers) {
		fmt.Println(i)
		conn, err := net.Dial("tcp", lynk.Peers[i].IP+":"+lynk.Peers[i].Port)
		if err == nil {
			sendFile(lynxutil.HomePath+lynk.Name+"/meta.info", conn)
		}
		fmt.Println(lynk.Peers[i].IP)
		i++
	}
}

// PurgeOldIPs - This function tries to connect to every peer in the swarm.info file and removes
// them if unable to connect.
func PurgeOldIPs() {
	// Loops through all tracker lynks.
	for _, lynk := range tLynks {

		// Loops through all peers of a given lynk
		i := 0
		for i < len(lynk.Peers) {
			conn, err := net.Dial("tcp", lynk.Peers[i].IP+":"+lynk.Peers[i].Port)

			// If we cannot connect, remove the peer
			if err != nil {
				deletePeer(lynk.Peers[i].IP, lynk.Name)
			}
			conn.Close()
			i++
		}
	}
}

// TransferTracker - This function transfers the needed tracker files (swarm/meta.info) to the
// specified IP and then deletes the local copies of these files.
func TransferTracker(lynkName, owner, IP string) error {
	conn, _ := net.Dial("tcp", IP+":"+lynxutil.TrackerPort)

	// Sends the new peer the needed tracker files
	err := sendFile(lynxutil.HomePath+lynkName+"/"+lynkName+"_Tracker/swarm.info", conn)
	if err != nil {
		return err
	}

	err = sendFile(lynxutil.HomePath+lynkName+"/"+lynkName+"_Tracker/meta.info", conn)
	if err != nil {
		return err
	}

	// Removes the tracker directory from this computer
	os.RemoveAll(lynxutil.HomePath + lynkName + "/" + lynkName + "_Tracker/")

	return nil // No errors if we reach this point
}
