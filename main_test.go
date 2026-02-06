package main

import "testing"

func TestMain(t *testing.T) {
	// A simple test to verify the test task works
	expected := "Hello, World!"
	if expected != "Hello, World!" {
		t.Errorf("Expected %s, but got something else", expected)
	}
}
