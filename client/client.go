/**
 *
 *	The client side of the Lynx application. It is responsible for receiving data and maintaining the lynx files.
 *
 *	 @author: Michael Bruce
 *	 @author: Max Kernchen
 *
 *	 @verison: 2/17/2016
 */

package client

import (
	"bufio"
	"bytes"
	"../mycrypt"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/textproto"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

/**	A struct which represents a Peer of the client */
type Peer struct {
	IP   string
	Port string
}

/** A struct which holds all the information about a specific Lynk. */
type Lynk struct {
	Name    string
	Owner   string
	Synced  string
	Tracker string
	Files   []File
	Peers   []Peer
	FileNames []string
	FileSize []int
}

/** A struct based which represents a File in a Lynk's directory. It is based
upon BitTorrent protocol dictionaries */
type File struct {
	length      int
	path        string // Might not need path
	name        string
	chunks      string
	chunkLength int
}

/** An array of the lynks found from parsing the lynks.txt file */
var lynks []Lynk

/** The location of the user's root directory */
var homePath string

/** An array of the files found from parsing the metainfo file */
//var files []File

/** The IP Address / Port of our tracker */
//var tracker string

/** An array of all the client's peers */
//var peers []Peer

/** A special symbol we use to denote the end of 1 entry in the metainfo file */
const END_OF_ENTRY = ":#!"

/** The array index of our metainfo values */
const META_VALUE_INDEX = 1

/**
 * Function that deletes an entry from our files array.
 * @param string nameToDelete - This is the name of the file we want to delete
 * @param string lynkName - The lynk we want to delete it from
 */
func deleteEntry(nameToDelete, lynkName string) {
	lynk := getLynk(lynks, lynkName)

	i := 0
	for i < len(lynk.Files) {
		if nameToDelete == lynk.Files[i].name {
			lynk.Files = append(lynk.Files[:i], lynk.Files[i+1:]...)
		}
		i++
	}

	fmt.Println(lynk.Files)
}

/**
 * Deletes the current meta.info and replaces it with a new version that
 * accurately reflects the array of Files after they have been modified
 * @return error - An error can be produced when issues arise from trying to create
 * or remove the meta file - otherwise error will be nil.
 */
func updateMetainfo(metaPath string) error {
	parseMetainfo(metaPath)
	lynkName := GetLynkName(metaPath)
	lynk := getLynk(lynks, lynkName)

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
		newMetainfo.WriteString("length:::" + strconv.Itoa(lynk.Files[i].length) + "\n") //convert to str
		newMetainfo.WriteString("path:::" + lynk.Files[i].path + "\n")
		newMetainfo.WriteString("name:::" + lynk.Files[i].name + "\n")
		newMetainfo.WriteString("chunkLength:::" + strconv.Itoa(lynk.Files[i].chunkLength) + "\n")
		newMetainfo.WriteString("chunks:::" + lynk.Files[i].chunks + "\n")
		newMetainfo.WriteString(END_OF_ENTRY + "\n")
		i++
	}

	return newMetainfo.Close()
}

/**
 * Parses the information in meta.info file and places each entry into a File
 * struct and appends that struct to the array of structs
 * @param string metaPath - The path to the metainfo file
 * @return error - An error can be produced when issues arise from trying to access
 * the meta file or from an invalid meta file type - otherwise error will be nil.
 */
func parseMetainfo(metaPath string) error {
	lynkName := GetLynkName(metaPath)
	//fmt.Println(metaPath)
	//fmt.Println(lynkName)

	lynk := getLynk(lynks, lynkName)
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
	tempFile := File{}

	// Scan each line
	for scanner.Scan() {

		line := strings.TrimSpace(scanner.Text()) // Trim helps with errors in \n
		split := strings.Split(line, ":::")

		if split[0] == "announce" {
			lynk.Tracker = split[META_VALUE_INDEX]
		} else if split[0] == "owner" {
			lynk.Owner = split[META_VALUE_INDEX]
		} else if split[0] == "lynkName" {
			lynk.Name = split[META_VALUE_INDEX]
		} else if split[0] == "chunkLength" {
			tempFile.chunkLength, _ = strconv.Atoi(split[META_VALUE_INDEX])
		} else if split[0] == "length" {
			tempFile.length, _ = strconv.Atoi(split[META_VALUE_INDEX])
		} else if strings.Contains(line, "path") {
			tempFile.path = split[META_VALUE_INDEX]
		} else if split[0] == "name" {
			tempFile.name = split[META_VALUE_INDEX]
		} else if strings.Contains(line, "chunks") {
			tempFile.chunks = split[META_VALUE_INDEX]
		} else if strings.Contains(line, END_OF_ENTRY) {
			lynk.Files = append(lynk.Files, tempFile) // Append the current file to the file array
			tempFile = File{}                         // Empty the current file
		}

	}

	//fmt.Println(lynk)
	return metaFile.Close()

	/////////////////////////////////////////////////
	/*files = nil // Resets files array

	metaFile, err = os.Open(metaPath)
	if err != nil {
		return err
		//} else if metaPath != "../resources/meta.info" {
		//	return errors.New("Invalid File Type")
	}

	scanner = bufio.NewScanner(metaFile)
	tempFile = File{}

	// Scan each line
	for scanner.Scan() {

		line := strings.TrimSpace(scanner.Text()) // Trim helps with errors in \n
		split := strings.Split(line, ":::")

		if split[0] == "announce" {
			tracker = split[META_VALUE_INDEX]
		} else if split[0] == "chunkLength" {
			tempFile.chunkLength, _ = strconv.Atoi(split[META_VALUE_INDEX])
		} else if split[0] == "length" {
			tempFile.length, _ = strconv.Atoi(split[META_VALUE_INDEX])
		} else if strings.Contains(line, "path") {
			tempFile.path = split[META_VALUE_INDEX]
		} else if strings.Contains(line, "name") {
			tempFile.name = split[META_VALUE_INDEX]
		} else if strings.Contains(line, "chunks") {
			tempFile.chunks = split[META_VALUE_INDEX]
		} else if strings.Contains(line, END_OF_ENTRY) {
			files = append(files, tempFile) // Append the current file to the file array
			tempFile = File{}               // Empty the current file
		}

	}

	return metaFile.Close()*/
}

/**
 * Adds a file to the meta.info by parsing that file's information
 * @param string addPath - the path of the file to be added
 * @param string metaPath - the path of the metainfo file - must be full path from root.
 * @return error - An error can be produced when issues arise from trying to access
 * the meta file or if the file to be added already exists in the meta file - otherwise
 * error will be nil.
 */
func addToMetainfo(addPath, metaPath string) error {
	fmt.Println("add to meta")
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
	fmt.Println(addStat.Name() + " this is addstat")
	parseMetainfo(metaPath)
	lynkName := GetLynkName(metaPath)
	lynk := getLynk(lynks, lynkName)

	i := 0
	fmt.Println(lynk.Files)
	for i < len(lynk.Files) {
		if lynk.Files[i].name == addStat.Name() {
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
	metaFile.WriteString(END_OF_ENTRY + "\n")
	return metaFile.Close()
}

/**
 * Copies a file from src to dst
 * @param string src - the file that will be copied
 * @param string dst - the destination of the file to be copied
 * @return error - An error can be produced when issues arise from trying to access,
 * create, and write from either the src or dst files - otherwise error will be nil.
 */
func FileCopy(src, dst string) error {
	in, err := os.Open(src) // Opens input
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst) // Opens output
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in) // Copies the file contents
	if err != nil {
		return err
	}

	return out.Close() // Checks for close error
}

/**
 * Function used to drive and test our other client functions
 */
func main() {

	/*fmt.Println("Hello World!")
	fmt.Println("Cool Beans!")
	err := fileCopy(os.Args[1], os.Args[2])

	if err != nil {
		fmt.Println("You suck")
	} else {
		fmt.Println(os.Args[1] + " copied to " + os.Args[2])
	}*/

	//parseMetainfo(os.Args[1])
	/*addToMetainfo("test.txt", "../resources/meta.info")
	addToMetainfo("test2.txt", "../resources/meta.info")
	addToMetainfo("file1.txt", "../resources/meta.info")
	parseMetainfo("../resources/meta.info")

	i := 0
	for i < len(files) {
		fmt.Println(files[i])
		if files[i].name == "test.txt" {

		}
		i++
	}*/

}

/**
 * Checks to see if we have the passed in file.
 * @param string filePath - The name of the file to check for - This includes the lynk name.
 * E.G. - 'Cool_Lynk/coolFile.txt'
 * @return bool - A boolean indicating whether or not we have a file in our
 * files array.
 */
func HaveFile(filePath string) bool {
	have := false

	lynkInfo := strings.Split(filePath, "/")
	if len(lynkInfo) != 2 {
		fmt.Println(filePath + " is an invalid filepath")
		return have
	}
	lynkName := lynkInfo[0]
	fileName := lynkInfo[1]
	metaPath := homePath + lynkName + "/meta.info"
	//fmt.Println("META:::::::::::::" + metaPath)
	//fmt.Println("FNAME:::::::::::::" + fileName)
	//fmt.Println("LNAME:::::::::::::" + lynkName)
	//parseMetainfo("../resources/meta.info")
	//fmt.Println(lynks)
	parseMetainfo(metaPath)
	lynk := getLynk(lynks, lynkName)
	fmt.Println(lynk.Files)

	i := 0
	for i < len(lynk.Files) && !have {
		if lynk.Files[i].name == fileName {
			have = true
		}
		i++
	}

	return have
}

/**
 * Simply returns the tracker associated with the passed in Lynk
 * @return string - A string representing the tracker's IP address.
 */
func GetTracker(metaPath string) string {
	parseMetainfo(metaPath)
	lynkName := GetLynkName(metaPath)
	lynk := getLynk(lynks, lynkName)
	return lynk.Tracker
}

/**
 * Gets a file from the peer(s)
 * @param string fileName - The name of the file to find in the peers
 * @return error - An error can be produced if there are connection issues,
 * problems creating or writing to the file, or from not being able to get there
 * desired file - otherwise error will be nil.
 */
func getFile(fileName, metaPath string) error {
	// Will parseMetainfo file and then ask tracker for list of peers when tracker is implemented
	parseMetainfo(metaPath)
	lynkName := GetLynkName(metaPath)
	lynk := getLynk(lynks, lynkName)
	askTrackerForPeers(lynkName)

	i := 0
	gotFile := false
	fmt.Println(lynk.Peers)

	for i < len(lynk.Peers) && !gotFile {
		conn, err := net.Dial("tcp", lynk.Peers[i].IP+":"+lynk.Peers[i].Port)
		if err == nil {
			fmt.Fprintf(conn, "Do_You_Have_FileName:"+lynkName+"/"+fileName+"\n") //Can just append lynkName here

			reply, err := bufio.NewReader(conn).ReadString('\n') // Waits for a String ending in newline
			reply = strings.TrimSpace(reply)

			// Has file and no errors
			if reply != "NO" && err == nil {
				file, err := os.Create(homePath + lynkName + "/" + fileName + "_Network") // + "_Network" is for TESTING that this was a file sent over the network
				if err != nil {
					return err
				}
				defer file.Close()

				/*reply, err = bufio.NewReader(conn).ReadString('\n') // Waits for a String ending in newline
				reply = strings.TrimSpace(reply)
				length, _ := strconv.Atoi(reply)*/

				bufIn := make([]byte, 512) // Will later set this to chunk length instead of 512
				n, err := conn.Read(bufIn)

				// Decrypt
				key := []byte("abcdefghijklmnopqrstuvwxyz123456")
				var plainFile []byte
				if plainFile, err = mycrypt.Decrypt(key, bufIn); err != nil {
					log.Fatal(err)
				}

				// Decompress
				//tempBuf := bytes.NewBuffer(bufIn)
				r, err := gzip.NewReader(bytes.NewBuffer(plainFile))
				bufOut := make([]byte, 512) // Will later set this to chunk length instead of 512
				r.Read(bufOut)
				//io.Copy(os.Stdout, r)
				r.Close()

				fmt.Println(n, "bytes received")
				file.Write(bufOut)
				/*_, err = io.Copy(file, conn)
				if err != nil {
					return err
				}*/
				gotFile = true
			}
		}
		fmt.Println(i)
		i++
	}

	if gotFile {
		return nil
	} else {
		return errors.New("Did not receive File")
	}
}

/**
 * Asks the tracker for a list of peers and then places them into a lynk's peers array
 */
func askTrackerForPeers(lynkName string) {
	lynk := getLynk(lynks, lynkName)
	// Connects to tracker
	conn, err := net.Dial("tcp", lynk.Tracker)
	if err != nil {
		return
	}

	fmt.Fprintf(conn, "Swarm_Request:"+findPCsIP()+":8080:"+lynkName+"\n") // Server runs on 8080 by default
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)

	reply, err := tp.ReadLine()
	//fmt.Println(reply)

	// Tracker will close connection when finished - which will produce error and break us out of this loop
	for err == nil {
		peerArray := strings.Split(reply, ":::")
		tmpPeer := Peer{IP: peerArray[0], Port: peerArray[1]}
		if !contains(lynk.Peers, tmpPeer) {
			lynk.Peers = append(lynk.Peers, tmpPeer)
		}
		reply, err = tp.ReadLine()
	}

	//fmt.Println(peers)
}

/**
 * Simple helper method that checks peers array for specific peer.
 * @param s []peers - The peers array
 * @param e Peer - The peer we are checking for
 */
func contains(s []Peer, e Peer) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

/**
 * Simple helper method that checks lynks array for specific lynk.
 * @param l []Lynk - The lynks array
 * @param lynkName string - The lynk we are checking for
 */
func getLynk(l []Lynk, lynkName string) *Lynk {
	for i, a := range l {
		if a.Name == lynkName {
			return &l[i]
		}
	}
	return nil // Don't have Lynk
}

/**
  Function which creates a new metainfo file for use within the gui server
  @param dirPath string the directory to be added as a lynk
  @param name string the name of the new lynk
*/
func CreateMeta(name string) error {
	tDir, err := os.Stat(homePath + name) // Checks to see if the directory exists
	fmt.Println(tDir.Name())
	if err != nil || !tDir.IsDir() {
		fmt.Println("ERROR!")
		return errors.New("Directory " + name + "does not exist in the Lynx directory.")
	}

	metaFile, err := os.Create(homePath + name + "/meta.info")
	//metaFile, err := os.OpenFile("temp_meta.info", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return err
	}

	currentUser, err := user.Current()
	metaFile.WriteString("announce:::" + findPCsIP() + ":9000\n") //add current ip and tracker is port 9000
	//metaFile.WriteString("port:::4005\n")
	metaFile.WriteString("lynkName:::" + name + "\n")
	metaFile.WriteString("owner:::" + currentUser.Name + "\n")
	//metaFile.WriteString("downloadsdir:::" + downloadsdir + "\n")

	addLynk(name, currentUser.Name)
	//startWalk(name)
	filepath.Walk(homePath+name, visitFiles)

	//FileCopy("temp_meta.info", dirPath+"meta.info")

	//err2 := os.Remove("temp_meta.info") move removal to shutdown process cannot remove
	// due to in use by other proc?
	return nil // Everything was fine if we reached this point
}

/**
Function which visits each file within a directory
@param:path:the path where the root directory is located
@param:f:each file within the root or inner directories
@param:err: any error we way encoutner along the way
*/
func visitFiles(path string, file os.FileInfo, err error) error {
	//dont add directories to meta.info
	if !file.IsDir() && !strings.Contains(path, "_Tracker") && file.Name() != "meta.info" {
		fmt.Println(file.Name())
		slashes := strings.Replace(path, "\\", "/", -1)
		fmt.Println(slashes)
		tmpStr := strings.TrimPrefix(slashes, homePath)
		tmpArr := strings.Split(tmpStr, "/")
		addToMetainfo(path, homePath+tmpArr[0]+"/meta.info")
	}

	return nil
}

/**
Function which visits each directory within a directory
@param:path:the path where the root directory is located
@param:f:each file within the root or inner directories
@param:err: any error we way encoutner along the way
*/
func visitDirectories(path string, file os.FileInfo, err error) error {
	slashes := strings.Replace(path, "\\", "/", -1)
	base := strings.TrimPrefix(slashes, homePath)


	if file.IsDir() && !strings.Contains(base, "/") && base != "" {
		fmt.Println(file.Name())
		lynks = append(lynks, Lynk{Name: file.Name()})
	}

	return nil
}

/**
Function which walks through all the files in the directory and calls visit
@param:root: the root directory to start our walking procedure
*/
/*func startWalk(root string) {
	filepath.Walk(root, visitFiles)
}*/

/**
* Finds the ip of the current pc
* @return error - The single string ip
 */
func findPCsIP() string {
	var onlyfirstip = false //only need first ip address
	var ipstring = ""
	ifaces, err := net.Interfaces()
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			fmt.Println(err)
		}
		for _, addrs := range addrs {
			if ipnet, ok := addrs.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					if !onlyfirstip {
						onlyfirstip = true
						ipstring = ipnet.IP.String()
					}

				}
			}

		}

	}
	if err != nil {
		fmt.Println(err)
	}
	return ipstring
}

/**
  Function which adds a lynk to list of lynks and also will added it to lynks.txt file as well
  @param: name - the name of the lynk
  @param: owner- the owner of the lynk
*/
func addLynk(name, owner string) error {

	lynkFile, err := os.OpenFile(homePath+"lynks.txt", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		// create file if not real
	}

	i := 0
	// look here
	ParseLynks(homePath + "lynks.txt")

	for i < len(lynks) {
		if lynks[i].Name == name {
			return errors.New("Can't Add Duplicate Lynk")
		}
		i++
	}

	lynkFile.WriteString(name + ":::unsynced:::" + owner + "\n")

	//look here
	filepath.Walk(homePath, visitDirectories)
	// look here
	genLynks()

	fmt.Println(lynks)

	return lynkFile.Close()

}

/**
 * Parses the information in meta.info file and places each entry into a File
 * struct and appends that struct to the array of structs
 * @param string metaPath - The path to the metainfo file
 * @return error - An error can be produced when issues arise from trying to access
 * the meta file or from an invalid meta file type - otherwise error will be nil.
 */
func ParseLynks(lynksFilePath string) error {
	lynks = nil // Resets files array

	lynksFile, err := os.Open(lynksFilePath)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(lynksFile)
	tempLynk := Lynk{}

	// Scan each line
	for scanner.Scan() {

		line := strings.TrimSpace(scanner.Text()) // Trim helps with errors in \n
		split := strings.Split(line, ":::")
		tempLynk.Name = split[0]
		tempLynk.Synced = split[1]
		tempLynk.Owner = split[2]

		lynks = append(lynks, tempLynk) // Append the current file to the file array
		tempLynk = Lynk{}               // Empty the current file
	}

	return lynksFile.Close()
}

/**
 * Function which deletes a Lynk based upon its name from the list of lynks
 * @param nameToDelete string - the lynk we want to remove
 */
func DeleteLynk(nameToDelete string) {

	i := 0
	for i < len(lynks) {
		if nameToDelete == lynks[i].Name {
			lynks = append(lynks[:i], lynks[i+1:]...)
		}
		i++
	}
	updateLynksFile()

}

/**
 *  Function which parses the list of lynks and updates the lynks.txt file
 * 	@returns error - will produce an error if we cannot open the lynks.txt file.
 */
func updateLynksFile() error {

	newLynks, err := os.Create(homePath + "lynks.txt")
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

/**
 * Function which will allow a user to join an existing link by way of its meta.info file
 * @param metaPath string - the path to the meta.info file which will be used to find the information
 *  		    about the lynk
 * @param downloadsdir string - the place where all the files will be downloaded after they have been found
 *			in the meta.ifno file
 */
func JoinLynk(metaPath, downloadsdir string) {
	metaFile, err := os.Open(metaPath)
	if err != nil {
		fmt.Println(err)
		//} else if metaPath != "../resources/meta.info" {
		//	return errors.New("Invalid File Type")
	}
	lynkName := ""
	owner := ""
	scanner := bufio.NewScanner(metaFile)
	tempPeer := Peer{}
	// Scan each line
	for scanner.Scan() {

		line := strings.TrimSpace(scanner.Text()) // Trim helps with errors in \n
		split := strings.Split(line, ":::")

		if split[0] == "announce" {
			tempPeer.IP = split[META_VALUE_INDEX]
		} else if split[0] == "port" {
			tempPeer.Port = split[META_VALUE_INDEX]
		} else if split[0] == "lynkname" {
			lynkName = split[META_VALUE_INDEX]
		} else if split[0] == "owner" {
			owner = split[META_VALUE_INDEX]
		}
	}

	//peers = append(peers, tempPeer)
	//fmt.Println(peers)

	addLynk(lynkName, owner)
	// getFile("3HLxd.jpg", lynkName)

}

/**
 * Function init runs before main and allows us to create an array of Lynks.
 */
func init() {
	currentusr, _ := user.Current()
	homePath = currentusr.HomeDir + "/Lynx/"
	homePath = strings.Replace(homePath, "\\", "/", -1)
	filepath.Walk(homePath, visitDirectories)
	genLynks()
	fmt.Println(lynks)
	//lynk := getLynk(lynks, "Tests")
	//fmt.Println(lynk.Files)
}

/**
 * Helper function that generates all the data for our lynx array by parsing each corresponding meta.info file.
 */
func genLynks() {
	i := 0
	for i < len(lynks) {
		parseMetainfo(homePath + lynks[i].Name + "/meta.info")
		i++
	}
}

/**
 * Helper function that returns our Lynk name if we pass in its metaPath.
 * @returns string - The lynk name
 */
func GetLynkName(metaPath string) string {
	return strings.TrimSuffix(strings.TrimPrefix(metaPath, homePath), "/meta.info")
}

func GetLynks() []Lynk{
	return lynks
}
func GetLynksLen() int{
	return len(lynks)
}
func PopulateFilesAndSize() {
	i := 0
	for i < len(lynks) {
		files := lynks[i].Files
		j := 0
		if (len(lynks[i].FileNames) == 0 && len(lynks[i].FileSize) == 0) {
		for j < len(files) {

			lynks[i].FileNames = append(lynks[i].FileNames, files[j].name)
			lynks[i].FileSize = append(lynks[i].FileSize, files[j].length)
			j++
			}
		}
		i++
	}

}

func reorderLynks() error {

	newLynks, err := os.Create(homePath + "lynks.txt")
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





