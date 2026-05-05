//go:build acctest

package acc

import (
	"encoding/json"
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
type CMAVars struct {
	PrivateApps []Ref `json:"private_apps"`
	Users       []Ref `json:"users"`
}

const (
	accTestVariable = "TFACC_TEST_VARS"
)

var cmaVars CMAVars

// initialize the test package
func init() { //nolint:gochecknoinits
	readVarsFromEnv()
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
