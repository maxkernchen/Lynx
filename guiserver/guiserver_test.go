// The unit tests for our guiserver - which are essentially our system tests since the guiserver
// is the entry point into the Lynx application
// @author: Michael Bruce
// @author: Max Kernchen
// @verison: 2/17/2016
package main

import (
	"capstone/client"
	"capstone/lynxutil"
	"capstone/tracker"
	"errors"
	"flag"
	"fmt"
	"os/user"
	"testing"
	"time"
)

// Gets user's home directory
var cU, _ = user.Current()

// Adds "Lynx" to home directory string
var hPath = cU.HomeDir + "/Lynx/"

// Uses homePath and our Tests Lynk to create mPath
var mPath = hPath + "SysTests/meta.info"

// Path to the image file
var imgPath = cU.HomeDir + "/funny.jpg"

// Path to the joining meta.info file
var joinPath = cU.HomeDir + "/meta.info"

// Count of the # of successful tests.
var successful = 0

// Count of the # of original files in systemtests.
var original = 0

// Total # of the tests.
const total = 2

// If delay > 0 - we will start Lynx after the delay specified
var delay int

// How Lynx is started. 0:Simple Start, 1:Create, 2:Join
var startUp int

// Lynx's operation. -1:Remove, 0:Simple Run, 1:Add, 2:Verify Received
var op int

// How long Lynx waits to perform op.
var opDelay int

// Unit tests for our Encrypt and Decrypt functions.
// @param *testing.T t - The wrapper for the test
func TestMain(t *testing.T) {
	defer func() {
		if e := recover(); e != nil {
			// e is the interface{} typed-value we passed to panic()
			t.Error(e)
		}
	}()

	if delay > 0 {
		time.Sleep(time.Duration(delay) * time.Minute) // Waits X amount of time and then continues
	}

	go launch() // Fires up web server

	err := testStartUp()
	if err != nil {
		t.Error("Start Up Failed: ", err)
	}

	err = testOp()
	if err != nil {
		t.Error("Operation Failed: ", err)
	}

	//fmt.Println("\nSuccess on ", successful, "/", total, " tests.")
}

// Helper function that verifies there has been a change in the files for the system tests lynk.
// @returns True if there has been a change, False if there has not been a change
func checkChanges() bool {
	changed := len(lynxutil.GetLynk(client.GetLynks(), "SysTests").Files)
	result := false
	if changed != original {
		original = changed
		result = true
	}

	return result
}

// Helper function that tests the Lynx start up
// @returns nil if successful, error if unsuccessful
func testStartUp() error {
	var err error

	if startUp == 1 {
		err = testCreate()
	} else if startUp == 2 {
		err = testJoin()
	} else {
		//fmt.Println("\n----------------SimpleStart----------------")
		successful++
	}

	return err
}

// Helper function that tests the create functionality of Lynx
// @returns nil if successful, error if unsuccessful
func testCreate() error {
	//fmt.Println("----------------TestCreate----------------")
	err := client.CreateMeta("SysTests")
	tracker.CreateSwarm("SysTests")
	if err != nil {
		fmt.Println("Test failed, expected no errors. Got " + err.Error())
	} else {
		//fmt.Println("Successfully Created Lynk")
		successful++
	}

	return err
}

// Helper function that tests the create functionality of Lynx
// @returns nil if successful, error if unsuccessful
func testJoin() error {
	//fmt.Println("\n----------------TestJoin----------------")
	err := client.JoinLynk(joinPath)
	if err != nil {
		fmt.Println("Test failed, expected no errors. Got " + err.Error())
	} else {
		//fmt.Println("Successfully Joined Another Lynk")
		successful++
	}

	return err
}

// Helper function that checks a Lynx was successfully changed
// @returns True if there has been a change, False if there has not been a change
func testReceive() bool {
	// Sleeps For An Extra 30 Minutes
	time.Sleep(30 * time.Minute)

	//fmt.Println("\n----------------TestReceive----------------")
	result := checkChanges()

	if !result {
		fmt.Println("Test failed, expected to see a change in the Lynk")
	} else {
		//fmt.Println("Successfully Added A File To The Lynk")
		successful++
	}

	return result
}

// Helper function that checks a file was successfully added to a Lynk
// @returns nil if successful, error if unsuccessful
func testAdd() error {
	//fmt.Println("\n----------------TestAddFile----------------")
	err := client.AddToMetainfo(imgPath, mPath)
	err = client.UpdateMetainfo(mPath)

	if err != nil {
		fmt.Println("Test failed, expected no errors. Got " + err.Error())
	} else {
		//fmt.Println("Successfully Added A File To The Lynk")
		successful++
	}

	return err
}

// Helper function that checks a file was successfully removed from a Lynk
// @returns nil if successful, error if unsuccessful
func testRemove() error {
	//fmt.Println("\n----------------TestRemoveFile----------------")
	err := client.DeleteFile("funny.jpg", "SysTests")

	if err != nil {
		fmt.Println("Test failed, expected no errors. Got " + err.Error())
	} else {
		//fmt.Println("Successfully Removed A File From The Lynk")
		successful++
	}

	return err
}

// Helper function that tests the Lynx operation
// @returns nil if successful, error if unsuccessful
func testOp() error {
	if op != 0 {
		time.Sleep(time.Duration(opDelay) * time.Hour)
	}

	var err error

	if op == 1 {
		err = testAdd()
	} else if op == -1 {
		err = testRemove()
	} else if op == 2 {
		if !testReceive() {
			err = errors.New("Lynk Did Not Change")
		}
	} else {
		//fmt.Println("\n----------------SimpleRun----------------")
		successful++
		time.Sleep(12 * time.Hour) // Simple run lasts 12 hours by default
	}

	return err
}

// Initializes our testing env and sets up varous flags
func init() {
	delayPtr := flag.Int("delay", 0, "If delay > 0 - we will start Lynx after the delay specified")
	startUpPtr := flag.Int("startUp", 0, "How Lynx is started. 0:Simple Start, 1:Create, 2:Join")
	opPtr := flag.Int("op", 0, "Lynx's operation. -1:Remove, 0:Simple Run, 1:Add, 2:Receive Test")
	opDelayPtr := flag.Int("opDelay", 4, "How long Lynx waits to perform op - in hours.")

	flag.Parse()
	//fmt.Println("delay:", *delayPtr)
	//fmt.Println("startUp:", *startUpPtr)
	//fmt.Println("op:", *opPtr)
	//fmt.Println("opDelay:", *opDelayPtr)
	delay = *delayPtr
	startUp = *startUpPtr
	op = *opPtr
	opDelay = *opDelayPtr
}
