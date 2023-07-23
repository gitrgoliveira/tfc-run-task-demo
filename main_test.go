package main

import (
	"reflect"
	"testing"
)

func TestReadRegexPatterns(t *testing.T) {
	// Test case 1: Test with a valid file path
	filePath := "patternsFile.txt"
	expectedPatterns := []string{
		"(?s)provisioner\\s+\"remote-exec\"\\s+\\{.*?\\}",
		"(?s)provisioner\\s+\"local-exec\"\\s+\\{.*?\\}",
		"(?s)data\\s+\"external\"\\s+\\{.*?\\}",
		"(?s)data\\s+\"http\\s+\\{.*?\\}",
		"(?s)resource\\s+\"toolbox\\s+\\{.*?\\}",
	}
	patterns, err := readRegexPatterns(filePath)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(patterns, expectedPatterns) {
		t.Errorf("Expected patterns %v, but got %v", expectedPatterns, patterns)
	}

	// Test case 2: Test with an empty file
	filePath = "patternsFile_empty.txt"
	expectedPatterns = []string{}
	patterns, err = readRegexPatterns(filePath)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(patterns, expectedPatterns) {
		t.Errorf("Expected patterns %v, but got %v", expectedPatterns, patterns)
	}

	// Test case 3: Test with a non-existent file
	filePath = "/path/to/non_existent_file.txt"
	expectedError := "open /path/to/non_existent_file.txt: no such file or directory" //os.ErrNotExist
	_, err = readRegexPatterns(filePath)
	if err == nil {
		t.Errorf("Expected error %v, but got nil", expectedError)
	} else if err.Error() != expectedError {
		t.Errorf("Expected error %v, but got %v", expectedError, err)
	}
}
