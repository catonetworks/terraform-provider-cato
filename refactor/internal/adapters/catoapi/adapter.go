package catoapi

import cato "github.com/catonetworks/cato-go-sdk"

// Adapter wraps the SDK. Use cases consume only their required method subset.
type Adapter struct{ client *cato.Client }

func NewAdapter(client *cato.Client) *Adapter { return &Adapter{client: client} }
