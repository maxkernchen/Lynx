package main

import (
	"testing"
)

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