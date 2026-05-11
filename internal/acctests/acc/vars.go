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
type CMAVars map[string][]Ref

// type CMAVars struct {
// 	Users            []Ref `json:"users"`
// 	GlobalIPRanges   []Ref `json:"global_ip_ranges"`
// 	FloatingRanges   []Ref `json:"floating_ranges"`
// 	UserGroups       []Ref `json:"user_groups"`
// 	SystemGroups     []Ref `json:"system_groups"`
// 	DevicePostures   []Ref `json:"device_postures"`
// 	CustomApps       []Ref `json:"custom_apps"`
// 	CustomCategories []Ref `json:"custom_categories"`
// }

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
