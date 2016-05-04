// The unit tests for our lynxutil helper functions
// @author: Michael Bruce
// @author: Max Kernchen
// @verison: 5/1/2016
package lynxutil

import (
	"fmt"
	"os/user"
	"testing"
)

// Count of the # of successful tests.
var successful = 0

// Total # of the tests.
const total = 7

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

	result := FileCopy("test.txt", "test2.txt")

	if result != nil {
		t.Error("Test failed, expected no errors. Got ", result)
	} else {
		fmt.Println("Successfully Copied File")
		successful++
	}

	// Tests that overwriting a file is fine
	result = FileCopy("test.txt", "test2.txt")

	if result != nil {
		t.Error("Test failed, expected no errors. Got ", result)
	} else {
		fmt.Println("Successfully Overwrote File")
		successful++
	}

	result = FileCopy("fake.txt", "test2.txt")

	if result == nil {
		t.Error("Test failed, expected failure due to non-existent file fake.txt. Got ", result)
	} else {
		fmt.Println("Successfully Produced Non-Existent File Error")
		successful++
	}

	result = FileCopy("nopermission.txt", "test2.txt")

	if result == nil {
		t.Error("Test failed, expected failure due to permissions on nopermission.txt. Got ", result)
	} else {
		fmt.Println("Successfully Produced File Permission Error")
		successful++
	}
}

// Unit tests for our GetIP function.
// @param *testing.T t - The wrapper for the test
func TestGetIP(t *testing.T) {
	fmt.Println("\n----------------TestGetIP----------------")

	result := GetIP()

	if result != GetIP() {
		t.Error("Test failed, expected to get IP: "+GetIP()+". Got ", result)
	} else {
		fmt.Println("Successfully Got IP Address")
		successful++
	}
}

// Unit tests for our GetLynk function.
// @param *testing.T t - The wrapper for the test
func TestGetLynk(t *testing.T) {
	fmt.Println("\n----------------TestGetIP----------------")
	testLynks := make([]Lynk, 3)
	testLynks = append(testLynks, Lynk{Name: "test"}, Lynk{Name: "cool"}, Lynk{Name: "test2"})

	lynk := GetLynk(testLynks, "nonExistent")
	if lynk != nil {
		t.Error("Test failed, expected to get nil. Got ", lynk)
	} else {
		fmt.Println("Successfully Got Nil For Non-Existent Lynk")
		successful++
	}

	lynk = GetLynk(testLynks, "cool")
	if lynk != nil && lynk.Name == "cool" {
		fmt.Println("Successfully Got 'cool' Lynk")
		successful++
	} else {
		t.Error("Test failed, expected to get cool lynk. Got ", lynk)
	}

	fmt.Println("\nSuccess on ", successful, "/", total, " tests.")
}
