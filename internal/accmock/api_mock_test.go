package accmock

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockItem struct {
	name    string
	request string
	exp     string
}

var mockItems = []mockItem{
	{
		name:    "createMock",
		request: `{"operationName":"createMock","variables":{"newMock":{"name":"res01"}}}`,
		exp:     `{"data":{"mock":{"addMock":{"mock":{"id":"1000"}}}}}`,
	},
	{
		name:    "readMock 1",
		request: `{"operationName":"readMock","variables":{"ref":{"input":1000}}}`,
		exp:     `{"data":{"mock":{"groupName":"example-group","id":"1000"}}}`,
	},
	{
		name:    "updateMock 2",
		request: `{"operationName":"updateMock","variables":{"updateMock":{"id":"1000"}}}`,
		exp:     `{"data":{"mock":{"updateMock":{"mock":{"id":"1000"}}}}}`,
	},
	{
		name:    "readMock 2",
		request: `{"operationName":"readMock","variables":{"ref":{"input":"1000"}}}`,
		exp:     `{"data":{"mock":{"groupName":"new-group","id":"1000"}}}`,
	},
	{
		name:    "deleteMock",
		request: `{"operationName":"deleteMock","variables":{"ref":{"input":"1000"}}}`,
		exp:     `{"data":{"mock":{"removeMock":{"mock":{"id":"1000"}}}}}`,
	},
}

func TestMockServer(t *testing.T) {
	t.Parallel()

	server, err := RunMockServer(testMockDataDir())
	if err != nil {
		t.Fatalf("RunMockServer() error = %v", err)
	}
	defer server.Close()

	for _, tc := range mockItems {
		t.Run(tc.name, func(t *testing.T) {
			res, err := http.Post(server.URL(), "application/json", strings.NewReader(tc.request))
			require.NoError(t, err, "failed to send request for %q", tc.name)
			got, err := io.ReadAll(res.Body)
			require.NoError(t, err, "failed to read response body for %q", tc.name)
			if err := res.Body.Close(); err != nil {
				t.Logf("failed to close response body for %q: %v", tc.name, err)
			}
			assert.Equal(t, tc.exp, string(got))
		})
	}
}

func testMockDataDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(dir, "data", "mockTest")
}
