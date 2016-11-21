// The unit tests for our client
// @author: Michael Bruce
// @author: Max Kernchen
// @verison: 2/17/2016
package client

import (
	"capstone/lynxutil"
	"fmt"
	"io/ioutil"
	"os/user"
	"strings"
	"testing"
)

// Count of the # of successful tests.
var successful = 0

// Total # of the tests.
const total = 19

// Gets user's home directory
var cU, _ = user.Current()

// Adds "Lynx" to home directory string
var hPath = cU.HomeDir + "/Lynx/"

// Uses homePath and our Tests Lynk to create mPath
var mPath = hPath + "Tests/meta.info"

// Unit tests for our FileCopy function.
// @param *testing.T t - The wrapper for the test
func TestFileCopy(t *testing.T) {
	fmt.Println("\n----------------TestFileCopy----------------")

	result := lynxutil.FileCopy("test.txt", "test2.txt")

	if result != nil {
		t.Error("Test failed, expected no errors. Got ", result)
	} else {
		fmt.Println("Successfully Copied File")
		successful++
	}

	// Tests that overwriting a file is fine
	result = lynxutil.FileCopy("test.txt", "test2.txt")

	if result != nil {
		t.Error("Test failed, expected no errors. Got ", result)
	} else {
		fmt.Println("Successfully Overwrote File")
		successful++
	}

	result = lynxutil.FileCopy("fake.txt", "test2.txt")

	if result == nil {
		t.Error("Test failed, expected failure due to non-existent file fake.txt. Got ", result)
	} else {
		fmt.Println("Successfully Produced Non-Existent File Error")
		successful++
	}

	result = lynxutil.FileCopy("nopermission.txt", "test2.txt")

	if result == nil {
		t.Error("Test failed, expected failure due to permissions on nopermission.txt. Got ", result)
	} else {
		fmt.Println("Successfully Produced File Permission Error")
		successful++
	}
}

// Unit tests for addToMetainfo function
// @param *testing.T t - The wrapper for the test
func TestAddToMetainfo(t *testing.T) {
	fmt.Println("\n----------------TestAddToMetainfo----------------")

	parseMetainfo(mPath)
	hasTest := false
	tLynk := lynxutil.GetLynk(lynks, "Tests")

	i := 0
	for i < len(tLynk.Files) {
		//fmt.Print(files[i].name)
		if tLynk.Files[i].Name == "test.txt" {
			//fmt.Print(files[i].name)
			hasTest = true
		}
		i++
	}

	// add test.txt to the metainfo
	result := AddToMetainfo("test.txt", mPath)

	if result != nil && !hasTest {
		t.Error("Test failed, expected no errors. Got ", result)
	} else {
		fmt.Println("Successfully Added To Metainfo")
		successful++
	}

	parseMetainfo(mPath)

	// check that test.txt is in the File struct list
	i = 0
	for i < len(tLynk.Files) {
		if tLynk.Files[i].Name == "test.txt" {
			hasTest = true
		}
		i++
	}

	if !hasTest {
		t.Error("test.txt was not added to metainfo.")
	} else {
		fmt.Println("Successfully Added test.txt To Metainfo")
		successful++
	}

	result = AddToMetainfo("test.txt", mPath)

	if result == nil {
		t.Error("Test failed, expected failure due to duplicates. Got ", result)
	} else {
		fmt.Println("Successfully Avoided Adding Duplicates")
		successful++
	}
}

// Unit tests for parseMetainfo function
// @param *testing.T t - The wrapper for the test
func TestParseMetainfo(t *testing.T) {
	fmt.Println("\n----------------TestParseMetainfo----------------")

	result := parseMetainfo("fake")

	if result == nil {
		t.Error("Test failed, expected failure due non-existent file. Got ", result)
	} else {
		fmt.Println("Successfully Produced Non-existent File Error")
		successful++
	}

	result = parseMetainfo("test.txt")

	if result == nil {
		t.Error("Test failed, expected failure due incorrect file. Got ", result)
	} else {
		fmt.Println("Successfully Produced Incorrect File Error")
		successful++
	}

	result = parseMetainfo(mPath)

	if result != nil {
		t.Error("Test failed, expected no errors. Got ", result)
	} else {
		fmt.Println("Successfully Parsed The meta.info File")
		successful++
	}
}

// Unit tests for updateMetainfo function
// @param *testing.T t - The wrapper for the test
func TestUpdateMetainfo(t *testing.T) {
	fmt.Println("\n----------------TestUpdateMetainfo----------------")

	parseMetainfo(mPath)

	result := UpdateMetainfo(mPath)

	if result != nil {
		t.Error("Test failed, expected no errors. Got ", result)
	} else {
		fmt.Println("Successfully Updated The meta.info File")
		successful++
	}

}

// Unit tests for deleteEntry function
// @param *testing.T t - The wrapper for the test
func TestDeleteFile(t *testing.T) {
	fmt.Println("\n----------------TestDeleteFiley----------------")
	failed := false

	parseMetainfo(mPath)
	lynkName := GetLynkName(mPath)
	DeleteFile("test.txt", lynkName)

	tLynk := lynxutil.GetLynk(lynks, lynkName)

	i := 0
	for i < len(tLynk.Files) {
		if tLynk.Files[i].Name == "test.txt" {
			t.Error("Error, test.txt is still in files")
			failed = true
		}
		i++
	}

	if !failed {
		fmt.Println("Successfully Removed The test.txt File")
		successful++
	} else {
		failed = false
	}

	DeleteFile("test11.txt", lynkName)

	i = 0
	for i < len(tLynk.Files) {
		if tLynk.Files[i].Name == "test11.txt" {
			t.Error("Error, test11.txt is still in files")
			failed = true
		}
		i++
	}

	if !failed {
		fmt.Println("Successfully Removed The test11.txt File")
		successful++
	}
}

// Unit tests for getFile function
// @param *testing.T t - The wrapper for the test
func TestGetFile(t *testing.T) {
	fmt.Println("\n----------------TestGetFile----------------")

	err := getFile("test.txt", mPath)

	if err != nil {
		t.Error(err.Error())
	} else {
		fmt.Println("Successfully Got File 'test.txt'")
		successful++
	}

	err = getFile("non-existent.txt", mPath)

	if err != nil {
		fmt.Println("Successfully Produced Non-Existent File Error")
		successful++
	} else {
		t.Error(err.Error())
	}

}

// Unit tests for HaveFile function
// @param *testing.T t - The wrapper for the test
func TestHaveFile(t *testing.T) {
	fmt.Println("\n----------------TestHaveFile----------------")

	result := HaveFile("Tests/test.txt")

	if result {
		fmt.Println("Successfully Found 'test.txt'")
		successful++
	} else {
		t.Error("Could Not Find 'test.txt'")
	}

	result = HaveFile("Tests/non-existent.txt")

	if !result {
		fmt.Println("Successfully Produced False For Non-Existent File")
		successful++
	} else {
		t.Error("Incorrectly Found A Non-Existent File")
	}

}

// Unit tests for GetTrackerIP function
// @param *testing.T t - The wrapper for the test
func TestGetTracker(t *testing.T) {
	fmt.Println("\n----------------TestGetTracker----------------")

	ip := GetTracker(mPath) // Should be 127.0.0.1 during testing
	content, _ := ioutil.ReadFile(mPath)
	s := string(content)

	if strings.Contains(s, ip) {
		fmt.Println("Successfully Got Tracker")
		successful++
	} else {
		t.Error("Found Incorrect Tracker IP: " + ip)
	}
}

// Unit tests for GetTrackerIP function
// @param *testing.T t - The wrapper for the test
func TestAskTrackerForPeers(t *testing.T) {
	fmt.Println("\n----------------TestAskTracker----------------")

	lynkName := GetLynkName(mPath)
	lynk := lynxutil.GetLynk(lynks, lynkName)
	askTrackerForPeers(lynkName)

	if len(lynk.Peers) <= 0 {
		t.Error("Did Not Get Correct List Of Peers")
	} else {
		fmt.Println("Successfully Got Peers")
		successful++
	}

	fmt.Println("\nSuccess on ", successful, "/", total, " tests.")
}
