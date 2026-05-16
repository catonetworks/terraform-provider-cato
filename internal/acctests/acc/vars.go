//go:build acctest

package acc

import (
	"encoding/json"
	"fmt"
	"os"
)

// CMAVars defines vriables for CMA tests, used as a workaround until the API is fixed
//
//	TFACC_TEST_VARS='{
//	   "private_apps": [
//	        { "name": "acctest_private_app_1", "id": "219" },
//	        { "name": "acctest_private_app_2", "id": "220" }
//	   ],
//	   "users": [ ]
//	   ...
//	}'
type CMAVars map[string][]Ref

// SkipTests defines variables for skipping tests, used as a workaround until the API is fixed
// Map [TestFunctionName]Reason
//
//	TFACC_TEST_SKIP='{
//	  "TestAccPrivAccessPolicy": "ENG-184376 - Rule has an invalid entity: SubscriptionMailingList"
//	}'
type SkipTests map[string]string

const (
	accTestVariable = "TFACC_TEST_VARS"
	accTestSkipVar  = "TFACC_TEST_SKIP"
)

var cmaVars CMAVars
var skipTests SkipTests

// initialize the test package
func init() { //nolint:gochecknoinits
	readVarsFromEnv()
	readSkipFromEnv()
}

func readVarsFromEnv() {
	v := os.Getenv(accTestVariable)
	if v == "" {
		return
	}
	err := json.Unmarshal([]byte(v), &cmaVars)
	if err != nil {
		panic(err)
	}
}

func readSkipFromEnv() {
	skipTests = make(SkipTests)
	v := os.Getenv(accTestSkipVar)
	if v == "" {
		return
	}
	err := json.Unmarshal([]byte(v), &skipTests)
	if err != nil {
		fmt.Printf("error unmarshalling skip tests variable '%s': %v\n", accTestSkipVar, err)
	}
}
