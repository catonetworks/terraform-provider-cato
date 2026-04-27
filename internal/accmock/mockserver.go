//go:build acctest

/*
Package accmock provides a mock HTTP server for testing Terraform providers that interact with APIs.
The mock server is designed to simulate API responses based on predefined fixtures, allowing for consistent and repeatable tests
without relying on actual external services.

Key features of the accmock package include:
  - Configurable operations: Define the expected API operations (CREATE, READ, UPDATE, DELETE) and their corresponding request
    and response structures in a YAML configuration file.
  - Fixture-based responses: Store mock responses in YAML files organized by resource type, name/ID, and version,
    enabling easy management of test scenarios.
  - State management: The mock server maintains an internal state of resources, allowing it to simulate realistic API behavior
    such as resource creation, updates, and deletions.
  - Request recording: All incoming requests and outgoing responses are recorded for later verification in tests.

The AccMock runs a mock web server, which
checks GraphQL operation (e.g. privateAppCreatePrivateApp), and based on the config for that operation (from config.yaml), it checks
if it's CREATE, READ, UPDATE or DELETE type, and how to extract resource name and ID from the request body.
  - CREATE: creates a resource in memory, version 0; returns response from file {resourceType}/{resourceName}/000_create.yaml
  - READ: gets current version from memory, returns file {resourceType}/{resourceName}/{version}_read.yaml
  - UPDATE:  gets current version from memory, returns file {resourceType}/{resourceName}/{version}_update.yaml;
    increments the version in memory
  - DELETE:  gets current version from memory, returns file {resourceType}/{resourceName}/{version}_zap.yaml
*/
package accmock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"go.yaml.in/yaml/v3"
)

var ACCMockActive bool

type MockServer struct {
	calls       []Call
	resources   resourceStates
	baseDir     string
	recordedDir string
	server      *httptest.Server
	cfg         *config
	t           *testing.T
	mu          sync.Mutex
}

type resourceStates map[string]resourceState

type Call struct {
	RequestBody  []byte
	ResponseBody []byte
	Error        error
}

type config struct {
	Operations map[string]Operation `yaml:"Operations"`
}

type Operation struct {
	Type         string                   `yaml:"Type"`
	Resource     string                   `yaml:"Resource"`
	ResourcePath string                   `yaml:"ResourcePath"`
	IDPath       string                   `yaml:"IDPath"`
	NamePath     string                   `yaml:"NamePath"`
	Static       bool                     `yaml:"Static"`
	Subtypes     map[string]OperationType `yaml:"Subtypes"`
}

const (
	OpCreate    = "CREATE"
	OpRead      = "READ"
	OpUpdate    = "UPDATE"
	OpDelete    = "DELETE"
	OpNoContent = "NO_CONTENT"
)

var operationTypes = []string{OpCreate, OpRead, OpUpdate, OpDelete, OpNoContent}

type OperationType struct {
	Static bool `yaml:"Static"`
}

type resourceState struct {
	ID      string
	Name    string
	Version int
}

type mockFixture struct {
	ResourceID string  `yaml:"ResourceID"`
	GraphQL    graphql `yaml:"GraphQL"`
}

type graphql struct {
	StatusCode int           `yaml:"StatusCode"`
	Delay      time.Duration `yaml:"Delay"`
	Body       any           `yaml:"Body"`
}

func NewMockServer(t *testing.T, testName string) *MockServer {
	projectRoot := getProjectRoot()
	baseDir := filepath.Join(projectRoot, "test_data", testName)
	recordedDir := filepath.Join(projectRoot, "tmp_recorded", testName)
	mockServer := &MockServer{
		resources:   make(map[string]resourceState),
		baseDir:     baseDir,
		recordedDir: recordedDir,
		t:           t,
	}
	return mockServer
}

func (s *MockServer) Run() {
	if s == nil {
		return
	}
	if err := os.Setenv("TF_API_DUMP_DIR", s.recordedDir); err != nil {
		s.t.Fatalf("failed to set env TF_API_DUMP_DIR: %v", err)
	}
	_ = os.RemoveAll(s.recordedDir)

	if !ACCMockActive {
		return
	}

	cfg, err := readConfig(filepath.Join(s.baseDir, "config.yaml"))
	if err != nil {
		s.t.Fatalf("failed to read config: %v", err)
		return
	}

	s.cfg = cfg
	s.server = httptest.NewServer(s)

	if err := os.Setenv("CATO_BASEURL", s.URL()); err != nil {
		s.t.Fatalf("failed to set env CATO_BASEURL: %v", err)
	}
}

// URL returns the base URL of the mock server, which can be used to configure the API client in tests.
func (s *MockServer) URL() string { return s.server.URL }

// Calls returns a copy of the recorded calls
func (s *MockServer) Calls() []Call {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Call(nil), s.calls...)
}

// processGraphQLRequest finds the mock response and returns it
// - parse request body, find graphql operation name
// - find given operation in config, learn its op-type, resource, etc. If not found return error
//   - if ResourcePath is defined, use Subtypes (e.g. for entityLookup/location)
//
// - handle CREATE, READ, UPDATE, DELETE operations:
func (s *MockServer) processGraphQLRequest(data []byte) (*mockFixture, error) {
	operationName, parsedData, err := getItem(data, "operationName")
	if err != nil {
		return nil, fmt.Errorf("failed to parse graphql request body: %w", err)
	}

	operation, ok := s.cfg.Operations[operationName]
	if !ok {
		return nil, fmt.Errorf("operation %q not found in config", operationName)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	switch operation.Type {
	case OpCreate:
		return s.processCreate(operationName, operation, parsedData)
	case OpRead:
		return s.processRead(operationName, operation, parsedData)
	case OpUpdate:
		return s.processUpdate(operationName, operation, parsedData)
	case OpDelete:
		return s.processDelete(operationName, operation, parsedData)
	case OpNoContent:
		return &mockFixture{GraphQL: graphql{StatusCode: http.StatusNoContent}}, nil
	default:
		return nil, fmt.Errorf("unsupported operation type %q for %q", operation.Type, operationName)
	}
}

// processCreate handles CREATE operations by:
//   - extract resource name from request body parsedData
//   - find response file {resourceType}/{resourceName}/000_create.yaml
//   - create the object in s.Resources with ID and Name from the response file;
//     if it already exists, return error
//   - return the response body from the file
func (s *MockServer) processCreate(operationName string, op Operation, parsedData any) (*mockFixture, error) {
	resourceName, err := extractItem(parsedData, op.NamePath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract resource name: %w", err)
	}
	if resourceName == "" {
		return nil, fmt.Errorf("cannot find resource name for operation %q", operationName)
	}
	if _, exists := s.resources[resourceName]; exists {
		return nil, fmt.Errorf("%s resource %q already exists", operationName, resourceName)
	}

	fixturePath := s.fixturePath(op.Resource, resourceName, "", 0, "create")
	fixture, err := s.readFixture(fixturePath)
	if err != nil {
		return nil, err
	}
	if fixture.ResourceID == "" {
		return nil, fmt.Errorf("%s fixture %q does not include ResourceID", operationName, fixturePath)
	}

	s.resources[resourceName] = resourceState{ID: fixture.ResourceID, Name: resourceName, Version: 0}
	return fixture, nil
}

// processRead handles READ operations,
//   - extract resource Name or ID from request parsedData
//   - extract resource subtype from parsedData if ResourcePath is defined
//   - if Static == true: use version 000, otherwise use version defined in s.Resources
//   - find response file {resourceType or Subtype}/{resourceName or ID}/{version}_read.yaml
//     if resourceName or ID is not found, use "default"
func (s *MockServer) processRead(operationName string, op Operation, parsedData any) (fixture *mockFixture, err error) {
	resourceID, resourceName := s.findIDName(op, parsedData)

	// extract resource type
	resourceType := op.Resource
	isStatic := op.Static
	if resourceType == "" { // generic entity lookup, find resource type in the request body
		if resourceType, err = extractItem(parsedData, op.ResourcePath); err != nil {
			return nil, fmt.Errorf("%s failed to extract resource type: %w", operationName, err)
		}
		subtypeInfo, ok := op.Subtypes[resourceType]
		if !ok {
			return nil, fmt.Errorf("%s unsupported resource sub-type %q extracted from path %q",
				operationName, resourceType, op.ResourcePath)
		}
		isStatic = subtypeInfo.Static
	}

	// figure out resource name and version
	resourceVersion := 0
	if !isStatic {
		resourceState, err := s.getCurrentResource(operationName, resourceID, resourceName)
		if err != nil {
			return nil, err
		}
		resourceName, resourceVersion = resourceState.Name, resourceState.Version
	}

	return s.readFixture(s.fixturePath(resourceType, resourceName, resourceID, resourceVersion, "read"))
}

// processUpdate handles UPDATE operations,
//   - extract resource Name or ID from request body
//   - find the resource in s.Resources by ID or Name and get its version
//   - read response file {resourceType}/{resourceName}/{version}_update.yaml
//   - increment the version
func (s *MockServer) processUpdate(operationName string, op Operation, parsedData any) (fixture *mockFixture, err error) {
	resourceID, resourceName := s.findIDName(op, parsedData)

	resourceVersion := 0
	if !op.Static {
		resourceState, err := s.getCurrentResource(operationName, resourceID, resourceName)
		if err != nil {
			return nil, err
		}
		resourceName, resourceVersion = resourceState.Name, resourceState.Version
		resourceState.Version++
		s.resources[resourceName] = resourceState
	}

	return s.readFixture(s.fixturePath(op.Resource, resourceName, resourceID, resourceVersion, "update"))
}

// processDelete handles DELETE operations,
//   - extract resource Name or ID from request body
//   - find the resource in s.Resources by ID or Name and get its version
//   - read response file {resourceType}/{resourceName}/{version}_zap.yaml
//   - remove from s.Resources
func (s *MockServer) processDelete(operationName string, op Operation, parsedData any) (fixture *mockFixture, err error) {
	resourceID, resourceName := s.findIDName(op, parsedData)

	resourceVersion := 0
	if !op.Static {
		resourceState, err := s.getCurrentResource(operationName, resourceID, resourceName)
		if err != nil {
			return nil, err
		}
		resourceName, resourceVersion = resourceState.Name, resourceState.Version
		delete(s.resources, resourceName)
	}

	return s.readFixture(s.fixturePath(op.Resource, resourceName, resourceID, resourceVersion, "zap"))
}

func (s *MockServer) findIDName(op Operation, parsedData any) (resourceID, resourceName string) {
	if op.IDPath != "" {
		resourceID, _ = extractItem(parsedData, op.IDPath)
	}
	if resourceID == "" && op.NamePath != "" {
		resourceName, _ = extractItem(parsedData, op.NamePath)
	}
	return resourceID, resourceName
}

// getCurrentResource find the resource in s.Resources by ID or Name and return its state; if not found, return error
func (s *MockServer) getCurrentResource(operationName, resourceID, resourceName string) (resState resourceState, err error) {
	var ok bool

	switch {
	case resourceName != "":
		resState, ok = s.resources.ByName(resourceName)
		if !ok {
			return resourceState{}, fmt.Errorf("%s resource with name %q not found", operationName, resourceName)
		}
	case resourceID != "":
		resState, ok = s.resources.ByID(resourceID)
		if !ok {
			return resourceState{}, fmt.Errorf("%s resource with ID %q not found", operationName, resourceID)
		}
	default:
		return resourceState{}, fmt.Errorf("%s for non-static read ID or Name must be provided", operationName)
	}
	return resState, nil
}

// fixturePath constructs the path to the fixture file based on the resource type, name, ID, version, and action.
// {baseDir}/{resourceType}/{resourceName or ID}/{version}_{action}.yaml
func (s *MockServer) fixturePath(resourceType, resourceName, resourceID string, version int, action string) string {
	nameIDDirectory := "default"
	switch {
	case resourceName != "":
		nameIDDirectory = resourceName
	case resourceID != "":
		nameIDDirectory = resourceID
	}
	fileName := fmt.Sprintf("%03d_%s.yaml", version, action)
	return filepath.Join(s.baseDir, resourceType, nameIDDirectory, fileName)
}

// readFixture reads the fixture file at the specified path and unmarshals it into a mockFixture struct.
func (s *MockServer) readFixture(path string) (*mockFixture, error) {
	contents, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("read fixture %q: %w", path, err)
	}

	var fixture mockFixture
	if err := yaml.Unmarshal(contents, &fixture); err != nil {
		return nil, fmt.Errorf("unmarshal fixture %q: %w", path, err)
	}

	return &fixture, nil
}

// Close shuts down the mock server and releases any associated resources.
func (s *MockServer) Close() {
	if s != nil && s.server != nil {
		s.server.Close()
	}
}

// ServeHTTP handles incoming HTTP requests to the mock server,
// processes them according to the defined operations, and records the calls for later verification.
func (s *MockServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request, err := io.ReadAll(r.Body)
	defer func() { _ = r.Body.Close() }()
	if err != nil {
		s.fail(w, fmt.Errorf("failed to read request body: %w", err))
		return
	}

	fixture, err := s.processGraphQLRequest(request)
	if err != nil {
		s.fail(w, fmt.Errorf("failed to process GraphQL request: %w", err), request)
		return
	}
	body, err := json.Marshal(fixture.GraphQL.Body)
	if err != nil {
		s.fail(w, fmt.Errorf("failed to marshal fixture body: %w", err), request)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if fixture.GraphQL.StatusCode == 0 {
		fixture.GraphQL.StatusCode = http.StatusOK
	}
	w.WriteHeader(fixture.GraphQL.StatusCode)
	if fixture.GraphQL.Delay != 0 {
		time.Sleep(fixture.GraphQL.Delay)
	}

	_, err = w.Write(body)

	s.addCall(Call{
		RequestBody:  request,
		ResponseBody: body,
		Error:        err,
	})
}

// fail records the call and responds with http error
func (s *MockServer) fail(w http.ResponseWriter, err error, requestResponse ...[]byte) {
	call := Call{Error: err}
	if len(requestResponse) > 0 {
		call.RequestBody = requestResponse[0]
	}
	if len(requestResponse) > 1 {
		call.ResponseBody = requestResponse[1]
	}
	s.addCall(call)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// addCall safely appends a call to the mock server's call log.
func (s *MockServer) addCall(c Call) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, c)
}

// readConfig reads the mock server configuration from a YAML file and returns a Config struct.
func readConfig(path string) (*config, error) {
	contents, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	var cfg config
	if err := yaml.Unmarshal(contents, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config %q: %w", path, err)
	}

	// validate config
	for name, operation := range cfg.Operations {
		if operation.Type == OpNoContent {
			continue // skip validation for NO_CONTENT operations
		}
		if !slices.Contains(operationTypes, operation.Type) {
			return nil, fmt.Errorf("invalid operation type %q for operation %q; expected one of %s",
				operation.Type, name, strings.Join(operationTypes, ", "))
		}

		hasResource := operation.Resource != ""
		hasResourcePath := operation.ResourcePath != ""
		if hasResource == hasResourcePath {
			return nil, fmt.Errorf("invalid config for operation %q: exactly one of Resource and ResourcePath must be set", name)
		}

		if hasResourcePath && len(operation.Subtypes) == 0 {
			return nil, fmt.Errorf("invalid config for operation %q: Subtypes must be set when ResourcePath is used", name)
		}
	}

	return &cfg, nil
}

func (r resourceStates) ByName(name string) (resourceState, bool) {
	rs, ok := r[name]
	return rs, ok
}

func (r resourceStates) ByID(id string) (resourceState, bool) {
	for _, rs := range r {
		if rs.ID == id {
			return rs, true
		}
	}
	return resourceState{}, false
}

func getProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0) //nolint:dogsled
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

func init() { //nolint:gochecknoinits
	if m := os.Getenv("TF_ACC_MOCK"); m == "1" || m == "true" || m == "yes" {
		ACCMockActive = true
	}
}
