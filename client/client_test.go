/**
 *
 *	The unit tests for our client
 *
 *	 @author: Michael Bruce
 *	 @author: Max Kernchen
 *
 *	 @verison: 2/17/2016
 */

package main

import (
	"fmt"
	"testing"
)

/** Count of the # of successful tests. */
var successful = 0

/**
 * Unit tests for our filecopy function.
 */
func TestFileCopy(t *testing.T) {
	fmt.Println("\n----------------TestFileCopy----------------")

	result := fileCopy("test.txt", "test2.txt")

	if result != nil {
		t.Error("Test failed, expected no errors. Got ", result)
	} else {
		fmt.Println("Successfully Copied File")
		successful++
	}

	// Tests that overwriting a file is fine
	result = fileCopy("test.txt", "test2.txt")

	if result != nil {
		t.Error("Test failed, expected no errors. Got ", result)
	} else {
		fmt.Println("Successfully Overwrote File")
		successful++
	}

	result = fileCopy("fake.txt", "test2.txt")

	if result == nil {
		t.Error("Test failed, expected failure due to non-existent file fake.txt. Got ", result)
	} else {
		fmt.Println("Successfully Produced Non-Existent File Error")
		successful++
	}

	result = fileCopy("nopermission.txt", "test2.txt")

	if result == nil {
		t.Error("Test failed, expected failure due to permissions on nopermission.txt. Got ", result)
	} else {
		fmt.Println("Successfully Produced File Permission Error")
		successful++
	}
}

/**
 * Unit tests for addToMetainfo function
 */
func TestAddToMetainfo(t *testing.T) {
	fmt.Println("\n----------------TestAddToMetainfo----------------")

	parseMetainfo("meta.info")
	hasTest := false

	i := 0
	for i < len(files) {
		//fmt.Print(files[i].name)
		if files[i].name == "test.txt" {
			//fmt.Print(files[i].name)
			hasTest = true
		}
		i++
	}

	// add test.txt to the metainfo
	result := addToMetainfo("test.txt", "meta.info")

	if result != nil && !hasTest {
		t.Error("Test failed, expected no errors. Got ", result)
	} else {
		fmt.Println("Successfully Added To Metainfo")
		successful++
	}

	parseMetainfo("meta.info")

	// check that test.txt is in the File struct list
	i = 0
	for i < len(files) {
		if files[i].name == "test.txt" {
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

	result = addToMetainfo("test.txt", "meta.info")

	if result == nil {
		t.Error("Test failed, expected failure due to duplicates. Got ", result)
	} else {
		fmt.Println("Successfully Avoided Adding Duplicates")
		successful++
	}
}

/**
 * Unit tests for parseMetainfo function
 */
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

	result = parseMetainfo("meta.info")

	if result != nil {
		t.Error("Test failed, expected no errors. Got ", result)
	} else {
		fmt.Println("Successfully Parsed The meta.info File")
		successful++
	}
}

/**
 * Unit tests for updateMetainfo function
 */
func TestUpdateMetainfo(t *testing.T) {
	fmt.Println("\n----------------TestUpdateMetainfo----------------")

	parseMetainfo("meta.info")

	result := updateMetainfo()

	if result != nil {
		t.Error("Test failed, expected no errors. Got ", result)
	} else {
		fmt.Println("Successfully Updated The meta.info File")
		successful++
	}

}

/**
 * Unit tests for deleteEntry function
 */
func TestDeleteEntry(t *testing.T) {
	fmt.Println("\n----------------TestDeleteEntry----------------")
	failed := false

	parseMetainfo("meta.info")
	deleteEntry("test.txt")

	i := 0
	for i < len(files) {
		if files[i].name == "test.txt" {
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

	deleteEntry("test11.txt")

	i = 0
	for i < len(files) {
		if files[i].name == "test11.txt" {
			t.Error("Error, test11.txt is still in files")
			failed = true
		}
		i++
	}

	if !failed {
		fmt.Println("Successfully Removed The test11.txt File")
		successful++
	} else {
		failed = false
	}

}
