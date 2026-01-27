package provider

import (
	"fmt"
	"testing"
)

func TestFetch(t *testing.T) {
	c, err := fetchConnectors()
	if err != nil {
		t.Errorf("fetch failed: %v", err)
	}
	fmt.Printf("%v", c)
}

func TestCreate(t *testing.T) {
	newC := Connector{
		Name:      "test-connector-100",
		GroupName: "conn-group-100",
		Type:      "VIRTUAL",
		Timezone:  "America/New_York",
		PreferredPopLocation: PreferredPopLocation{
			Automatic:     false,
			PreferredOnly: false,
			Primary: &PopLocation{
				By:    "NAME",
				Input: "New York_Sta",
			},
		},
		Address: Address{
			Country: Country{
				By:    "NAME",
				Input: "United States",
			},
			StateName: ptr("Virginia"),
			CityName:  "Richmond",
		},
	}
	id, err := createConnector(&newC)
	if err != nil {
		t.Errorf("create failed: %v", err)
	}
	fmt.Printf("Created connector: %v\n", id)
}

func TestDelete(t *testing.T) {
	id := "1500000016"
	err := deleteConnector(id)
	if err != nil {
		t.Errorf("delete failed: %v", err)
	}
	fmt.Printf("Deleted connector: %v\n", id)
}
