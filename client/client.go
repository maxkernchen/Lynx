package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"net"
	"bufio"
	"strings"
)

type Peer struct {
	IP   string
	Port string
}

var trackerIP string // Will be set after parsing Metainfo
var peers []Peer

var metamap map[string]map[string]string
var infodict map[string]string

func metainfo(src string){
	metamap = make(map[string]map[string]string)
	infodict = make(map[string]string)

	fileInfo, err := os.Stat(src)
	if err != nil {
	}
	metamap["announce"] = map[string]string{"announceInner": "url of tracker annouce"}
	temp := fileInfo.ModTime().String()
	metamap["creation_date"] =  map[string]string{"creation_dateInner": temp}

	infodict["length"] = strconv.FormatInt(fileInfo.Size(), 10) //covert to string
	infodict["name"] = fileInfo.Name()
	infodict["piece_length"] = "fileInfo.Size()/constant"
	infodict["pieces"] = "sha1.Sum(chunckBuffer)"
	metamap["info"] = infodict

	fmt.Println(metamap["announce"])
	fmt.Println(metamap["creation_date"])

	fmt.Println(metamap["info"])
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
	}
	*/

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

