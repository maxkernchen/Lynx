// The unit tests for our tracker
// @author: Michael Bruce
// @author: Max Kernchen
// @verison: 2/17/2016
package tracker

import (
	"bufio"
	"capstone/lynxutil"
	"fmt"
	"io/ioutil"
	"net"
	"os/user"
	"strings"
	"testing"
)

// Count of the # of successful tests.
var successful = 0

// Total # of the tests.
const total = 8

// Gets user's home directory */
var cU, _ = user.Current()

// Adds "Lynx" to home directory string */
var hPath = cU.HomeDir + "/Lynx/"

// Uses homePath and our Tests Lynk to create swarm path */
var sPath = hPath + "Tests/Tests_Tracker/swarm.info"

// Uses homePath and our Tests Lynk to create meta path */
var mPath = hPath + "Tests/Tests_Tracker/meta.info"

// Unit tests for listen, handle, and send functions
// @param *testing.T t - The wrapper for the test
func TestListenHandleSend(t *testing.T) {
	fmt.Println("\n----------------TestListen----------------")

	conn, err := net.Dial("tcp", "127.0.0.1:9000")

	if err != nil {
		t.Error(err.Error())
	} else {
		fmt.Println("Successfully Connected To Tracker")
		successful++
	}

	fmt.Println("\n----------------TestHandleRequest----------------")

	fmt.Fprintf(conn, "lol:fake.txt\n")

	reply, err := bufio.NewReader(conn).ReadString('\n') // Waits for a String ending in newline
	reply = strings.TrimSpace(reply)

	if err != nil {
		fmt.Println("Successfully Handled Invalid Request")
		successful++
	} else {
		t.Error("Test failed, expected to have server respond 'NO'. Got", reply)
	}
	conn.Close()

	conn, err = net.Dial("tcp", "127.0.0.1:9000")

	fmt.Fprintf(conn, "Swarm_Request:813.444.555.111:7500:Tests\n")

	reply, err = bufio.NewReader(conn).ReadString('\n') // Waits for a String ending in newline
	reply = strings.TrimSpace(reply)
	content, _ := ioutil.ReadFile(sPath)
	s := string(content)

	if strings.Contains(s, reply) {
		fmt.Println("Successfully Handled Valid Request")
		successful++
	} else {
		t.Error("Test failed, expected to receive swarm.info. Got", reply)
	}

	conn.Close()

	fmt.Println("\n----------------TestSendFile----------------")

	conn, err = net.Dial("tcp", "127.0.0.1:9000")

	fmt.Fprintf(conn, "Meta_Request:813.444.555.111:7500:Tests\n")

	reply, err = bufio.NewReader(conn).ReadString('\n') // Waits for a String ending in newline
	reply = strings.TrimSpace(reply)
	content, _ = ioutil.ReadFile(mPath)
	s = string(content)

	if strings.Contains(s, reply) {
		fmt.Println("Successfully Sent A File")
		successful++
	} else {
		t.Error("Test failed, expected to receive first line of meta.info. Got", reply)
	}

}

// Unit tests for parsing, updating, and adding to swarm.info
// @param *testing.T t - The wrapper for the test
func TestSwarminfo(t *testing.T) {
	fmt.Println("\n----------------TestParseSwarminfo----------------")

	result := parseSwarminfo(sPath)

	if result != nil {
		t.Error("Test failed, expected no error. Got ", result)
	} else {
		fmt.Println("Successfully Parsed Swarm Info")
		successful++
	}

	fmt.Println("\n----------------TestAddToSwarminfo----------------")

	p1 := lynxutil.Peer{IP: "124.123.563.186", Port: "4500"}
	result = addToSwarminfo(p1, sPath)

	if result == nil {
		t.Error("*Run Test Twice If This Is First Time* Test failed, expected duplicate error. Got ",
			result)
	} else {
		fmt.Println("Successfully Avoided Duplicate")
		successful++
	}

	fmt.Println("\n----------------TestUpdateSwarminfo----------------")

	result = updateSwarminfo(sPath)

	if result != nil {
		t.Error("Test failed, expected no error. Got ", result)
	} else {
		fmt.Println("Successfully Updated Swarm Info")
		successful++
	}

	fmt.Println("\n----------------TestBroadcastIP----------------")

	BroadcastNewIP(sPath)

	if result != nil {
		t.Error("Test failed, expected no error. Got ", result)
	} else {
		fmt.Println("Successfully Broadcast New IP")
		successful++
	}

	fmt.Println("\nSuccess on ", successful, "/", total, " tests.")
}
