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
var imgPath = cU.HomeDir + "funny.jpg"

// Path to the joining meta.info file
var joinPath = cU.HomeDir + "meta.info"

// Count of the # of successful tests.
var successful = 0

// Count of the # of original files in systemtests.
var original = 0

// Total # of the tests.
const total = 3

// If delay > 0 - we will start Lynx after the delay specified
var delay int

// Whether we create or join. True = Create | False = Join
var create bool

// Unit tests for our Encrypt and Decrypt functions.
// @param *testing.T t - The wrapper for the test
func TestMain(t *testing.T) {
	defer func() {
		if e := recover(); e != nil {
			// e is the interface{} typed-value we passed to panic()
			t.Error(e) // Prints "Whoops: boom!"
		}
	}()

	var wait1 time.Duration = 12
	var wait2 time.Duration = 16

	if delay > 0 {
		time.Sleep(time.Duration(delay) * time.Minute) // Waits X amount of time and then continues
	}

	go launch() // Fires up web server

	if create { // If we want to create
		fmt.Println("----------------TestCreate----------------")
		err := client.CreateMeta("SysTests")
		tracker.CreateSwarm("SysTests")
		if err != nil {
			t.Error("Test failed, expected no errors. Got ")
		} else {
			fmt.Println("Successfully Created Lynk")
			successful++
		}

		time.Sleep(wait1 * time.Hour) // Waits X amount of time and then continues

		fmt.Println("\n----------------TestAddFile----------------")
		err = client.AddToMetainfo(imgPath, mPath)
		err = client.UpdateMetainfo(mPath)

		if err != nil {
			t.Error("Test failed, expected no errors. Got ")
		} else {
			fmt.Println("Successfully Added A File To The Lynk")
			successful++
		}

		time.Sleep(wait2 * time.Hour) // Waits X amount of time and then continues

		fmt.Println("\n----------------TestRemoveFile----------------")
		client.DeleteFile("funny.jpg", "SysTests")

		if err != nil {
			t.Error("Test failed, expected no errors. Got ")
		} else {
			fmt.Println("Successfully Removed A File From The Lynk")
			successful++
		}

	} else {
		fmt.Println("\n----------------TestJoin----------------")
		err := client.JoinLynk(mPath)
		if err != nil {
			t.Error("Test failed, expected no errors. Got ")
		} else {
			fmt.Println("Successfully Joined Another Lynk")
			successful++
		}

		// Adds an Hour to time1
		time.Sleep(wait1*time.Hour + time.Hour)
		fmt.Println("\n----------------TestReceivedAddFile----------------")

		if checkChanges() {
			t.Error("Test failed, expected no errors. Got ")
		} else {
			fmt.Println("Successfully Added A File To The Lynk")
			successful++
		}

		// Adds an Hour to time2
		time.Sleep(wait2*time.Hour + time.Hour)
		fmt.Println("\n----------------TestReceivedRemoveFile----------------")

		if checkChanges() {
			t.Error("Test failed, expected no errors. Got ")
		} else {
			fmt.Println("Successfully Removed A File From The Lynk")
			successful++
		}

	}

	if delay <= 0 {
		fmt.Println("\nSuccess on ", successful, "/", total, " tests.")
	} else {
		fmt.Println("Delayed: ")
	}
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

// Initializes our testing env and sets up varous flags
func init() {
	numbPtr := flag.Int("delay", 0, "If delay > 0 - we will start Lynx after the delay specified")
	boolPtr := flag.Bool("create", true, "Whether we create or join. True = Create | False = Join")

	flag.Parse()
	fmt.Println("delay:", *numbPtr)
	fmt.Println("create:", *boolPtr)
	delay = *numbPtr
	create = *boolPtr
}
