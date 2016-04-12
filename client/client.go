/**
 *
 *	The client side of the Lynx application. Currently handles file copying, metainfo parsing,
 *	metainfo entry addition and deletion, and getting files from peers.
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
	"path/filepath"
	"strconv"
	"strings"
	"os/user"
)

/**	A struct which represents a Peer of the client */
type Peer struct {
	IP   string
	Port string
}
/**
   A struct which holds all the lynks within a connection
 */
type Lynk struct{
	Name string
	Owner string
	Synced string
}

/** A struct based which represents a File in our Lynx directory. It is based
upon BitTorrent protocol dictionaries */
type File struct {
	length      int
	path        string
	name        string
	chunks      string
	chunkLength int
}

/** An array of the lynks found from parsing the lynks.txt file */
var lynks []Lynk

/** An array of the files found from parsing the metainfo file */
var files []File

/** The IP Address / Port of our tracker */
var tracker string

/** An array of all the client's peers */
var peers []Peer

/** A special symbol we use to denote the end of 1 entry in the metainfo file */
const END_OF_ENTRY = ":#!"

/** The array index of our metainfo values */
const META_VALUE_INDEX = 1

/**
 * Function that deletes an entry from our files array.
 * @param string nameToDelete - This is the name of the file we want to delete
 */
func deleteEntry(nameToDelete string) {

	i := 0
	for i < len(files) {
		if nameToDelete == files[i].name {
			files = append(files[:i], files[i+1:]...)
		}
		i++
	}

}

/**
 * Deletes the current meta.info and replaces it with a new version that
 * accurately reflects the array of Files after they have been modified
 * @return error - An error can be produced when issues arise from trying to create
 * or remove the meta file - otherwise error will be nil.
 */
func updateMetainfo() error {
	parseMetainfo("../resources/meta.info")

	err := os.Remove("../resources/meta.info")
	if err != nil {
		fmt.Println(err)
		return err
	}

	newMetainfo, err := os.Create("../resources/meta.info")
	if err != nil {
		fmt.Println(err)
		return err
	}

	newMetainfo.WriteString("announce:::" + tracker + "\n") // Write tracker IP
	i := 0
	for i < len(files) {
		newMetainfo.WriteString("length:::" + strconv.Itoa(files[i].length) + "\n") //convert to str
		newMetainfo.WriteString("path:::" + files[i].path + "\n")
		newMetainfo.WriteString("name:::" + files[i].name + "\n")
		newMetainfo.WriteString("chunkLength:::" + strconv.Itoa(files[i].chunkLength) + "\n")
		newMetainfo.WriteString("chunks:::" + files[i].chunks + "\n")
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
	files = nil // Resets files array

	metaFile, err := os.Open(metaPath)
	if err != nil {
		return err
	} else if metaPath != "../resources/meta.info" {
		return errors.New("Invalid File Type")
	}

	scanner := bufio.NewScanner(metaFile)
	tempFile := File{}

	// Scan each line
	for scanner.Scan() {

		line := strings.TrimSpace(scanner.Text()) // Trim helps with errors in \n
		fmt.Println(line)
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

	return metaFile.Close()
}

/**
 * Adds a file to the meta.info by parsing that file's information
 * @param string addPath - the path of the file to be added
 * @param string metaPath - the path of the metainfo file
 * @return error - An error can be produced when issues arise from trying to access
 * the meta file or if the file to be added already exists in the meta file - otherwise
 * error will be nil.
 */
func addToMetainfo(addPath, metaPath string) error {
	metaFile, err := os.OpenFile(metaPath, os.O_APPEND|os.O_WRONLY, 0644) // Opens for appending
	if err != nil {
		return err
	}

	addStat, err := os.Stat(addPath)
	if err != nil {
		return err
	}

	parseMetainfo(metaPath)

	i := 0
	for i < len(files) {
		if files[i].name == addStat.Name() {
			return errors.New("Can't Add Duplicates To Metainfo")
		}
		i++
	}

	//tempSize := addStat.Size()                 // Write length
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
	//defer out.Close()

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
	addToMetainfo("test.txt", "../resources/meta.info")
	addToMetainfo("test2.txt", "../resources/meta.info")
	addToMetainfo("file1.txt", "../resources/meta.info")
	parseMetainfo("../resources/meta.info")

	i := 0
	for i < len(files) {
		fmt.Println(files[i])
		if files[i].name == "test.txt" {

		}
		i++
	}

}

/**
 * Checks to see if we have the passed in file. This function works based on
 * a filepath relative to where the executable using it is run.
 * @param string fileName - The name of the file to check for
 * @return bool - A boolean indicating whether or not we have a file in our
 * files array.
 */
func HaveFile(fileName string) bool {
	have := false

	parseMetainfo("../resources/meta.info")

	i := 0
	for i < len(files) && !have {
		if files[i].name == fileName {
			have = true
		}
		i++
	}

	return have
}

/**
 * Simply returns the tracker global variable after parsing the meta.info file
 * @return string - A string representing the tracker's IP address.
 */
func GetTracker() string {
	parseMetainfo("../resources/meta.info")

	return tracker
}

/**
 * Gets a file from the peer(s)
 * @param string fileName - The name of the file to find in the peers
 * @return error - An error can be produced if there are connection issues,
 * problems creating or writing to the file, or from not being able to get there
 * desired file - otherwise error will be nil.
 */
func getFile(fileName string) error {
	// Will parseMetainfo file and then ask tracker for list of peers when tracker is implemented
	parseMetainfo("../resources/meta.info")
	askTrackerForPeers()
	//peers = append(peers, Peer{IP: "127.0.0.1", Port: "8080"}) // For testing ONLY - Hardcodes myself as a peer

	i := 0
	gotFile := false
	fmt.Println(peers)

	for i < len(peers) && !gotFile {
		conn, err := net.Dial("tcp", peers[i].IP+":"+peers[i].Port)
		if err == nil {
			fmt.Fprintf(conn, "Do_You_Have_FileName:"+fileName+"\n")

			reply, err := bufio.NewReader(conn).ReadString('\n') // Waits for a String ending in newline
			reply = strings.TrimSpace(reply)

			// Has file and no errors
			if reply != "NO" && err == nil {
				file, err := os.Create(fileName + "_Network") // + "_Network" is for TESTING that this was a file sent over the network
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
 * Asks the tracker for a list of peers and then places them into peers array
 */
func askTrackerForPeers() {
	// Connects to tracker
	conn, err := net.Dial("tcp", tracker)
	if err != nil {
		return
	}

	fmt.Fprintf(conn, "Swarm_Request:127.0.0.1:8080\n") // Need to write function in server which let's us get its port/ip
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)

	reply, err := tp.ReadLine()
	//fmt.Println(reply)

	// Tracker will close connection when finished - which will produce error and break us out of this loop
	for err == nil {
		peerArray := strings.Split(reply, ":::")
		tmpPeer := Peer{IP: peerArray[0], Port: peerArray[1]}
		if !contains(peers, tmpPeer) {
			peers = append(peers, tmpPeer)
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
  Function which creates a new metainfo file for use within the gui server

  @param:downloadsdir: the directory where the files exisit to be added to the lynk
  @param:name: the name of the new lynk
 */
func CreateMeta(downloadsdir, name string){
	os.Create("temp_meta.info")

	metaFile,err := os.OpenFile("temp_meta.info", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}

	currentUser, err := user.Current()
	metaFile.WriteString("announce:::"+findPCsIP()+"\n") //add current ip
	metaFile.WriteString("lynkname:::" + name + "\n")
	metaFile.WriteString("owner:::"+currentUser.Name +"\n")
	metaFile.WriteString("downloadsdir:::"+downloadsdir +"\n")

	addLynk(name,currentUser.Name)

	startWalk(downloadsdir)

	FileCopy("temp_meta.info", downloadsdir + "meta.info")

	//err2 := os.Remove("temp_meta.info") move removal to shutdown process cannot remove
	// due to in use by other proc?
}
/**
 Function which visits each file within a directory
 @param:path:the path where the root directory is located
 @param:f:each file within the root or inner directories
 @param:err: any error we way encoutner along the way
 */
func visit(path string, file os.FileInfo, err error) error {
	//dont add directories to meta.info
	if(!file.IsDir()){
		addToMetainfo(path,"temp_meta.info")
	}

	return nil
}
/**
 Function which walks through all the files in the directory and calls visit
 @param:root: the root directory to start our walking procedure
*/
func startWalk(root string) {
	filepath.Walk(root, visit)
}

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
		if err != nil{
			fmt.Println(err)
		}
		for _, addrs := range addrs {
			if ipnet, ok := addrs.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					if(!onlyfirstip){
						onlyfirstip = true
						ipstring=ipnet.IP.String()
					}


				}
			}

		}

	}
	if err != nil{
		fmt.Println(err)
	}
	return ipstring
}



func addLynk(name, owner string) error{

	lynkFile,err := os.OpenFile("resources/lynks.txt", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}
	i := 0
	for i < len(lynks) {
		if lynks[i].Name == name {
			return errors.New("Can't Add Duplicate Lynk")
		}
		i++
	}
	lynkFile.WriteString(name + ":::" +"unsynced:::" + owner + "\n")

	ParseLynks("resources/lynks.txt")
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
		tempLynk = Lynk{}           // Empty the current file
	}


	return lynksFile.Close()
}
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

func updateLynksFile() error {


	newLynks, err := os.Create("resources/lynks.txt")
	if err != nil {
		fmt.Println(err)
		return err
	}

	i := 0
	for i < len(lynks) {
		newLynks.WriteString(lynks[i].Name+":::" +lynks[i].Synced+":::" +
		lynks[i].Owner + "\n")

		i++
	}

	return newLynks.Close()

}


