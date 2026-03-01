package tests

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

func getRandName(resource string) string {
	const length = 10
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = charset[r.Intn(len(charset))]
	}
	return "test_" + resource + "_" + string(bytes)
}

func checkCMAVars(t *testing.T) func() {
	return func() {
		for _, envVar := range []string{"CATO_TOKEN", "CATO_ACCOUNT_ID", "CATO_ENDPOINT"} {
			if os.Getenv(envVar) == "" {
				t.Fatalf("ERROR: env variable '%s' not set", envVar)
			}
		}
	}
}

func getRandIP() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("10.%d.%d.%d", 2+r.Intn(252), 2+r.Intn(252), 2+r.Intn(252))
}
