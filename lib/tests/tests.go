// Copyright (c) 2019 Daniel Oaks <daniel@danieloaks.net>
// released under the MIT license

package tests

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/ircdocs/ircts/lib/conn"
	"github.com/ircdocs/ircts/lib/utils"
)

// TestReturn is a return value for a Test.
type TestReturn int

const (
	// Success indicates the test succeeded.
	Success TestReturn = iota
	// Failure indicates the test failed (due to server noncompliance).
	Failure
	// NotApplicable indicates the server does not fulfil the requirements of the test,
	// for example not implementing a required client capability or no accounts being defined.
	NotApplicable
)

var (
	// TestReturnStrings returns an appropriate string for a TestReturn.
	TestReturnStrings = map[TestReturn]string{
		Success:       "Success",
		Failure:       "Failed",
		NotApplicable: "N/A",
	}
)

// TestGroupCollection holds multiple TestGroups.
type TestGroupCollection []TestGroup

// TestGroup is a collection of related tests.
type TestGroup struct {
	// Name of the test collection.
	Name string

	// Description contains a summary of what the tests in this group do.
	Description string

	// Tests that are a part of this group.
	Tests []Test
}

// Test is a single test that's done on a server.
type Test struct {
	// Name to use for this specific test (TestGroup name is automatically prepended).
	Name string

	// Description contains a summary of what this test does.
	Description string

	// ClientsRequiredAtStart is the number of clients required on test start.
	ClientsRequiredAtStart int

	// RequiredCaps are the required client capabilities for this test to run.
	RequiredCaps []string

	// Handler is the function that runs the test and returns the result.
	Handler func(name string, rm *RunManager) bool
}

// TestResults holds test results.
type TestResults struct {
	resultCode    map[string]TestReturn
	resultText    map[string]string
	resultTraffic map[string]string
}

// NewTestResults returns a new TestResults.
func NewTestResults() *TestResults {
	var tr TestResults
	tr.resultCode = make(map[string]TestReturn)
	tr.resultText = make(map[string]string)
	tr.resultTraffic = make(map[string]string)
	return &tr
}

// Set sets a single test's results.
func (tr *TestResults) Set(name string, code TestReturn, text, traffic string) {
	tr.resultCode[name] = code
	tr.resultText[name] = text
	tr.resultTraffic[name] = text
}

// Print shows the results of a single test.
func (tr *TestResults) Print(name string) {
	result := tr.resultCode[name]
	fmt.Println(name, "-", TestReturnStrings[result])
	fmt.Println("  ", tr.resultText[name])
	if result == Failure {
		fmt.Println("---")
		fmt.Println(tr.resultTraffic[name])
		fmt.Println("---")
	}
}

// RunManager holds info to help run tests.
type RunManager struct {
	Config  *utils.Config
	Pool    *conn.ConnectionPool
	Results TestResults
}

// NewRunManager returns a RunManager.
func NewRunManager() *RunManager {
	var rm RunManager
	rm.Results = *NewTestResults()
	return &rm
}

// NewNick returns a new nickname.
func (rm *RunManager) NewNick() string {
	len := 9
	buff := make([]byte, len)
	rand.Read(buff)
	str := base64.StdEncoding.EncodeToString(buff)
	// Base 64 can be longer than len
	return str[:len]
}

// AllTests contains all of our test groups.
var AllTests TestGroupCollection

func init() {
	AllTests = TestGroupCollection{
		saslTests,
	}
}
