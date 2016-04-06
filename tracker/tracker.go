/**
 *
 *	 The tracker for the Lynx application. Currently handles
 *
 *	 @author: Michael Bruce
 *	 @author: Max Kernchen
 *
 *	 @verison: 2/17/2016
 */

package tracker

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
	"os"
	"strings"
	"os/user"
)

/**	A struct which represents a Peer of the client */
type Peer struct {
	IP   string
	Port string
}

/** An array of all the client's peers */
var peers []Peer

/**
 * Function that deletes an entry from our peers array.
 * @param Peer peerToDelete - This is the peer struct we want to delete
 */
func deleteEntry(peerToDelete Peer) {
	i := 0
	for i < len(peers) {
		if peerToDelete.IP == peers[i].IP && peerToDelete.Port == peers[i].Port {
			peers = append(peers[:i], peers[i+1:]...)
		}
		i++
	}
}

/**
 * Deletes the current swarm.info and replaces it with a new version that
 * accurately reflects the array of Peers after they have been modified
 * @return error - An error can be produced when issues arise from trying to create
 * or remove the swarm file - otherwise error will be nil.
 */
func updateSwarminfo() error {
	parseSwarminfo("../resources/swarm.info")

	err := os.Remove("../resources/swarm.info")
	if err != nil {
		fmt.Println(err)
		return err
	}

	newSwarmInfo, err := os.Create("../resources/swarm.info")
	if err != nil {
		fmt.Println(err)
		return err
	}

	i := 0
	for i < len(peers) {
		newSwarmInfo.WriteString(peers[i].IP + ":::" + peers[i].Port + "\n")
		i++
	}

	return newSwarmInfo.Close()
}

/**
 * Parses the information in swarm.info file and places each entry into a Peer
 * struct and appends that struct to the array of peers
 * @param string swarmPath - The path to the swarminfo file
 * @return error - An error can be produced when issues arise from trying to access
 * the swarm file or from an invalid swarm file type - otherwise error will be nil.
 */
func parseSwarminfo(swarmPath string) error {
	peers = nil // Resets peers array

	swarmFile, err := os.Open(swarmPath)
	if err != nil {
		return err
	} else if swarmPath != "../resources/swarm.info" {
		return errors.New("Invalid File Type")
	}

	scanner := bufio.NewScanner(swarmFile)
	tempPeer := Peer{}

	// Scan each line
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text()) // Trim helps with errors in \n
		split := strings.Split(line, ":::")

		tempPeer.IP = split[0]
		tempPeer.Port = split[1]
		peers = append(peers, tempPeer) // Append the current file to the file array
	}

	return swarmFile.Close()
}

/**
 * Adds a peer to the swarm.info file
 * @param string addPath - the path of the file to be added
 * @param string swarmPath - the path of the swarminfo file
 * @return error - An error can be produced when issues arise from trying to access
 * the swarm file or if the file to be added already exists in the swarm file - otherwise
 * error will be nil.
 */
func addToSwarminfo(addPeer Peer, swarmPath string) error {
	swarmFile, err := os.OpenFile(swarmPath, os.O_APPEND|os.O_WRONLY, 0644) // Opens for appending
	if err != nil {
		return err
	}

	parseSwarminfo(swarmPath)

	i := 0
	for i < len(peers) {
		if peers[i].IP == addPeer.IP && peers[i].Port == addPeer.Port {
			return errors.New("Can't Add Duplicates To Swarminfo")
		}
		i++
	}

	// Write to swarminfo file using ::: to IP and Port
	swarmFile.WriteString(addPeer.IP + ":::" + addPeer.Port + "\n")

	return swarmFile.Close()
}

/**
 * Function used to drive and test our other client functions
 */
func main() {
	p1 := Peer{IP: "124.123.563.186", Port: "4500"}
	p2 := Peer{IP: "812.333.444.555", Port: "6000"}
	addToSwarminfo(p1, "../resources/swarm.info")
	addToSwarminfo(p2, "../resources/swarm.info")
	addToSwarminfo(p1, "../resources/swarm.info")
	parseSwarminfo("../resources/swarm.info")

	i := 0
	for i < len(peers) {
		fmt.Println(peers[i].IP)
		i++
	}

	Listen()
}

/**
 * Creates a welcomeSocket that listens for TCP connections - once someone connects a goroutine is spawned
 * to handle the request
 */
func Listen() {

	fmt.Println("Starting Tracker on Port 9000")

	welcomeSocket, wErr := net.Listen("tcp", ":9000") // Will later need to set port dynamically

	if wErr != nil {
		// handle error
	}

	var cErr error

	for cErr == nil {
		conn, cErr := welcomeSocket.Accept()
		if cErr != nil {
			// handle error
		}
		go handleRequest(conn)
	}

}

/**
 * Handles a request / push sent by a client, can either be a swarm or meta request or a push
 * of an updated meta.info file - also adds the requesting client to the swarm.info file
 * @param net.Conn conn - The socket which the client is asking on
 * @return error - An error can be produced when trying to send a file or if there is incorrect
 * syntax in the request - otherwise error will be nil.
 */
func handleRequest(conn net.Conn) error {
	request, err := bufio.NewReader(conn).ReadString('\n') // Waits for a String ending in newline
	if err != nil {
		return err
	}

	request = strings.TrimSpace(request)
	fmt.Println(request)
	// Makes sure we are dealing with a request and not a push
	if request != "Meta_Push" {
		// Client syntax for request is "Swarm_Request:<IP>:<Port>\n"
		tmpArr := strings.Split(request, ":")
		if len(tmpArr) != 3 {
			conn.Close()
			return errors.New("Invalid Request Syntax")
		}

		fileToSend := ""
		// Checks to see if we are dealing w/ a Swarm or Meta Request
		if tmpArr[0] == "Swarm_Request" {
			fileToSend = "../resources/swarm.info"
		} else if tmpArr[0] == "Meta_Request" {
			fileToSend = "../resources/meta.info" // Should encrypt & compress when sending meta.info
		} else {
			conn.Close()
			return errors.New("Invalid Request Syntax")
		}

		tmpPeer := Peer{IP: strings.TrimSpace(tmpArr[1]), Port: strings.TrimSpace(tmpArr[2])}
		fmt.Println("New Peer: ", tmpPeer)

		err = sendFile(fileToSend, conn) // Sending The file
		if err != nil {
			conn.Close()
			return err
		}

		addToSwarminfo(tmpPeer, "../resources/swarm.info") // So we only add peer to swarmlist on success
		fmt.Println("No Errors")
	} else { // We are receiving a meta.info file
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

		/*reader := bufio.NewReader(conn)
		tp := textproto.NewReader(reader)

		reply, err := tp.ReadLine()
		fmt.Println(reply)

		for err == nil {
			reply, err = tp.ReadLine()
			fmt.Println(reply)
		}*/

		/*n, err := io.Copy(newMetainfo, conn)
		if err != nil {
			fmt.Println(err)
			return err
		}*/

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
		newMetainfo.Write(bufOut)

		//fmt.Println(n, " bytes copied")
	}

	return conn.Close()
}

/**
 * Sends a file across the network to a peer.
 * @param string fileName - The name of the file to send to the peer
 * @param net.Conn conn - The socket over which we will send the file
 * @return error - An error can be produced when trying open a file or write over
 * the network - otherwise error will be nil.
 */
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
/**
 * Creates a new swarm.info upon clicking of create button in gui
 * @param string downloadsdir - the directory where all files within it will be put into the lynk
 * @param string lynkname - the name of the lynk
 */
func CreateSwarm(downloadsdir, lynkname string){
	p1 := Peer{IP: "", Port: "4500"}

	os.Create("temp_swarm.info")

	swarmFile,err := os.OpenFile("temp_swarm.info", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}

	swarmFile.WriteString("locationofdownloads:::"+downloadsdir + "\n")

	currentuser, err := user.Current()
	trackerdir := lynkname+"_tracker"
	os.Mkdir(currentuser.HomeDir+"/"+trackerdir,0644)

	ipstring := findPCsIP()
	p1.IP=ipstring
	addToSwarminfo(p1,"temp_swarm.info")

	fileCopy("temp_swarm.info", currentuser.HomeDir+"/"+trackerdir+"/swarm.info")
	fileCopy("temp_meta.info", currentuser.HomeDir+"/"+trackerdir+"/meta.info")


}

/**
 * Copies a file from src to dst
 * @param string src - the file that will be copied
 * @param string dst - the destination of the file to be copied
 * @return error - An error can be produced when issues arise from trying to access,
 * create, and write from either the src or dst files - otherwise error will be nil.
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

	return out.Close() // Checks for close error
}
/**
* Finds the ip of the current pc
* @return error - The single string ip
*/
func findPCsIP() string {
	var onlyfirstip = false
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




