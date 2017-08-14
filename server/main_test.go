package server

import (
	"os"
	"testing"

	"github.com/ikedam/gaetest/testutil"
)

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	testutil.Setup()
	defer testutil.Teardown()
	return m.Run()
}
