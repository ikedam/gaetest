package testutil

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	Setup()
	defer Teardown()
	return m.Run()
}
