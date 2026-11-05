package fqdncontainer

import (
	"context"
	"fmt"

	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

type ContainerRecord = cato.ContainerList_Container_List_Containers

// DownloadRecord contains the identity and encoded member file returned by the API.
type DownloadRecord = cato.ContainerFqdnDownloadFile_Container_Fqdn_DownloadFile

// ResourceClient is the narrow account-scoped container API contract.
type ResourceClient interface {
	Find(context.Context, string, string) (*cato.ContainerList_Container_List_Containers, error)
	CreateInline(context.Context, string, m.CreateFqdnContainerFromListInput) (*cato.ContainerList_Container_List_Containers, error)
	CreateURL(context.Context, string, m.CreateFqdnContainerFromURLInput) (*cato.ContainerList_Container_List_Containers, error)
	UpdateInline(context.Context, string, m.UpdateFqdnContainerFromListInput) (*cato.ContainerList_Container_List_Containers, error)
	UpdateURL(context.Context, string, m.UpdateFqdnContainerFromURLInput) (*cato.ContainerList_Container_List_Containers, error)
	Download(context.Context, string, string) (*cato.ContainerFqdnDownloadFile_Container_Fqdn_DownloadFile, error)
	Delete(context.Context, string, string) (*cato.ContainerList_Container_List_Containers, error)
}
type sdkClient struct{ client *catoclient.Service }

const metadataFields = `id name description size syncData {
fileType url timeInterval timeUnit notifications {
mailingList{
id name
} subscriptionGroup{
id name
} webhook{
id name
}
}
}`

func (c *sdkClient) Find(ctx context.Context, account, id string) (*ContainerRecord, error) {
	// Exact, type-filtered list has reliable empty-list semantics, unlike the nullable search resolver.
	var result cato.ContainerList
	doc := `query containerList($accountId:ID!,$containerSearchInput:ContainerSearchInput!){
container(accountId:$accountId){
list(input:$containerSearchInput){
containers{
` + metadataFields + `}}}}`
	err := c.client.V2().Client.Post(ctx,
		"containerList",
		doc,
		&result,
		map[string]interface{}{"accountId": account,
			"containerSearchInput": m.ContainerSearchInput{Refs: []*m.ContainerRefInput{{By: m.ObjectRefByID,
				Input: id}},
				Types: []m.ContainerType{m.ContainerTypeFqdn}}})
	if err != nil {
		return nil,
			err
	}
	list := result.GetContainer().GetList()
	if list == nil ||
		list.Containers == nil {
		return nil,
			fmt.Errorf("API did not return a container list")
	}
	if len(list.Containers) == 0 {
		return nil,
			nil
	}
	if len(list.Containers) != 1 ||
		list.Containers[0] == nil ||
		list.Containers[0].ID != id {
		return nil,
			fmt.Errorf("API did not confirm the requested container identity")
	}
	return list.Containers[0],
		nil
}
func (c *sdkClient) CreateInline(ctx context.Context, account string, in m.CreateFqdnContainerFromListInput) (*ContainerRecord, error) {
	var result struct {
		Container struct {
			Fqdn struct {
				Result struct {
					Container *ContainerRecord `json:"container" graphql:"container"`
				} `json:"createFromList" graphql:"createFromList"`
			} `json:"fqdn" graphql:"fqdn"`
		} `json:"container" graphql:"container"`
	}
	doc := `mutation containerFqdnCreateFromList($accountId:ID!,$createFqdnContainerFromListInput:CreateFqdnContainerFromListInput!){
container(accountId:$accountId){
fqdn{
createFromList(input:$createFqdnContainerFromListInput){
container{
` + metadataFields + `}}}}}`
	err := c.client.V2().Client.Post(ctx,
		"containerFqdnCreateFromList",
		doc,
		&result,
		map[string]interface{}{"accountId": account,
			"createFqdnContainerFromListInput": in})
	return result.Container.Fqdn.Result.Container,
		err
}
func (c *sdkClient) CreateURL(ctx context.Context, account string, in m.CreateFqdnContainerFromURLInput) (*ContainerRecord, error) {
	var result struct {
		Container struct {
			Fqdn struct {
				Result struct {
					Container *ContainerRecord `json:"container" graphql:"container"`
				} `json:"createFromURL" graphql:"createFromURL"`
			} `json:"fqdn" graphql:"fqdn"`
		} `json:"container" graphql:"container"`
	}
	doc := `mutation containerFqdnCreateFromURL($accountId:ID!,$createFqdnContainerFromUrlInput:CreateFqdnContainerFromUrlInput!){
container(accountId:$accountId){
fqdn{
createFromURL(input:$createFqdnContainerFromUrlInput){
container{
` + metadataFields + `}}}}}`
	err := c.client.V2().Client.Post(ctx,
		"containerFqdnCreateFromURL",
		doc,
		&result,
		map[string]interface{}{"accountId": account,
			"createFqdnContainerFromUrlInput": in})
	return result.Container.Fqdn.Result.Container,
		err
}
func (c *sdkClient) UpdateInline(ctx context.Context, account string, in m.UpdateFqdnContainerFromListInput) (*ContainerRecord, error) {
	var result struct {
		Container struct {
			Fqdn struct {
				Result struct {
					Container *ContainerRecord `json:"container" graphql:"container"`
				} `json:"updateFromList" graphql:"updateFromList"`
			} `json:"fqdn" graphql:"fqdn"`
		} `json:"container" graphql:"container"`
	}
	doc := `mutation containerFqdnUpdateFromList($accountId:ID!,$updateFqdnContainerFromListInput:UpdateFqdnContainerFromListInput!){
container(accountId:$accountId){
fqdn{
updateFromList(input:$updateFqdnContainerFromListInput){
container{
` + metadataFields + `}}}}}`
	err := c.client.V2().Client.Post(ctx,
		"containerFqdnUpdateFromList",
		doc,
		&result,
		map[string]interface{}{"accountId": account,
			"updateFqdnContainerFromListInput": map[string]interface{}{"description": in.Description,
				"ref":    in.Ref,
				"values": in.Values}})
	return result.Container.Fqdn.Result.Container,
		err
}
func (c *sdkClient) UpdateURL(ctx context.Context, account string, in m.UpdateFqdnContainerFromURLInput) (*ContainerRecord, error) {
	var result struct {
		Container struct {
			Fqdn struct {
				Result struct {
					Container *ContainerRecord `json:"container" graphql:"container"`
				} `json:"updateFromURL" graphql:"updateFromURL"`
			} `json:"fqdn" graphql:"fqdn"`
		} `json:"container" graphql:"container"`
	}
	doc := `mutation containerFqdnUpdateFromURL($accountId:ID!,$updateFqdnContainerFromUrlInput:UpdateFqdnContainerFromUrlInput!){
container(accountId:$accountId){
fqdn{
updateFromURL(input:$updateFqdnContainerFromUrlInput){
container{
` + metadataFields + `}}}}}`
	err := c.client.V2().Client.Post(ctx,
		"containerFqdnUpdateFromURL",
		doc,
		&result,
		map[string]interface{}{"accountId": account,
			"updateFqdnContainerFromUrlInput": in})
	return result.Container.Fqdn.Result.Container,
		err
}
func (c *sdkClient) Download(ctx context.Context, account, id string) (*DownloadRecord, error) {
	var result struct {
		Container struct {
			Fqdn struct {
				Download *DownloadRecord `json:"downloadFile" graphql:"downloadFile"`
			} `json:"fqdn" graphql:"fqdn"`
		} `json:"container" graphql:"container"`
	}
	doc := `query containerFqdnDownloadFile($accountId:ID!,$downloadFqdnContainerFileInput:DownloadFqdnContainerFileInput!){
container(accountId:$accountId){
fqdn{
downloadFile(input:$downloadFqdnContainerFileInput){
id name encodedFile
}
}
}
}`
	err := c.client.V2().Client.Post(ctx,
		"containerFqdnDownloadFile",
		doc,
		&result,
		map[string]interface{}{"accountId": account,
			"downloadFqdnContainerFileInput": m.DownloadFqdnContainerFileInput{By: m.ObjectRefByID,
				Input: id}})
	return result.Container.Fqdn.Download,
		err
}
func (c *sdkClient) Delete(ctx context.Context, account, id string) (*ContainerRecord, error) {
	var result struct {
		Container struct {
			Delete struct {
				Container *ContainerRecord `json:"container" graphql:"container"`
			} `json:"delete" graphql:"delete"`
		} `json:"container" graphql:"container"`
	}
	doc := `mutation containerDelete($accountId:ID!,$deleteContainerInput:DeleteContainerInput!){
container(accountId:$accountId){
delete(input:$deleteContainerInput){
container{
id name size description
}
}
}
}`
	err := c.client.V2().Client.Post(ctx,
		"containerDelete",
		doc,
		&result,
		map[string]interface{}{"accountId": account,
			"deleteContainerInput": m.DeleteContainerInput{Ref: &m.ContainerRefInput{By: m.ObjectRefByID,
				Input: id}}})
	return result.Container.Delete.Container,
		err
}
