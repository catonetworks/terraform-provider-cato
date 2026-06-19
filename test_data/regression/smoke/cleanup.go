/*
Delete sites created by terraform tests, "TF-Test-*"
*/
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

const testSitePrefix = "TF-Test-"

type Cfg struct {
	BaseURL   string
	Token     string
	AccountID string
}

type NameID struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

const (
	CatoBaseURL   = "CATO_BASEURL"
	CatoToken     = "CATO_TOKEN"
	CatoAccountID = "CATO_ACCOUNT_ID"
)

func main() {
	ctx := context.Background()
	if err := cleanUpTestSites(ctx); err != nil {
		fmt.Printf("Error cleaning up test sites: %v\n", err)
		os.Exit(1)
	}
}

func cleanUpTestSites(ctx context.Context) error {
	var deleteErr error
	cfg, err := getCfg()
	if err != nil {
		return err
	}
	client, err := newClient(cfg)
	if err != nil {
		return err
	}
	sites, err := getSites(ctx, client, cfg.AccountID)
	if err != nil {
		return err
	}
	for _, site := range sites {
		if !strings.HasPrefix(site.Name, testSitePrefix) {
			continue
		}
		if _, err = client.SiteRemoveSite(ctx, site.ID, cfg.AccountID); err != nil {
			deleteErr = fmt.Errorf("deleting site '%s' (%s): %v", site.Name, site.ID, err)
		}
	}
	return deleteErr
}

func getCfg() (*Cfg, error) {
	cfg := Cfg{
		BaseURL:   os.Getenv(CatoBaseURL),
		Token:     os.Getenv(CatoToken),
		AccountID: os.Getenv(CatoAccountID),
	}
	if cfg.BaseURL == "" || cfg.Token == "" || cfg.AccountID == "" {
		return nil, fmt.Errorf("missing env vars: %s,%s,%s", CatoBaseURL, CatoToken, CatoAccountID)
	}
	return &cfg, nil
}

func newClient(cfg *Cfg) (*cato.Client, error) {
	headers := map[string]string{"User-Agent": "tf-test"}
	catoClient, err := cato.New(cfg.BaseURL, cfg.Token, cfg.AccountID, &http.Client{}, headers)
	return catoClient, err
}

func getSites(ctx context.Context, client *cato.Client, accountID string) (sites []NameID, err error) {
	limit := int64(100)
	result, err := client.EntityLookupMinimal(ctx, accountID, cato_models.EntityTypeSite, &limit, nil, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	for _, site := range result.GetEntityLookup().GetItems() {
		sites = append(sites, NameID{
			ID:   site.GetEntity().GetID(),
			Name: *site.GetEntity().GetName(),
		})
	}
	return sites, nil
}
