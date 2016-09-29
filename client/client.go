// Package client is responsible for receiving data and maintaining / manipulating a lynk's
// directory.
// @author: Michael Bruce
// @author: Max Kernchen
// @verison: 2/17/2016
package client

import (
	"bufio"
	"bytes"
	"capstone/lynxutil"
	"capstone/mycrypt"
	"compress/gzip"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/textproto"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

// An array of the lynks found from parsing the lynks.txt file
var lynks []lynxutil.Lynk

// A special symbol we use to denote the end of 1 entry in the metainfo file
const endOfEntry = ":#!"

// The array index of our metainfo values
const metaValueIndex = 1

// Function that deletes an entry from a lynk's files array.
// @param string nameToDelete - This is the name of the file we want to delete
// @param string lynkName - The lynk we want to delete it from
func deleteFile(nameToDelete, lynkName string) {
	lynk := lynxutil.GetLynk(lynks, lynkName)

	i := 0
	for i < len(lynk.Files) {
		if nameToDelete == lynk.Files[i].Name {
			lynk.Files = append(lynk.Files[:i], lynk.Files[i+1:]...)
		}
		i++
	}

	fmt.Println(lynk.Files)
}

// Deletes the current meta.info and replaces it with a new version that
// accurately reflects the array of Files after they have been modified.
// @return error - An error can be produced when issues arise from trying to create
// or remove the meta file - otherwise error will be nil.
func updateMetainfo(metaPath string) error {
	parseMetainfo(metaPath)
	lynkName := GetLynkName(metaPath)
	lynk := lynxutil.GetLynk(lynks, lynkName)

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

	newMetainfo.WriteString("announce:::" + lynk.Tracker + "\n") // Write tracker IP
	newMetainfo.WriteString("lynkName:::" + lynk.Name + "\n")
	newMetainfo.WriteString("owner:::" + lynk.Owner + "\n")
	i := 0
	for i < len(lynk.Files) {
		newMetainfo.WriteString("length:::" + strconv.Itoa(lynk.Files[i].Length) + "\n") // str conv
		newMetainfo.WriteString("path:::" + lynk.Files[i].Path + "\n")
		newMetainfo.WriteString("name:::" + lynk.Files[i].Name + "\n")
		newMetainfo.WriteString("chunkLength:::" + strconv.Itoa(lynk.Files[i].ChunkLength) + "\n")
		newMetainfo.WriteString("chunks:::" + lynk.Files[i].Chunks + "\n")
		newMetainfo.WriteString(endOfEntry + "\n")
		i++
	}

	return newMetainfo.Close()
}

// Parses the information in meta.info file and places each entry into a File
// struct and appends that struct to the array of structs
// @param string metaPath - The path to the metainfo file
// @return error - An error can be produced when issues arise from trying to access
// the meta file or from an invalid meta file type - otherwise error will be nil.
func parseMetainfo(metaPath string) error {
	lynk := lynxutil.GetLynk(lynks, GetLynkName(metaPath))
	if lynk == nil {
		return errors.New("Lynk Not Found")
	}
	lynk.Files = nil // Resets files array

	metaFile, err := os.Open(metaPath)
	if err != nil {
		return err
	} else if !strings.Contains(metaPath, "meta.info") {
		return errors.New("Invalid File Type")
	}

	scanner := bufio.NewScanner(metaFile)
	tempFile := lynxutil.File{}
	for scanner.Scan() { // Scan each line
		split := strings.Split(strings.TrimSpace(scanner.Text()), ":::")
		if split[0] == "announce" {
			lynk.Tracker = split[metaValueIndex]
		} else if split[0] == "owner" {
			lynk.Owner = split[metaValueIndex]
		} else if split[0] == "lynkName" {
			lynk.Name = split[metaValueIndex]
		} else if split[0] == "chunkLength" {
			tempFile.ChunkLength, _ = strconv.Atoi(split[metaValueIndex])
		} else if split[0] == "length" {
			tempFile.Length, _ = strconv.Atoi(split[metaValueIndex])
		} else if split[0] == "path" {
			tempFile.Path = split[metaValueIndex]
		} else if split[0] == "name" {
			tempFile.Name = split[metaValueIndex]
		} else if split[0] == "chunks" {
			tempFile.Chunks = split[metaValueIndex]
		} else if split[0] == endOfEntry {
			lynk.Files = append(lynk.Files, tempFile) // Append the current file to the file array
			tempFile = lynxutil.File{}                // Empty the current file
		}
	}
	return metaFile.Close()
}

// Adds a file to the meta.info by parsing that file's information
// @param string addPath - the path of the file to be added
// @param string metaPath - the path of the metainfo file - must be full path from root.
// @return error - An error can be produced when issues arise from trying to access
// the meta file or if the file to be added already exists in the meta file - otherwise
// error will be nil.
func addToMetainfo(addPath, metaPath string) error {
	metaFile, err := os.OpenFile(metaPath, os.O_APPEND|os.O_WRONLY, 0644) // Opens for appending
	if err != nil {
		fmt.Println(err)
		return err
	}

	addStat, err := os.Stat(addPath)
	if err != nil {
		fmt.Println(err)
		return err
	}

	parseMetainfo(metaPath)
	lynkName := GetLynkName(metaPath)
	lynk := lynxutil.GetLynk(lynks, lynkName)

	i := 0
	for i < len(lynk.Files) {
		if lynk.Files[i].Name == addStat.Name() {
			return errors.New("Can't Add Duplicates To Metainfo")
		}
		i++
	}

	lengthStr := strconv.FormatInt(addStat.Size(), 10) // Convert int64 to string
	metaFile.WriteString("length:::" + lengthStr + "\n")

	tempPath, err := filepath.Abs(addPath) // Find the path of the current file
	if err != nil {
		return err
	}

	// Write to metainfo file using ::: to separate keys and values
	metaFile.WriteString("path:::" + tempPath + "\n")
	metaFile.WriteString("name:::" + addStat.Name() + "\n")
	metaFile.WriteString("chunkLength:::-1\n")
	metaFile.WriteString("chunks:::chunking not currently implemented\n")
	metaFile.WriteString(endOfEntry + "\n")
	return metaFile.Close()
}

// HaveFile - Checks to see if we have the passed in file.
// @param string filePath - The name of the file to check for - This includes the lynk name.
// E.G. - 'Cool_Lynk/coolFile.txt'
// @return bool - A boolean indicating whether or not we have a file in our
// files array.
func HaveFile(filePath string) bool {
	have := false

	lynkInfo := strings.Split(filePath, "/")
	if len(lynkInfo) != 2 {
		fmt.Println(filePath + " is an invalid filepath")
		return have
	}

	lynkName := lynkInfo[0]
	fileName := lynkInfo[1]
	metaPath := lynxutil.HomePath + lynkName + "/meta.info"
	parseMetainfo(metaPath)
	lynk := lynxutil.GetLynk(lynks, lynkName)

	i := 0
	for i < len(lynk.Files) && !have {
		if lynk.Files[i].Name == fileName {
			have = true
		}
		i++
	}

	return have
}

// GetTracker - Simply returns the tracker associated with the passed in Lynk
// @param string metaPath - The meta.info path associated with the lynk we're interested in
// @return string - A string representing the tracker's IP address.
func GetTracker(metaPath string) string {
	parseMetainfo(metaPath)
	lynkName := GetLynkName(metaPath)
	lynk := lynxutil.GetLynk(lynks, lynkName)
	return lynk.Tracker
}

// Gets a file from the peer(s)
// @param string fileName - The name of the file to find in the peers
// @param string metaPath - The meta.info path associated with the lynk we're interested in
// @return error - An error can be produced if there are connection issues,
// problems creating or writing to the file, or from not being able to get there
// desired file - otherwise error will be nil.
func getFile(fileName, metaPath string) error {
	// Will parseMetainfo file and then ask tracker for list of peers
	parseMetainfo(metaPath)
	lynkName := GetLynkName(metaPath)
	lynk := lynxutil.GetLynk(lynks, lynkName)
	//fmt.Println("Asking For File From: " + metaPath)
	askTrackerForPeers(lynkName)
	//fmt.Println(lynk.Peers)

	i := 0
	gotFile := false
	for i < len(lynk.Peers) && !gotFile {
		conn, err := net.Dial("tcp", lynk.Peers[i].IP+":"+lynk.Peers[i].Port)
		// We don't want to return on err because we might be able to connect to next peer.
		if err == nil {
			gotFile = askForFile(lynkName, fileName, conn)
		}
		//fmt.Println(i)
		i++
	}

	if gotFile {
		return nil
	}

	return errors.New("Did not receive file") // If we got here - we didn't have the file.
}

// The function responsible for actually asking for a file from a peer
// @param string lynkName - The name of the lynk we're asking about
// @param string fileName - The name of the file to find in the peers
// @param net.Conn conn - The connection to the peer
// @return bool - True or false is returned based on whether or not we successfully received a file
func askForFile(lynkName, fileName string, conn net.Conn) bool {
	fmt.Fprintf(conn, "Do_You_Have_FileName:"+lynkName+"/"+fileName+"\n")

	reply, err := bufio.NewReader(conn).ReadString('\n') // Waits for a String ending in newline
	reply = strings.TrimSpace(reply)
	gotFile := false

	// Has file and no errors
	if reply != "NO" && err == nil {
		file, err := os.Create(lynxutil.HomePath + lynkName + "/" + fileName)
		if err != nil {
			return gotFile
		}
		defer file.Close()

		bufIn, err := ioutil.ReadAll(conn)
		if err != nil {
			return gotFile
		}

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
		file.Write(bufOut)
		gotFile = true
	}

	return gotFile
}

// Asks the tracker for a list of peers and then places them into a lynk's peers array
// @param string lynkName - The name of the lynk we're interested in
func askTrackerForPeers(lynkName string) error {
	lynk := lynxutil.GetLynk(lynks, lynkName)
	// Connects to tracker
	conn, err := net.Dial("tcp", lynk.Tracker)
	if err != nil {
		i := 0
		for i < len(lynk.Peers) && err != nil {
			pConn, _ := net.Dial("tcp", lynk.Peers[i].IP+":"+lynk.Peers[i].Port)
			fmt.Fprintf(pConn, "Tracker_Request:"+lynkName+"/\n")
			reply := ""
			reply, err = bufio.NewReader(pConn).ReadString('\n') // Waits for a String ending in newline
			reply = strings.TrimSpace(reply)

			conn, err = net.Dial("tcp", reply)
			i++
		}

		// We could not connect to the tracker
		if err != nil {
			return err
		}
	}

	// Gives IP and ServerPort So It Can Be Added To swarm.info
	fmt.Fprintf(conn, "Swarm_Request:"+lynxutil.GetIP()+":"+lynxutil.ServerPort+":"+lynkName+"\n")
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)

	reply, err := tp.ReadLine()
	//fmt.Println(reply)

	// Tracker will close connection when finished - which will break us out of this loop
	for err == nil {
		peerArray := strings.Split(reply, ":::")
		tmpPeer := lynxutil.Peer{IP: peerArray[0], Port: peerArray[1]}
		if !contains(lynk.Peers, tmpPeer) {
			lynk.Peers = append(lynk.Peers, tmpPeer)
		}
		reply, err = tp.ReadLine()
	}

	return nil // Did not have an error if we reached this point
}

// Simple helper method that checks peers array for specific peer.
// @param s []peers - The peers array
// @param e Peer - The peer we are checking for
func contains(s []lynxutil.Peer, e lynxutil.Peer) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// CreateMeta - This function creates a new metainfo file for use within the GUI server
// @param name string - The name of the new lynk
func CreateMeta(name string) error {
	tDir, err := os.Stat(lynxutil.HomePath + name) // Checks to see if the directory exists
	if err != nil || !tDir.IsDir() {
		fmt.Println("ERROR!")
		return errors.New("Directory " + name + "does not exist in the Lynx directory.")
	}

	metaFile, err := os.Create(lynxutil.HomePath + name + "/meta.info")
	if err != nil {
		fmt.Println(err)
		return err
	}

	currentUser, _ := user.Current()
	metaFile.WriteString("announce:::" + lynxutil.GetIP() + ":" + lynxutil.TrackerPort + "\n")
	metaFile.WriteString("lynkName:::" + name + "\n")
	metaFile.WriteString("owner:::" + currentUser.Name + "\n")

	addLynk(name, currentUser.Name)
	filepath.Walk(lynxutil.HomePath+name, visitFiles)

	parseMetainfo(lynxutil.HomePath + name + "/meta.info")

	return nil // Everything was fine if we reached this point
}

// Function which visits each file within a directory
// @param path string - the path where the root directory is located
// @param file os.FileInfo - each file within the root or inner directories
// @param err error - any error we way encoutner along the way
// @return error - An error can produced if we encounter an invalid file.
func visitFiles(path string, file os.FileInfo, err error) error {
	// Don't add directories, trackers, or a meta.info file to the new meta.info
	if !file.IsDir() && !strings.Contains(path, "_Tracker") && file.Name() != "meta.info" {
		//fmt.Println(file.Name())
		slashes := strings.Replace(path, "\\", "/", -1)
		//fmt.Println(slashes)
		tmpStr := strings.TrimPrefix(slashes, lynxutil.HomePath)
		tmpArr := strings.Split(tmpStr, "/")
		addToMetainfo(path, lynxutil.HomePath+tmpArr[0]+"/meta.info")
	}

	return nil
}

// Function which adds a lynk to list of lynks and also will added it to lynks.txt file as well
// @param name string - the name of the lynk
// @param owner string - the owner of the lynk
// @return error - An error can be produced if the lynks.txt file cannot be opened
func addLynk(name, owner string) error {

	lynkFile, err := os.OpenFile(lynxutil.HomePath+"lynks.txt", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		// create file if not real
	}

	i := 0
	for i < len(lynks) {
		// Will have to validate directory names
		if lynks[i].Name+lynks[i].Owner == name+owner {
			return errors.New("Can't Add Duplicate Lynk")
		}
		i++
	}

	lynkFile.WriteString(name + ":::Unsynced:::" + owner + "\n")

	ParseLynks(lynxutil.HomePath + "lynks.txt")
	genLynks()

	// fmt.Println(lynks)
	return lynkFile.Close()
}

// ParseLynks - Parses the information in lynks file and places each entry into a the lynks array.
// @param string lynksFilePath - The path to the lynks.txt file
// @return error - An error can be produced when issues arise from trying to access
// the lynks.txt file.
func ParseLynks(lynksFilePath string) error {
	lynks = nil // Resets files array

	lynksFile, err := os.Open(lynksFilePath)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(lynksFile)
	tempLynk := lynxutil.Lynk{}

	// Scan each line
	for scanner.Scan() {

		line := strings.TrimSpace(scanner.Text()) // Trim helps with errors in \n
		split := strings.Split(line, ":::")
		tempLynk.Name = split[0]
		tempLynk.Synced = split[1]
		tempLynk.Owner = split[2]

		lynks = append(lynks, tempLynk) // Append the current file to the file array
		tempLynk = lynxutil.Lynk{}      // Empty the current file
	}

	return lynksFile.Close()
}

// DeleteLynk - This function deletes a Lynk based upon its name from the list of lynks
// @param nameToDelete string - the lynk we want to remove
func DeleteLynk(nameToDelete string, deleteLocal bool) {
	i := 0
	for i < len(lynks) {
		if nameToDelete == lynks[i].Name {
			// Removes this peer from swarm.info file
			conn, _ := net.Dial("tcp", lynks[i].Tracker)
			fmt.Fprintf(conn, "Disconnect:"+lynxutil.GetIP()+":"+lynks[i].Name+"\n")

			lynks = append(lynks[:i], lynks[i+1:]...)
		}
		i++
	}
	updateLynksFile()

	if deleteLocal {
		os.RemoveAll(lynxutil.HomePath + nameToDelete)
	}
}

// Function which removes the lynks.txt file and creates a new one based on the current lynks array
// @returns error - will produce an error if we cannot open the lynks.txt file.
func updateLynksFile() error {
	newLynks, err := os.Create(lynxutil.HomePath + "lynks.txt")
	if err != nil {
		fmt.Println(err)
		return err
	}

	i := 0
	for i < len(lynks) {
		newLynks.WriteString(lynks[i].Name + ":::" + lynks[i].Synced + ":::" +
			lynks[i].Owner + "\n")

		i++
	}

	return newLynks.Close()
}

// JoinLynk - Function which will allow a user to join an existing link by way of its meta.info file
// @param metaPath string - the path to the meta.info file which will be used to find the
// information about the lynk
func JoinLynk(metaPath string) error {
	metaFile, err := os.Open(metaPath)
	if err != nil {
		return err
	}
	lynkName := ""
	owner := ""
	scanner := bufio.NewScanner(metaFile)
	tempPeer := lynxutil.Peer{}

	// Scan each line
	for scanner.Scan() {

		line := strings.TrimSpace(scanner.Text()) // Trim helps with errors in \n
		split := strings.Split(line, ":::")

		if split[0] == "announce" {
			tempPeer.IP = split[metaValueIndex]
		} else if split[0] == "port" {
			tempPeer.Port = split[metaValueIndex]
		} else if split[0] == "lynkName" {
			lynkName = split[metaValueIndex]
		} else if split[0] == "owner" {
			owner = split[metaValueIndex]
		} else {
			return errors.New("Invalid Meta")
		}

	}

	createJoin(lynkName, metaPath)
	addLynk(lynkName, owner)

	return UpdateLynk(lynkName) // Gets all of the files for the lynk over the network
}

// UpdateLynk - Function which will update the files of a Lynk with the current versions.
// @param lynkName string - the name of the Lynk we want to update
func UpdateLynk(lynkName string) error {
	// We actually get the files we need over the network.
	lynk := lynxutil.GetLynk(lynks, lynkName)
	var err error // Creates nil error
	for _, file := range lynk.Files {
		err = getFile(file.Name, lynxutil.HomePath+lynkName+"/meta.info")
		// If we fail to get the file the first time, we attempt again.
		if err.Error() == "Did not receive file" {
			for i := 0; i < lynxutil.ReconnAttempts; i++ {
				err = getFile(file.Name, lynxutil.HomePath+lynkName+"/meta.info")
			}
		}
	}

	return err
}

// Function which creates the directory for a newly joined lynk.
// @params name string - the name of the new lynk
// @params oldMetaPath string - the name of the metaPath we are using to create our new metaPath
func createJoin(name, oldMetaPath string) error {
	tDir, err := os.Stat(lynxutil.HomePath + name)
	// Checks to see if the directory exists so we don't overwrite
	if err == nil && tDir.IsDir() {
		fmt.Println("ERROR!" + tDir.Name() + " Already Exists")
		return errors.New("Directory " + name + " Already Exists")
	}

	newLynkDir := lynxutil.HomePath + name
	os.Mkdir(newLynkDir, 0755)

	err = lynxutil.FileCopy(oldMetaPath, newLynkDir+"/meta.info")
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil // Everything was fine if we reached this point
}

// Function init runs as soon as this class is imported and allows us to create an array of Lynks.
func init() {
	ParseLynks(lynxutil.HomePath + "lynks.txt")
	genLynks()
	//fmt.Println(lynks)
	//lynk := lynxutil.GetLynk(lynks, "Tests")
	//fmt.Println(lynk.Files)
}

// Helper function that generates all the data for our lynks array by parsing each corresponding
// meta.info file.
func genLynks() {
	i := 0
	for i < len(lynks) {
		parseMetainfo(lynxutil.HomePath + lynks[i].Name + "/meta.info")
		i++
	}
}

// GetLynkName - Helper function that returns our Lynk name if we pass in its metaPath.
// @param string metaPath - The meta.info path associated with the lynk we're interested in
// @returns string - The lynk name
func GetLynkName(metaPath string) string {
	return strings.TrimSuffix(strings.TrimPrefix(metaPath, lynxutil.HomePath), "/meta.info")
}

// GetLynks - Simply returns our current lynks array.
// @returns - The current lynks array.
func GetLynks() []lynxutil.Lynk {
	return lynks
}

// GetLynksLen - Returns the size of our lynks array.
// @returns - The current size of our lynks array.
func GetLynksLen() int {
	return len(lynks)
}

// PopulateFilesAndSize - Fills Our Lynks Array With File And Size Information
func PopulateFilesAndSize() {
	i := 0
	for i < len(lynks) {
		files := lynks[i].Files
		j := 0
		if len(lynks[i].FileNames) == 0 && len(lynks[i].FileSize) == 0 {
			for j < len(files) {

				lynks[i].FileNames = append(lynks[i].FileNames, files[j].Name)
				lynks[i].FileSize = append(lynks[i].FileSize, files[j].Length)
				j++
			}
		}
		i++
	}

}

// IsDownloading - Returns whether or not the client associated the specified lynk is downloading
// @param lynkName
// @returns - Returns whether or not the client associated the specified lynk is downloading
func IsDownloading(lynkName string) bool {
	lynk := lynxutil.GetLynk(lynks, lynkName)

	return lynk.DLing
}

// StopDownload - Sets a boolean to stop the lynk from downloading
func StopDownload(lynkName string) {
	lynk := lynxutil.GetLynk(lynks, lynkName)
	lynk.DLing = false
}
