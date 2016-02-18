/**
 *
 *	The client side of the Lynx application. Currently handles file copying, metainfo parsing,
 *	metainfo entry addition and deletion, and experimentally handles peer and file retrieval
 *
 *	 @author: Michael Bruce
 *	 @author: Max Kernchen
 *
 *	 @verison: 2/17/2016
 */

package main

import (
	"fmt"
	"io"
	"os"
	"net"
	"bufio"
	"strings"
	"strconv"
	"path/filepath"
	"errors"
)

/**	A struct which represents a Peer of the client */
type Peer struct {
	IP   string
	Port string
}

/** A struct based which represents a File in our Lynx directory. It is based
upon BitTorrent protocol dictionaries */
type File struct {
	length       int
	path         string
	name         string
	pieces       string
	piece_length int
}

/** An array of the files found from parsing the metainfo file */
var files []File
/** The IP Address of our tracker */
var trackerIP string
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
	for i < len(files){
		if(nameToDelete == files[i].name){
			if len(files) > 2 {
				files = append(files[:i], files[i+1:]...)
			} else if i == 0 {
				files = append(files[i:])
			} else if i == 1 {
				files = append(files[:i])
			}
		}
		i++
	}

}

/**
 * Deletes the current meta.info and replaces it with a new version that
 * accurately reflects the array of Files after they have been modified
 */
func updateMetainfo() error {

	err := os.Remove("meta.info")
	if err != nil {
		fmt.Println(err)
		return err
	}

	newMetainfo, err := os.Create("meta.info")
	if err != nil {
		fmt.Println(err)
		return err
	}

	newMetainfo.WriteString("annouce:::" + trackerIP + "\n") // Write tracker IP
	i := 0
	for i < len(files) {
		newMetainfo.WriteString("length:::" + strconv.Itoa(files[i].length) + "\n") //convert to str
		newMetainfo.WriteString("path:::" + files[i].path + "\n")
		newMetainfo.WriteString("name:::" + files[i].name + "\n")
		newMetainfo.WriteString("pieces_length:::" + strconv.Itoa(files[i].piece_length) + "\n")
		newMetainfo.WriteString("pieces:::" + files[i].pieces + "\n")
		newMetainfo.WriteString(END_OF_ENTRY + "\n")

		i++
	}

	return newMetainfo.Close()

}

/**
 * Parses the information in meta.info file and places each entry into a File
 * struct and appends that struct to the array of structs
 * @param string metaPath - The path to the metainfo file
 */
func parseMetainfo(metaPath string) error {

	metainfo_file, err := os.Open(metaPath)
	if err != nil {
		return err
	} else if metaPath != "meta.info" {
		return errors.New("Invalid File Type")
	}

	scanner   := bufio.NewScanner(metainfo_file)
	temp_file := File{}

	// Scan each line
	for scanner.Scan() {

		line  := strings.TrimSpace(scanner.Text()) // Trim helps with errors in \n
		split := strings.Split(line, ":::")

		if split[0] == "announce" { // If the line is announce
			trackerIP = split[META_VALUE_INDEX]
		} else if split[0] == "pieces_length" {
			temp_int, err := strconv.Atoi(split[META_VALUE_INDEX])
			temp_file.piece_length = temp_int
			if err != nil {
				return err
			}
		} else if split[0] == "length" {
			temp_int, err := strconv.Atoi(split[META_VALUE_INDEX])
			temp_file.length = temp_int
			if err != nil {
				return err
			}
		} else if (strings.Contains(line, "path")) {
			temp_file.path = split[META_VALUE_INDEX]
		} else if (strings.Contains(line, "name")) {
			temp_file.name = split[META_VALUE_INDEX]
		} else if (strings.Contains(line, "pieces")) {
			temp_file.pieces = split[META_VALUE_INDEX]
		} else if (strings.Contains(line, END_OF_ENTRY)) {
			files = append(files, temp_file)  // Append the current file to the list of structs
			temp_file = File{} // Empty the current file
		}

	}

	return metainfo_file.Close()
}

/**
 * Adds a file to the meta.info by parsing that file's information
 * @param string addPath - the path of the file to be added
 * @param string metaPath - the path of the metainfo file
 */
func addToMetainfo(addPath, metaPath string) error {
	metainfo_file, err := os.OpenFile(metaPath, os.O_APPEND | os.O_WRONLY, 0644) // Opens for appending
	if err != nil{
		return err
	}

	add_info,err := os.Stat(addPath)
	if err != nil{
		return err
	}

	parseMetainfo(metaPath)
	i := 0
	for i < len(files) {
		if files[i].name == add_info.Name() {
			return errors.New("Can't Add Duplicates To Metainfo")
		}
		i++
	}

	temp_size := add_info.Size()	// Write length
	temp_str  := strconv.FormatInt(temp_size,10) 	// Convert int64 to string
	metainfo_file.WriteString("length:::" + temp_str + "\n")

	temp_file_path, err := filepath.Abs(addPath) 	// Find the path of the current file
	if err != nil{
		return err
	}

	// Write to metainfo file using ::: to separate keys and values
	metainfo_file.WriteString("path:::" + temp_file_path + "\n")
	metainfo_file.WriteString("name:::" + add_info.Name() + "\n")
	metainfo_file.WriteString("pieces_length:::-1\n")
	metainfo_file.WriteString("pieces:::chuncking not currently implemented\n")
	metainfo_file.WriteString(END_OF_ENTRY + "\n")

	return metainfo_file.Close()
}

/**
 * Copies a file from src to dst
 * @param string src - the file that will be copied
 * @param string dst - the destination of the file to be copied
 */
func fileCopy(src, dst string) error {
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

	cerr := out.Close() // Checks for close error
	return cerr
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
	parseMetainfo("meta.info")
	i := 0
	for i < len(files) {
		fmt.Println(files[i])
		if files[i].name == "test.txt" {
			
		}
		i++
	}

	//addToMetainfo("test2.txt", "meta.info")
	//deleteEntry("test.txt")
	//updateMetainfo()

}

// ------------------------- CODE BELOW THIS LINE IS UNTESTED AND DANGEROUS ------------------------- \\

/**
 * Asks the tracker for a list of peers and then places them into peers array
 */
func askTrackerForPeers() {
	// Connets to tracker
	conn, err := net.Dial("tcp", trackerIP);
	if err != nil {
		return
	}

	fmt.Fprintf(conn, "Announce_Request: <Stuff>")

	reply, err := bufio.NewReader(conn).ReadString('\n') // Waits for a String ending in newline

	for err != nil {
		peerArray := strings.Split(reply, ":::")
		peers = append(peers, Peer{IP:peerArray[0], Port:peerArray[1]})
		reply, err = bufio.NewReader(conn).ReadString('\n') // Waits for a String ending in newline
	}

}

/**
 * Gets a file from the peer(s)
 * @param string fileName - The name of the file to find in the peers
 */
func getFile(fileName string) {

	i := 0
	gotFile := false
	for i < len(peers) && !gotFile {
		conn, err := net.Dial("tcp", peers[i].IP);
		if err != nil {
			return
		}

		fmt.Fprintf(conn, "Do_You_Have_FileName:" + fileName)

		reply, err := bufio.NewReader(conn).ReadString('\n') // Waits for a String ending in newline

		// Has file and no errors
		if reply != "NO" && err == nil {
			file, err := os.Create(fileName)
			if err != nil {
				break // could set boolean instead
			}
			defer file.Close();

			n, err := io.Copy(conn, file)
			if err != nil {
				break // could set boolean instead
			}
			fmt.Println(n, "this was sent")
			gotFile = true
		}

		i++
	}

}
