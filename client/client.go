package main

import (
	"fmt"
	"io"
	"os"
	"net"
	"bufio"
	"strings"
	"strconv"
)

type Peer struct {
	IP   string
	Port string
}

type File struct {
	length int
	path string
	name string
	piece_length int
	pieces string
}
var files []File
var trackerIP string // Will be set after parsing Metainfo
var peers []Peer

const END_OF_FILE = ":#!"



func metainfo(metainfo_path string){
	metainfo_file, err := os.Open(metainfo_path)
	if err != nil{

	}
	defer metainfo_file.Close()

	scanner := bufio.NewScanner(metainfo_file)
	temp_file := File{}

	for scanner.Scan() {
		temp_line := scanner.Text()
		split_arry := strings.Split(temp_line, ":::")

		if(strings.Contains(temp_line,"announce")){
			trackerIP = split_arry[1]
		}else if(strings.Contains(temp_line, "length")){
			temp_int,err := strconv.Atoi(split_arry[1])
			temp_file.length = temp_int
			if err != nil{
			}
		}else if(strings.Contains(temp_line, "path")){
			temp_file.path = split_arry[1]
		}else if(strings.Contains(temp_line,"name")){
			temp_file.name = split_arry[1]
		}else if(strings.Contains(temp_line,"piece_length")){
			temp_int,err := strconv.Atoi(split_arry[1])
			temp_file.piece_length = temp_int
			if err != nil{
			}
		}else if(strings.Contains(temp_line,"pieces")){
			temp_file.pieces = split_arry[1]
		}else if(strings.Contains(temp_line,END_OF_FILE)){
			files = append(files, temp_file)
			temp_file = File{}
		}

	}
	fmt.Printf("%v", files)





}


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



func main() {
	/*fmt.Println("Hello World!")
	fmt.Println("Cool Beans!")
	err := fileCopy(os.Args[1], os.Args[2])

	if err != nil {
		fmt.Println("You suck")
	} else {
		fmt.Println(os.Args[1] + " copied to " + os.Args[2])
	}*/


	metainfo(os.Args[1]);

}

// ------------------------- CODE BELOW THIS LINE IS UNTESTED AND DANGEROUS ------------------------- \\

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

func getFile(fileID string) {

	i := 0
	for i < len(peers) {
		conn, err := net.Dial("tcp", peers[i].IP);
		if err != nil {
			return
		}

		fmt.Fprintf(conn, "Do_You_Have_FileID:" + fileID)

		reply, err := bufio.NewReader(conn).ReadString('\n') // Waits for a String ending in newline

		if reply == "NO" || err != nil {
			break // could set boolean instead
		} else {
			file, err := os.Create(fileID) // Should use name instead of id
			if err != nil {
				break // could set boolean instead
			}
			defer file.Close();

			n, err := io.Copy(conn, file)
			if err != nil {
				break // could set boolean instead
			}
			fmt.Println(n, "this was sent")
		}


	}

}

