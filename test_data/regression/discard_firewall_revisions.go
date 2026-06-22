/*
Discard unpublished Internet Firewall, WAN Firewall, and WAN Network policy revisions.

This tool reads CATO_BASEURL, CATO_TOKEN, and CATO_ACCOUNT_ID from the
environment. By default it only lists open revisions; pass --execute to discard.
*/
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

const (
	catoBaseURL   = "CATO_BASEURL"
	catoToken     = "CATO_TOKEN"
	catoAccountID = "CATO_ACCOUNT_ID"
)

type cfg struct {
	baseURL   string
	token     string
	accountID string
}

type revision struct {
	policy      string
	id          string
	name        string
	description string
	changes     int64
	createdTime string
	updatedTime string
}

func main() {
	execute := flag.Bool("execute", false, "discard all listed firewall and WAN policy revisions")
	configPath := flag.String("config", "", "optional regression env config path; account ID is inferred from a '# ... account <id>' comment when CATO_ACCOUNT_ID is unset")
	flag.Parse()

	ctx := context.Background()
	if err := run(ctx, *execute, *configPath); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, execute bool, configPath string) error {
	cfg, err := getCfg(configPath)
	if err != nil {
		return err
	}

	client, err := newClient(cfg)
	if err != nil {
		return err
	}

	revisions, err := listFirewallRevisions(ctx, client, cfg.accountID)
	if err != nil {
		return err
	}

	fmt.Printf("account_id=%s open_firewall_wan_revisions=%d execute=%t\n", cfg.accountID, len(revisions), execute)
	for _, rev := range revisions {
		fmt.Printf("%s revision id=%s name=%q changes=%d created=%s updated=%s\n",
			rev.policy, rev.id, rev.name, rev.changes, rev.createdTime, rev.updatedTime)
	}

	if !execute {
		fmt.Println("dry-run only; rerun with --execute to discard these revisions")
		return nil
	}

	var discardErr error
	for _, rev := range revisions {
		if err := discardRevision(ctx, client, cfg.accountID, rev); err != nil {
			discardErr = errors.Join(discardErr, err)
			continue
		}
		fmt.Printf("%s revision id=%s discarded\n", rev.policy, rev.id)
	}

	return discardErr
}

func getCfg(configPath string) (*cfg, error) {
	cfg := cfg{
		baseURL:   os.Getenv(catoBaseURL),
		token:     os.Getenv(catoToken),
		accountID: os.Getenv(catoAccountID),
	}
	if cfg.accountID == "" && configPath != "" {
		accountID, err := accountIDFromConfig(configPath)
		if err != nil {
			return nil, err
		}
		cfg.accountID = accountID
	}
	if cfg.baseURL == "" || cfg.token == "" || cfg.accountID == "" {
		return nil, fmt.Errorf("missing env vars: %s,%s,%s", catoBaseURL, catoToken, catoAccountID)
	}
	return &cfg, nil
}

func accountIDFromConfig(path string) (string, error) {
	contents, err := os.ReadFile(path) //nolint:gosec // operator-provided local config path
	if err != nil {
		return "", fmt.Errorf("read config %q: %w", path, err)
	}
	matches := regexp.MustCompile(`(?m)^#.*account ([0-9]+)\b`).FindSubmatch(contents)
	if len(matches) != 2 {
		return "", fmt.Errorf("could not infer account ID from %q", path)
	}
	return string(matches[1]), nil
}

func newClient(cfg *cfg) (*cato.Client, error) {
	return cato.New(cfg.baseURL, cfg.token, cfg.accountID, &http.Client{},
		map[string]string{"User-Agent": "terraform-provider-cato-discard-firewall-revisions"})
}

func listFirewallRevisions(ctx context.Context, client *cato.Client, accountID string) ([]revision, error) {
	resp, err := client.Policy(ctx, nil, nil, accountID)
	if err != nil {
		return nil, err
	}

	var revisions []revision
	for _, rev := range resp.GetPolicy().GetInternetFirewall().GetRevisionsInternetFirewallPolicyQueries().GetRevision() {
		revisions = append(revisions, revision{
			policy:      "internet_firewall",
			id:          rev.GetID(),
			name:        rev.GetName(),
			description: rev.GetDescription(),
			changes:     rev.GetChanges(),
			createdTime: rev.GetCreatedTime(),
			updatedTime: rev.GetUpdatedTime(),
		})
	}
	for _, rev := range resp.GetPolicy().GetWanFirewall().GetRevisionsWanFirewallPolicyQueries().GetRevision() {
		revisions = append(revisions, revision{
			policy:      "wan_firewall",
			id:          rev.GetID(),
			name:        rev.GetName(),
			description: rev.GetDescription(),
			changes:     rev.GetChanges(),
			createdTime: rev.GetCreatedTime(),
			updatedTime: rev.GetUpdatedTime(),
		})
	}

	wanNetwork, err := client.WanNetworkPolicy(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if rev := wanNetwork.GetPolicy().GetWanNetwork().GetPolicy().GetRevision(); rev != nil && rev.GetID() != "" {
		revisions = append(revisions, revision{
			policy:      "wan_network",
			id:          rev.GetID(),
			name:        rev.GetName(),
			description: rev.GetDescription(),
			changes:     rev.GetChanges(),
			createdTime: rev.GetCreatedTime(),
			updatedTime: rev.GetUpdatedTime(),
		})
	}

	return revisions, nil
}

func discardRevision(ctx context.Context, client *cato.Client, accountID string, rev revision) error {
	input := &cato_models.PolicyDiscardRevisionInput{ID: &rev.id}

	switch rev.policy {
	case "internet_firewall":
		resp, err := client.PolicyInternetFirewallDiscardPolicyRevision(ctx,
			&cato_models.InternetFirewallPolicyMutationInput{
				Revision: &cato_models.PolicyMutationRevisionInput{ID: &rev.id},
			},
			input,
			accountID,
		)
		if err != nil {
			return fmt.Errorf("%s revision %s: %w", rev.policy, rev.id, err)
		}
		errors := resp.GetPolicy().GetInternetFirewall().GetDiscardPolicyRevision().GetErrors()
		if len(errors) > 0 {
			return fmt.Errorf("%s revision %s: %s", rev.policy, rev.id, formatInternetFirewallErrors(errors))
		}
	case "wan_firewall":
		resp, err := client.PolicyWanFirewallDiscardPolicyRevision(ctx, input, accountID)
		if err != nil {
			return fmt.Errorf("%s revision %s: %w", rev.policy, rev.id, err)
		}
		errors := resp.GetPolicy().GetWanFirewall().GetDiscardPolicyRevision().GetErrors()
		if len(errors) > 0 {
			return fmt.Errorf("%s revision %s: %s", rev.policy, rev.id, formatWanFirewallErrors(errors))
		}
	case "wan_network":
		resp, err := client.PolicyWanNetworkDiscardPolicyRevision(ctx, accountID)
		if err != nil {
			return fmt.Errorf("%s revision %s: %w", rev.policy, rev.id, err)
		}
		errors := resp.GetPolicy().GetWanNetwork().GetDiscardPolicyRevision().GetErrors()
		if len(errors) > 0 {
			return fmt.Errorf("%s revision %s: %s", rev.policy, rev.id, formatWanNetworkErrors(errors))
		}
	default:
		return fmt.Errorf("unknown policy type %q for revision %s", rev.policy, rev.id)
	}

	return nil
}

func formatInternetFirewallErrors(errors []*cato.PolicyInternetFirewallDiscardPolicyRevision_Policy_InternetFirewall_DiscardPolicyRevision_Errors) string {
	return formatErrors(len(errors), func(i int) (*string, *string) {
		return errors[i].GetErrorCode(), errors[i].GetErrorMessage()
	})
}

func formatWanFirewallErrors(errors []*cato.PolicyWanFirewallDiscardPolicyRevision_Policy_WanFirewall_DiscardPolicyRevision_Errors) string {
	return formatErrors(len(errors), func(i int) (*string, *string) {
		return errors[i].GetErrorCode(), errors[i].GetErrorMessage()
	})
}

func formatWanNetworkErrors(errors []*cato.PolicyWanNetworkDiscardPolicyRevision_Policy_WanNetwork_DiscardPolicyRevision_Errors) string {
	return formatErrors(len(errors), func(i int) (*string, *string) {
		return errors[i].GetErrorCode(), errors[i].GetErrorMessage()
	})
}

func formatErrors(n int, get func(int) (*string, *string)) string {
	out := ""
	for i := 0; i < n; i++ {
		code, message := get(i)
		if i > 0 {
			out += "; "
		}
		out += fmt.Sprintf("code=%s message=%s", stringPtr(code), stringPtr(message))
	}
	return out
}

func stringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
