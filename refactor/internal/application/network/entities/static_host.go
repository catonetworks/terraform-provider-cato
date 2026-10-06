package entities

// StaticHost is the application data passed to the SDK adapter.
// A nil MAC address means no MAC address was configured.
type StaticHost struct {
	Name       string
	IP         string
	MacAddress *string
}
