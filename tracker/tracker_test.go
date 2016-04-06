/**
 *
 *	 The unit tests for our server
 *
 *	 @author: Michael Bruce
 *	 @author: Max Kernchen
 *
 *	 @verison: 2/17/2016
 */

package tracker

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
)

/** Count of the # of successful tests. */
var successful = 0

/** Total # of the tests. */
const total = 7

/**
 * Unit tests for listen, handle, and send functions
 * @param *testing.T t - The wrapper for the test
 */
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

	fmt.Fprintf(conn, "Swarm_Request:813.444.555.111:7500\n")

	reply, err = bufio.NewReader(conn).ReadString('\n') // Waits for a String ending in newline
	reply = strings.TrimSpace(reply)

	if reply == "127.0.0.1:::8080" {
		fmt.Println("Successfully Handled Valid Request")
		successful++
	} else {
		t.Error("Test failed, expected to receive first line of swarm.info. Got", reply)
	}

	conn.Close()

	fmt.Println("\n----------------TestSendFile----------------")

	conn, err = net.Dial("tcp", "127.0.0.1:9000")

	fmt.Fprintf(conn, "Meta_Request:813.444.555.111:7500\n")

	reply, err = bufio.NewReader(conn).ReadString('\n') // Waits for a String ending in newline
	reply = strings.TrimSpace(reply)

	if reply == "announce:::127.0.0.1:9000" {
		fmt.Println("Successfully Sent A File")
		successful++
	} else {
		t.Error("Test failed, expected to receive first line of meta.info. Got", reply)
	}

}

/**
 * Unit tests for parsing, updating, and adding to swarm.info
 * @param *testing.T t - The wrapper for the test
 */
func TestSwarminfo(t *testing.T) {
	fmt.Println("\n----------------TestParseSwarminfo----------------")

	result := parseSwarminfo("../resources/swarm.info")

	if result != nil {
		t.Error("Test failed, expected duplicate error. Got ", result)
	} else {
		fmt.Println("Successfully Parsed Swarm Info")
		successful++
	}

	fmt.Println("\n----------------TestAddToSwarminfo----------------")

	p1 := Peer{IP: "124.123.563.186", Port: "4500"}
	result = addToSwarminfo(p1, "../resources/swarm.info")

	if result == nil {
		t.Error("Test failed, expected duplicate error. Got ", result)
	} else {
		fmt.Println("Successfully Avoided Duplicate")
		successful++
	}

	fmt.Println("\n----------------TestUpdateSwarminfo----------------")

	result = updateSwarminfo()

	if result != nil {
		t.Error("Test failed, expected no error. Got ", result)
	} else {
		fmt.Println("Successfully Updated Swarm Info")
		successful++
	}

	fmt.Println("\nSuccess on ", successful, "/", total, " tests.")
}
