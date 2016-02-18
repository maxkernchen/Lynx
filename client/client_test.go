package main

import (
	"testing"
)

/**
 * Unit tests for our filecopy function.
 */
func TestFileCopy(t *testing.T) {
	result := fileCopy("test.txt", "test2.txt")

	if result != nil {
		t.Error("Test failed, expected no errors. Got ", result)
	}

	// Tests that overwriting a file is fine
	result = fileCopy("test.txt", "test2.txt")

	if result != nil {
		t.Error("Test failed, expected no errors. Got ", result)
	}

	result = fileCopy("fake.txt", "test2.txt")

	if result == nil {
		t.Error("Test failed, expected failure due to non-existent file fake.txt. Got ", result)
	}

	result = fileCopy("nopermission.txt", "test2.txt")

	if result == nil {
		t.Error("Test failed, expected failure due to permissions on nopermission.txt. Got ", result)
	}
}

/**
 * Unit tests for addToMetainfo function
 */
func TestAddToMetainfo(t *testing.T) {
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
	}

	result = addToMetainfo("test.txt", "meta.info")

	if result == nil {
		t.Error("Test failed, expected failure due to duplicates. Got ", result)
	}
}

/**
 * Unit tests for parseMetainfo function
 */
func TestParseMetainfo(t *testing.T) {
	result := parseMetainfo("fake")

	if result == nil {
		t.Error("Test failed, expected failure due non-existent file. Got ", result)
	}

	result = parseMetainfo("test.txt")

	if result == nil {
		t.Error("Test failed, expected failure due incorrect file. Got ", result)
	}

	result = parseMetainfo("meta.info")

	if result != nil {
		t.Error("Test failed, expected no errors. Got ", result)
	}
}

/**
 * Unit tests for updateMetainfo function
 */
func TestUpdateMetainfo(t *testing.T) {
	parseMetainfo("meta.info")

	result := updateMetainfo()

	if result != nil {
		t.Error("Test failed, expected no errors. Got ", result)
	}
}

/**
 * Unit tests for deleteEntry function
 */
func TestDeleteEntry(t *testing.T) {

	parseMetainfo("meta.info")

	deleteEntry("test.txt")

	i := 0
	for i < len(files) {
		if files[i].name == "test.txt" {
			t.Error("Error, test.txt is still in files")
		}
		i++
	}

	deleteEntry("test11.txt")

	i = 0
	for i < len(files) {
		if files[i].name == "test11.txt" {
			t.Error("Error, test.txt is still in files")
		}
		i++
	}

}
