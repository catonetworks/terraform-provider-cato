package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Resp struct {
	Data Data `json:"data"`
}
type Data struct {
	ZtnaAppConnector ztnaAppConnector `json:"ztnaAppConnector"`
}

type ztnaAppConnector struct {
	ZtnaAppConnectorList ztnaAppConnectorList `json:"ztnaAppConnectorList"`
}

type ztnaAppConnectorList struct {
	ZtnaAppConnector []Connector `json:"ztnaAppConnector"`
}

type Connector struct {
	ID                   string               `json:"id,omitempty"`
	Name                 string               `json:"name"`
	Description          *string              `json:"description,omitempty"`
	SerialNumber         *string              `json:"serialNumber,omitempty"`
	SocketModel          *string              `json:"socketModel,omitempty"`
	Type                 string               `json:"type,omitempty"`
	GroupName            string               `json:"groupName,omitempty"`
	Address              Address              `json:"address,omitempty"`
	Timezone             string               `json:"timezone,omitempty"`
	PreferredPopLocation PreferredPopLocation `json:"preferredPopLocation,omitempty"`
	PrivateAppRef        []interface{}        `json:"privateAppRef,omitempty"`
}

type Address struct {
	AddressValidated string  `json:"addressValidated,omitempty"`
	CityName         string  `json:"cityName,omitempty"`
	Country          Country `json:"country,omitempty"`
	StateName        *string `json:"stateName,omitempty"`
	Street           *string `json:"street,omitempty"`
	ZipCode          *string `json:"zipCode,omitempty"`
}

type Country struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	By    string `json:"by,omitempty"`
	Input string `json:"input,omitempty"`
}

type PreferredPopLocation struct {
	PreferredOnly bool         `json:"preferredOnly"`
	Automatic     bool         `json:"automatic"`
	Primary       *PopLocation `json:"primary,omitempty"`
	Secondary     *PopLocation `json:"secondary,omitempty"`
}

type PopLocation struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	By    string `json:"by,omitempty"`
	Input string `json:"input,omitempty"`
}

func fetchConnector(id string) (*Connector, error) {
	connectors, err := fetchConnectors()
	if err != nil {
		return nil, err
	}
	for _, connector := range connectors {
		if connector.ID == id {
			return &connector, nil
		}
	}
	return nil, fmt.Errorf("Connector not found")
}

func createConnector(c *Connector) (id string, err error) {
	type Variables struct {
		AccountID    string    `json:"accountId"`
		NewConnector Connector `json:"newConnector"`
	}
	type gqlData struct {
		Query         string    `json:"query"`
		Variables     Variables `json:"variables"`
		OperationName string    `json:"operationName"`
	}

	type respCreate struct {
		Data struct {
			ZtnaAppConnector struct {
				AddZtnaAppConnector struct {
					ZtnaAppConnector struct {
						ID string `json:"id"`
					} `json:"ztnaAppConnector"`
				} `json:"addZtnaAppConnector"`
			} `json:"ztnaAppConnector"`
		} `json:"data"`
	}

	type ztnaAppConnector struct {
		ZtnaAppConnectorList ztnaAppConnectorList `json:"ztnaAppConnectorList"`
	}
	accountID := os.Getenv("CATO_ACCOUNT_ID")
	apiKey := os.Getenv("CATO_TOKEN")
	apiURL := os.Getenv("CATO_ENDPOINT")
	data := gqlData{
		Query: gqlCreateConnector,
		Variables: Variables{
			AccountID:    accountID,
			NewConnector: *c,
		},
		OperationName: "CreateConnector",
	}
	body, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	respBytes, err := postJSON(apiURL, body, apiKey)
	if err != nil {
		return "", err
	}
	var resp respCreate
	if err = json.Unmarshal(respBytes, &resp); err != nil {
		return "", err
	}
	if resp.Data.ZtnaAppConnector.AddZtnaAppConnector.ZtnaAppConnector.ID == "" {
		return "", fmt.Errorf("error creating connector: %s", string(respBytes))
	}
	return resp.Data.ZtnaAppConnector.AddZtnaAppConnector.ZtnaAppConnector.ID, nil
}

func deleteConnector(id string) (err error) {
	type ztnaAppCon struct {
		Input string `json:"input"`
	}
	type connectorRef struct {
		ZtnaAppConnector ztnaAppCon `json:"ztnaAppConnector"`
	}
	type Variables struct {
		AccountID    string       `json:"accountId"`
		ConnectorRef connectorRef `json:"connectorRef"`
	}
	type gqlData struct {
		Query         string    `json:"query"`
		Variables     Variables `json:"variables"`
		OperationName string    `json:"operationName"`
	}

	type respDelete struct {
		Data struct {
			ZtnaAppConnector struct {
				RemoveZtnaAppConnector struct {
					ZtnaAppConnector struct {
						ID string `json:"id"`
					} `json:"ztnaAppConnector"`
				} `json:"removeZtnaAppConnector"`
			} `json:"ztnaAppConnector"`
		} `json:"data"`
	}

	type ztnaAppConnector struct {
		ZtnaAppConnectorList ztnaAppConnectorList `json:"ztnaAppConnectorList"`
	}
	accountID := os.Getenv("CATO_ACCOUNT_ID")
	apiKey := os.Getenv("CATO_TOKEN")
	apiURL := os.Getenv("CATO_ENDPOINT")
	data := gqlData{
		Query: gqlDeleteConnector,
		Variables: Variables{
			AccountID: accountID,
			ConnectorRef: connectorRef{
				ZtnaAppConnector: ztnaAppCon{
					Input: id,
				},
			},
		},
		OperationName: "DeleteConnector",
	}
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	respBytes, err := postJSON(apiURL, body, apiKey)
	if err != nil {
		return err
	}
	var resp respDelete
	if err = json.Unmarshal(respBytes, &resp); err != nil {
		return err
	}
	if resp.Data.ZtnaAppConnector.RemoveZtnaAppConnector.ZtnaAppConnector.ID == "" {
		return fmt.Errorf("error deleting connector: %s", string(respBytes))
	}
	return nil
}

func fetchConnectors() ([]Connector, error) {
	type gqlData struct {
		Query     string `json:"query"`
		Variables struct {
			AccountID string `json:"accountId"`
		} `json:"variables"`
		OperationName string `json:"operationName"`
	}

	accountID := os.Getenv("CATO_ACCOUNT_ID")
	apiKey := os.Getenv("CATO_TOKEN")
	apiURL := os.Getenv("CATO_ENDPOINT")
	data := gqlData{
		Query: gqlQuery,
		Variables: struct {
			AccountID string "json:\"accountId\""
		}{AccountID: accountID},
		OperationName: "ztnaAppConnector",
	}
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	respBytes, err := postJSON(apiURL, body, apiKey)
	if err != nil {
		return nil, err
	}
	var resp Resp
	if err = json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}
	return resp.Data.ZtnaAppConnector.ZtnaAppConnectorList.ZtnaAppConnector, nil
}

func postJSON(reqURL string, body []byte, key string) ([]byte, error) {
	// Create request
	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", key)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Optional: check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed: %s (%s)", resp.Status, string(respBody))
	}

	return respBody, nil
}

const gqlQuery = `query ztnaAppConnector($accountId: ID!) {
  ztnaAppConnector(accountId: $accountId) {
    ztnaAppConnectorList(input: {}) {
      ztnaAppConnector {
        id
        name
        serialNumber
        type
        groupName
        address {
          addressValidated
          cityName
          country {
            id
            name
          }
          stateName
          street
          zipCode
        }
        timezone
        preferredPopLocation {
          preferredOnly
          primary {
            id
            name
          }
          secondary {
            id
            name
          }
        }
        privateAppRef {
          id
          name
        }
      }
    }
  }
}
`

const gqlCreateConnector = `
mutation CreateConnector($accountId: ID!, $newConnector: AddZtnaAppConnectorInput!) {
  ztnaAppConnector(accountId: $accountId) {
    addZtnaAppConnector(input: $newConnector) {
      ztnaAppConnector {
        id
      } 
    }
  }
}
`

const gqlDeleteConnector = `
mutation DeleteConnector($accountId: ID!, $connectorRef: RemoveZtnaAppConnectorInput!) {
  ztnaAppConnector(accountId: $accountId) {
    removeZtnaAppConnector(input: $connectorRef) {
      ztnaAppConnector {
        id
        name
      }
    }
  }
}
`
